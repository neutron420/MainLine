"use client";

import { useState, useMemo, useEffect } from "react";
import { Plus, Search, FolderKanban, ExternalLink, MoreHorizontal, ChevronLeft, ChevronRight } from "lucide-react";
import Link from "next/link";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
} from "@/components/ui/dropdown-menu";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";

interface ProjectCard {
  id: string;
  name: string;
  repository: string;
  description: string;
  team: string;
  tech: string;
  status: "active" | "inProgress" | "onHold";
  updated: string;
  members: { initials: string }[];
}

const projects: ProjectCard[] = [
  { id: "p1", name: "User Service Schema", repository: "github.com/mainline/user-service", description: "User authentication and profile data models with Row-Level Security policies", team: "Platform", tech: "PostgreSQL", status: "active", updated: "2 hours ago", members: [{ initials: "AL" }, { initials: "BO" }] },
  { id: "p2", name: "Payment DB Migration", repository: "github.com/mainline/payment-db", description: "Payment transaction tables, ledger entries, and reconciliation views", team: "Payments", tech: "PostgreSQL", status: "inProgress", updated: "5 hours ago", members: [{ initials: "CH" }] },
  { id: "p3", name: "Analytics Warehouse", repository: "github.com/mainline/analytics-db", description: "Materialized views for sales analytics, cohort analysis, and reporting", team: "Data", tech: "PostgreSQL", status: "active", updated: "1 day ago", members: [{ initials: "DI" }, { initials: "EV" }] },
  { id: "p4", name: "Legacy CRM Schema", repository: "github.com/mainline/crm-migration", description: "Customer relationship tables being migrated from MySQL with data validation", team: "Core", tech: "PostgreSQL", status: "onHold", updated: "3 days ago", members: [{ initials: "FR" }, { initials: "GR" }] },
  { id: "p5", name: "Notification Queue", repository: "github.com/mainline/notif-db", description: "Push notification scheduling, delivery status tracking, and retry logic tables", team: "Infra", tech: "PostgreSQL", status: "active", updated: "5 days ago", members: [{ initials: "HA" }] },
  { id: "p6", name: "Search Index Schema", repository: "github.com/mainline/search-db", description: "Full-text search index tables, trigram indexes, and ranking configurations", team: "Search", tech: "PostgreSQL", status: "inProgress", updated: "1 week ago", members: [{ initials: "IV" }, { initials: "JA" }] },
  { id: "p7", name: "Inventory Service", repository: "github.com/mainline/inventory-db", description: "Product catalog, stock levels, warehouse locations, and inventory snapshots", team: "Platform", tech: "PostgreSQL", status: "active", updated: "1 week ago", members: [{ initials: "KA" }, { initials: "LE" }] },
  { id: "p8", name: "Billing System Schema", repository: "github.com/mainline/billing-db", description: "Invoice generation, subscription plans, usage metering, and payment reconciliation", team: "Payments", tech: "PostgreSQL", status: "onHold", updated: "2 weeks ago", members: [{ initials: "MI" }] },
  { id: "p9", name: "Auth Provider Sync", repository: "github.com/mainline/auth-sync", description: "OAuth provider token storage, session management, and refresh token rotation", team: "Platform", tech: "PostgreSQL", status: "active", updated: "2 weeks ago", members: [{ initials: "NO" }, { initials: "OL" }] },
  { id: "p10", name: "Content CMS Schema", repository: "github.com/mainline/cms-db", description: "Page builder content types, media library, and versioned publishing workflow tables", team: "Core", tech: "PostgreSQL", status: "inProgress", updated: "3 weeks ago", members: [{ initials: "PE" }] },
  { id: "p11", name: "ML Feature Store", repository: "github.com/mainline/feature-store", description: "Feature engineering pipelines, model inference logs, and experiment tracking tables", team: "Data", tech: "PostgreSQL", status: "active", updated: "3 weeks ago", members: [{ initials: "QU" }, { initials: "RI" }] },
  { id: "p12", name: "Support Ticket DB", repository: "github.com/mainline/support-db", description: "Ticket lifecycle, SLA tracking, customer history, and agent assignment queues", team: "Core", tech: "PostgreSQL", status: "onHold", updated: "1 month ago", members: [{ initials: "SA" }] },
];

