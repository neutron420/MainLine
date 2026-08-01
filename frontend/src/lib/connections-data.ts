export type ConnectionStatus = "connected" | "error" | "delayed" | "disabled";

export interface DbConnection {
  id: string;
  name: string;
  host: string;
  database: string;
  user: string;
  port: number;
  env: string;
  status: ConnectionStatus;
  latency: string;
  lastSync: string;
  version: string;
  ssl: "required" | "prefer" | "disable";
}

export const connectionStatusConfig: Record<
  ConnectionStatus,
  { label: string; badge: "default" | "secondary" | "destructive" | "outline"; dot: string }
> = {
  connected: { label: "Connected", badge: "default", dot: "bg-emerald-500" },
  error: { label: "Error", badge: "destructive", dot: "bg-red-500" },
  delayed: { label: "Delayed", badge: "secondary", dot: "bg-amber-500" },
  disabled: { label: "Disabled", badge: "outline", dot: "bg-muted-foreground/40" },
};

export const dbConnections: DbConnection[] = [
  {
    id: "c1",
    name: "Production",
    host: "ep-prod-42.aws.neon.tech",
    database: "mainline_prod",
    user: "mainline_admin",
    port: 5432,
    env: "Production",
    status: "connected",
    latency: "12ms",
    lastSync: "2 min ago",
    version: "PostgreSQL 16.4",
    ssl: "required",
  },
  {
    id: "c2",
    name: "Staging",
    host: "ep-staging-71.aws.neon.tech",
    database: "mainline_staging",
    user: "mainline_admin",
    port: 5432,
    env: "Staging",
    status: "connected",
    latency: "18ms",
    lastSync: "5 min ago",
    version: "PostgreSQL 16.4",
    ssl: "required",
  },
  {
    id: "c3",
    name: "Development",
    host: "localhost",
    database: "mainline_dev",
    user: "postgres",
    port: 5432,
    env: "Development",
    status: "delayed",
    latency: "340ms",
    lastSync: "23 min ago",
    version: "PostgreSQL 15.6",
    ssl: "disable",
  },
  {
    id: "c4",
    name: "Analytics Warehouse",
    host: "ep-warehouse-90.aws.neon.tech",
    database: "analytics",
    user: "analytics_ro",
    port: 5432,
    env: "Production",
    status: "error",
    latency: "—",
    lastSync: "3 hours ago",
    version: "PostgreSQL 16.1",
    ssl: "prefer",
  },
];
