"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, AlertTriangle, CheckCircle2, Eye, Loader2, RefreshCw } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
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
import { useConnections } from "@/lib/api/hooks/use-connections";
import {
  useDriftEvents,
  useCheckDrift,
} from "@/lib/api/hooks/use-drift";
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

export default function DriftPage() {
  const params = useParams();
  const projectId = params.id as string;

  const { data: connections } = useConnections(projectId);
  const [connectionId, setConnectionId] = useState<string>("");
  const effectiveConnectionId = connectionId || connections?.[0]?.id || "";

  const { data: events, isLoading, error } = useDriftEvents({
    connectionId: effectiveConnectionId || undefined,
  });
  const checkDrift = useCheckDrift();

  const [search, setSearch] = useState("");
  const filteredEvents = events?.filter((d) => {
    if (search === "") return true;
    const haystack = `${d.objectName} ${d.objectType} ${d.driftType} ${d.status}`.toLowerCase();
    return haystack.includes(search.toLowerCase());
  });

  const unresolved = events?.filter((d) => d.status === "unresolved").length ?? 0;

  return (
            <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search drift..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
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
                  <AlertTriangle className="size-6" />
                  Schema Drift
                </h1>
                <Badge variant="outline" className="text-[11px]">{events?.length ?? 0} reports</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {unresolved > 0 && <span className="text-red-500 font-medium">{unresolved} unresolved · </span>}
                Differences between tracked and live schema state
              </p>
            </div>
          </div>

          {/* Controls */}
          <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
            <Select value={effectiveConnectionId} onValueChange={setConnectionId}>
              <SelectTrigger className="h-10 w-72">
                <SelectValue placeholder="Select connection" />
              </SelectTrigger>
              <SelectContent>
                {(connections ?? []).map((conn) => (
                  <SelectItem key={conn.id} value={conn.id}>
                    {conn.name} ({conn.databaseName})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button
              variant="outline"
              className="h-10 gap-2 w-fit"
              disabled={!effectiveConnectionId || checkDrift.isPending}
              onClick={() => checkDrift.mutate({ connectionId: effectiveConnectionId })}
            >
              {checkDrift.isPending ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
              {checkDrift.isPending ? "Checking…" : "Check for Drift"}
            </Button>
            {checkDrift.isSuccess && (
              <span className="text-sm text-muted-foreground">
                {checkDrift.data.totalDrifts > 0 ? (
                  <span className="text-red-500 font-medium">{checkDrift.data.totalDrifts} drift(s) found</span>
                ) : (
                  <span className="text-green-500 font-medium">No drift detected</span>
                )}
              </span>
            )}
            {checkDrift.isError && (
              <span className="text-sm text-red-500">{getApiErrorMessage(checkDrift.error)}</span>
            )}
          </div>

          <Card>
            <CardHeader className="border-0 pb-0">
              <CardTitle className="text-base flex items-center gap-2">
                <AlertTriangle className="size-4" />
                Drift reports
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              {!effectiveConnectionId ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  Add a connection to this project to start detecting drift.
                </p>
              ) : isLoading ? (
                <p className="text-sm text-muted-foreground py-6 text-center">Loading drift reports...</p>
              ) : error ? (
                <p className="text-sm text-red-500 py-6 text-center">{getApiErrorMessage(error)}</p>
              ) : !events || events.length === 0 ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  No drift detected. Run a check to verify the current state.
                </p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[30%]">Object</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Severity</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead className="text-right">Detected</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredEvents?.map((drift) => {
                      const severity = severityConfig[drift.severity] ?? severityConfig.medium;
                      const status = statusConfig[drift.status] ?? statusConfig.unresolved;
                      return (
                        <TableRow key={drift.id} className="cursor-pointer">
                          <TableCell>
                            <Link href={`/projects/${projectId}/drift/${drift.id}`} className="flex items-center gap-2.5">
                              <span className="font-mono text-sm">{drift.objectName || drift.objectType}</span>
                            </Link>
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">{drift.driftType}</TableCell>
                          <TableCell>
                            <Badge variant={severity.badge} className="text-[10px] px-1.5 py-0">{severity.label}</Badge>
                          </TableCell>
                          <TableCell>
                            <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 gap-1">
                              {status.label === "Unresolved" ? <AlertTriangle className="size-2.5" /> : status.label === "Acknowledged" ? <Eye className="size-2.5" /> : <CheckCircle2 className="size-2.5" />}
                              {status.label}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-right text-sm text-muted-foreground">
                            {relativeTime(drift.detectedAt)}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
  );
}
