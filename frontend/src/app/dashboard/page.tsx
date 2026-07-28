"use client";

import { ArrowDown, ArrowUp, MoreHorizontal, Pin, Settings, Share2, Trash, TriangleAlert, ListFilter, Columns, Plus, GitBranch, Database, UserPlus, Clock, CheckCircle2, GitPullRequest, AlertCircle } from "lucide-react";
import { useState, useMemo } from "react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
import { Input } from "@/components/ui/input";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { ProjectDataTable, Project } from "@/components/ui/project-data-table";

const stats = [
  { title: "Total Schemas", value: 1248, delta: 12.5, lastMonth: 1109, positive: true, prefix: "", suffix: "" },
  { title: "Active Projects", value: 36, delta: 8.3, lastMonth: 33, positive: true, prefix: "", suffix: "" },
  { title: "Pending Reviews", value: 14, delta: -5.2, lastMonth: 19, positive: false, prefix: "", suffix: "" },
  { title: "Team Members", value: 24, delta: 4.1, lastMonth: 23, positive: true, prefix: "", suffix: "" },
];

function formatNumber(n: number) {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + "M";
  if (n >= 1_000) return n.toLocaleString();
  return n.toString();
}

const mockProjects: Project[] = [
  { id: "p1", name: "User Service Schema", repository: "https://github.com/mainline/user-service", team: "Platform", tech: "PostgreSQL", createdAt: "2026-07-15", contributors: [{ src: "https://i.pravatar.cc/150?u=a1", alt: "Alice", fallback: "AL" }, { src: "https://i.pravatar.cc/150?u=a2", alt: "Bob", fallback: "BO" }], status: { text: "Active", variant: "active" } },
  { id: "p2", name: "Payment DB Migration", repository: "https://github.com/mainline/payment-db", team: "Payments", tech: "PostgreSQL", createdAt: "2026-07-10", contributors: [{ src: "https://i.pravatar.cc/150?u=a3", alt: "Charlie", fallback: "CH" }], status: { text: "In Progress", variant: "inProgress" } },
  { id: "p3", name: "Analytics Warehouse", repository: "https://github.com/mainline/analytics-db", team: "Data", tech: "PostgreSQL", createdAt: "2026-06-28", contributors: [{ src: "https://i.pravatar.cc/150?u=a4", alt: "Diana", fallback: "DI" }, { src: "https://i.pravatar.cc/150?u=a5", alt: "Eve", fallback: "EV" }], status: { text: "Active", variant: "active" } },
  { id: "p4", name: "Legacy CRM Schema", repository: "https://github.com/mainline/crm-migration", team: "Core", tech: "PostgreSQL", createdAt: "2026-06-20", contributors: [{ src: "https://i.pravatar.cc/150?u=a6", alt: "Frank", fallback: "FR" }, { src: "https://i.pravatar.cc/150?u=a7", alt: "Grace", fallback: "GR" }], status: { text: "On Hold", variant: "onHold" } },
  { id: "p5", name: "Notification Queue", repository: "https://github.com/mainline/notif-db", team: "Infra", tech: "PostgreSQL", createdAt: "2026-06-15", contributors: [{ src: "https://i.pravatar.cc/150?u=a8", alt: "Hank", fallback: "HA" }], status: { text: "Active", variant: "active" } },
  { id: "p6", name: "Search Index Schema", repository: "https://github.com/mainline/search-db", team: "Search", tech: "PostgreSQL", createdAt: "2026-06-01", contributors: [{ src: "https://i.pravatar.cc/150?u=a9", alt: "Ivy", fallback: "IV" }, { src: "https://i.pravatar.cc/150?u=a10", alt: "Jack", fallback: "JA" }], status: { text: "In Progress", variant: "inProgress" } },
];

const allColumns: (keyof Project)[] = ["name", "repository", "team", "tech", "createdAt", "contributors", "status"];

const quickActions = [
  { label: "New Project", icon: Plus, variant: "default" as const },
  { label: "New Schema", icon: Database, variant: "outline" as const },
  { label: "Run Migration", icon: GitBranch, variant: "outline" as const },
  { label: "Invite Team", icon: UserPlus, variant: "outline" as const },
];

const activities = [
  { type: "migration", icon: GitBranch, project: "User Service", description: "Migration v1.2.0 deployed to production", user: "Alice", time: "2 hours ago", color: "bg-blue-500" },
  { type: "review", icon: CheckCircle2, project: "Payment DB", description: "Schema review approved update payment_status", user: "Bob", time: "3 hours ago", color: "bg-green-500" },
  { type: "pr", icon: GitPullRequest, project: "CRM", description: "New pull request drop legacy zip_code column", user: "Charlie", time: "5 hours ago", color: "bg-purple-500" },
  { type: "alert", icon: AlertCircle, project: "Analytics", description: "Schema drift detected warehouse.analytics differs from staging", user: "System", time: "1 day ago", color: "bg-amber-500" },
  { type: "migration", icon: GitBranch, project: "Search Index", description: "Migration v0.8.0 rolled back due to index conflict", user: "Diana", time: "2 days ago", color: "bg-red-500" },
];

