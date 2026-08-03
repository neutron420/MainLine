# SchemaHub Backup & Disaster Recovery

> Companion to `docs/DEPLOYMENT.md` (§ Backup Strategy). This document is the
> operational runbook: what is backed up, how to restore it, what we promise
> (RTO/RPO), and how to rehearse the recovery.

## 1. What needs protecting

| Component | State | Backup mechanism | Owner |
|---|---|---|---|
| Control-plane PostgreSQL (RDS on AWS / Neon) | SchemaHub's own data: users, projects, connections, schemas, migrations, audit logs, drift events | RDS automated snapshots + PITR (or Neon PITR + branching) | Platform |
| Redis (ElastiCache / compose) | Cache + pub/sub + rate-limit counters | Redis AOF + RDB snapshots (or ElastiCache snapshots) | Platform |
| Migrations + schema metadata | Rebuildable from DB (source of truth is Postgres itself) | Included in the DB backups | — |
| Envoy / backend / frontend config | Stateless; configuration is code | Git (IaC) | Platform |
| Connection credentials (JWT keys, master key) | Secrets | Secret manager (SecretsManager/SSM), Git-crypt-free — never in repo | Platform |

**Application state**: the backend and frontend are stateless containers. There
is no application data to back up; recovery is "redeploy the last good image".

## 2. Backup strategy

### 2.1 PostgreSQL — RDS (AWS deployment)

- **Automated snapshots**: AWS takes daily snapshots automatically.
  `backup_retention_period = 7` (Terraform: `infra/terraform/modules/rds`).
- **PITR**: RDS enables point-in-time recovery **up to the end of the backup
  retention window** (7 days) using transaction logs — recover to any second
  within that window. RPO ≈ 5 minutes.
- **Manual snapshots**: take a manual snapshot before any destructive
  operation (major migration, bulk delete, schema change):

  ```bash
  aws rds create-db-snapshot \
    --db-instance-identifier schemahub-prod \
    --db-snapshot-identifier schemahub-pre-migration-$(date +%Y%m%d-%H%M%S)
  ```

- **pg_dump for portability**: snapshots are region-scoped. Keep a logical
  backup that can be restored anywhere (or into Neon):

  ```bash
  # On a host with network access to the DB (bastion / CI runner)
  pg_dump \
    "postgres://schemahub:REDACTED@schemahub-prod.xxxx.us-east-1.rds.amazonaws.com:5432/schemahub?sslmode=require" \
    --format=custom --no-owner --compress=9 \
    -f schemahub-$(date +%Y%m%d-%H%M%S).dump
  # Upload off-site (S3 bucket, versioned, cross-region replication)
  aws s3 cp schemahub-*.dump s3://schemahub-backups/pgdump/ --storage-class STANDARD_IA
  ```

  Schedule: nightly cron/systemd timer or the CI scheduler; retain 30 days on
  S3 + 90 days in Glacier.

### 2.2 PostgreSQL — Neon (primary recommended deployment)

Neon is the documented primary database for SchemaHub (`docs/TECH_STACK.md`).
Backups are managed by Neon:

- **PITR**: continuous WAL archiving gives point-in-time recovery for the
  configured retention period (7 days default on paid plans).
- **Snapshots**: Neon takes daily/weekly snapshots automatically (retention
  per plan — see the Neon console).
- **Branching as DR + workflow tool**: create a branch of the production
  database to:
  - restore to a point in time (branch **from a point**),
  - test migrations in isolation (`docs/DATABASE_DESIGN.md` § workspace
    isolation),
  - promote the branch back to primary if the original is compromised.

  ```bash
  # neonctl — create a PITR branch from 30 minutes ago
  neonctl branches create --name dr-drill-$(date +%s) \
    --parent schemahub-prod --parent-timestamp "$(date -u -d '30 minutes ago' +%Y-%m-%dT%H:%M:%SZ)"
  # promote branch (only when the drill proves the branch is sound)
  neonctl branches promote dr-drill-...
  ```

  Neon branching gives **database-level rollback with zero downtime** for the
  "data corruption" scenario — the fastest DR path in the portfolio.

### 2.3 Redis

- **Compose / self-managed** (`docker/redis/redis.conf`):
  - AOF enabled (`appendonly yes`, `appendfsync everysec`) — durability to ≤
    1 s of writes.
  - RDB snapshots (`save 900 1 / 300 10 / 60 10000`) for fast cold restores.
  - **Treat Redis as a cache**: caches repopulate from the DB; pub/sub
    subscribers reconnect and the audit log is in Postgres. RTO for Redis is
    "restart the container" — data loss beyond the last appendfsync is
    acceptable and designed for.
- **ElastiCache** (`infra/terraform/modules/redis`): `snapshot_retention_limit
  = 7` — automatic daily snapshots with 7-day retention.
  ```bash
  aws elasticache create-snapshot --replication-group-id schemahub-prod-redis \
    --snapshot-name schemahub-redis-$(date +%Y%m%d-%H%M%S)
  ```

## 3. Restore runbooks

### Runbook A — RDS PITR restore (corruption / bad migration)

