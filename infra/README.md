# SchemaHub Infrastructure

Production-grade IaC and monitoring for the SchemaHub platform.

## Layout

```
infra/
├── terraform/            # Terraform IaC (AWS)
│   ├── main.tf           # Root module: wires vpc + rds + redis + ecs
│   ├── variables.tf      # Root inputs (secrets have no defaults)
│   ├── outputs.tf        # ALB endpoint, endpoints, SSM parameter names
│   ├── provider.tf       # AWS provider + S3 backend note (commented)
│   ├── environments/     # dev / staging / production tfvars examples
│   └── modules/
│       ├── vpc/          # VPC, public/private subnets, IGW, NAT, route tables
│       ├── rds/          # RDS PostgreSQL 16, encrypted, backups 7d, PITR
│       ├── redis/        # ElastiCache Redis (cluster mode disabled), backups
│       └── ecs/          # Fargate backend + ALB (h2c gRPC), SecretsManager
└── monitoring/
    ├── prometheus.yml    # Scrape config (targets = compose service names)
    ├── alerts.yml        # Alert rules (backend down, gRPC error rate)
    ├── grafana/          # Grafana provisioning (datasource + dashboard provider)
    └── grafana-dashboards/
        └── schemahub.json # Full dashboard: per-service rate/latency/errors, DB pool, Redis, runtime
```

## What the backend exposes for monitoring

- **`/metrics`** on port **9091** (env `METRICS_PORT`): Prometheus endpoint via
  `promhttp`. Serves Go runtime/process metrics today; wire the standard
  grpc-go and pgxpool collectors to populate the gRPC/pool panels.
- Structured `slog` logs (JSON) — ship them to Loki/CloudWatch and alert on
  `level=error`.
- gRPC health via reflection (`/grpc.reflection.v1alpha.ServerReflection`).

## Terraform (AWS)

Modules implement: VPC with public/private subnets + NAT; RDS Postgres 16
(encrypted, 7-day backups, deletion protection); ElastiCache Redis; Fargate
backend behind an ALB (HTTP/2 h2c to the gRPC port, health check against
`/metrics`).

Secrets policy: nothing sensitive has a default. `db_password`,
`jwt_private_key`, `jwt_public_key`, `encryption_key` must be supplied in
`terraform.tfvars` (never committed). Connection URLs land in SSM SecureString;
app secrets land in SecretsManager and are referenced from the task definition
via `valueFrom` — never in the container env.

```bash
cd infra/terraform
cp environments/dev/terraform.tfvars.example terraform.tfvars
# edit terraform.tfvars with real values
terraform init
terraform plan
terraform apply
```

Never commit `*.tfstate` or `*.tfplan` (gitignored).

## Prometheus + Grafana (local compose)

`docker-compose.yml` includes `prometheus` (scraping `backend:9091` and
`envoy:9901`) and `grafana` (port 3001, auto-provisioned datasource +
dashboard):

```bash
docker compose -f docker/docker-compose.yml up prometheus grafana
# Grafana: http://localhost:3001 (admin / GRAFANA_ADMIN_PASSWORD, default admin)
# Dashboard: SchemaHub
```

For a managed deployment, run Prometheus against the same `prometheus.yml`
and import `schemahub.json` with the `DS_PROMETHEUS` variable pointing at it.
The Redis panels expect a redis_exporter scrape of the ElastiCache endpoint
(job is commented out in `prometheus.yml` until an exporter is deployed).
