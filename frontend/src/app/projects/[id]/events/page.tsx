"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Zap, GitBranch, AlertTriangle, GitPullRequest, PlugZap, Rocket, Radio } from "lucide-react";

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
import { useEventStream } from "@/lib/api/hooks/use-realtime";

const eventIcons = {
  migration: GitBranch,
  drift: AlertTriangle,
  review: GitPullRequest,
  connection: PlugZap,
  deploy: Rocket,
  schema: GitBranch,
} as const;

const filterTypes = ["all", "migration", "drift", "review", "connection", "deploy", "schema"] as const;

function eventCategory(type: string): keyof typeof eventIcons {
  if (type.includes("drift")) return "drift";
  if (type.includes("review")) return "review";
  if (type.includes("connection")) return "connection";
  if (type.includes("deploy")) return "deploy";
  if (type.includes("schema")) return "schema";
  return "migration";
}

function relativeTime(iso: string | undefined): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

export default function EventsPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [filter, setFilter] = useState<(typeof filterTypes)[number]>("all");

  const { events, connected } = useEventStream({ projectIds: [projectId], maxEvents: 100 });

  const filtered =
    filter === "all" ? events : events.filter((e) => eventCategory(e.type) === filter);

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
                <Badge variant="outline" className="text-[11px]">{events.length} events</Badge>
                <Badge variant={connected ? "default" : "secondary"} className="text-[10px] px-1.5 py-0 gap-1">
                  <span className={`size-1.5 rounded-full ${connected ? "bg-green-500 animate-pulse" : "bg-muted-foreground/50"}`} />
                  {connected ? "LIVE" : "Reconnecting…"}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Real-time activity across this project</p>
            </div>
          </div>

          {/* Filter tabs */}
          <div className="flex items-center gap-1 rounded-lg border bg-muted/50 p-1 w-fit overflow-x-auto">
            {filterTypes.map((f) => (
              <button
                key={f}
                onClick={() => setFilter(f)}
                className={`rounded-md px-3 py-1.5 text-sm capitalize whitespace-nowrap ${
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
                <Radio className="size-4" />
                Timeline
                <Badge variant="outline" className="text-[10px] px-1.5 py-0">{filtered.length}</Badge>
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              <div className="relative pl-6">
                <div className="absolute left-[7px] top-2 bottom-2 w-px bg-border" />
                <div className="space-y-0">
                  {filtered.length === 0 ? (
                    <p className="text-sm text-muted-foreground py-6 text-center pl-6">
                      {connected ? "Waiting for events..." : "Event stream disconnected — reconnecting..."}
                    </p>
                  ) : (
                    filtered.map((event) => {
                      const Icon = eventIcons[eventCategory(event.type)];
                      return (
                        <div key={event.id} className="relative pb-6 last:pb-0">
                          <div className="absolute -left-[26px] top-1 size-3.5 rounded-full border-2 border-background bg-primary" />
                          <div className="flex items-center gap-3">
                            <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-muted">
                              <Icon className="size-4 text-muted-foreground" />
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2 flex-wrap">
                                <p className="text-sm font-medium capitalize">{event.type.replace(/_/g, " ")}</p>
                                <Badge variant="outline" className="text-[10px] px-1.5 py-0">
                                  {event.resource?.type ?? "event"}
                                </Badge>
                              </div>
                              <p className="text-sm text-muted-foreground mt-0.5">
                                {event.actor?.email ?? "system"}
                                {event.resource?.id ? ` · ${event.resource.id.slice(0, 8)}` : ""}
                              </p>
                            </div>
                            <span className="text-xs text-muted-foreground shrink-0">{relativeTime(event.timestamp)}</span>
                          </div>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
