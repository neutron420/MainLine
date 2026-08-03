# VPC module — networking foundation for the SchemaHub platform.
#
# Creates a VPC with one public and one private subnet per availability zone,
# an Internet Gateway for the public tier, a NAT Gateway for outbound traffic
# from the private tier, and per-AZ route tables.
#
# Cost note: a single NAT Gateway is created by default. For production HA,
# set `nat_gateway_count` equal to the number of AZs (one NAT per AZ, each in
# its own public subnet) so a single-AZ failure does not take out egress for
# the whole private tier. If NAT cost is a concern for non-prod environments,
# a NAT instance (t3.nano with a static EIP and iptables MASQUERADE) is a
# documented alternative, but it is a single point of failure and must run in
# an autoscaling group — we recommend the managed gateway.

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  azs = length(var.azs) > 0 ? var.azs : slice(data.aws_availability_zones.available.names, 0, var.az_count)
  // Public subnets get the first N /24 blocks of the VPC CIDR, private
  // subnets the next N. Works for any /16 (or larger) vpc_cidr.
  public_cidrs  = [for i in range(var.az_count) : cidrsubnet(var.vpc_cidr, 8, i)]
  private_cidrs = [for i in range(var.az_count) : cidrsubnet(var.vpc_cidr, 8, var.az_count + i)]
}

resource "aws_vpc" "this" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = merge(var.tags, {
    Name = "${var.name}-vpc"
  })
}

# ---------------------------------------------------------------------------
# Public tier: subnets, Internet Gateway, route table
# ---------------------------------------------------------------------------

resource "aws_subnet" "public" {
  for_each = { for i, az in local.azs : az => {
    index = i
    cidr  = local.public_cidrs[i]
  } }

  vpc_id            = aws_vpc.this.id
  availability_zone = each.key
  cidr_block        = each.value.cidr

  # Public subnets auto-assign public IPs: needed by the NAT gateways and by
  # any future EC2-based services (jump hosts, NAT instances) in the public
  # tier. Fargate tasks and the ALB attach ENIs and never use this flag.
  map_public_ip_on_launch = true

  tags = merge(var.tags, {
    Name    = "${var.name}-public-${each.key}"
    Tier    = "public"
    Purpose = "load-balancer-nat"
  })
}

resource "aws_internet_gateway" "this" {
  vpc_id = aws_vpc.this.id

  tags = merge(var.tags, {
    Name = "${var.name}-igw"
  })
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.this.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.this.id
  }

  tags = merge(var.tags, {
    Name = "${var.name}-public-rt"
  })
}

resource "aws_route_table_association" "public" {
  for_each = aws_subnet.public

  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

# ---------------------------------------------------------------------------
# NAT gateways (one EIP per gateway) and the private tier
# ---------------------------------------------------------------------------

resource "aws_eip" "nat" {
  count = var.nat_gateway_count

  domain = "vpc"

  tags = merge(var.tags, {
    Name = "${var.name}-nat-eip-${count.index}"
  })
}

resource "aws_nat_gateway" "this" {
  count = var.nat_gateway_count

  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[local.azs[count.index % var.az_count]].id

  tags = merge(var.tags, {
    Name = "${var.name}-natgw-${count.index}"
  })
}

resource "aws_subnet" "private" {
  for_each = { for i, az in local.azs : az => {
    index = i
    cidr  = local.private_cidrs[i]
  } }

  vpc_id            = aws_vpc.this.id
  availability_zone = each.key
  cidr_block        = each.value.cidr

  tags = merge(var.tags, {
    Name    = "${var.name}-private-${each.key}"
    Tier    = "private"
    Purpose = "application"
  })
}

# One private route table per AZ; each routes to NAT gateway index
# `az_index % nat_gateway_count` so a single NAT serves all AZs (cost mode)
# or each AZ gets its own gateway (HA mode).
resource "aws_route_table" "private" {
  for_each = { for i, az in local.azs : az => i }

  vpc_id = aws_vpc.this.id

  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.this[each.value % var.nat_gateway_count].id
  }

  tags = merge(var.tags, {
    Name = "${var.name}-private-rt-${each.key}"
  })
}

resource "aws_route_table_association" "private" {
  for_each = aws_subnet.private

  subnet_id      = each.value.id
  route_table_id = aws_route_table.private[each.key].id
}
