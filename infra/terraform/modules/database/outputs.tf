# Database module — managed PostgreSQL (skeleton)
# TODO: implement aws_db_instance (or Neon), secret store the connection URL.

variable "name" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

output "connection_url" {
  description = "Postgres connection URL for the backend"
  value       = ""
  sensitive   = true
}
