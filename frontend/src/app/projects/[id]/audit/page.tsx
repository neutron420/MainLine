"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, FileText, Download } from "lucide-react";

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
import { auditLog, auditActionConfig } from "@/lib/audit-data";

export default function AuditPage() {
  const params = useParams();
  const projectId = params.id as string;

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
              <BreadcrumbItem><BreadcrumbPage>Audit Log</BreadcrumbPage></BreadcrumbItem>
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
                  <FileText className="size-6" />
                  Audit Log
                </h1>
                <Badge variant="outline" className="text-[11px]">{auditLog.length} entries</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Every schema-changing action, permanently recorded</p>
            </div>
            <Button variant="outline" className="h-10 gap-2 shrink-0">
              <Download className="size-4" />
              Export CSV
            </Button>
          </div>

          <Card>
            <CardHeader className="border-0 pb-0">
              <CardTitle className="text-base flex items-center gap-2">
                <FileText className="size-4" />
                Actions
              </CardTitle>
            </CardHeader>
            <CardContent className="pt-4">
              <div className="divide-y">
                {auditLog.map((entry) => {
                  const action = auditActionConfig[entry.action];
                  return (
                    <div key={entry.id} className="flex items-start gap-3 py-4">
                      <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted">
                        <span className="text-xs font-medium">{entry.actor.split(" ").map((n) => n[0]).join("").slice(0, 2)}</span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <p className="text-sm font-medium">{entry.actor}</p>
                          <Badge variant={action.badge} className="text-[10px] px-1.5 py-0">{action.label}</Badge>
                          <span className="text-sm text-muted-foreground">{entry.resource}</span>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">{entry.detail}</p>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">{entry.time}</span>
                    </div>
                  );
                })}
              </div>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