1. Stop write traffic: scale the ECS service to 0 or block the ALB listener.
2. Restore to a point **before** the corruption:
   ```bash
   aws rds restore-db-instance-to-point-in-time \
     --source-db-instance-identifier schemahub-prod \
     --target-db-instance-identifier schemahub-prod-restore \
     --restore-time 2026-08-03T14:30:00Z \
     --db-instance-class db.t4g.small \
     --multi-az \
     --storage-encrypted
   ```
3. Verify: `pg_isready`, row counts, checksum of latest migration record,
   spot-check audit log timestamps.
4. Point the app at the restored instance (update the SecretsManager
   `DATABASE_URL` secret), roll back ECS tasks.
5. Promote: update Route53 / listener target or rename the instance (rename
   is disruptive — prefer the secret update path).
6. Record the restore timestamp in the incident log.

### Runbook B — Neon PITR branch promotion (data corruption)

1. Create a branch from the restore point (see § 2.2 command).
2. Validate on the branch (read-only queries, schema checksum comparison via
   the SchemaHub schema-version endpoints).
3. Update `DATABASE_URL` (env/secret) to the branch's connection string and
   redeploy; or promote the branch to primary via the Neon console / neonctl.
4. Point the app at the (new) primary. Because Neon keeps the original
   database intact, rollback is trivial — promote again from another point.

### Runbook C — Restore pg_dump logical backup (cross-provider)

```bash
pg_restore --no-owner --no-privileges \
  --dbname "postgres://schemahub:REDACTED@<target>:5432/schemahub?sslmode=require" \
  schemahub-20260803-030000.dump
```

Prerequisites: target database exists, extensions enabled by migration 001
(`pgcrypto`), schema `public` writable. Run migrations 002+ afterward if the
dump predates them.

### Runbook D — Redis restore

```bash
# Self-managed: copy RDB/AOF into a fresh container data dir and start
docker compose -f docker/docker-compose.yml up -d redis
# ElastiCache:
aws elasticache copy-snapshot \
  --source-snapshot-name schemahub-redis-20260803 \
  --target-snapshot-name schemahub-redis-dr-restore
# then restore into a new replication group (ElastiCache restores into a
# new cluster; repoint REDIS_URL via the secret).
```

### Runbook E — Full regional recovery (region loss)

1. Fail over or recreate in the DR region from IaC: `terraform apply` with
   `region = us-east-2` (state must live in S3 for cross-region apply —
   `infra/terraform/provider.tf`).
2. Restore RDS from a cross-region snapshot copy (S3→RDS or RDS snapshot
   copy); or use Neon — Neon is region-agnostic by construction.
3. Restore SecretsManager/SSM secrets (backup the secret ARNs + values into
   the DR account or use AWS Backup).
4. Update DNS to the DR ALB DNS name.

## 4. RTO / RPO targets

| Scenario | RTO | RPO | Method |
|---|---|---|---|
| Single backend task dies | < 1 min | 0 | Fargate replaces; ALB drains |
| AZ failure (compute) | < 5 min | 0 | Service spread across AZs + autoscaling |
| RDS instance failure (RDS) | < 15 min | < 5 min | Multi-AZ failover or snapshot restore |
| Neon compute failure | < 5 min | 0 | Neon restarts compute automatically |
| Data corruption (RDS) | < 60 min | < 5 min | PITR restore (Runbook A) |
| Data corruption (Neon) | < 15 min | < 5 min | Branch from PITR point (Runbook B) |
| Region loss | < 4 h | < 1 h | Cross-region restore (Runbook E) |
| Redis loss | < 15 min | ≤ 1 s (AOF) / 0 (cache refill) | Restart / snapshot restore (Runbook D) |

These are **targets**, not guarantees — they hold only if a DR drill validated
them in the current quarter (see below).

## 5. DR drill checklist (quarterly)

- [ ] RDS: manual snapshot created and verified restorable (`aws rds
      describe-db-snapshots --snapshot-identifier ...` returns
      `status=available`).
- [ ] RDS PITR drill: restore to a point 30 minutes back, run
      `pg_restore`-less validation queries (Runbook A), tear down the restore.
- [ ] Neon: create a branch from yesterday's PITR point, run the migration
      suite against it, compare schema checksums with production.
- [ ] pg_dump: a fresh dump is present in the S3 backup bucket and its size
      is non-zero (alert if the job has not run in > 48 h).
- [ ] Redis: ElastiCache snapshot exists; compose Redis restarts cleanly from
      its AOF after `kill -9`.
- [ ] Secrets: `DATABASE_URL` / `REDIS_URL` restore path works (update secret
      → task redeploy → app healthy).
- [ ] Verify RTO: time the restore of a snapshot end-to-end; record actual
      vs. target in the incident log.
- [ ] Confirm backup retention is still aligned with compliance (SOC2/HIPAA
      backups-retention evidence).

## 6. Links

- `docs/DEPLOYMENT.md` — deployment topology and the original backup table
- `docs/NEON_COMPATIBILITY.md` — Neon-specific audit (branching, pooled mode)
- `infra/terraform/modules/rds` — snapshot/PITR/encryption configuration
- `infra/terraform/modules/redis` — ElastiCache snapshot configuration
- `docker/redis/redis.conf` — local Redis durability settings
