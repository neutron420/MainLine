"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import {
  ArrowLeft,
  Database,
  GitBranch,
  AlertTriangle,
  FileText,
  Calendar,
  Users,
  Settings,
  ExternalLink,
  Clock,
  CheckCircle2,
  XCircle,
  Search,
  RotateCcw,
  GitCompareArrows,
  Activity,
} from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";









import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";




import {  Tabs, TabsContent, TabsList, TabsTrigger  } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {  Avatar, AvatarFallback  } from "@/components/ui/avatar";

import {  useProject, useMembers  } from "@/lib/api/hooks/use-projects";
import {  useSchemas  } from "@/lib/api/hooks/use-schemas";
import {  useMigrations  } from "@/lib/api/hooks/use-migrations";
import {  useAuditEntries  } from "@/lib/api/hooks/use-audit";
import {  useEventStream  } from "@/lib/api/hooks/use-realtime";
import {  getApiErrorMessage  } from "@/lib/api/errors";

const tabs = [
  { value: "overview", label: "Overview" },
  { value: "schemas", label: "Schemas" },
  { value: "migrations", label: "Migrations" },
  { value: "audit", label: "Audit" },
  { value: "events", label: "Events" },
  { value: "settings", label: "Settings" },
];

function relativeTime(iso: string | undefined): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days}d ago`;
  return new Date(iso).toLocaleDateString();
}

function initialsOf(name: string): string {
  return name
    .split(/[\s@._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("") || "?";
}

const migrationStatusConfig: Record<
  string,
  { label: string; badge: "default" | "secondary" | "destructive" | "outline"; icon: typeof CheckCircle2 }
> = {
  draft: { label: "Draft", badge: "outline", icon: FileText },
  pending: { label: "Pending", badge: "secondary", icon: Clock },
  running: { label: "Running", badge: "secondary", icon: Clock },
  completed: { label: "Completed", badge: "default", icon: CheckCircle2 },
  failed: { label: "Failed", badge: "destructive", icon: XCircle },
  rolled_back: { label: "Rolled Back", badge: "outline", icon: RotateCcw },
};

const auditBadgeConfig: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  migration: "default",
  drift: "destructive",
  review: "secondary",
  auth: "outline",
  team: "outline",
};

export default function ProjectDetailPage() {
  const params = useParams();
  const id = params.id as string;

  const { data: project, isLoading, error } = useProject(id);
  const { data: schemas } = useSchemas(id);
  const { data: migrations } = useMigrations(id);
  const { data: members } = useMembers(id);
  const { data: auditEntries } = useAuditEntries();
  const { events, connected } = useEventStream({ projectIds: [id], maxEvents: 8 });
  const [search, setSearch] = useState("");

  const memberList = Array.isArray(members) ? members : [];

  const filteredSchemas = (schemas ?? []).filter((s) => {
    if (search === "") return true;
    const haystack = `${s.schemaName} ${s.connectionId}`.toLowerCase();
    return haystack.includes(search.toLowerCase());
  });

  if (isLoading) {
    return (
                <div className="flex h-full flex-col items-center justify-center gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">Loading project...</p>
          </div>
    );
  }

  if (!project) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Project not found</h2>
            <p className="text-sm text-muted-foreground">
              {error ? getApiErrorMessage(error) : "The project you are looking for does not exist."}
            </p>
            <Link href="/projects">
              <Button variant="outline">Back to Projects</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search schemas..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
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
                  <div className="bg-green-500 size-2.5 rounded-full" />
                  <Badge variant="outline" className="text-[11px]">PostgreSQL</Badge>
                  <Badge variant="secondary" className="text-[11px] capitalize">{project.visibility}</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {project.description || "No description provided."}
                </p>
                <div className="flex items-center gap-3 mt-2 text-sm text-muted-foreground">
                  <span className="flex items-center gap-1">
                    <ExternalLink className="size-3" />
                    {project.slug}
                  </span>
                  <span className="flex items-center gap-1">
                    <Clock className="size-3" />
                    Created {relativeTime(project.createdAt)}
                  </span>
                </div>
              </div>
            </div>
            <Link href={`/projects/${id}/settings`}>
              <Button variant="outline" className="h-11 gap-2 shrink-0">
                <Settings className="size-4" />
                Project Settings
              </Button>
            </Link>
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
                    <p className="text-2xl font-semibold">{schemas?.length ?? 0}</p>
                    <p className="text-xs text-muted-foreground mt-1">tracked in this project</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <GitBranch className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Migrations</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">{migrations?.length ?? 0}</p>
                    <p className="text-xs text-muted-foreground mt-1">
                      {migrations?.filter((m) => m.status === "completed").length ?? 0} completed
                    </p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <Users className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Members</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">{memberList.length || project.memberCount}</p>
                    <p className="text-xs text-muted-foreground mt-1">collaborators</p>
                  </CardContent>
                </Card>
                <Card>
                  <CardHeader className="flex flex-row items-center gap-3 pb-2 border-0">
                    <Activity className="size-4 text-muted-foreground" />
                    <CardTitle className="text-sm font-medium">Live Events</CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-2xl font-semibold">{events.length}</p>
                    <p className="text-xs font-medium mt-1">
                      {connected ? <span className="text-green-500">Stream connected</span> : <span className="text-amber-500">Connecting...</span>}
                    </p>
                  </CardContent>
                </Card>
              </div>

              {/* Team + Activity */}
              <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Users className="size-4" />
                      Team
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    {memberList.length === 0 ? (
                      <p className="text-muted-foreground text-sm py-4">No members yet. Invite your team to collaborate.</p>
                    ) : (
                      <div className="space-y-3">
                        {memberList.map((member) => (
                          <div key={member.userId || member.email} className="flex items-center gap-3">
                            <Avatar className="size-8">
                              <AvatarFallback className="text-xs">{initialsOf(member.email || member.userId)}</AvatarFallback>
                            </Avatar>
                            <div>
                              <p className="text-sm font-medium">{member.email || member.userId}</p>
                              <p className="text-xs text-muted-foreground capitalize">{member.role}</p>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <Activity className="size-4" />
                      Live Events
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    {events.length === 0 ? (
                      <p className="text-muted-foreground text-sm py-4">
                        {connected ? "Waiting for schema events..." : "Event stream disconnected — reconnecting..."}
                      </p>
                    ) : (
                      <div className="space-y-3">
                        {events.map((event) => (
                          <div key={event.id} className="flex items-start gap-3 pb-3 border-b last:border-b-0">
                            <CheckCircle2 className="bg-primary/10 text-primary size-7 shrink-0 rounded-full p-1.5" />
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium capitalize">{event.type.replace(/_/g, " ")}</p>
                              <div className="flex items-center gap-2 text-xs text-muted-foreground mt-0.5">
                                <span>{event.actor?.email ?? "system"}</span>
                                <span>·</span>
                                <span>{relativeTime(event.timestamp)}</span>
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
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
                      Schemas
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{schemas?.length ?? 0}</Badge>
                    </CardTitle>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  {!schemas || schemas.length === 0 ? (
                    <p className="text-muted-foreground text-sm py-6 text-center">
                      No schemas tracked yet. Connect a database to start introspecting.
                    </p>
                  ) : filteredSchemas.length === 0 ? (
                    <p className="text-muted-foreground text-sm py-6 text-center">
                      No schemas match your search.
                    </p>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[30%]">Schema</TableHead>
                          <TableHead>Connection</TableHead>
                          <TableHead>Version</TableHead>
                          <TableHead className="text-right">Last Updated</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {filteredSchemas.map((schema) => (
                          <TableRow key={schema.id}>
                            <TableCell>
                              <div className="flex items-center gap-2.5">
                                <div className="bg-green-500 size-1.5 rounded-full" />
                                <Link
                                  href={`/projects/${id}/schemas/${schema.id}`}
                                  className="font-mono text-sm hover:underline"
                                >
                                  {schema.schemaName}
                                </Link>
                              </div>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground">{schema.connectionId}</TableCell>
                            <TableCell>
                              <Link
                                href={`/projects/${id}/schemas/${schema.id}/compare`}
                                className="flex items-center gap-1.5 text-sm hover:underline"
                              >
                                <GitCompareArrows className="size-3.5" />
                                <span className="font-mono text-xs">{schema.currentVersionId?.slice(0, 8)}</span>
                              </Link>
                            </TableCell>
                            <TableCell className="text-right text-sm text-muted-foreground">
                              {relativeTime(schema.updatedAt)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
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
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{migrations?.length ?? 0}</Badge>
                    </CardTitle>
                    <Link href={`/projects/${id}/migrations/new`}>
                      <Button size="sm" className="h-9 gap-2">
                        <GitBranch className="size-4" />
                        New Migration
                      </Button>
                    </Link>
                  </div>
                </CardHeader>
                <CardContent className="pt-0">
                  {!migrations || migrations.length === 0 ? (
                    <p className="text-muted-foreground text-sm py-6 text-center">
                      No migrations yet. Create your first migration to start versioning schema changes.
                    </p>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[12%]">Version</TableHead>
                          <TableHead className="w-[35%]">Migration</TableHead>
                          <TableHead>Status</TableHead>
                          <TableHead className="text-right">Updated</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {migrations.map((migration) => {
                          const status = migrationStatusConfig[migration.status] ?? migrationStatusConfig.draft;
                          return (
                            <TableRow key={migration.id}>
                              <TableCell><span className="font-mono text-sm">{migration.version}</span></TableCell>
                              <TableCell>
                                <Link
                                  href={`/projects/${id}/migrations/${migration.id}`}
                                  className="flex items-center gap-2.5 hover:underline"
                                >
                                  <status.icon className="size-4 text-muted-foreground" />
                                  <span className="text-sm truncate">{migration.title}</span>
                                </Link>
                              </TableCell>
                              <TableCell>
                                <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                              </TableCell>
                              <TableCell className="text-right text-sm text-muted-foreground">
                                {relativeTime(migration.updatedAt || migration.createdAt)}
                              </TableCell>
                            </TableRow>
                          );
                        })}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="audit" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <FileText className="size-4" />
                    Audit Log
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">{auditEntries?.length ?? 0}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  {!auditEntries || auditEntries.length === 0 ? (
                    <p className="text-muted-foreground text-sm py-6 text-center">
                      No audit activity yet.
                    </p>
                  ) : (
                    <Table>
                      <TableHeader>
                        <TableRow>
                          <TableHead className="w-[25%]">Actor</TableHead>
                          <TableHead className="w-[25%]">Action</TableHead>
                          <TableHead>Type</TableHead>
                          <TableHead className="w-[30%]">Resource</TableHead>
                          <TableHead className="text-right">Time</TableHead>
                        </TableRow>
                      </TableHeader>
                      <TableBody>
                        {auditEntries.slice(0, 10).map((entry) => (
                          <TableRow key={entry.id}>
                            <TableCell>
                              <div className="flex items-center gap-2.5">
                                <Avatar className="size-7">
                                  <AvatarFallback className="text-[9px]">{initialsOf(entry.actorEmail || entry.actorId)}</AvatarFallback>
                                </Avatar>
                                <span className="text-sm">{entry.actorEmail || entry.actorId || "system"}</span>
                              </div>
                            </TableCell>
                            <TableCell className="text-sm">{entry.action}</TableCell>
                            <TableCell>
                              <Badge variant={auditBadgeConfig[entry.eventType] ?? "outline"} className="text-[10px] px-1.5 py-0">
                                {entry.eventType}
                              </Badge>
                            </TableCell>
                            <TableCell className="text-sm text-muted-foreground truncate">
                              {entry.resourceType} {entry.resourceId}
                            </TableCell>
                            <TableCell className="text-right text-sm text-muted-foreground">
                              {relativeTime(entry.createdAt)}
                            </TableCell>
                          </TableRow>
                        ))}
                      </TableBody>
                    </Table>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="events" className="mt-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base flex items-center gap-2">
                    <Calendar className="size-4" />
                    Live Events
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">{events.length}</Badge>
                  </CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  {events.length === 0 ? (
                    <p className="text-muted-foreground text-sm py-6 text-center">
                      {connected ? "Waiting for schema events..." : "Event stream disconnected — reconnecting..."}
                    </p>
                  ) : (
                    events.map((event) => (
                      <div key={event.id} className="flex gap-3 py-3.5 border-b last:border-b-0">
                        <div className="bg-primary/10 mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full text-primary">
                          <AlertTriangle className="size-4" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate capitalize">{event.type.replace(/_/g, " ")}</p>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span className="text-xs text-muted-foreground">{event.actor?.email ?? "system"}</span>
                            <span className="text-xs text-muted-foreground">·</span>
                            <span className="text-xs text-muted-foreground">
                              {event.resource?.type ?? ""} {event.resource?.id ?? ""}
                            </span>
                            <span className="text-xs text-muted-foreground ml-auto">{relativeTime(event.timestamp)}</span>
                          </div>
                        </div>
                      </div>
                    ))
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="settings" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <Settings className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground mb-4">
                    Manage project configuration, migration policy and members
                  </p>
                  <div className="flex items-center justify-center gap-3">
                    <Link href={`/projects/${id}/settings`}>
                      <Button variant="outline" className="gap-2">
                        <Settings className="size-4" />
                        Project Settings
                      </Button>
                    </Link>
                    <Link href={`/projects/${id}/settings/members`}>
                      <Button variant="outline" className="gap-2">
                        <Users className="size-4" />
                        Members
                      </Button>
                    </Link>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
