export type DriftSeverity = "critical" | "warning" | "info";
export type DriftStatus = "unresolved" | "acknowledged" | "resolved";

export interface DriftItem {
  id: string;
  table: string;
  schema: string;
  env: string;
  kind: string;
  detail: string;
  severity: DriftSeverity;
  status: DriftStatus;
  detected: string;
}

export interface DriftDiffRow {
  type: "added" | "removed" | "changed";
  column: string;
  staging: string;
  production: string;
}

export const driftSeverityConfig: Record<DriftSeverity, { label: string; badge: "destructive" | "secondary" | "outline" }> = {
  critical: { label: "Critical", badge: "destructive" },
  warning: { label: "Warning", badge: "secondary" },
  info: { label: "Info", badge: "outline" },
};

export const driftStatusConfig: Record<DriftStatus, { label: string; badge: "default" | "secondary" | "outline" }> = {
  unresolved: { label: "Unresolved", badge: "default" },
  acknowledged: { label: "Acknowledged", badge: "secondary" },
  resolved: { label: "Resolved", badge: "outline" },
};

export const driftItems: DriftItem[] = [
  {
    id: "d1",
    table: "sessions",
    schema: "auth",
    env: "Staging",
    kind: "Column added",
    detail: "sessions.user_agent added in staging but missing in production",
    severity: "critical",
    status: "unresolved",
    detected: "3 hours ago",
  },
  {
    id: "d2",
    table: "events",
    schema: "analytics",
    env: "Production",
    kind: "Index missing",
    detail: "idx_events_occurred missing in production",
    severity: "warning",
    status: "acknowledged",
    detected: "1 day ago",
  },
  {
    id: "d3",
    table: "users",
    schema: "public",
    env: "Development",
    kind: "Type mismatch",
    detail: "users.status is varchar(20) in dev vs enum in prod",
    severity: "warning",
    status: "resolved",
    detected: "3 days ago",
  },
];

export const driftDiffs: Record<string, DriftDiffRow[]> = {
  d1: [
    { type: "added", column: "user_agent", staging: "varchar(512)", production: "—" },
    { type: "changed", column: "created_at", staging: "timestamptz DEFAULT now()", production: "timestamptz" },
  ],
  d2: [
    { type: "added", column: "idx_events_occurred", staging: "BTREE (occurred)", production: "—" },
    { type: "added", column: "idx_events_payload", staging: "GIN (payload)", production: "—" },
  ],
  d3: [
    { type: "changed", column: "status", staging: "varchar(20)", production: "enum('active','inactive','suspended')" },
  ],
};
