"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Play, Loader2, CheckCircle2, AlertTriangle, GitBranch, RotateCcw, Terminal } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
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
import { useMigration ,
  useExecuteMigration,
  useRollbackMigration,
  useWatchMigration,
} from "@/lib/api/hooks/use-migrations";
import { useConnections } from "@/lib/api/hooks/use-connections";
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

function formatElapsed(ms: number | undefined): string {
  if (!ms) return "0s";
  if (ms < 1000) return `${ms}ms`;
  return `${(ms / 1000).toFixed(1)}s`;
}

export default function MigrationRunPage() {
  const params = useParams();
  const projectId = params.id as string;
  const migrationId = params.migrationId as string;

  const { data: migration, isLoading } = useMigration(projectId, migrationId);
  const { data: connections } = useConnections(projectId);

  const [connectionId, setConnectionId] = useState("");
  const [runId, setRunId] = useState<string | null>(null);

  const execute = useExecuteMigration();
  const rollback = useRollbackMigration();
  const { status, logs, connected } = useWatchMigration(runId ?? undefined);

  const alreadyRun = migration && ["completed", "rolled_back"].includes(migration.status);

  const startRun = () => {
    if (!connectionId || execute.isPending) return;
    execute.mutate(
      { migrationId, connectionId },
      {
        onSuccess: (run) => {
          if (run?.id) setRunId(run.id);
        },
      },
    );
  };

  const doRollback = () => {
    if (!runId) return;
    rollback.mutate({ migrationId, runId });
  };

  const isTerminal =
    status?.state === "completed" || status?.state === "failed" || status?.state === "rolled_back";

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
              <BreadcrumbItem>
                <BreadcrumbLink href={`/projects/${projectId}/migrations/${migrationId}`}>
                  {migration?.version ?? "Migration"}
                </BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>Run</BreadcrumbPage></BreadcrumbItem>
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
              <Link href={`/projects/${projectId}/migrations/${migrationId}`}>
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2.5 flex-wrap">
                  <Badge variant="outline" className="font-mono text-[11px]">{migration?.version ?? "—"}</Badge>
                  <h1 className="text-2xl font-semibold tracking-tight truncate">Run Migration</h1>
                  {migration && (
                    <Badge variant={(migrationStatusConfig[migration.status] ?? migrationStatusConfig.draft).badge} className="text-[11px]">
                      {(migrationStatusConfig[migration.status] ?? migrationStatusConfig.draft).label}
                    </Badge>
                  )}
                </div>
                <p className="text-sm text-muted-foreground mt-1">{migration?.title ?? (isLoading ? "Loading..." : "Migration")}</p>
              </div>
            </div>
          </div>

          {alreadyRun ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                <CheckCircle2 className="size-12 text-green-500 mb-4" />
                <h2 className="text-lg font-semibold">
                  {migration.status === "completed" ? "Already executed" : "Migration was rolled back"}
                </h2>
                <p className="text-sm text-muted-foreground mt-1 max-w-md">
                  {migration.status === "completed"
                    ? "This migration has already been applied. Create a new migration for further changes."
                    : "This migration is no longer active and cannot be re-run."}
                </p>
                <Link href={`/projects/${projectId}/migrations/${migrationId}`} className="mt-6">
                  <Button variant="outline">Back to Migration</Button>
                </Link>
              </CardContent>
            </Card>
          ) : isTerminal ? (
            status?.state === "completed" ? (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                  <CheckCircle2 className="size-12 text-green-500 mb-4" />
                  <h2 className="text-lg font-semibold">Migration executed successfully</h2>
                  <p className="text-sm text-muted-foreground mt-1">
                    {migration?.version} applied in {formatElapsed(Number(status.elapsedMs ?? 0))}
                  </p>
                  <div className="flex items-center gap-2 mt-6">
                    <Link href={`/projects/${projectId}/migrations/${migrationId}`}>
                      <Button variant="outline">View Migration</Button>
                    </Link>
                    <Link href={`/projects/${projectId}`}>
                      <Button>Back to Project</Button>
                    </Link>
                  </div>
                </CardContent>
              </Card>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                  <AlertTriangle className="size-12 text-red-500 mb-4" />
                  <h2 className="text-lg font-semibold">
                    {status?.state === "rolled_back" ? "Migration rolled back" : "Migration failed"}
                  </h2>
                  <p className="text-sm text-muted-foreground mt-1 max-w-lg">
                    {status?.errorMessage || rollback.data?.errorMessage || "Execution did not complete. Check the logs below."}
                  </p>
                  {status?.state !== "rolled_back" && (
                    <Button
                      className="mt-6 gap-2"
                      variant="destructive"
                      onClick={doRollback}
                      disabled={rollback.isPending || !runId}
                    >
                      {rollback.isPending ? (
                        <>
                          <Loader2 className="size-4 animate-spin" />
                          Rolling back…
                        </>
                      ) : (
                        <>
                          <RotateCcw className="size-4" />
                          Rollback Migration
                        </>
                      )}
                    </Button>
                  )}
                  <Link href={`/projects/${projectId}/migrations/${migrationId}`} className="mt-4">
                    <Button variant="ghost">Back to Migration</Button>
                  </Link>
                </CardContent>
              </Card>
            )
          ) : (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
              {/* Left */}
              <div className="space-y-6 lg:col-span-2">
                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <GitBranch className="size-4" />
                      Execution
                    </CardTitle>
                    <CardDescription>
                      {runId ? "Live progress via server stream" : "Choose a connection and run the migration"}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="pt-0">
                    {!runId ? (
                      <div className="flex flex-col items-center justify-center py-10 text-center gap-3">
                        <Play className="size-8 text-muted-foreground/40" />
                        <p className="text-sm text-muted-foreground max-w-sm">
                          Execution runs the up migration inside a transaction and streams progress back in real time.
                        </p>
                      </div>
                    ) : (
                      <div className="space-y-4">
                        {status && (
                          <div className="space-y-3">
                            <div className="flex items-center gap-3 flex-wrap">
                              <Badge variant="secondary" className="text-[10px] px-1.5 py-0 gap-1">
                                <span className="size-1.5 rounded-full bg-primary animate-pulse" />
                                {connected ? "Live" : "Connecting..."}
                              </Badge>
                              <span className="text-sm font-medium">{status.state}</span>
                              <span className="text-xs text-muted-foreground ml-auto">{formatElapsed(Number(status.elapsedMs ?? 0))}</span>
                            </div>
                            {status.totalStatements > 0 && (
                              <div>
                                <div className="flex items-center justify-between text-xs text-muted-foreground mb-1.5">
                                  <span>
                                    {status.completedStatements} / {status.totalStatements} statements
                                  </span>
                                  <span>
                                    {status.totalStatements > 0
                                      ? Math.round((status.completedStatements / status.totalStatements) * 100)
                                      : 0}
                                    %
                                  </span>
                                </div>
                                <div className="h-2 rounded-full bg-muted overflow-hidden">
                                  <div
                                    className="h-full bg-primary transition-all"
                                    style={{
                                      width: `${
                                        status.totalStatements > 0
                                          ? (status.completedStatements / status.totalStatements) * 100
                                          : 0
                                      }%`,
                                    }}
                                  />
                                </div>
                              </div>
                            )}
                            {status.currentStatement && (
                              <pre className="rounded-md bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed">
                                {status.currentStatement}
                              </pre>
                            )}
                            {status.errorMessage && (
                              <p className="text-sm text-red-500 font-mono break-words">{status.errorMessage}</p>
                            )}
                          </div>
                        )}
                        {logs.length > 0 && (
                          <div className="rounded-md border">
                            <div className="flex items-center gap-2 border-b px-3 py-2">
                              <Terminal className="size-3.5 text-muted-foreground" />
                              <span className="font-mono text-xs font-medium text-muted-foreground">Execution log</span>
                            </div>
                            <div className="max-h-56 overflow-y-auto">
                              {logs.map((log, i) => (
                                <div key={`${log.sequence}-${i}`} className="flex items-start gap-3 px-3 py-2 border-b last:border-b-0 font-mono text-xs">
                                  <span className="text-muted-foreground shrink-0">#{log.sequence}</span>
                                  <span className="text-muted-foreground shrink-0">
                                    {log.durationMs >= 0 ? `${formatElapsed(log.durationMs)}` : ""}
                                  </span>
                                  <span className="truncate flex-1">
                                    {log.sql || log.errorMessage || "statement"}
                                  </span>
                                  {log.rowsAffected > 0 && (
                                    <span className="text-muted-foreground shrink-0">{log.rowsAffected} rows</span>
                                  )}
                                </div>
                              ))}
                            </div>
                          </div>
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
                    <CardTitle className="text-base">Run Configuration</CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="grid gap-5">
                      <div className="grid gap-2">
                        <Label htmlFor="conn">Connection</Label>
                        <Select value={connectionId} onValueChange={setConnectionId} disabled={Boolean(runId)}>
                          <SelectTrigger id="conn" className="h-11">
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
                        {(!connections || connections.length === 0) && (
                          <p className="text-xs text-amber-500 flex items-center gap-1.5">
                            <AlertTriangle className="size-3.5" />
                            No connections in this project yet.
                          </p>
                        )}
                      </div>
                      <Separator />
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Migration</span>
                        <span className="font-mono">{migration?.version ?? "—"}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Checksum</span>
                        <span className="font-mono text-xs truncate max-w-[140px]">{migration?.checksum ?? "—"}</span>
                      </div>
                      <Button
                        className="h-11 gap-2 w-full"
                        disabled={!connectionId || execute.isPending || Boolean(runId)}
                        onClick={startRun}
                      >
                        {execute.isPending || runId ? (
                          <>
                            <Loader2 className="size-4 animate-spin" />
                            Executing…
                          </>
                        ) : (
                          <>
                            <Play className="size-4" />
                            Run Migration
                          </>
                        )}
                      </Button>
                      {execute.isError && (
                        <p className="text-sm text-red-500">{getApiErrorMessage(execute.error)}</p>
                      )}
                    </div>
                  </CardContent>
                </Card>
              </div>
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
