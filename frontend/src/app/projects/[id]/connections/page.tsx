"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Plus, Database, Globe, Server, User, Shield, Clock, PlugZap } from "lucide-react";

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
import { connectionStatusConfig, dbConnections } from "@/lib/connections-data";

export default function ConnectionsPage() {
  const params = useParams();
  const projectId = params.id as string;

  const connected = dbConnections.filter((c) => c.status === "connected").length;
  const withError = dbConnections.filter((c) => c.status === "error").length;

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
              <BreadcrumbItem><BreadcrumbPage>Connections</BreadcrumbPage></BreadcrumbItem>
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
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-2xl font-semibold tracking-tight">Connections</h1>
                  <Badge variant="outline" className="text-[11px]">{dbConnections.length} linked</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {connected} connected · {withError > 0 && <span className="text-red-500 font-medium">{withError} failing · </span>}
                  PostgreSQL databases linked to this project
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
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {dbConnections.map((conn) => {
              const status = connectionStatusConfig[conn.status];
              return (
                <Card key={conn.id}>
                  <CardHeader className="border-0 pb-0">
                    <div className="flex items-start justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <div className="flex size-10 items-center justify-center rounded-lg bg-muted">
                          <Database className="size-5 text-muted-foreground" />
                        </div>
                        <div>
                          <CardTitle className="text-base">{conn.name}</CardTitle>
                          <p className="text-xs text-muted-foreground mt-0.5">{conn.version}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Badge variant="outline" className="text-[10px] px-1.5 py-0">{conn.env}</Badge>
                        <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 gap-1">
                          <span className={`size-1.5 rounded-full ${status.dot}`} />
                          {status.label}
                        </Badge>
                      </div>
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
                        <span className="truncate font-mono text-xs">{conn.database}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <User className="size-4 shrink-0" />
                        <span className="truncate font-mono text-xs">{conn.user}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Shield className="size-4 shrink-0" />
                        <span className="text-xs">{conn.ssl === "required" ? "SSL required" : conn.ssl === "prefer" ? "SSL preferred" : "SSL disabled"}</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <PlugZap className="size-4 shrink-0" />
                        <span className="text-xs">{conn.latency} latency</span>
                      </div>
                      <div className="flex items-center gap-2 text-muted-foreground">
                        <Clock className="size-4 shrink-0" />
                        <span className="text-xs">Synced {conn.lastSync}</span>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 mt-4 pt-4 border-t">
                      <Button variant="outline" size="sm" className="h-8 text-xs">Test</Button>
                      <Button variant="outline" size="sm" className="h-8 text-xs">Sync Schema</Button>
                      <Button variant="ghost" size="sm" className="h-8 text-xs text-muted-foreground ml-auto">
                        Edit
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
