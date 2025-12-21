resource "random_password" "db_password" {
  length  = 16
  special = false
}

resource "aws_secretsmanager_secret" "db_password" {
  name = "vest-db-password"
}

resource "aws_secretsmanager_secret_version" "db_password" {
  secret_id     = aws_secretsmanager_secret.db_password.id
  secret_string = random_password.db_password.result
}

resource "aws_db_instance" "default" {
  identifier           = "vest-db"
  allocated_storage    = 20
  storage_type         = "gp3"
  engine               = "postgres"
  engine_version       = "18.1"
  instance_class       = "db.t3.micro"
  db_name              = var.db_name
  username             = var.db_username
  password             = random_password.db_password.result
  parameter_group_name = "default.postgres18"
  skip_final_snapshot  = true
  publicly_accessible  = false
  
  # VPC & Security Groups would go here in a real setup
  # vpc_security_group_ids = [aws_security_group.db.id]
  # db_subnet_group_name   = aws_db_subnet_group.default.name
}

# Output the endpoint
output "db_endpoint" {
  value = aws_db_instance.default.endpoint
}
