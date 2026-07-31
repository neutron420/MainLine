"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Database, GitBranch, GitPullRequest, AlertTriangle, FileText, Calendar, Users, Settings, ExternalLink, Clock, CheckCircle2, XCircle, Search, RotateCcw } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { NotificationsPopover } from "@/components/notifications-popover";

const mockDetail = {
  p1: {
    name: "User Service Schema",
    repo: "github.com/mainline/user-service",
    description: "User authentication and profile data models with Row-Level Security policies",
    team: "Platform",
    status: "active" as const,
    tech: "PostgreSQL",
    created: "2026-03-15",
    members: [
      { name: "Alice", initials: "AL", role: "Lead" },
      { name: "Bob", initials: "BO", role: "Contributor" },
    ],
  },
  p2: {
    name: "Payment DB Migration",
    repo: "github.com/mainline/payment-db",
    description: "Payment transaction tables, ledger entries, and reconciliation views",
    team: "Payments",
    status: "inProgress" as const,
    tech: "PostgreSQL",
    created: "2026-04-10",
    members: [
      { name: "Charlie", initials: "CH", role: "Lead" },
    ],
  },
};

const statusConfig = {
  active: { label: "Active", dot: "bg-green-500" },
  inProgress: { label: "In Progress", dot: "bg-yellow-500" },
  onHold: { label: "On Hold", dot: "bg-red-500" },
};

const tabs = [
  { value: "overview", label: "Overview" },
  { value: "schemas", label: "Schemas" },
  { value: "migrations", label: "Migrations" },
  { value: "drift", label: "Drift" },
  { value: "audit", label: "Audit" },
  { value: "events", label: "Events" },
  { value: "settings", label: "Settings" },
];

const tableStatusConfig = {
  verified: { label: "Verified", dot: "bg-green-500", badge: "default" as const },
  pending: { label: "Pending Review", dot: "bg-yellow-500", badge: "secondary" as const },
  drift: { label: "Drift", dot: "bg-red-500", badge: "destructive" as const },
};

const projectTables = {
  p1: [
    { name: "users", schema: "public", columns: 6, indexes: 3, status: "verified" as const, updated: "2h ago" },
    { name: "teams", schema: "public", columns: 4, indexes: 1, status: "verified" as const, updated: "1d ago" },
    { name: "memberships", schema: "public", columns: 5, indexes: 2, status: "pending" as const, updated: "5h ago" },
    { name: "sessions", schema: "auth", columns: 6, indexes: 2, status: "drift" as const, updated: "3h ago" },
    { name: "oauth_tokens", schema: "auth", columns: 6, indexes: 1, status: "verified" as const, updated: "2d ago" },
  ],
  p2: [
    { name: "payments", schema: "public", columns: 6, indexes: 2, status: "pending" as const, updated: "4h ago" },
    { name: "invoices", schema: "public", columns: 6, indexes: 2, status: "verified" as const, updated: "1d ago" },
    { name: "ledger_entries", schema: "public", columns: 5, indexes: 2, status: "verified" as const, updated: "2d ago" },
    { name: "events", schema: "analytics", columns: 5, indexes: 2, status: "drift" as const, updated: "3h ago" },
    { name: "daily_metrics", schema: "analytics", columns: 4, indexes: 1, status: "verified" as const, updated: "1w ago" },
  ],
};

const migrationStatusConfig = {
  deployed: { label: "Deployed", badge: "default" as const, icon: CheckCircle2 },
  inProgress: { label: "In Progress", badge: "secondary" as const, icon: Clock },
  failed: { label: "Failed", badge: "destructive" as const, icon: XCircle },
  rolledBack: { label: "Rolled Back", badge: "outline" as const, icon: RotateCcw },
};

const migrations = [
  { version: "v1.2.0", name: "Add users table composite index", author: "Alice", status: "deployed" as const, duration: "42s", applied: "2 hours ago" },
  { version: "v1.1.2", name: "Backfill email_verified_at", author: "Bob", status: "deployed" as const, duration: "3m 12s", applied: "1 day ago" },
  { version: "v1.1.1", name: "Drop legacy zip_code column", author: "Charlie", status: "failed" as const, duration: "—", applied: "2 days ago" },
  { version: "v1.1.0", name: "Add email verification columns", author: "Alice", status: "inProgress" as const, duration: "running", applied: "3 days ago" },
  { version: "v1.0.3", name: "Rename stock_level column", author: "Diana", status: "rolledBack" as const, duration: "1m 05s", applied: "5 days ago" },
  { version: "v1.0.2", name: "Add notification_preferences jsonb", author: "Eve", status: "deployed" as const, duration: "28s", applied: "1 week ago" },
];

