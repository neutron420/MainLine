"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Plus, Database, Globe, Server, User, Shield, Clock, PlugZap, Trash2, CheckCircle2, XCircle, Loader2 } from "lucide-react";

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
import {
  useConnections,
  useTestConnection,
  useDeleteConnection,
} from "@/lib/api/hooks/use-connections";
import { getApiErrorMessage } from "@/lib/api/errors";

function relativeTime(iso: string | undefined): string {
  if (!iso) return "Never";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "Never";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins}m ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

function sslLabel(sslMode: string): string {
  switch (sslMode.toLowerCase()) {
    case "required":
      return "SSL required";
    case "prefer":
      return "SSL preferred";
    case "disabled":
    case "disable":
    case "":
      return "SSL disabled";
    default:
      return `SSL ${sslMode}`;
  }
}

export default function ConnectionsPage() {
  const params = useParams();
  const projectId = params.id as string;

  const { data: connections, isLoading, error } = useConnections(projectId);
  const testConnection = useTestConnection();
  const deleteConnection = useDeleteConnection(projectId);

  const [search, setSearch] = useState("");
  const filteredConnections = (connections ?? []).filter((c) => {
    if (search === "") return true;
    const haystack = `${c.name} ${c.databaseName} ${c.host} ${c.connectionStatus}`.toLowerCase();
    return haystack.includes(search.toLowerCase());
  });

  const connected = (connections ?? []).filter((c) => c.connectionStatus === "connected").length;
  const withError = (connections ?? []).filter((c) => c.connectionStatus === "error").length;

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
              <BreadcrumbItem><BreadcrumbPage>Connections</BreadcrumbPage></BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="flex items-center gap-2 ml-auto">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search connections..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
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
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-2xl font-semibold tracking-tight">Connections</h1>
                  <Badge variant="outline" className="text-[11px]">{(connections ?? []).length} linked</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {isLoading ? "Loading connections..." : error ? (
                    <span className="text-red-500">{getApiErrorMessage(error)}</span>
                  ) : (
                    <>
                      {connected} connected · {withError > 0 && <span className="text-red-500 font-medium">{withError} failing · </span>}
                      PostgreSQL databases linked to this project
                    </>
                  )}
                </p>
              </div>
              <Link href={`/projects/${projectId}/connections/new`}>
                <Button className="h-10 gap-2 shrink-0">
                  <Plus className="size-4" />
                  New Connection
                </Button>
              </Link>
            </div>
          </div>

          {/* Connection cards */}
          {isLoading ? (
            <p className="text-sm text-muted-foreground">Loading connections...</p>
          ) : !connections || connections.length === 0 ? (
            <Card>
              <CardContent className="py-10 text-center">
                <Database className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                <p className="text-sm text-muted-foreground">
                  No connections yet. Link a PostgreSQL database to start tracking schemas.
                </p>
                <Link href={`/projects/${projectId}/connections/new`} className="inline-block mt-4">
                  <Button variant="outline" className="gap-2">
                    <Plus className="size-4" />
                    Add Connection
                  </Button>
                </Link>
              </CardContent>
            </Card>
          ) : filteredConnections.length === 0 ? (
            <Card>
              <CardContent className="py-10 text-center">
                <Database className="size-10 text-muted-foreground/40 mx-auto mb-3" />
                <p className="text-sm text-muted-foreground">
                  No connections match your search.
                </p>
              </CardContent>
            </Card>
          ) : (
            <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
              {filteredConnections.map((conn) => (
                <Card key={conn.id}>
                  <CardHeader className="border-0 pb-0">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                          <Database className="size-5 text-muted-foreground" />
                        </div>
                        <div>
                          <CardTitle className="text-base">{conn.name}</CardTitle>
                          <p className="text-xs text-muted-foreground mt-0.5">{conn.databaseName}</p>
                        </div>
                      </div>
                      <Badge
                        variant={conn.connectionStatus === "connected" ? "default" : conn.connectionStatus === "error" ? "destructive" : "secondary"}
                        className="text-[10px] px-1.5 py-0 gap-1"
                      >
                        <span
                          className={`size-1.5 rounded-full ${
                            conn.connectionStatus === "connected"
                              ? "bg-green-500"
                              : conn.connectionStatus === "error"
                                ? "bg-red-500"
                                : "bg-muted-foreground/50"
                          }`}
                        />
                        {conn.connectionStatus}
                      </Badge>
                    </div>
                  </CardHeader>
                  <CardContent className="pt-4">
                    <div className="grid grid-cols-2 gap-x-6 gap-y-3 text-sm">
                      <div className="flex items-center gap-2 text-muted-foreground min-w-0">
                        <Globe className="size-4 shrink-0" />
                        <span className="truncate font-mono text-xs">{conn.host}:{conn.port}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Server className="size-4 shrink-0" />
                        <span className="truncate font-mono text-xs">{conn.databaseName}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <User className="size-4 shrink-0" />
                        <span className="truncate font-mono text-xs">{conn.username}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Shield className="size-4 shrink-0" />
                        <span className="text-xs">{sslLabel(conn.sslMode)}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <PlugZap className="size-4 shrink-0" />
                        <span className="text-xs">Last test {relativeTime(conn.lastConnectedAt)}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Clock className="size-4 shrink-0" />
                        <span className="text-xs">Created {relativeTime(conn.createdAt)}</span>
                      </div>
                    </div>
                    {testConnection.isPending && testConnection.variables?.connectionId === conn.id && (
                      <div className="mt-3 flex items-center gap-2 text-xs text-muted-foreground">
                        <Loader2 className="size-3.5 animate-spin" />
                        Testing connection...
                      </div>
                    )}
                    {testConnection.data && testConnection.variables?.connectionId === conn.id && (
                      <div className={`mt-3 flex items-center gap-2 text-xs ${testConnection.data.success ? "text-green-500" : "text-red-500"}`}>
                        {testConnection.data.success ? <CheckCircle2 className="size-3.5" /> : <XCircle className="size-3.5" />}
                        {testConnection.data.success
                          ? `Connected · ${testConnection.data.serverVersion || testConnection.data.databaseName} · ${testConnection.data.latencyMs}ms`
                          : testConnection.data.error || "Connection failed"}
                      </div>
                    )}
                    {testConnection.isError && testConnection.variables?.connectionId === conn.id && (
                      <div className="mt-3 flex items-center gap-2 text-xs text-red-500">
                        <XCircle className="size-3.5" />
                        {getApiErrorMessage(testConnection.error)}
                      </div>
                    )}
                    <div className="flex items-center gap-3 mt-4 pt-4 border-t">
                      <Button
                        variant="outline"
                        size="sm"
                        className="h-8 text-xs gap-1.5"
                        disabled={testConnection.isPending}
                        onClick={() => testConnection.mutate({ projectId, connectionId: conn.id })}
                      >
                        <PlugZap className="size-3" />
                        Test
                      </Button>
                      <Link href={`/projects/${projectId}/schemas`}>
                        <Button variant="outline" size="sm" className="h-8 text-xs">
                          Sync Schema
                        </Button>
                      </Link>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-8 text-xs text-muted-foreground gap-1.5 ml-auto"
                        disabled={deleteConnection.isPending}
                        onClick={() => {
                          if (confirm(`Delete connection "${conn.name}"?`)) {
                            deleteConnection.mutate(conn.id);
                          }
                        }}
                      >
                        <Trash2 className="size-3" />
                        Delete
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
