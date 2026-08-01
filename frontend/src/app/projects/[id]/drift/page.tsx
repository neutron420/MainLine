"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, AlertTriangle, CheckCircle2, Eye } from "lucide-react";

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
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { driftItems, driftSeverityConfig, driftStatusConfig } from "@/lib/drift-data";

const envColors: Record<string, string> = {
  Production: "bg-blue-500/10 text-blue-600",
  Staging: "bg-purple-500/10 text-purple-600",
  Development: "bg-slate-500/10 text-slate-600",
};

export default function DriftPage() {
  const params = useParams();
  const projectId = params.id as string;
  const [filter, setFilter] = useState<"all" | "unresolved" | "acknowledged" | "resolved">("all");

  const unresolved = driftItems.filter((d) => d.status === "unresolved").length;
  const filtered = filter === "all" ? driftItems : driftItems.filter((d) => d.status === filter);

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
              <BreadcrumbItem><BreadcrumbPage>Drift</BreadcrumbPage></BreadcrumbItem>
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
                  <AlertTriangle className="size-6" />
                  Schema Drift
                </h1>
                <Badge variant="outline" className="text-[11px]">{driftItems.length} reports</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">
                {unresolved > 0 && <span className="text-red-500 font-medium">{unresolved} unresolved · </span>}
                Differences between connected environments
              </p>
            </div>
          </div>

          {/* Filter tabs */}
          <div className="flex items-center gap-1 rounded-lg border bg-muted/50 p-1 w-fit">
            {(["all", "unresolved", "acknowledged", "resolved"] as const).map((f) => (
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
                <AlertTriangle className="size-4" />
                Drift reports
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[30%]">Table</TableHead>
                    <TableHead>Kind</TableHead>
                    <TableHead>Env</TableHead>
                    <TableHead>Severity</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="text-right">Detected</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {filtered.map((drift) => {
                    const severity = driftSeverityConfig[drift.severity];
                    const status = driftStatusConfig[drift.status];
                    return (
                      <TableRow key={drift.id} className="cursor-pointer">
                        <TableCell>
                          <Link href={`/projects/${projectId}/drift/${drift.id}`} className="flex items-center gap-2.5">
                            <span className="font-mono text-sm">{drift.schema}.{drift.table}</span>
                          </Link>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{drift.kind}</TableCell>
                        <TableCell>
                          <span className={`inline-flex items-center rounded-md px-2 py-0.5 text-[11px] font-medium ${envColors[drift.env]}`}>
                            {drift.env}
                          </span>
                        </TableCell>
                        <TableCell>
                          <Badge variant={severity.badge} className="text-[10px] px-1.5 py-0">{severity.label}</Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 gap-1">
                            {status.label === "Unresolved" ? <AlertTriangle className="size-2.5" /> : status.label === "Acknowledged" ? <Eye className="size-2.5" /> : <CheckCircle2 className="size-2.5" />}
                            {status.label}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right text-sm text-muted-foreground">{drift.detected}</TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
