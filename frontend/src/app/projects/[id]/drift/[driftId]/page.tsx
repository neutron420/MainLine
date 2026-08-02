"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, AlertTriangle, CheckCircle2, Eye, GitBranch, Loader2 } from "lucide-react";

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
import { useDriftEvent, useResolveDriftEvent } from "@/lib/api/hooks/use-drift";
import { getApiErrorMessage } from "@/lib/api/errors";

const severityConfig: Record<string, { label: string; badge: "default" | "destructive" | "secondary" | "outline" }> = {
  critical: { label: "Critical", badge: "destructive" },
  high: { label: "High", badge: "destructive" },
  medium: { label: "Medium", badge: "secondary" },
  low: { label: "Low", badge: "outline" },
};

const statusConfig: Record<string, { label: string; badge: "default" | "secondary" | "destructive" | "outline" }> = {
  unresolved: { label: "Unresolved", badge: "destructive" },
  acknowledged: { label: "Acknowledged", badge: "secondary" },
  resolved: { label: "Resolved", badge: "default" },
  false_positive: { label: "False Positive", badge: "outline" },
};

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

export default function DriftDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const driftId = params.driftId as string;

  const { data: drift, isLoading, error } = useDriftEvent(driftId);
  const resolve = useResolveDriftEvent();

  if (isLoading) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex h-full flex-col items-center justify-center gap-4 p-8">
            <AlertTriangle className="size-12 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">Loading drift report...</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  if (!drift) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <AlertTriangle className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Drift not found</h2>
            <p className="text-sm text-muted-foreground">
              {error ? getApiErrorMessage(error) : "This drift report does not exist."}
            </p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const severity = severityConfig[drift.severity] ?? severityConfig.medium;
  const statusCfg = statusConfig[drift.status] ?? statusConfig.unresolved;
  const hasResolved = drift.status === "resolved" || drift.status === "false_positive";

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}/drift`}>Drift</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>{drift.objectName}</BreadcrumbPage></BreadcrumbItem>
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
            <Link href={`/projects/${projectId}/drift`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight font-mono">{drift.objectName}</h1>
                <Badge variant={severity.badge} className="text-[10px] px-1.5 py-0">{severity.label}</Badge>
                <Badge variant={statusCfg.badge} className="text-[10px] px-1.5 py-0">{statusCfg.label}</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {drift.driftType} · detected {relativeTime(drift.detectedAt)}
              </p>
            </div>
            <div className="flex items-center gap-2 shrink-0">
              <Button
                variant="outline"
                className="h-10 gap-2"
                disabled={resolve.isPending || hasResolved}
                onClick={() => resolve.mutate({ eventId: drift.id, status: "acknowledged" })}
              >
                {resolve.isPending ? <Loader2 className="size-4 animate-spin" /> : <Eye className="size-4" />}
                Acknowledge
              </Button>
              <Button
                className="h-10 gap-2"
                disabled={resolve.isPending || hasResolved}
                onClick={() => resolve.mutate({ eventId: drift.id, status: "resolved" })}
              >
                <CheckCircle2 className="size-4" />
                Mark Resolved
              </Button>
            </div>
          </div>

          {drift.severity === "critical" && !hasResolved && (
            <div className="flex items-start gap-2.5 rounded-lg border border-red-500/40 bg-red-500/5 p-3.5 text-sm">
              <AlertTriangle className="size-4 text-red-500 shrink-0 mt-0.5" />
              <p className="text-muted-foreground">
                Critical drift on <span className="font-mono text-foreground">{drift.objectName}</span>. The live
                database no longer matches the tracked schema version. Resolve this before running migrations.
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
                <div className="grid grid-cols-[1fr_1fr] gap-px overflow-hidden rounded-lg border bg-muted">
                  <div className="bg-background px-3 py-2 text-xs font-medium">Expected (tracked)</div>
                  <div className="bg-background px-3 py-2 text-xs font-medium">Actual (live)</div>
                  <div className="bg-background px-3 py-4 font-mono text-xs whitespace-pre-wrap break-words">
                    {drift.expectedDefinition || "—"}
                  </div>
                  <div className="bg-background px-3 py-4 font-mono text-xs whitespace-pre-wrap break-words">
                    {drift.actualDefinition || "—"}
                  </div>
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
                    <dt className="text-xs text-muted-foreground">Object</dt>
                    <dd className="mt-0.5 font-mono text-xs">{drift.objectName}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Type</dt>
                    <dd className="mt-0.5">{drift.objectType}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Drift Kind</dt>
                    <dd className="mt-0.5">{drift.driftType}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Detected</dt>
                    <dd className="mt-0.5">{relativeTime(drift.detectedAt)}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Expected Version</dt>
                    <dd className="mt-0.5 font-mono text-xs">{drift.expectedVersionId.slice(0, 8)}</dd>
                  </div>
                  <div>
                    <dt className="text-xs text-muted-foreground">Summary</dt>
                    <dd className="mt-0.5 text-muted-foreground">{drift.diffSummary || "—"}</dd>
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