const driftStatusConfig = {
  unresolved: { label: "Unresolved", badge: "destructive" as const },
  acknowledged: { label: "Acknowledged", badge: "secondary" as const },
  resolved: { label: "Resolved", badge: "default" as const },
};

const severityConfig = {
  critical: { label: "Critical", badge: "destructive" as const },
  warning: { label: "Warning", badge: "secondary" as const },
};

const driftEvents = [
  { id: "d1", table: "sessions", schema: "auth", env: "Staging", kind: "Column added", detail: "sessions.user_agent added in staging but missing in production", severity: "critical" as const, status: "unresolved" as const, detected: "3 hours ago" },
  { id: "d2", table: "events", schema: "analytics", env: "Production", kind: "Index missing", detail: "idx_events_occurred missing in production", severity: "warning" as const, status: "acknowledged" as const, detected: "1 day ago" },
  { id: "d3", table: "users", schema: "public", env: "Development", kind: "Type mismatch", detail: "users.status is varchar(20) in dev vs enum in prod", severity: "warning" as const, status: "resolved" as const, detected: "3 days ago" },
];

const auditCategoryConfig = {
  auth: { label: "Auth", badge: "outline" as const },
  migration: { label: "Migration", badge: "default" as const },
  review: { label: "Review", badge: "secondary" as const },
  drift: { label: "Drift", badge: "destructive" as const },
  team: { label: "Team", badge: "outline" as const },
};

const auditLog = [
  { actor: "Alice", initials: "AL", action: "Migration deployed", category: "migration" as const, resource: "users · v1.2.0", time: "2 hours ago" },
  { actor: "Bob", initials: "BO", action: "Review approved", category: "review" as const, resource: "payments · update status enum", time: "3 hours ago" },
  { actor: "System", initials: "SY", action: "Drift detected", category: "drift" as const, resource: "auth.sessions", time: "3 hours ago" },
  { actor: "Alice", initials: "AL", action: "Signed in", category: "auth" as const, resource: "rk@mainline.dev", time: "4 hours ago" },
  { actor: "Charlie", initials: "CH", action: "Migration failed", category: "migration" as const, resource: "customers · v1.1.1", time: "2 days ago" },
  { actor: "Diana", initials: "DI", action: "Member invited", category: "team" as const, resource: "eve@mainline.dev", time: "3 days ago" },
];

const eventHistory = [
  { title: "Migration v1.2.0 deployed", detail: "User Service · production", icon: CheckCircle2, color: "bg-green-500", time: "2 hours ago" },
  { title: "Review approved update payment status", detail: "Payment DB", icon: GitPullRequest, color: "bg-purple-500", time: "3 hours ago" },
  { title: "Schema drift detected", detail: "auth.sessions differs from staging", icon: AlertTriangle, color: "bg-amber-500", time: "3 hours ago" },
  { title: "Migration v1.0.3 rolled back", detail: "inventory.products", icon: RotateCcw, color: "bg-red-500", time: "5 days ago" },
  { title: "Eve joined the team", detail: "Invited by Diana", icon: Users, color: "bg-blue-500", time: "3 days ago" },
];

