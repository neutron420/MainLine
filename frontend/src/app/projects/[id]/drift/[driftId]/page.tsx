"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, AlertTriangle, CheckCircle2, Eye, GitBranch } from "lucide-react";

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
import { driftItems, driftDiffs, driftSeverityConfig, driftStatusConfig } from "@/lib/drift-data";

const typeStyles = {
  added: {
    label: "Added",
    text: "text-emerald-600",
    bg: "bg-emerald-500/10",
    border: "border-l-emerald-500",
    sign: "+",
  },
  removed: {
    label: "Removed",
    text: "text-red-600",
    bg: "bg-red-500/10",
    border: "border-l-red-500",
    sign: "−",
  },
  changed: {
    label: "Changed",
    text: "text-amber-600",
    bg: "bg-amber-500/10",
    border: "border-l-amber-500",
    sign: "~",
  },
} as const;

export default function DriftDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const driftId = params.driftId as string;

  const [status, setStatus] = useState<"unresolved" | "acknowledged" | "resolved">(
    driftItems.find((d) => d.id === driftId)?.status ?? "unresolved"
  );

  const drift = driftItems.find((d) => d.id === driftId);
  if (!drift) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <AlertTriangle className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Drift not found</h2>
            <p className="text-sm text-muted-foreground">This drift report does not exist.</p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const severity = driftSeverityConfig[drift.severity];
  const statusCfg = driftStatusConfig[status];
  const diffRows = driftDiffs[drift.id] ?? [];

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>Drift</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>{drift.table}</BreadcrumbPage></BreadcrumbItem>
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
                <h1 className="text-2xl font-semibold tracking-tight font-mono">{drift.schema}.{drift.table}</h1>
                <Badge variant={severity.badge} className="text-[10px] px-1.5 py-0">{severity.label}</Badge>
                <Badge variant={statusCfg.badge} className="text-[10px] px-1.5 py-0">{statusCfg.label}</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {drift.kind} · {drift.env} · detected {drift.detected}
              </p>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Button variant="outline" className="h-10 gap-2" onClick={() => setStatus("acknowledged")} disabled={status === "resolved"}>
                <Eye className="size-4" />
                Acknowledge
              </Button>
              <Button className="h-10 gap-2" onClick={() => setStatus("resolved")}>
                <CheckCircle2 className="size-4" />
                Mark Resolved
              </Button>
            </div>
          </div>

          {drift.severity === "critical" && status === "unresolved" && (
            <div className="flex items-start gap-2.5 rounded-lg border border-red-500/40 bg-red-500/5 p-3.5 text-sm">
              <AlertTriangle className="size-4 text-red-500 shrink-0 mt-0.5" />
              <p className="text-muted-foreground">
                Critical drift — <span className="font-mono text-foreground">{drift.table}.{diffRows[0]?.column ?? ""}</span> exists in {drift.env} but not in the reference environment. Queries referencing it will fail against production.
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Diff */}
            <Card className="lg:col-span-2">
              <CardHeader className="border-0">
                <CardTitle className="text-base flex items-center gap-2">
                  <GitBranch className="size-4" />
                  Schema Diff
                </CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="grid grid-cols-[auto_1fr_1fr] gap-px overflow-hidden rounded-lg border bg-muted">
                  <div className="bg-background px-3 py-2 text-xs font-medium text-muted-foreground">Object</div>
                  <div className="bg-background px-3 py-2 text-xs font-medium">{drift.env}</div>
                  <div className="bg-background px-3 py-2 text-xs font-medium">Production</div>
                  {diffRows.map((row) => {
                    const style = typeStyles[row.type];
                    return (
                      <div key={row.column} className={`contents`}>
                        <div className={`flex items-center gap-2 bg-background px-3 py-2.5 border-l-2 ${style.border} ${style.bg}`}>
                          <span className={`text-xs font-mono font-semibold w-4 ${style.text}`}>{style.sign}</span>
                          <div>
                            <p className="text-xs font-mono">{row.column}</p>
                            <span className={`text-[10px] uppercase tracking-wide ${style.text}`}>{style.label}</span>
                          </div>
                        </div>
                        <div className={`bg-background px-3 py-2.5 font-mono text-xs ${style.text}`}>{row.staging}</div>
                        <div className={`bg-background px-3 py-2.5 font-mono text-xs ${style.text}`}>{row.production}</div>
                      </div>
                    );
                  })}
                </div>
              </CardContent>
            </Card>

            {/* Details */}
            <Card>
              <CardHeader className="border-0">
                <CardTitle className="text-base">Details</CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <dl className="space-y-4 text-sm">
                  <div>
                    <dt className="text-xs text-muted-foreground">Environment</dt>
                    <dd className="mt-0.5 flex items-center gap-2">
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{drift.env}</Badge>
                    </dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Kind</dt>
                    <dd className="mt-0.5">{drift.kind}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Table</dt>
                    <dd className="mt-0.5 font-mono text-xs">{drift.schema}.{drift.table}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Detected</dt>
                    <dd className="mt-0.5">{drift.detected}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Summary</dt>
                    <dd className="mt-0.5 text-muted-foreground">{drift.detail}</dd>
                  </div>
                </dl>
                <Separator className="my-4" />
                <Link href={`/projects/${projectId}/migrations/new`}>
                  <Button variant="outline" className="w-full h-9 gap-2">
                    <GitBranch className="size-4" />
                    Create Migration to Fix
                  </Button>
                </Link>
              </CardContent>
            </Card>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
