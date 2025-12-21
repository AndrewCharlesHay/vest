resource "aws_db_instance" "default" {
  identifier           = "vest-db"
  allocated_storage    = 20
  storage_type         = "gp3"
  engine               = "postgres"
  engine_version       = "18.1"
  instance_class       = "db.t3.micro"
  db_name              = var.db_name
  username             = var.db_username
  password             = var.db_password
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
