terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
  backend "s3" {
    bucket         = "vest-tf-state-backend-ops-interview-ah"
    key            = "terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true

  }
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_ecr_repository" "app" {
  name = "vest-app"
}

import {
  to = aws_ecr_repository.app
  id = "vest-app"
}

resource "aws_ecs_cluster" "main" {
  name = "vest-cluster"
}

resource "random_password" "api_key" {
  length  = 32
  special = false
}

resource "aws_ecs_task_definition" "app" {
  family                   = "vest-app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = 256
  memory                   = 512
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn

  container_definitions = jsonencode([
    {
      name  = "vest-app"
      image = "${aws_ecr_repository.app.repository_url}:latest"
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
        }
      ]
      environment = [
        { name = "DB_HOST", value = aws_db_instance.default.address },
        { name = "DB_USER", value = var.db_username },
        { name = "DB_NAME", value = var.db_name },
        { name = "API_KEY", value = random_password.api_key.result },
        # SFTP Config
        { name = "SFTP_HOST", value = "localhost:22" },
        { name = "SFTP_USER", value = "vest" },
        { name = "SFTP_PASS", value = "pass" },
        { name = "SFTP_DIR", value = "/upload" }
      ]
      secrets = [
        {
          name      = "DB_PASSWORD"
          valueFrom = aws_secretsmanager_secret.db_password.arn
        }
      ]
      mountPoints = [
        {
          sourceVolume  = "sftp-data"
          containerPath = "/upload"
        }
      ]
    },
    {
      name  = "sftp"
      image = "atmoz/sftp"
      # Format: user:pass:uid:gid:dir
      # Default uid 1001 for user.
      command = ["vest:pass:1001:1001:upload"]
      portMappings = [
        {
          containerPort = 22
          hostPort      = 22
        }
      ]
      mountPoints = [
        {
          sourceVolume  = "sftp-data"
          containerPath = "/home/vest/upload"
        }
      ]
    }
  ])

  volume {
    name = "sftp-data"
  }
}

import {
  to = aws_iam_role.ecs_task_execution_role
  id = "vest-ecs-task-execution-role"
}

resource "aws_iam_role" "ecs_task_execution_role" {
  name = "vest-ecs-task-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}


