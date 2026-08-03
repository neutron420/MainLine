output "alb_dns_name" {
  description = "DNS name of the ALB (gRPC-Web endpoint)"
  value       = aws_lb.this.dns_name
}

output "alb_arn" {
  description = "ARN of the ALB"
  value       = aws_lb.this.arn
}

output "target_group_arn" {
  description = "ARN of the backend target group"
  value       = aws_lb_target_group.backend.arn
}

output "cluster_name" {
  description = "ECS cluster name"
  value       = aws_ecs_cluster.this.name
}

output "cluster_id" {
  description = "ECS cluster id"
  value       = aws_ecs_cluster.this.id
}

output "service_name" {
  description = "ECS service name"
  value       = aws_ecs_service.backend.name
}

output "task_definition_arn" {
  description = "ARN of the backend task definition"
  value       = aws_ecs_task_definition.backend.arn
}

output "app_security_group_id" {
  description = "Application security group id (feed back into the RDS/Redis modules)"
  value       = aws_security_group.app.id
}

output "execution_role_arn" {
  description = "ARN of the task execution role"
  value       = aws_iam_role.task_execution.arn
}

output "log_group_name" {
  description = "CloudWatch log group for backend logs"
  value       = aws_cloudwatch_log_group.this.name
}
