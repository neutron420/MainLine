export type MigrationStatus = "deployed" | "inReview" | "inProgress" | "failed" | "rolledBack" | "draft";

export interface MigrationChange {
  action: "add" | "remove" | "modify";
  column: string;
  type: string;
  note?: string;
}

export interface MigrationTimelineEntry {
  label: string;
  user: string;
  time: string;
}

export interface Migration {
  id: string;
  version: string;
  name: string;
  author: string;
  initials: string;
  status: MigrationStatus;
  environment: string;
  database: string;
  table: string;
  created: string;
  duration: string;
  applied: string;
  summary: string;
  sql: string;
  changes: MigrationChange[];
  timeline: MigrationTimelineEntry[];
}

export const migrationStatusConfig: Record<MigrationStatus, { label: string; dot: string; badge: "default" | "secondary" | "destructive" | "outline" }> = {
  deployed: { label: "Deployed", dot: "bg-green-500", badge: "default" },
  inReview: { label: "In Review", dot: "bg-purple-500", badge: "secondary" },
  inProgress: { label: "In Progress", dot: "bg-yellow-500", badge: "secondary" },
  failed: { label: "Failed", dot: "bg-red-500", badge: "destructive" },
  rolledBack: { label: "Rolled Back", dot: "bg-muted-foreground/50", badge: "outline" },
  draft: { label: "Draft", dot: "bg-muted-foreground/50", badge: "outline" },
};

export const environments = ["Production", "Staging", "Development"];

export const migrationsData: Migration[] = [
  {
    id: "m1",
    version: "v1.2.0",
    name: "Add users table composite index",
    author: "Alice",
    initials: "AL",
    status: "deployed",
    environment: "Production",
    database: "user_service_prod",
    table: "users",
    created: "2 hours ago",
    duration: "42s",
    applied: "2 hours ago",
    summary: "Queries filtering by email and status are slow on the users table. A composite B-tree index on (email, status) covers the auth lookup path.",
    sql: `CREATE INDEX idx_users_email_status\n  ON public.users (email, status);`,
    changes: [
      { action: "add", column: "idx_users_email_status", type: "index", note: "B-tree on (email, status)" },
    ],
    timeline: [
      { label: "Created", user: "Alice", time: "3 hours ago" },
      { label: "Submitted for review", user: "Alice", time: "3 hours ago" },
      { label: "Approved", user: "Bob", time: "2 hours ago" },
      { label: "Deployed to Production", user: "System", time: "2 hours ago" },
    ],
  },
  {
    id: "m2",
    version: "v1.1.0",
    name: "Add email verification columns",
    author: "Alice",
    initials: "AL",
    status: "inReview",
    environment: "Staging",
    database: "user_service_prod",
    table: "users",
    created: "3 days ago",
    duration: "—",
    applied: "—",
    summary: "Email verification flow stores the timestamp and a one-time token. Token must be unique and expire after 24 hours.",
    sql: `ALTER TABLE public.users\n  ADD COLUMN email_verified_at timestamptz NULL,\n  ADD COLUMN email_verification_token uuid NULL;`,
    changes: [
      { action: "add", column: "email_verified_at", type: "timestamptz", note: "NULL until verified" },
      { action: "add", column: "email_verification_token", type: "uuid", note: "One-time token" },
    ],
    timeline: [
      { label: "Created", user: "Alice", time: "3 days ago" },
      { label: "Submitted for review", user: "Alice", time: "3 days ago" },
      { label: "Requested changes", user: "Bob", time: "2 days ago" },
      { label: "Re-submitted", user: "Alice", time: "1 day ago" },
    ],
  },
  {
    id: "m3",
    version: "v1.0.3",
    name: "Rename stock_level column",
    author: "Diana",
    initials: "DI",
    status: "rolledBack",
    environment: "Production",
    database: "inventory_prod",
    table: "products",
    created: "5 days ago",
    duration: "1m 05s",
    applied: "4 days ago",
    summary: "stock_level renamed to available_stock with a view shim so the old name keeps working for one release cycle. Rolled back after a query regression.",
    sql: `ALTER TABLE public.products\n  RENAME COLUMN stock_level TO available_stock;\n\nCREATE VIEW public.products_v1 AS\n  SELECT id, available_stock AS stock_level, name\n  FROM public.products;`,
    changes: [
      { action: "modify", column: "stock_level", type: "rename", note: "→ available_stock with view shim" },
    ],
    timeline: [
      { label: "Created", user: "Diana", time: "6 days ago" },
      { label: "Approved", user: "Frank", time: "5 days ago" },
      { label: "Deployed to Production", user: "System", time: "4 days ago" },
      { label: "Rolled back", user: "System", time: "4 days ago" },
    ],
  },
];
