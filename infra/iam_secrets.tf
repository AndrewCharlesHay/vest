resource "aws_iam_policy" "secrets" {
  name        = "vest-secrets-policy"
  description = "Allow reading DB password secret"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = [
          aws_secretsmanager_secret.db_password.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "secrets" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = aws_iam_policy.secrets.arn
}
