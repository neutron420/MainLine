variable "name" {
  description = "Resource name prefix"
  type        = string
  default     = "schemahub"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "backend_image" {
  description = "Backend container image (ghcr.io/...)"
  type        = string
  default     = "ghcr.io/neutron420/mainline/backend:latest"
}

variable "jwt_private_key" {
  description = "JWT RS256 private key (PEM)"
  type        = string
  sensitive   = true
}

variable "jwt_public_key" {
  description = "JWT RS256 public key (PEM)"
  type        = string
  sensitive   = true
}
