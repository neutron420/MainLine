export type AuditAction = "create" | "update" | "delete" | "approve" | "reject" | "run" | "login";

export interface AuditEntry {
  id: string;
  actor: string;
  action: AuditAction;
  resource: string;
  detail: string;
  time: string;
}

export const auditActionConfig: Record<AuditAction, { label: string; badge: "default" | "secondary" | "destructive" | "outline" }> = {
  create: { label: "Create", badge: "default" },
  update: { label: "Update", badge: "secondary" },
  delete: { label: "Delete", badge: "destructive" },
  approve: { label: "Approve", badge: "default" },
  reject: { label: "Reject", badge: "destructive" },
  run: { label: "Run", badge: "outline" },
  login: { label: "Login", badge: "outline" },
};

export const auditLog: AuditEntry[] = [
  { id: "a1", actor: "Aarav Mehta", action: "approve", resource: "Migration m_0042", detail: "Approved add_ltv_column migration", time: "10 min ago" },
  { id: "a2", actor: "Priya Sharma", action: "run", resource: "Migration m_0041", detail: "Applied migration to Staging", time: "45 min ago" },
  { id: "a3", actor: "Rahul Verma", action: "create", resource: "Connection", detail: "Linked Analytics Warehouse database", time: "2 hours ago" },
  { id: "a4", actor: "Priya Sharma", action: "update", resource: "Table users", detail: "Changed column users.status type", time: "5 hours ago" },
  { id: "a5", actor: "Sneha Patel", action: "reject", resource: "Migration m_0039", detail: "Rejected sessions_ip_index migration", time: "1 day ago" },
  { id: "a6", actor: "Aarav Mehta", action: "delete", resource: "Table old_campaigns", detail: "Dropped table in Development", time: "2 days ago" },
  { id: "a7", actor: "Vikram Singh", action: "login", resource: "Session", detail: "Signed in from Mumbai, IN", time: "3 days ago" },
  { id: "a8", actor: "Priya Sharma", action: "approve", resource: "Migration m_0038", detail: "Approved add_events_payload_gin migration", time: "4 days ago" },
];
