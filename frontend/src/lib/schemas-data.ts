export type TableStatus = "verified" | "pending" | "drift";

export interface SchemaColumn {
  name: string;
  type: string;
  nullable: boolean;
  key?: "PK" | "FK";
  default?: string;
}

export interface SchemaIndex {
  name: string;
  columns: string[];
  type: string;
  unique: boolean;
}

export interface SchemaRelation {
  name: string;
  from: string;
  to: string;
}

export interface SchemaTable {
  id: string;
  name: string;
  status: TableStatus;
  columns: SchemaColumn[];
  indexes: SchemaIndex[];
  relations: SchemaRelation[];
}

export interface SchemaGroup {
  id: string;
  name: string;
  tables: SchemaTable[];
}

export interface SchemaProject {
  id: string;
  name: string;
  schemas: SchemaGroup[];
}

export const schemaProjects: SchemaProject[] = [
  {
    id: "p1",
    name: "User Service",
    schemas: [
      {
        id: "s1",
        name: "public",
        tables: [
          {
            id: "t-users",
            name: "users",
            status: "verified",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "email", type: "varchar(255)", nullable: false },
              { name: "password_hash", type: "text", nullable: false },
              { name: "status", type: "varchar(20)", nullable: false, default: "'active'" },
              { name: "email_verified_at", type: "timestamptz", nullable: true },
              { name: "created_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_users_email", columns: ["email"], type: "btree", unique: true },
              { name: "idx_users_email_status", columns: ["email", "status"], type: "btree", unique: false },
              { name: "idx_users_created_at", columns: ["created_at"], type: "btree", unique: false },
            ],
            relations: [],
          },
          {
            id: "t-teams",
            name: "teams",
            status: "verified",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "name", type: "varchar(100)", nullable: false },
              { name: "slug", type: "varchar(60)", nullable: false },
              { name: "created_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_teams_slug", columns: ["slug"], type: "btree", unique: true },
            ],
            relations: [],
          },
          {
            id: "t-memberships",
            name: "memberships",
            status: "pending",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "user_id", type: "uuid", nullable: false, key: "FK" },
              { name: "team_id", type: "uuid", nullable: false, key: "FK" },
              { name: "role", type: "varchar(20)", nullable: false, default: "'member'" },
              { name: "joined_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_memberships_user", columns: ["user_id"], type: "btree", unique: false },
              { name: "idx_memberships_team", columns: ["team_id"], type: "btree", unique: false },
            ],
            relations: [
              { name: "memberships_user_id_fkey", from: "memberships.user_id", to: "users.id" },
              { name: "memberships_team_id_fkey", from: "memberships.team_id", to: "teams.id" },
            ],
          },
        ],
      },
      {
        id: "s2",
        name: "auth",
        tables: [
          {
            id: "t-sessions",
            name: "sessions",
            status: "drift",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "user_id", type: "uuid", nullable: false, key: "FK" },
              { name: "refresh_token_hash", type: "text", nullable: false },
              { name: "expires_at", type: "timestamptz", nullable: false },
              { name: "revoked_at", type: "timestamptz", nullable: true },
              { name: "created_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_sessions_user", columns: ["user_id"], type: "btree", unique: false },
              { name: "idx_sessions_refresh", columns: ["refresh_token_hash"], type: "btree", unique: true },
            ],
            relations: [
              { name: "sessions_user_id_fkey", from: "sessions.user_id", to: "users.id" },
            ],
          },
          {
            id: "t-oauth_tokens",
            name: "oauth_tokens",
            status: "verified",
            columns: [
              { name: "id", type: "bigint", nullable: false, key: "PK", default: "identity" },
              { name: "provider", type: "varchar(30)", nullable: false },
              { name: "provider_user_id", type: "varchar(120)", nullable: false },
              { name: "access_token", type: "text", nullable: true },
              { name: "token_expires_at", type: "timestamptz", nullable: true },
              { name: "created_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_oauth_provider_user", columns: ["provider", "provider_user_id"], type: "btree", unique: true },
            ],
            relations: [],
          },
        ],
      },
    ],
  },
  {
    id: "p2",
    name: "Payment DB",
    schemas: [
      {
        id: "s3",
        name: "public",
        tables: [
          {
            id: "t-payments",
            name: "payments",
            status: "pending",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "user_id", type: "uuid", nullable: false, key: "FK" },
              { name: "amount", type: "numeric(12,2)", nullable: false },
              { name: "currency", type: "char(3)", nullable: false, default: "'USD'" },
              { name: "status", type: "varchar(20)", nullable: false, default: "'pending'" },
              { name: "paid_at", type: "timestamptz", nullable: true },
            ],
            indexes: [
              { name: "idx_payments_user", columns: ["user_id"], type: "btree", unique: false },
              { name: "idx_payments_status", columns: ["status"], type: "btree", unique: false },
            ],
            relations: [
              { name: "payments_user_id_fkey", from: "payments.user_id", to: "users.id" },
            ],
          },
          {
            id: "t-invoices",
            name: "invoices",
            status: "verified",
            columns: [
              { name: "id", type: "uuid", nullable: false, key: "PK", default: "gen_random_uuid()" },
              { name: "payment_id", type: "uuid", nullable: true, key: "FK" },
              { name: "number", type: "varchar(32)", nullable: false },
              { name: "amount", type: "numeric(12,2)", nullable: false },
              { name: "issued_at", type: "timestamptz", nullable: false, default: "now()" },
              { name: "status", type: "varchar(20)", nullable: false, default: "'draft'" },
            ],
            indexes: [
              { name: "idx_invoices_number", columns: ["number"], type: "btree", unique: true },
              { name: "idx_invoices_payment", columns: ["payment_id"], type: "btree", unique: false },
            ],
            relations: [
              { name: "invoices_payment_id_fkey", from: "invoices.payment_id", to: "payments.id" },
            ],
          },
          {
            id: "t-ledger_entries",
            name: "ledger_entries",
            status: "verified",
            columns: [
              { name: "id", type: "bigint", nullable: false, key: "PK", default: "identity" },
              { name: "account_id", type: "uuid", nullable: false },
              { name: "amount", type: "numeric(14,4)", nullable: false },
              { name: "entry_type", type: "varchar(10)", nullable: false },
              { name: "created_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_ledger_account", columns: ["account_id"], type: "btree", unique: false },
              { name: "idx_ledger_created", columns: ["created_at"], type: "btree", unique: false },
            ],
            relations: [],
          },
        ],
      },
      {
        id: "s4",
        name: "analytics",
        tables: [
          {
            id: "t-events",
            name: "events",
            status: "drift",
            columns: [
              { name: "id", type: "bigint", nullable: false, key: "PK", default: "identity" },
              { name: "event_name", type: "varchar(80)", nullable: false },
              { name: "properties", type: "jsonb", nullable: false, default: "'{}'" },
              { name: "occurred_at", type: "timestamptz", nullable: false },
              { name: "ingested_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_events_name", columns: ["event_name"], type: "btree", unique: false },
              { name: "idx_events_occurred", columns: ["occurred_at"], type: "brin", unique: false },
            ],
            relations: [],
          },
          {
            id: "t-daily_metrics",
            name: "daily_metrics",
            status: "verified",
            columns: [
              { name: "metric_date", type: "date", nullable: false, key: "PK" },
              { name: "metric_name", type: "varchar(60)", nullable: false, key: "PK" },
              { name: "value", type: "numeric(20,4)", nullable: false },
              { name: "updated_at", type: "timestamptz", nullable: false, default: "now()" },
            ],
            indexes: [
              { name: "idx_metrics_date", columns: ["metric_date"], type: "btree", unique: false },
            ],
            relations: [],
          },
        ],
      },
    ],
  },
];

export const schemaStatusConfig = {
  verified: { label: "Verified", dot: "bg-green-500", badge: "default" as const, icon: "shield" as const },
  pending: { label: "Pending Review", dot: "bg-yellow-500", badge: "secondary" as const, icon: "clock" as const },
  drift: { label: "Drift", dot: "bg-red-500", badge: "destructive" as const, icon: "alert" as const },
};

export const projectOptions = ["User Service", "Payment DB"];
