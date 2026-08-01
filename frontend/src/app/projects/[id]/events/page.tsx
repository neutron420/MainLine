"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Zap, GitBranch, AlertTriangle, GitPullRequest, PlugZap, Rocket } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { projectEvents, eventTypeConfig } from "@/lib/events-data";

const eventIcons = {
  migration: GitBranch,
  drift: AlertTriangle,
  review: GitPullRequest,
  connection: PlugZap,
  deploy: Rocket,
} as const;

export default function EventsPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [filter, setFilter] = useState<"all" | "migration" | "drift" | "review" | "connection" | "deploy">("all");

  const filtered = filter === "all" ? projectEvents : projectEvents.filter((e) => e.type === filter);

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
              <BreadcrumbItem><BreadcrumbLink href="/projects">Projects</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>SchemaHub</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>Events</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header */}
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight flex items-center gap-2">
                  <Zap className="size-6" />
                  Events
                </h1>
                <Badge variant="outline" className="text-[11px]">{projectEvents.length} events</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Real-time activity across this project</p>
            </div>
          </div>

          {/* Filter tabs */}
          <div className="flex items-center gap-1 rounded-lg border bg-muted/50 p-1 w-fit">
            {(["all", "migration", "drift", "review", "connection", "deploy"] as const).map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`rounded-md px-3 py-1.5 text-sm capitalize ${
                  filter === f ? "bg-background shadow-sm font-medium" : "text-muted-foreground hover:text-foreground"
                }`}
              >
                {f}
              </button>
            ))}
          </div>

          <Card>
            <CardHeader className="border-0 pb-0">
              <CardTitle className="text-base flex items-center gap-2">
                <Zap className="size-4" />
                Timeline
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              <div className="relative pl-6">
                <div className="absolute left-[7px] top-2 bottom-2 w-px bg-border" />
                <div className="space-y-0">
                  {filtered.map((event) => {
                    const config = eventTypeConfig[event.type];
                    const Icon = eventIcons[event.type];
                    return (
                      <div key={event.id} className="relative pb-6 last:pb-0">
                        <div className={`absolute -left-[26px] top-1 size-3.5 rounded-full border-2 border-background ${config.dot}`} />
                        <div className="flex items-center gap-3">
                          <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
                            <Icon className="size-4 text-muted-foreground" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <div className="flex items-center gap-2 flex-wrap">
                              <p className="text-sm font-medium">{event.title}</p>
                              <Badge variant="outline" className="text-[10px] px-1.5 py-0">{config.label}</Badge>
                            </div>
                            <p className="text-sm text-muted-foreground mt-0.5">{event.detail}</p>
                          </div>
                          <span className="text-xs text-muted-foreground shrink-0">{event.time}</span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
