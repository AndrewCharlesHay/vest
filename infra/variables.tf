

variable "db_username" {
  description = "The username for the database"
  type        = string
  default     = "vest_user"
}

variable "db_name" {
  description = "The name of the database"
  type        = string
  default     = "vest"
}

variable "alert_email" {
  description = "Email address for CloudWatch alerts"
  type        = string
  default     = "noreply@example.com" # Placeholder until user provides one
}
