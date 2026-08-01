export type EventType = "migration" | "drift" | "review" | "connection" | "deploy";

export interface ProjectEvent {
  id: string;
  type: EventType;
  title: string;
  detail: string;
  actor: string;
  time: string;
}

export const eventTypeConfig: Record<EventType, { label: string; dot: string }> = {
  migration: { label: "Migration", dot: "bg-blue-500" },
  drift: { label: "Drift", dot: "bg-amber-500" },
  review: { label: "Review", dot: "bg-emerald-500" },
  connection: { label: "Connection", dot: "bg-purple-500" },
  deploy: { label: "Deploy", dot: "bg-slate-500" },
};

export const projectEvents: ProjectEvent[] = [
  { id: "e1", type: "migration", title: "m_0042 approved", detail: "add_ltv_column approved by Aarav Mehta", actor: "Aarav Mehta", time: "10 min ago" },
  { id: "e2", type: "migration", title: "m_0041 applied", detail: "Applied to Staging by Priya Sharma", actor: "Priya Sharma", time: "45 min ago" },
  { id: "e3", type: "drift", title: "Drift detected on sessions", detail: "user_agent column missing in Production", actor: "SchemaHub Bot", time: "3 hours ago" },
  { id: "e4", type: "connection", title: "New connection linked", detail: "Analytics Warehouse added by Rahul Verma", actor: "Rahul Verma", time: "2 hours ago" },
  { id: "e5", type: "review", title: "Review #12 requested", detail: "users_status_type change submitted by Priya Sharma", actor: "Priya Sharma", time: "5 hours ago" },
  { id: "e6", type: "deploy", title: "Production branch created", detail: "branch-prod-2026-01 created for release v1.4.0", actor: "Aarav Mehta", time: "8 hours ago" },
  { id: "e7", type: "review", title: "Review #11 closed", detail: "sessions_ip_index rejected by Sneha Patel", actor: "Sneha Patel", time: "1 day ago" },
  { id: "e8", type: "migration", title: "m_0038 rolled back", detail: "Rolled back from Production automatically", actor: "SchemaHub Bot", time: "2 days ago" },
];
