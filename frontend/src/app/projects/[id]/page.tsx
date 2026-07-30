"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Database, GitBranch, GitPullRequest, AlertTriangle, FileText, Calendar, Users, Settings, ExternalLink, Clock, CheckCircle2, XCircle, Search } from "lucide-react";

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
                <CardContent className="py-8 text-center">
                  <Database className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Schema management coming soon</p>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="migrations" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <GitBranch className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Migration history coming soon</p>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="drift" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <AlertTriangle className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Drift detection coming soon</p>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="audit" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <FileText className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Audit log coming soon</p>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="events" className="mt-6">
              <Card>
                <CardContent className="py-8 text-center">
                  <Calendar className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                  <p className="text-sm text-muted-foreground">Event history coming soon</p>
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
