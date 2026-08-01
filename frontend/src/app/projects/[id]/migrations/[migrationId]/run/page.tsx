"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Play, Loader2, CheckCircle2, AlertTriangle, GitBranch } from "lucide-react";

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
import { Tooltip } from "@heroui/react";
import { migrationsData, migrationStatusConfig, environments } from "@/lib/migrations-data";

const runSteps = [
  { label: "Acquire advisory lock", detail: "Prevents concurrent DDL on the same table" },
  { label: "Execute migration SQL", detail: "Runs the migration in a single transaction" },
  { label: "Verify schema state", detail: "Confirms columns and indexes match the plan" },
];

export default function MigrationRunPage() {
  const params = useParams();
  const projectId = params.id as string;
  const migrationId = params.migrationId as string;
  const migration = migrationsData.find((m) => m.id === migrationId);

  const [environment, setEnvironment] = useState("Staging");
  const [phase, setPhase] = useState<"idle" | "running" | "done">("idle");
  const [step, setStep] = useState(0);

  if (!migration) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <GitBranch className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Migration not found</h2>
            <p className="text-sm text-muted-foreground">The migration you are looking for does not exist.</p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const status = migrationStatusConfig[migration.status];

  const startRun = () => {
    setPhase("running");
    setStep(0);
    setTimeout(() => setStep(1), 900);
    setTimeout(() => setStep(2), 1800);
    setTimeout(() => setPhase("done"), 2700);
  };

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${projectId}/migrations/${migration.id}`}>{migration.version}</BreadcrumbLink></BreadcrumbItem>
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
              <Link href={`/projects/${projectId}/migrations/${migration.id}`}>
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2.5 flex-wrap">
                  <Badge variant="outline" className="font-mono text-[11px]">{migration.version}</Badge>
                  <h1 className="text-2xl font-semibold tracking-tight truncate">Run Migration</h1>
                  <Badge variant={status.badge} className="text-[11px]">{status.label}</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">{migration.name}</p>
              </div>
            </div>
          </div>

          {migration.status === "deployed" || migration.status === "rolledBack" ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                <CheckCircle2 className="size-12 text-green-500 mb-4" />
                <h2 className="text-lg font-semibold">
                  {migration.status === "deployed" ? "Already deployed" : "Migration was rolled back"}
                </h2>
                <p className="text-sm text-muted-foreground mt-1 max-w-md">
                  {migration.status === "deployed"
                    ? `This migration was applied to ${migration.environment} ${migration.applied}.`
                    : "This migration is no longer active and cannot be re-run."}
                </p>
                <Link href={`/projects/${projectId}/migrations/${migration.id}`} className="mt-6">
                  <Button variant="outline">Back to Migration</Button>
                </Link>
              </CardContent>
            </Card>
          ) : phase === "done" ? (
            <Card>
              <CardContent className="flex flex-col items-center justify-center py-16 text-center">
                <CheckCircle2 className="size-12 text-green-500 mb-4" />
                <h2 className="text-lg font-semibold">Migration executed successfully</h2>
                <p className="text-sm text-muted-foreground mt-1">
                  {migration.version} applied to <span className="font-medium text-foreground">{environment}</span> · {migration.database}
                </p>
                <div className="flex items-center gap-2 mt-6">
                  <Link href={`/projects/${projectId}/migrations/${migration.id}`}>
                    <Button variant="outline">View Migration</Button>
                  </Link>
                  <Link href={`/projects/${projectId}`}>
                    <Button>Back to Project</Button>
                  </Link>
                </div>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
              {/* Left */}
              <div className="space-y-6 lg:col-span-2">
                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base">Pre-flight Checks</CardTitle>
                    <CardDescription>Verified before the migration is executed</CardDescription>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <div className="divide-y">
                      {[
                        { label: "Schema matches approved review", detail: "No drift detected since approval" },
                        { label: "Advisory lock available", detail: "No blocking DDL in progress" },
                        { label: "Rollback script ready", detail: "Inverse migration generated" },
                      ].map((check, i) => (
                        <div key={i} className="flex items-start gap-3 py-3.5">
                          <CheckCircle2 className="size-4 mt-0.5 shrink-0 text-green-500" />
                          <div>
                            <p className="text-sm font-medium">{check.label}</p>
                            <p className="text-xs text-muted-foreground mt-0.5">{check.detail}</p>
                          </div>
                        </div>
                      ))}
                    </div>
                  </CardContent>
                </Card>

                <Card>
                  <CardHeader className="border-0">
                    <CardTitle className="text-base flex items-center gap-2">
                      <GitBranch className="size-4" />
                      Execution Steps
                    </CardTitle>
                  </CardHeader>
                  <CardContent className="pt-0">
                    {runSteps.map((stepItem, i) => (
                      <div key={i} className="flex items-start gap-3 py-3 border-b last:border-b-0">
                        {phase === "running" && step === i ? (
                          <Loader2 className="size-4 mt-0.5 shrink-0 text-primary animate-spin" />
                        ) : phase === "running" && step > i ? (
                          <CheckCircle2 className="size-4 mt-0.5 shrink-0 text-green-500" />
                        ) : (
                          <span className="size-4 mt-0.5 shrink-0 rounded-full border text-[9px] text-muted-foreground flex items-center justify-center">
                            {i + 1}
                          </span>
                        )}
                        <div className={phase === "running" && step === i ? "text-foreground" : ""}>
                          <p className="text-sm font-medium">{stepItem.label}</p>
                          <p className="text-xs text-muted-foreground mt-0.5">{stepItem.detail}</p>
                        </div>
                      </div>
                    ))}
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
                        <Label htmlFor="env">Environment</Label>
                        <Select value={environment} onValueChange={setEnvironment}>
                          <SelectTrigger id="env" className="h-11">
                            <SelectValue />
                          </SelectTrigger>
                          <SelectContent>
                            {environments.map((env) => (
                              <SelectItem key={env} value={env}>{env}</SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </div>
                      {environment === "Production" && (
                        <div className="flex items-start gap-2.5 rounded-md border border-destructive/40 bg-destructive/5 px-3 py-2.5">
                          <AlertTriangle className="size-4 text-destructive shrink-0 mt-0.5" />
                          <div>
                            <p className="text-sm font-medium text-destructive">Production environment</p>
                            <p className="text-xs text-muted-foreground mt-0.5">
                              This action affects live data and is tracked in the audit log. No backup is configured.
                            </p>
                          </div>
                        </div>
                      )}
                      <Separator />
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Migration</span>
                        <span className="font-mono">{migration.version}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Database</span>
                        <span className="font-mono text-xs">{migration.database}</span>
                      </div>
                      <div className="flex items-center justify-between text-sm">
                        <span className="text-muted-foreground">Table</span>
                        <span className="font-mono text-xs">{migration.table}</span>
                      </div>
                      <Button
                        className="h-11 gap-2 w-full"
                        disabled={phase === "running"}
                        onClick={startRun}
                      >
                        {phase === "running" ? (
                          <>
                            <Loader2 className="size-4 animate-spin" />
                            Executing...
                          </>
                        ) : (
                          <>
                            <Play className="size-4" />
                            Run Migration
                          </>
                        )}
                      </Button>
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
