"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Copy, Check, FileCode2, GitBranch, Clock, Database, Table2, Play } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
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
import { migrationsData, migrationStatusConfig } from "@/lib/migrations-data";

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
  const migration = migrationsData.find((m) => m.id === migrationId);
  const [activeTab, setActiveTab] = useState("sql");

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
                  <h1 className="text-2xl font-semibold tracking-tight truncate">{migration.name}</h1>
                  <Badge variant={status.badge} className="text-[11px]">{status.label}</Badge>
                </div>
                <div className="flex items-center gap-2 mt-1.5 text-sm text-muted-foreground flex-wrap">
                  <Avatar className="size-5">
                    <AvatarFallback className="text-[8px]">{migration.initials}</AvatarFallback>
                  </Avatar>
                  <span className="text-foreground font-medium">{migration.author}</span>
                  <span>created {migration.created}</span>
                  {migration.applied !== "—" && (
                    <>
                      <span>·</span>
                      <span>{migration.environment}</span>
                      <span>·</span>
                      <span>{migration.applied}</span>
                    </>
                  )}
                </div>
              </div>
              {migration.status !== "deployed" && migration.status !== "rolledBack" && (
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
                      <TabsTrigger value="changes" className="gap-1.5">
                        <GitBranch className="size-4" />
                        Changes
                        <span className="text-xs text-muted-foreground">{migration.changes.length}</span>
                      </TabsTrigger>
                      <TabsTrigger value="timeline" className="gap-1.5">
                        <Clock className="size-4" />
                        Timeline
                      </TabsTrigger>
                    </TabsList>
                  </Tabs>
                </CardHeader>
                <CardContent className="pt-4">
                  {activeTab === "sql" && (
                    <div className="rounded-md border">
                      <div className="flex items-center justify-between border-b px-3 py-2">
                        <span className="font-mono text-xs font-medium text-muted-foreground">
                          {migration.database} · {migration.table}
                        </span>
                        <CopyButton text={migration.sql} />
                      </div>
                      <pre className="bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed">
                        {migration.sql}
                      </pre>
                    </div>
                  )}
                  {activeTab === "changes" && (
                    <div className="flex flex-col gap-2">
                      {migration.changes.map((change, i) => (
                        <div key={i} className="flex items-start gap-3 rounded-md border px-3 py-2.5">
                          <span
                            className={`font-mono text-sm leading-5 ${
                              change.action === "add" ? "text-green-600" :
                              change.action === "remove" ? "text-red-600" : "text-amber-600"
                            }`}
                          >
                            {change.action === "add" ? "+" : change.action === "remove" ? "-" : "~"}
                          </span>
                          <div className="min-w-0">
                            <p className="font-mono text-sm">
                              {change.column}
                              <span className="text-muted-foreground"> {change.type}</span>
                            </p>
                            {change.note && <p className="text-xs text-muted-foreground mt-0.5">{change.note}</p>}
                          </div>
                        </div>
                      ))}
                    </div>
                  )}
                  {activeTab === "timeline" && (
                    <div>
                      {migration.timeline.map((entry, i) => (
                        <div key={i} className="flex items-start gap-3 py-3 border-b last:border-b-0">
                          <div className={`mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-full ${i === migration.timeline.length - 1 && migration.status === "deployed" ? "bg-green-500 text-white" : "bg-muted text-muted-foreground"}`}>
                            {i === migration.timeline.length - 1 && migration.status === "deployed" ? (
                              <Check className="size-4" />
                            ) : (
                              <Clock className="size-4" />
                            )}
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm">
                              <span className="font-medium">{entry.label}</span>
                              <span className="text-muted-foreground"> by {entry.user}</span>
                            </p>
                            <p className="text-xs text-muted-foreground mt-0.5">{entry.time}</p>
                          </div>
                        </div>
                      ))}
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
                        <p className="text-xs text-muted-foreground">Database</p>
                        <p className="font-mono text-xs truncate">{migration.database}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-sm">
                      <Table2 className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Table</p>
                        <p className="font-mono text-xs">{migration.table}</p>
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
                      <Clock className="size-4 text-muted-foreground shrink-0" />
                      <div className="min-w-0">
                        <p className="text-xs text-muted-foreground">Duration</p>
                        <p className="text-sm">{migration.duration}</p>
                      </div>
                    </div>
                    <Separator />
                    <div className="flex items-center justify-between">
                      <span className="text-xs text-muted-foreground">Environment</span>
                      <Badge variant="outline" className="text-[10px] px-1.5 py-0">{migration.environment}</Badge>
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
