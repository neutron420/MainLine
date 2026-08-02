"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Copy, Check, FileCode2, GitBranch, Clock, Database, Play, History } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { useMigration, useMigrationRuns } from "@/lib/api/hooks/use-migrations";
import { getApiErrorMessage } from "@/lib/api/errors";

const migrationStatusConfig: Record<
  string,
  { label: string; badge: "default" | "secondary" | "destructive" | "outline" }
> = {
  draft: { label: "Draft", badge: "outline" },
  pending: { label: "Pending", badge: "secondary" },
  running: { label: "Running", badge: "secondary" },
  completed: { label: "Completed", badge: "default" },
  failed: { label: "Failed", badge: "destructive" },
  rolled_back: { label: "Rolled Back", badge: "outline" },
};

const runStatusConfig: Record<
  string,
  { label: string; badge: "default" | "secondary" | "destructive" | "outline" }
> = {
  pending: { label: "Pending", badge: "secondary" },
  running: { label: "Running", badge: "secondary" },
  completed: { label: "Completed", badge: "default" },
  failed: { label: "Failed", badge: "destructive" },
  rolled_back: { label: "Rolled Back", badge: "outline" },
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
  return new Date(iso).toLocaleDateString();
}

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <Button
      variant="ghost"
      size="sm"
      className="h-7 gap-1.5 text-xs"
      onClick={() => {
        navigator.clipboard.writeText(text);
        setCopied(true);
        setTimeout(() => setCopied(false), 1500);
      }}
    >
      {copied ? <Check className="size-3.5 text-green-500" /> : <Copy className="size-3.5" />}
      {copied ? "Copied" : "Copy"}
    </Button>
  );
}

