# Networking module — VPC + subnets (skeleton)
# TODO: implement aws_vpc, aws_subnet (public/private), igw, natgw,
#       security groups, route tables.

variable "name" {
  type = string
}

variable "vpc_cidr" {
  type = string
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = []
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = []
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = ""
}
