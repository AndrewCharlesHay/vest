variable "backend_bucket_prefix" {
  default = "vest-tf-state"
}

resource "random_id" "bucket_suffix" {
  byte_length = 8
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = "vest-tf-state-backend-ops-interview-ah"
  
  lifecycle {
    prevent_destroy = true
  }
}

import {
  to = aws_s3_bucket.terraform_state
  id = "vest-tf-state-backend-ops-interview-ah"
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"
  }
}

output "s3_bucket_name" {
  value = aws_s3_bucket.terraform_state.id
}