export default function MigrationDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const migrationId = params.migrationId as string;

  const { data: migration, isLoading, error } = useMigration(projectId, migrationId);
  const { data: runs } = useMigrationRuns(migrationId);
  const [activeTab, setActiveTab] = useState("sql");

  if (isLoading) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex h-full flex-col items-center justify-center gap-4 p-8">
            <GitBranch className="size-12 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">Loading migration...</p>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  if (!migration) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <GitBranch className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Migration not found</h2>
            <p className="text-sm text-muted-foreground">
              {error ? getApiErrorMessage(error) : "The migration you are looking for does not exist."}
            </p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const status = migrationStatusConfig[migration.status] ?? migrationStatusConfig.draft;
  const canRun = !["completed", "rolled_back", "running", "pending"].includes(migration.status);

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}`}>Project</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>{migration.version}</BreadcrumbPage></BreadcrumbItem>
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
          <div className="flex flex-col gap-4">
            <div className="flex items-start gap-4">
              <Link href={`/projects/${projectId}`}>
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2.5 flex-wrap">
                  <Badge variant="outline" className="font-mono text-[11px]">{migration.version}</Badge>
                  <h1 className="text-2xl font-semibold tracking-tight truncate">{migration.title}</h1>
                  <Badge variant={status.badge} className="text-[11px]">{status.label}</Badge>
                </div>
                <div className="flex items-center gap-2 mt-1.5 text-sm text-muted-foreground flex-wrap">
                  <span>by <span className="text-foreground font-medium">{migration.createdBy}</span></span>
                  <span>· created {relativeTime(migration.createdAt)}</span>
                  {migration.updatedAt && migration.updatedAt !== migration.createdAt && (
                    <>
                      <span>·</span>
                      <span>updated {relativeTime(migration.updatedAt)}</span>
                    </>
                  )}
                </div>
                {migration.description && (
                  <p className="text-sm text-muted-foreground mt-1.5">{migration.description}</p>
                )}
              </div>
              {canRun && (
                <Link href={`/projects/${projectId}/migrations/${migration.id}/run`} className="shrink-0">
                  <Button className="h-11 gap-2">
                    <Play className="size-4" />
                    Run Migration
                  </Button>
                </Link>
              )}
            </div>
          </div>

          {/* Main grid */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
            {/* Left */}
            <div className="space-y-6 lg:col-span-2">
              <Card>
                <CardHeader className="border-0 pb-0">
                  <Tabs value={activeTab} onValueChange={setActiveTab}>
                    <TabsList variant="line" className="w-full justify-start gap-4">
                      <TabsTrigger value="sql" className="gap-1.5">
                        <FileCode2 className="size-4" />
                        SQL
                      </TabsTrigger>
                      <TabsTrigger value="runs" className="gap-1.5">
                        <History className="size-4" />
                        Runs
                        <span className="text-xs text-muted-foreground">{runs?.length ?? 0}</span>
                      </TabsTrigger>
                    </TabsList>
                  </Tabs>
                </CardHeader>
                <CardContent className="pt-4">
                  {activeTab === "sql" && (
                    <div className="space-y-4">
                      <div className="rounded-md border">
                        <div className="flex items-center justify-between border-b px-3 py-2">
                          <span className="font-mono text-xs font-medium text-muted-foreground">Up migration</span>
                          <CopyButton text={migration.upSql} />
                        </div>
                        <pre className="bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed whitespace-pre-wrap">
                          {migration.upSql || "-- no SQL defined"}
                        </pre>
                      </div>
                      {migration.downSql && (
                        <div className="rounded-md border">
                          <div className="flex items-center justify-between border-b px-3 py-2">
                            <span className="font-mono text-xs font-medium text-muted-foreground">Down migration</span>
                            <CopyButton text={migration.downSql} />
                          </div>
                          <pre className="bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed whitespace-pre-wrap">
                            {migration.downSql}
                          </pre>
                        </div>
                      )}
                    </div>
                  )}
                  {activeTab === "runs" && (
                    <div>
                      {!runs || runs.length === 0 ? (
                        <p className="text-sm text-muted-foreground py-6 text-center">
                          This migration has not been run yet.
                        </p>
                      ) : (
                        runs.map((run) => {
                          const runStatus = runStatusConfig[run.status] ?? runStatusConfig.pending;
                          return (
                            <div key={run.id} className="flex items-start gap-3 py-3 border-b last:border-b-0">
                              <div className={`mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full ${run.status === "completed" ? "bg-green-500/15 text-green-600" : run.status === "failed" || run.status === "rolled_back" ? "bg-red-500/15 text-red-600" : "bg-muted text-muted-foreground"}`}>
                                {run.status === "completed" ? <Check className="size-4" /> : <Clock className="size-4" />}
                              </div>
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center gap-2 flex-wrap">
                                  <p className="text-sm font-medium">{run.direction}</p>
                                  <Badge variant={runStatus.badge} className="text-[10px] px-1.5 py-0">{runStatus.label}</Badge>
                                </div>
                                <p className="text-xs text-muted-foreground mt-0.5">
                                  by {run.executedBy} · started {relativeTime(run.startedAt)}
                                  {run.completedAt ? ` · finished ${relativeTime(run.completedAt)}` : ""}
                                  {run.durationMs > 0 ? ` · ${Math.round(run.durationMs / 1000)}s` : ""}
                                </p>
                                {run.errorMessage && (
                                  <p className="text-xs text-red-500 mt-1 font-mono break-words">{run.errorMessage}</p>
                                )}
                              </div>
                            </div>
                          );
                        })
                      )}
                    </div>
                  )}
                </CardContent>
              </Card>
            </div>

            {/* Right */}
            <div className="space-y-6">
              <Card>
                <CardHeader className="border-0">
                  <CardTitle className="text-base">Details</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <div className="space-y-3">
                    <div className="flex items-center gap-3 text-sm">
                      <Database className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Project</p>
                        <p className="font-mono text-xs truncate">{migration.projectId}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <GitBranch className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Version</p>
                        <p className="font-mono text-xs">{migration.version}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <FileCode2 className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Checksum</p>
                        <p className="font-mono text-xs truncate">{migration.checksum || "—"}</p>
                      </div>
                    </div>
                    <Separator />
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">Status</span>
                      <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
