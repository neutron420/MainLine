variable "name" {
  description = "Resource name prefix"
  type        = string
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
}

variable "azs" {
  description = "Availability zones to place subnets in. Empty means 'use 3 healthy AZs in the provider region'."
  type        = list(string)
  default     = []
}

variable "az_count" {
  description = "Number of public/private subnet pairs to create"
  type        = number
  default     = 3

  validation {
    condition     = var.az_count >= 1 && var.az_count <= 3
    error_message = "az_count must be between 1 and 3."
  }
}

variable "nat_gateway_count" {
  description = "NAT gateways to create (1 = cost-optimized single NAT; 3 = HA one-per-AZ)"
  type        = number
  default     = 1

  validation {
    condition     = var.nat_gateway_count >= 1 && var.nat_gateway_count <= var.az_count
    error_message = "nat_gateway_count must be at least 1 and no more than az_count."
  }
}

variable "tags" {
  description = "Additional tags applied to all resources"
  type        = map(string)
  default     = {}
}
