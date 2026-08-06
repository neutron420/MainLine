"use client";

import {
  MoreHorizontal,
  Pin,
  Settings,
  Share2,
  Trash,
  TriangleAlert,
  ListFilter,
  Columns,
  Plus,
  GitBranch,
  Database,
  UserPlus,
  Clock,
  AlertCircle,
  Search,
  Activity,
} from "lucide-react";
import { useState, useMemo } from "react";
import Link from "next/link";


import {  Badge  } from "@/components/ui/badge";







import {  Button  } from "@/components/ui/button";
import {  Card, CardContent, CardHeader, CardTitle  } from "@/components/ui/card";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu";






import {  Input  } from "@/components/ui/input";
import {  ProjectDataTable, Project  } from "@/components/ui/project-data-table";

import {  useProjects  } from "@/lib/api/hooks/use-projects";
import {  useAuditEntries  } from "@/lib/api/hooks/use-audit";
import {  useEventStream  } from "@/lib/api/hooks/use-realtime";
import {  getApiErrorMessage  } from "@/lib/api/errors";

function formatNumber(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return n.toLocaleString();
  return n.toString();
}

function relativeTime(iso: string | undefined): string {
  if (!iso) return "recently";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "recently";
  const diffMs = Date.now() - then;
  const mins = Math.floor(diffMs / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} hour${hours > 1 ? "s" : ""} ago`;
  const days = Math.floor(hours / 24);
  return `${days} day${days > 1 ? "s" : ""} ago`;
}

function toTableProject(
  project: { id: string; name: string; createdAt: string; visibility: string },
): Project {
  return {
    id: project.id,
    name: project.name,
    repository: "",
    team: project.visibility,
    tech: "PostgreSQL",
    createdAt: project.createdAt.slice(0, 10),
    contributors: [],
    status: { text: "Active", variant: "active" },
  };
}

const allColumns: (keyof Project)[] = ["name", "repository", "team", "tech", "createdAt", "contributors", "status"];

export default function Page() {
  const { data: projects, isLoading, error } = useProjects();
  const { data: auditEntries } = useAuditEntries();
  const { events, connected } = useEventStream({ maxEvents: 8 });
  const [projectSearch, setProjectSearch] = useState("");

  const activityCount = auditEntries?.length ?? 0;
  const liveEventCount = events.length;

  const firstProject = projects?.[0];

  const now = Date.now();
  const monthAgo = now - 30 * 24 * 60 * 60 * 1000;
  const dayAgo = now - 24 * 60 * 60 * 1000;
  const projectsThisMonth = projects?.filter((p) => new Date(p.createdAt).getTime() >= monthAgo).length ?? 0;
  const auditToday = auditEntries?.filter((e) => new Date(e.createdAt).getTime() >= dayAgo).length ?? 0;

  const stats = [
    { title: "Total Projects", value: projects?.length ?? 0, delta: projectsThisMonth, lastMonth: 0, positive: true, prefix: "", suffix: "", detail: `${projectsThisMonth} new this month` },
    { title: "Live Events", value: liveEventCount, delta: 0, lastMonth: 0, positive: connected, prefix: "", suffix: "", detail: connected ? "Real-time stream connected" : "Connecting to event stream..." },
    { title: "Audit Entries", value: activityCount, delta: auditToday, lastMonth: 0, positive: true, prefix: "", suffix: "", detail: `${auditToday} in the last 24h` },
    { title: "Team Members", value: projects?.reduce((sum, p) => sum + p.memberCount, 0) ?? 0, delta: 0, lastMonth: 0, positive: true, prefix: "", suffix: "", detail: "Across all projects" },
  ];

  const quickActions = [
    { label: "New Project", icon: Plus, variant: "default" as const, href: "/projects/new" },
    { label: "New Schema", icon: Database, variant: "outline" as const, href: firstProject ? `/projects/${firstProject.id}/schemas` : "/projects" },
    { label: "Run Migration", icon: GitBranch, variant: "outline" as const, href: firstProject ? `/projects/${firstProject.id}/migrations` : "/projects" },
    { label: "Invite Team", icon: UserPlus, variant: "outline" as const, href: "/settings" },
  ];

  return (
            <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search projects..."
                value={projectSearch}
                onChange={(e) => setProjectSearch(e.target.value)}
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
          {/* Stats */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat, index) => (
              <Card key={index}>
                <CardHeader className="flex flex-row items-center justify-between border-0">
                  <CardTitle className="text-muted-foreground text-sm font-medium">{stat.title}</CardTitle>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon"><MoreHorizontal /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" side="bottom">
                      <DropdownMenuItem><Settings /> Settings</DropdownMenuItem>
                      <DropdownMenuItem><TriangleAlert /> Add Alert</DropdownMenuItem>
                      <DropdownMenuItem><Pin /> Pin to Dashboard</DropdownMenuItem>
                      <DropdownMenuItem><Share2 /> Share</DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem><Trash /> Remove</DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </CardHeader>
                <CardContent className="space-y-2.5">
                  <div className="flex items-center gap-2.5">
                    <span className="text-2xl font-medium text-foreground tracking-tight">
                      {stat.prefix + formatNumber(stat.value) + stat.suffix}
                    </span>
                    {stat.title === "Live Events" && (
                      <Badge variant={connected ? "default" : "secondary"} className="text-[10px]">
                        {connected ? "LIVE" : "OFFLINE"}
                      </Badge>
                    )}
                  </div>
                  <div className="text-xs text-muted-foreground mt-2 border-t pt-2.5">
                    {stat.detail}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Quick Actions */}
          <div className="flex flex-wrap gap-3">
            {quickActions.map((action) => (
              <Button
                key={action.label}
                variant={action.variant}
                className="h-11 gap-2"
                asChild
              >
                <Link href={action.href}>
                  <action.icon className="size-4" />
                  {action.label}
                </Link>
              </Button>
            ))}
          </div>

          {/* Live events + Audit row */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Live Events */}
            <Card>
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <Activity className="size-4" />
                  Live Events
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                {events.length === 0 ? (
                  <p className="text-muted-foreground text-sm py-6 text-center">
                    {connected
                      ? "Waiting for schema events... (migrations, drift, connections)"
                      : "Event stream disconnected — reconnecting..."}
                  </p>
                ) : (
                  <div className="space-y-0">
                    {events.map((event) => (
                      <div key={event.id} className="flex gap-3 py-3 border-b last:border-b-0">
                        <div className="bg-primary/10 mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full text-primary">
                          <GitBranch className="size-4" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate capitalize">
                            {event.type.replace(/_/g, " ")}
                          </p>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span className="text-xs text-muted-foreground">
                              {event.actor?.email ?? "system"}
                            </span>
                            <span className="text-xs text-muted-foreground">·</span>
                            <span className="text-xs text-muted-foreground">
                              {event.resource?.type ?? "schema"} {event.resource?.id ?? ""}
                            </span>
                            <span className="text-xs text-muted-foreground ml-auto">
                              {relativeTime(event.timestamp)}
                            </span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Recent Audit Activity */}
            <Card>
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <Clock className="size-4" />
                  Recent Activity
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                {!auditEntries || auditEntries.length === 0 ? (
                  <p className="text-muted-foreground text-sm py-6 text-center">
                    No audit activity yet. Actions like migrations and schema changes will appear here.
                  </p>
                ) : (
                  <div className="space-y-0">
                    {auditEntries.slice(0, 6).map((entry) => (
                      <div key={entry.id} className="flex gap-3 py-3 border-b last:border-b-0">
                        <div className="bg-primary/10 mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full text-primary">
                          <AlertCircle className="size-4" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium truncate">
                            {entry.action} {entry.resourceType} {entry.resourceId}
                          </p>
                          <div className="flex items-center gap-2 mt-0.5">
                            <span className="text-xs text-muted-foreground">
                              {entry.actorEmail || entry.actorId || "system"}
                            </span>
                            <span className="text-xs text-muted-foreground">·</span>
                            <span className="text-xs text-muted-foreground ml-auto">
                              {relativeTime(entry.createdAt)}
                            </span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </div>

          {/* Projects Table */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold tracking-tight">All Projects</h2>
              {error && (
                <span className="text-destructive text-sm">{getApiErrorMessage(error)}</span>
              )}
            </div>
            <FilterableProjectTable
              projects={(projects ?? []).map(toTableProject)}
              isLoading={isLoading}
              search={projectSearch}
            />
          </div>
        </div>
  );
}

function FilterableProjectTable({
  projects,
  isLoading,
  search,
}: {
  projects: Project[];
  isLoading: boolean;
  search: string;
}) {
  const [techFilter, setTechFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [visibleColumns, setVisibleColumns] = useState<Set<keyof Project>>(new Set(allColumns));

  const filteredProjects = useMemo(() => {
    return projects.filter((project) => {
      const techMatch = techFilter === "" || project.tech.toLowerCase().includes(techFilter.toLowerCase());
      const statusMatch = statusFilter === "all" || project.status.variant === statusFilter;
      const searchMatch =
        search === "" ||
        project.name.toLowerCase().includes(search.toLowerCase()) ||
        project.team.toLowerCase().includes(search.toLowerCase());
      return techMatch && statusMatch && searchMatch;
    });
  }, [projects, techFilter, statusFilter, search]);

  const toggleColumn = (column: keyof Project) => {
    setVisibleColumns((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(column)) newSet.delete(column);
      else newSet.add(column);
      return newSet;
    });
  };

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="flex flex-1 gap-4">
          <Input placeholder="Filter by technology..." value={techFilter} onChange={(e) => setTechFilter(e.target.value)} className="max-w-xs h-11" />
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="outline" className="flex items-center gap-2 h-11">
                <ListFilter className="h-4 w-4" />
                <span>Status</span>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              <DropdownMenuLabel>Filter by Status</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuCheckboxItem checked={statusFilter === "all"} onCheckedChange={() => setStatusFilter("all")}>All</DropdownMenuCheckboxItem>
              <DropdownMenuCheckboxItem checked={statusFilter === "active"} onCheckedChange={() => setStatusFilter("active")}>Active</DropdownMenuCheckboxItem>
              <DropdownMenuCheckboxItem checked={statusFilter === "inProgress"} onCheckedChange={() => setStatusFilter("inProgress")}>In Progress</DropdownMenuCheckboxItem>
              <DropdownMenuCheckboxItem checked={statusFilter === "onHold"} onCheckedChange={() => setStatusFilter("onHold")}>On Hold</DropdownMenuCheckboxItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="flex items-center gap-2 h-11">
              <Columns className="h-4 w-4" />
              <span>Columns</span>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent>
            <DropdownMenuLabel>Toggle Columns</DropdownMenuLabel>
            <DropdownMenuSeparator />
            {allColumns.map((column) => (
              <DropdownMenuCheckboxItem key={column} className="capitalize" checked={visibleColumns.has(column)} onCheckedChange={() => toggleColumn(column)}>
                {column}
              </DropdownMenuCheckboxItem>
            ))}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      {isLoading ? (
        <div className="border-border flex h-40 items-center justify-center rounded-lg border">
          <p className="text-muted-foreground text-sm">Loading projects...</p>
        </div>
      ) : filteredProjects.length === 0 ? (
        <div className="border-border flex h-40 flex-col items-center justify-center gap-3 rounded-lg border">
          <p className="text-muted-foreground text-sm">
            No projects yet.
          </p>
          <Button asChild variant="outline" className="gap-2">
            <Link href="/projects/new"><Plus className="size-4" /> Create your first project</Link>
          </Button>
        </div>
      ) : (
        <ProjectDataTable projects={filteredProjects} visibleColumns={visibleColumns} />
      )}
    </div>
  );
}
