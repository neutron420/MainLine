export interface Change {
  action: "add" | "remove" | "modify";
  column: string;
  type: string;
  note?: string;
}

export interface Reviewer {
  name: string;
  initials: string;
  decision: "approved" | "changes" | "pending";
}

export interface ReviewItem {
  id: string;
  title: string;
  project: string;
  table: string;
  version: string;
  database: string;
  author: string;
  initials: string;
  status: "pending" | "approved" | "changes" | "rejected";
  priority: "high" | "medium" | "low";
  time: string;
  description: string;
  changes: Change[];
  reviewers: Reviewer[];
}

export const reviews: ReviewItem[] = [
  {
    id: "r1",
    title: "Add users table composite index",
    project: "User Service",
    table: "users",
    version: "v1.2.0",
    database: "user_service_prod",
    author: "Alice",
    initials: "AL",
    status: "pending",
    priority: "high",
    time: "1 hour ago",
    description: "Queries filtering by email and status are slow on the users table. A composite B-tree index on (email, status) should cover the auth lookup path.",
    changes: [
      { action: "add", column: "idx_users_email_status", type: "index", note: "B-tree on (email, status)" },
    ],
    reviewers: [
      { name: "Bob", initials: "BO", decision: "pending" },
      { name: "Diana", initials: "DI", decision: "pending" },
    ],
  },
  {
    id: "r2",
    title: "Update payment status enum values",
    project: "Payment DB",
    table: "payments",
    version: "v0.9.1",
    database: "payments_prod",
    author: "Bob",
    initials: "BO",
    status: "pending",
    priority: "medium",
    time: "3 hours ago",
    description: "Finance workflow needs refunded and disputed states to track chargebacks. The status enum gets two new values; no existing rows are affected.",
    changes: [
      { action: "modify", column: "status", type: "enum", note: "Add 'refunded' and 'disputed' values" },
    ],
    reviewers: [
      { name: "Charlie", initials: "CH", decision: "pending" },
    ],
  },
  {
    id: "r3",
    title: "Drop legacy zip_code column",
    project: "CRM",
    table: "customers",
    version: "v2.1.0",
    database: "crm_prod",
    author: "Charlie",
    initials: "CH",
    status: "pending",
    priority: "low",
    time: "1 day ago",
    description: "zip_code was replaced by address_lookup_id in v2.0. All reads have been migrated; safe to drop the legacy column.",
    changes: [
      { action: "remove", column: "zip_code", type: "varchar(10)", note: "Replaced by address_lookup_id" },
    ],
    reviewers: [
      { name: "Alice", initials: "AL", decision: "pending" },
    ],
  },
  {
    id: "r4",
    title: "Add email verification column",
    project: "User Service",
    table: "users",
    version: "v1.2.1",
    database: "user_service_prod",
    author: "Alice",
    initials: "AL",
    status: "changes",
    priority: "medium",
    time: "2 days ago",
    description: "Email verification flow stores the timestamp and a one-time token. Token must be unique and expire after 24 hours.",
    changes: [
      { action: "add", column: "email_verified_at", type: "timestamptz", note: "NULL until verified" },
      { action: "add", column: "email_verification_token", type: "uuid", note: "One-time token" },
    ],
    reviewers: [
      { name: "Bob", initials: "BO", decision: "changes" },
      { name: "Diana", initials: "DI", decision: "approved" },
    ],
  },
  {
    id: "r5",
    title: "Add orders checkout_flow column",
    project: "Billing System",
    table: "orders",
    version: "v1.0.0",
    database: "billing_prod",
    author: "Diana",
    initials: "DI",
    status: "approved",
    priority: "medium",
    time: "5 hours ago",
    description: "New checkout flow needs to tag orders with the flow variant for A/B reporting. Defaults to checkout_v2.",
    changes: [
      { action: "add", column: "checkout_flow", type: "varchar(32)", note: "checkout_v2 default" },
    ],
    reviewers: [
      { name: "Eve", initials: "EV", decision: "approved" },
    ],
  },
  {
    id: "r6",
    title: "Rename stock_level to available_stock",
    project: "Inventory Service",
    table: "products",
    version: "v3.0.2",
    database: "inventory_prod",
    author: "Eve",
    initials: "EV",
    status: "approved",
    priority: "high",
    time: "1 day ago",
    description: "stock_level is ambiguous with reserved stock. Rename with a view shim so the old name keeps working for one release cycle.",
    changes: [
      { action: "modify", column: "stock_level", type: "rename", note: "→ available_stock with view shim" },
    ],
    reviewers: [
      { name: "Frank", initials: "FR", decision: "approved" },
      { name: "Alice", initials: "AL", decision: "approved" },
    ],
  },
  {
    id: "r7",
    title: "Add search_vector generated column",
    project: "Search Index",
    table: "documents",
    version: "v0.8.1",
    database: "search_prod",
    author: "Ivy",
    initials: "IV",
    status: "approved",
    priority: "medium",
    time: "2 days ago",
    description: "Generated tsvector column from title + body for fast full-text search, backed by a GIN index.",
    changes: [
      { action: "add", column: "search_vector", type: "tsvector", note: "Generated from title + body" },
      { action: "add", column: "idx_documents_search", type: "GIN index" },
    ],
    reviewers: [
      { name: "Jack", initials: "JA", decision: "approved" },
    ],
  },
  {
    id: "r8",
    title: "Remove deprecated session_token column",
    project: "Auth Provider Sync",
    table: "sessions",
    version: "v1.4.0",
    database: "auth_prod",
    author: "Jack",
    initials: "JA",
    status: "rejected",
    priority: "high",
    time: "3 days ago",
    description: "session_token is unused since refresh tokens landed. Legacy mobile clients still call an endpoint that reads it, so removal is risky.",
    changes: [
      { action: "remove", column: "session_token", type: "text", note: "Still referenced by legacy clients" },
    ],
    reviewers: [
      { name: "Hank", initials: "HA", decision: "changes" },
      { name: "Ivy", initials: "IV", decision: "approved" },
    ],
  },
  {
    id: "r9",
    title: "Add notification_preferences jsonb",
    project: "Notification Queue",
    table: "users",
    version: "v0.5.0",
    database: "notif_prod",
    author: "Hank",
    initials: "HA",
    status: "pending",
    priority: "low",
    time: "4 days ago",
    description: "Store per-user channel and digest toggles as jsonb. Flexible enough to add channels without migrations.",
    changes: [
      { action: "add", column: "notification_preferences", type: "jsonb", note: "Channel + digest toggles" },
    ],
    reviewers: [
      { name: "Grace", initials: "GR", decision: "pending" },
    ],
  },
  {
    id: "r10",
    title: "Add metric_events partitioning",
    project: "Analytics Warehouse",
    table: "events",
    version: "v2.2.0",
    database: "analytics_prod",
    author: "Frank",
    initials: "FR",
    status: "approved",
    priority: "high",
    time: "5 days ago",
    description: "events is the largest table in the warehouse. Range partitioning by created_at monthly improves maintenance and query pruning.",
    changes: [
      { action: "modify", column: "events", type: "partitioned", note: "Range partition by created_at, monthly" },
    ],
    reviewers: [
      { name: "Diana", initials: "DI", decision: "approved" },
      { name: "Grace", initials: "GR", decision: "approved" },
    ],
  },
  {
    id: "r11",
    title: "Add support ticket priority enum",
    project: "Support Ticket DB",
    table: "tickets",
    version: "v1.1.0",
    database: "support_prod",
    author: "Grace",
    initials: "GR",
    status: "changes",
    priority: "medium",
    time: "1 week ago",
    description: "Priority levels for SLA routing. Enum with low / medium / high / urgent values.",
    changes: [
      { action: "add", column: "priority", type: "enum", note: "low / medium / high / urgent" },
    ],
    reviewers: [
      { name: "Hank", initials: "HA", decision: "changes" },
    ],
  },
];

export const statusConfig = {
  pending: { label: "Pending", dot: "bg-yellow-500", badge: "secondary" as const },
  approved: { label: "Approved", dot: "bg-green-500", badge: "default" as const },
  changes: { label: "Changes Requested", dot: "bg-amber-500", badge: "outline" as const },
  rejected: { label: "Rejected", dot: "bg-red-500", badge: "destructive" as const },
};

export const priorityConfig = {
  high: { label: "High", variant: "destructive" as const },
  medium: { label: "Medium", variant: "default" as const },
  low: { label: "Low", variant: "secondary" as const },
};

export const projectOptions = [
  "User Service",
  "Payment DB",
  "CRM",
  "Billing System",
  "Inventory Service",
  "Search Index",
  "Auth Provider Sync",
  "Notification Queue",
  "Analytics Warehouse",
  "Support Ticket DB",
];