const statusConfig = {
  active: { label: "Active", dot: "bg-green-500", badge: "default" as const },
  inProgress: { label: "In Progress", dot: "bg-yellow-500", badge: "secondary" as const },
  onHold: { label: "On Hold", dot: "bg-red-500", badge: "destructive" as const },
};

const PER_PAGE = 8;

export default function ProjectsPage() {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [teamFilter, setTeamFilter] = useState<string>("all");
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);

  const teams = useMemo(() => [...new Set(projects.map((p) => p.team))], []);

  const filtered = useMemo(() => {
    return projects.filter((p) => {
      const matchSearch = p.name.toLowerCase().includes(search.toLowerCase()) || p.description.toLowerCase().includes(search.toLowerCase());
      const matchStatus = statusFilter === "all" || p.status === statusFilter;
      const matchTeam = teamFilter === "all" || p.team === teamFilter;
      return matchSearch && matchStatus && matchTeam;
    });
  }, [search, statusFilter, teamFilter]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PER_PAGE));
  const safePage = Math.min(page, totalPages);
  const paginated = useMemo(() => {
    const start = (safePage - 1) * PER_PAGE;
    return filtered.slice(start, start + PER_PAGE);
  }, [filtered, safePage]);

  useEffect(() => {
    if (page > totalPages) setPage(1);
  }, [page, totalPages]);

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <Tooltip delay={0}>
            <SidebarTrigger className="-ml-1 size-9" />
            <Tooltip.Content>
              <p>Toggle sidebar</p>
            </Tooltip.Content>
          </Tooltip>
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbPage>Projects</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header + New button */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Projects</h1>
              <p className="text-sm text-muted-foreground mt-1">Manage your database schema projects</p>
            </div>
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Tooltip delay={0}>
                  <Button className="h-11 gap-2">
                    <Plus className="size-4" />
                    New Project
                  </Button>
                  <Tooltip.Content>
                    <p>Create a new schema project</p>
                  </Tooltip.Content>
                </Tooltip>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[500px]">
                <DialogHeader>
                  <DialogTitle>Create Project</DialogTitle>
                  <DialogDescription>
                    Set up a new database schema project for your team.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="name">Project Name</Label>
                    <Input id="name" placeholder="e.g. User Service Schema" className="h-11" />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="desc">Description</Label>
                    <Textarea id="desc" placeholder="Brief description of the project..." className="min-h-[80px]" />
                  </div>
                  <div className="grid grid-cols-2 gap-4">
                    <div className="grid gap-2">
                      <Label htmlFor="team">Team</Label>
                      <Select>
                        <SelectTrigger id="team" className="h-11">
                          <SelectValue placeholder="Select team" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="platform">Platform</SelectItem>
                          <SelectItem value="payments">Payments</SelectItem>
                          <SelectItem value="data">Data</SelectItem>
                          <SelectItem value="core">Core</SelectItem>
                          <SelectItem value="infra">Infra</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="repo">Repository URL</Label>
                      <Input id="repo" placeholder="github.com/org/repo" className="h-11" />
                    </div>
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                  <Button onClick={() => setCreateOpen(false)}>Create Project</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Search + Filters */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input
                placeholder="Search projects..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 h-11"
              />
            </div>
            <div className="flex gap-3">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="h-11">
                    <FolderKanban className="size-4 mr-2" />
                    {statusFilter === "all" ? "Status" : statusConfig[statusFilter as keyof typeof statusConfig]?.label}
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
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" className="h-11">
                    {teamFilter === "all" ? "Team" : teamFilter}
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuLabel>Filter by Team</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuCheckboxItem checked={teamFilter === "all"} onCheckedChange={() => setTeamFilter("all")}>All Teams</DropdownMenuCheckboxItem>
                  {teams.map((team) => (
                    <DropdownMenuCheckboxItem key={team} checked={teamFilter === team} onCheckedChange={() => setTeamFilter(team)}>{team}</DropdownMenuCheckboxItem>
                  ))}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>

          {/* Results count */}
          <p className="text-sm text-muted-foreground">
            Showing {paginated.length} of {filtered.length} project{filtered.length !== 1 ? "s" : ""}
          </p>

          {/* Project Cards Grid */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
            {paginated.map((project) => {
              const status = statusConfig[project.status];
              return (
                <Link key={project.id} href={`/projects/${project.id}`}>
                  <Card className="h-full transition-all hover:shadow-md hover:border-primary/50 cursor-pointer">
                    <CardHeader className="flex flex-row items-start justify-between gap-4 pb-3">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="font-semibold truncate">{project.name}</h3>
                          <div className={`size-2 shrink-0 rounded-full ${status.dot}`} />
                        </div>
                        <p className="text-xs text-muted-foreground mt-1 line-clamp-2">{project.description}</p>
                      </div>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild onClick={(e) => e.stopPropagation()}>
                          <Button variant="ghost" size="icon" className="size-8 shrink-0">
                            <MoreHorizontal className="size-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem>View Details</DropdownMenuItem>
                          <DropdownMenuItem>Edit</DropdownMenuItem>
                          <DropdownMenuItem>Duplicate</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem className="text-destructive">Archive</DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </CardHeader>
                    <CardContent className="space-y-3 pt-0">
                      {/* Repo + Tech badges */}
                      <div className="flex items-center gap-2 flex-wrap">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            window.open(`https://${project.repository}`, "_blank", "noopener,noreferrer");
                          }}
                          className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors cursor-pointer bg-transparent border-0 p-0"
                        >
                          <ExternalLink className="size-3" />
                          {project.repository}
                        </button>
                        <Badge variant="outline" className="text-[10px] px-1.5 py-0">{project.tech}</Badge>
                      </div>

                      {/* Status + Team row */}
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2">
                          <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                          <span className="text-xs text-muted-foreground">{project.team}</span>
                        </div>
                        <span className="text-xs text-muted-foreground">{project.updated}</span>
                      </div>

                      {/* Contributors */}
                      <div className="flex items-center justify-between pt-1 border-t">
                        <div className="flex -space-x-1.5">
                          {project.members.map((m, i) => (
                            <Avatar key={i} className="size-6 border-2 border-background">
                              <AvatarFallback className="text-[9px]">{m.initials}</AvatarFallback>
                            </Avatar>
                          ))}
                        </div>
                      </div>
                    </CardContent>
                  </Card>
                </Link>
              );
            })}
          </div>

          {/* Pagination */}
          {totalPages > 1 && filtered.length > 0 && (
            <div className="flex items-center justify-center gap-2 pt-2">
              <Tooltip delay={0}>
                <Button
                  variant="outline"
                  size="icon"
                  className="size-10"
                  disabled={page === 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  <ChevronLeft className="size-4" />
                </Button>
                <Tooltip.Content>
                  <p>Previous page</p>
                </Tooltip.Content>
              </Tooltip>
              {Array.from({length: totalPages}, (_, i) => i + 1).map((p) => (
                <Tooltip key={p} delay={0}>
                  <Button
                    variant={p === page ? "default" : "outline"}
                    size="icon"
                    className="size-10"
                    onClick={() => setPage(p)}
                  >
                    {p}
                  </Button>
                  <Tooltip.Content>
                    <p>Page {p}</p>
                  </Tooltip.Content>
                </Tooltip>
              ))}
              <Tooltip delay={0}>
                <Button
                  variant="outline"
                  size="icon"
                  className="size-10"
                  disabled={page === totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  <ChevronRight className="size-4" />
                </Button>
                <Tooltip.Content>
                  <p>Next page</p>
                </Tooltip.Content>
              </Tooltip>
            </div>
          )}

          {/* Empty state */}
          {filtered.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <FolderKanban className="size-12 text-muted-foreground/40 mb-4" />
              <h3 className="text-lg font-medium">No projects found</h3>
              <p className="text-sm text-muted-foreground mt-1">Try adjusting your search or filters</p>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
