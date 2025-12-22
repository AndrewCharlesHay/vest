resource "random_id" "bucket_suffix" {
  byte_length = 8
}

resource "aws_s3_bucket" "terraform_state" {
  bucket = "vest-tf-state-${random_id.bucket_suffix.hex}"
  
  # Prevent accidental deletion of this S3 bucket
  lifecycle {
    prevent_destroy = true
  }
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
