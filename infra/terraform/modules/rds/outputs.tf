output "connection_url" {
  description = "Postgres connection URL for the backend (sensitive)"
  value       = aws_ssm_parameter.connection_url.value
  sensitive   = true
}

output "connection_url_ssm_parameter" {
  description = "Name of the SSM SecureString parameter holding the connection URL"
  value       = aws_ssm_parameter.connection_url.name
}

output "host" {
  description = "RDS endpoint hostname"
  value       = aws_db_instance.this.address
}

output "port" {
  description = "RDS port (5432)"
  value       = aws_db_instance.this.port
}

output "db_name" {
  description = "Database name"
  value       = aws_db_instance.this.db_name
}

output "db_username" {
  description = "Database master username"
  value       = aws_db_instance.this.username
}

output "security_group_id" {
  description = "Security group id of the RDS instance"
  value       = aws_security_group.this.id
}

output "arn" {
  description = "ARN of the RDS instance"
  value       = aws_db_instance.this.arn
}