const pendingReviews = [
  { title: "Add users table composite index", project: "User Service", author: "Alice", status: "changes", time: "1 hour ago", priority: "high" as const },
  { title: "Update payment status enum values", project: "Payment DB", author: "Bob", status: "pending", time: "3 hours ago", priority: "medium" as const },
  { title: "Drop legacy zip_code column", project: "CRM", author: "Charlie", status: "pending", time: "1 day ago", priority: "low" as const },
  { title: "Add email verification column", project: "User Service", author: "Alice", status: "pending", time: "2 days ago", priority: "medium" as const },
];

export default function Page() {
  return (
    <SidebarProvider
      style={{ "--sidebar-width": "350px" } as React.CSSProperties}
    >
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex shrink-0 items-center gap-2 border-b bg-background p-4">
          <SidebarTrigger className="-ml-1" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbPage>Dashboard</BreadcrumbPage></BreadcrumbItem>
              <BreadcrumbSeparator className="hidden md:block" />
              <BreadcrumbItem className="hidden md:block"><BreadcrumbPage>Overview</BreadcrumbPage></BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
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
                    <Badge variant={stat.positive ? "default" : "destructive"}>
                      {stat.delta > 0 ? <ArrowUp /> : <ArrowDown />}
                      {Math.abs(stat.delta)}%
                    </Badge>
                  </div>
                  <div className="text-xs text-muted-foreground mt-2 border-t pt-2.5">
                    Vs last month:{" "}
                    <span className="font-medium text-foreground">
                      {stat.prefix + formatNumber(stat.lastMonth) + stat.suffix}
                    </span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Quick Actions */}
          <div className="flex flex-wrap gap-3">
            {quickActions.map((action) => (
              <Button key={action.label} variant={action.variant} className="h-11 gap-2">
                <action.icon className="size-4" />
                {action.label}
              </Button>
            ))}
          </div>

          {/* Activity + Reviews row */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Activity Feed */}
            <Card>
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <Clock className="size-4" />
                  Recent Activity
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="space-y-0">
                  {activities.map((activity, i) => (
                    <div key={i} className="flex gap-3 py-3 border-b last:border-b-0">
                      <div className={`mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full ${activity.color} text-white`}>
                        <activity.icon className="size-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{activity.description}</p>
                        <div className="flex items-center gap-2 mt-0.5">
                          <span className="text-xs text-muted-foreground">{activity.project}</span>
                          <span className="text-xs text-muted-foreground">·</span>
                          <span className="text-xs text-muted-foreground">{activity.user}</span>
                          <span className="text-xs text-muted-foreground ml-auto">{activity.time}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>

            {/* Pending Reviews */}
            <Card>
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <GitPullRequest className="size-4" />
                  Pending Reviews
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="space-y-0">
                  {pendingReviews.map((review, i) => (
                    <div key={i} className="flex items-start gap-3 py-3 border-b last:border-b-0">
                      <Avatar className="size-8">
                        <AvatarFallback className="text-xs">{review.author.slice(0, 2).toUpperCase()}</AvatarFallback>
                      </Avatar>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium truncate">{review.title}</p>
                        <div className="flex items-center gap-2 mt-0.5">
                          <span className="text-xs text-muted-foreground">{review.project}</span>
                          <span className="text-xs text-muted-foreground">·</span>
                          <span className="text-xs text-muted-foreground">{review.author}</span>
                        </div>
                      </div>
                      <div className="flex flex-col items-end gap-1 shrink-0">
                        <Badge variant={review.priority === "high" ? "destructive" : review.priority === "medium" ? "default" : "secondary"} className="text-[10px] px-1.5 py-0">
                          {review.priority}
                        </Badge>
                        {review.status === "changes" && (
                          <span className="text-[10px] text-amber-500 font-medium">Changes requested</span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </div>

          {/* Projects Table */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h2 className="text-xl font-semibold tracking-tight">All Projects</h2>
            </div>
            <FilterableProjectTable />
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}

function FilterableProjectTable() {
  const [techFilter, setTechFilter] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [visibleColumns, setVisibleColumns] = useState<Set<keyof Project>>(new Set(allColumns));

  const filteredProjects = useMemo(() => {
    return mockProjects.filter((project) => {
      const techMatch = techFilter === "" || project.tech.toLowerCase().includes(techFilter.toLowerCase());
      const statusMatch = statusFilter === "all" || project.status.variant === statusFilter;
      return techMatch && statusMatch;
    });
  }, [techFilter, statusFilter]);

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
      <ProjectDataTable projects={filteredProjects} visibleColumns={visibleColumns} />
    </div>
  );
}
