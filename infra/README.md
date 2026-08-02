# SchemaHub Infrastructure (Skeleton)

> **Status: Skeleton only.** This directory is scaffolding for post-MVP
> provisioning. No resource is deployed yet — see `terraform/environments/*`.

## Layout

```
infra/
├── terraform/            # Terraform IaC
│   ├── environments/     # dev / staging / production workspaces
│   └── modules/          # database, redis, backend-service, networking
└── monitoring/
    ├── prometheus.yml    # Prometheus scrape config
    └── grafana-dashboards/
        └── schemahub.json # Grafana dashboard (placeholder)
```

## What the backend exposes for monitoring

- gRPC Prometheus metrics via `interceptor.MetricsInterceptor` (register a
  prometheus registry in `cmd/server` and add the standard grpc-go metrics).
- Structured `slog` logs (JSON) — ship them to Loki/CloudWatch and alert on
  `level=error`.
- Healthcheck: gRPC reflection is enabled (`/grpc.reflection.v1alpha.ServerReflection`).

## Terraform

The modules in `terraform/modules/` are declared but empty. To provision:

1. Fill in each module's `main.tf`/`variables.tf`/`outputs.tf`
   (`database` → Neon/managed Postgres or RDS, `redis` → Upstash/ElastiCache,
   `backend-service` → Railway/Fly/AWS ECS, `networking` → VPC/security groups).
2. Copy `environments/dev/terraform.tfvars.example` → `terraform.tfvars`.
3. `terraform init && terraform plan && terraform apply`

Never commit `*.tfstate` or `*.tfplan` (already gitignored).

## Prometheus

`monitoring/prometheus.yml` is ready to scrape a locally-running backend that
exposes metrics on `:9090/metrics` (set `METRICS_ADDR` in the backend once the
metrics endpoint is wired).

## Grafana

`monitoring/grafana-dashboards/schemahub.json` is a placeholder dashboard
(Overview + Schema + Migrations + Real-time + Database rows). Import it into
Grafana once Prometheus is up.