export default function ProjectDetailPage() {
  const params = useParams();
  const id = params.id as string;
  const project = mockDetail[id as keyof typeof mockDetail];

  if (!project) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Project not found</h2>
            <p className="text-sm text-muted-foreground">The project you are looking for does not exist.</p>
            <Link href="/projects">
              <Button variant="outline">Back to Projects</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const status = statusConfig[project.status];
  const tables = projectTables[id as keyof typeof projectTables] ?? projectTables.p1;

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbLink href="/projects">Projects</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>{project.name}</BreadcrumbPage></BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="flex items-center gap-2 ml-auto">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search..."
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
            <NotificationsPopover />
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          {/* Project header */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-4">
              <Link href="/projects">
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div>
                <div className="flex items-center gap-3">
                  <h1 className="text-2xl font-semibold tracking-tight">{project.name}</h1>
                  <div className={`size-2.5 rounded-full ${status.dot}`} />
                  <Badge variant="outline" className="text-[11px]">{project.tech}</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">{project.description}</p>
                <div className="flex items-center gap-3 mt-2 text-sm text-muted-foreground">
                  <a href={`https://${project.repo}`} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1 hover:text-foreground transition-colors">
                    <ExternalLink className="size-3" />
                    {project.repo}
                  </a>
                </div>
              </div>
            </div>
            <Button variant="outline" className="h-11 gap-2 shrink-0">
              <Settings className="size-4" />
              Project Settings
            </Button>
          </div>

          {/* Tabs */}
          <Tabs defaultValue="overview" className="w-full">
            <div className="overflow-x-auto -mx-1">
              <TabsList className="h-11 w-max min-w-full">
                {tabs.map((tab) => (
                  <TabsTrigger key={tab.value} value={tab.value} className="h-11">{tab.label}</TabsTrigger>
                ))}
              </TabsList>
            </div>

            <TabsContent value="overview" className="mt-6 space-y-6">
              {/* Info cards */}
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <Database className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Schemas</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">12</p>
                    <p className="text-xs text-muted-foreground mt-1">3 pending review</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <GitBranch className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Migrations</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">47</p>
                    <p className="text-xs text-muted-foreground mt-1">2 in progress</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <GitPullRequest className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Reviews</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">5</p>
                    <p className="text-xs text-muted-foreground mt-1">2 pending approval</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <AlertTriangle className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Drift Events</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">1</p>
                    <p className="text-xs text-amber-500 font-medium mt-1">Requires attention</p>
                  </CardContent>
                </Card>
              </div>

              {/* Team + Info */}
              <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Users className="size-4" />
                      Team
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="space-y-3">
                      {project.members.map((member, i) => (
                        <div key={i} className="flex items-center gap-3">
                          <Avatar className="size-8">
                            <AvatarFallback className="text-xs">{member.initials}</AvatarFallback>
                          </Avatar>
                          <div>
                            <p className="text-sm font-medium">{member.name}</p>
                            <p className="text-xs text-muted-foreground">{member.role}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Clock className="size-4" />
                      Recent Activity
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="space-y-3">
                      {[
                        { action: "Migration v1.2.0 deployed", user: "Alice", time: "2 hours ago", icon: CheckCircle2, color: "text-green-500" },
                        { action: "Schema review approved", user: "Bob", time: "5 hours ago", icon: CheckCircle2, color: "text-green-500" },
                        { action: "Migration v1.1.0 rolled back", user: "System", time: "1 day ago", icon: XCircle, color: "text-red-500" },
                      ].map((item, i) => (
                        <div key={i} className="flex items-start gap-3 pb-3 border-b last:border-b-0">
                          <item.icon className={`size-4 mt-0.5 shrink-0 ${item.color}`} />
                          <div className="flex-1 min-w-0">
                            <p className="text-sm">{item.action}</p>
                            <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                              <span>{item.user}</span>
                              <span>{item.time}</span>
                            </div>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            <TabsContent value="schemas" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <div className="flex items-center justify-between gap-4">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Database className="size-4" />
                      Tables
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{tables.length}</Badge>
                    </CardTitle>
                    <Link href="/schemas">
                      <Button variant="outline" size="sm" className="h-9 gap-2">
                        <Database className="size-4" />
                        Open Explorer
                      </Button>
                    </Link>
                    <Link href={`/projects/${id}/schemas/public/erd`}>
                      <Button size="sm" className="h-9 gap-2">
                        <GitBranch className="size-4" />
                        View ERD
                      </Button>
                    </Link>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[35%]">Table</TableHead>
                        <TableHead>Schema</TableHead>
                        <TableHead>Columns</TableHead>
                        <TableHead>Indexes</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead className="text-right">Updated</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {tables.map((table) => {
                        const status = tableStatusConfig[table.status];
                        return (
                          <TableRow key={table.name}>
                            <TableCell>
                              <div className="flex items-center gap-2.5">
                                <div className={`size-1.5 rounded-full ${status.dot}`} />
                                <span className="font-mono text-sm">{table.name}</span>
                              </div>
                            </TableCell>
                            <TableCell><span className="font-mono text-xs text-muted-foreground">{table.schema}</span></TableCell>
                            <TableCell className="text-sm text-muted-foreground">{table.columns}</TableCell>
                            <TableCell className="text-sm text-muted-foreground">{table.indexes}</TableCell>
                            <TableCell>
                              <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                            </TableCell>
                            <TableCell className="text-right text-sm text-muted-foreground">{table.updated}</TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="migrations" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <div className="flex items-center justify-between gap-4">
                    <CardTitle className="text-base flex items-center gap-2">
                      <GitBranch className="size-4" />
                      Migration History
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{migrations.length}</Badge>
                    </CardTitle>
                    <Button size="sm" className="h-9 gap-2">
                      <GitBranch className="size-4" />
                      Run Migration
                    </Button>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[15%]">Version</TableHead>
                        <TableHead className="w-[35%]">Migration</TableHead>
                        <TableHead>Author</TableHead>
                        <TableHead>Status</TableHead>
                        <TableHead>Duration</TableHead>
                        <TableHead className="text-right">Applied</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {migrations.map((migration) => {
                        const status = migrationStatusConfig[migration.status];
                        return (
                          <TableRow key={migration.version}>
                            <TableCell><span className="font-mono text-sm">{migration.version}</span></TableCell>
                            <TableCell>
                              <div className="flex items-center gap-2.5">
                                <status.icon className="size-4 text-muted-foreground" />
                                <span className="text-sm truncate">{migration.name}</span>
                              </div>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground">{migration.author}</TableCell>
                            <TableCell>
                              <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground">{migration.duration}</TableCell>
                            <TableCell className="text-right text-sm text-muted-foreground">{migration.applied}</TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="drift" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <AlertTriangle className="size-4" />
                    Drift Events
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">{driftEvents.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  {driftEvents.map((drift) => {
                    const status = driftStatusConfig[drift.status];
                    const severity = severityConfig[drift.severity];
                    return (
                      <div key={drift.id} className="flex items-start gap-3 py-4 border-b last:border-b-0">
                        <div className="mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full bg-muted">
                          <AlertTriangle className="size-4 text-amber-500" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <p className="text-sm font-medium">{drift.kind}</p>
                            <Badge variant={severity.badge} className="text-[10px] px-1.5 py-0">{severity.label}</Badge>
                            <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                          </div>
                          <p className="text-sm text-muted-foreground mt-1">{drift.detail}</p>
                          <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1">
                            <span className="font-mono">{drift.schema}.{drift.table}</span>
                            <span>·</span>
                            <span>{drift.env}</span>
                            <span>·</span>
                            <span>{drift.detected}</span>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="audit" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <FileText className="size-4" />
                    Audit Log
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">{auditLog.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead className="w-[25%]">Actor</TableHead>
                        <TableHead className="w-[25%]">Action</TableHead>
                        <TableHead>Category</TableHead>
                        <TableHead className="w-[30%]">Resource</TableHead>
                        <TableHead className="text-right">Time</TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {auditLog.map((entry, i) => {
                        const category = auditCategoryConfig[entry.category];
                        return (
                          <TableRow key={i}>
                            <TableCell>
                              <div className="flex items-center gap-2.5">
                                <Avatar className="size-7">
                                  <AvatarFallback className="text-[9px]">{entry.initials}</AvatarFallback>
                                </Avatar>
                                <span className="text-sm">{entry.actor}</span>
                              </div>
                            </TableCell>
                            <TableCell className="text-sm">{entry.action}</TableCell>
                            <TableCell>
                              <Badge variant={category.badge} className="text-[10px] px-1.5 py-0">{category.label}</Badge>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground truncate">{entry.resource}</TableCell>
                            <TableCell className="text-right text-sm text-muted-foreground">{entry.time}</TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="events" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <Calendar className="size-4" />
                    Event History
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">{eventHistory.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  {eventHistory.map((event, i) => (
                    <div key={i} className="flex gap-3 py-3.5 border-b last:border-b-0">
                      <div className={`mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full ${event.color} text-white`}>
                        <event.icon className="size-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{event.title}</p>
                        <div className="flex items-center gap-2 mt-0.5">
                          <span className="text-xs text-muted-foreground">{event.detail}</span>
                          <span className="text-xs text-muted-foreground ml-auto">{event.time}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="settings" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <Settings className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Project settings coming soon</p>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
