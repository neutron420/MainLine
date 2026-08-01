"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Table2, Columns3, Hash, Link2, GitCompareArrows, GitBranch, Database , MoreHorizontal } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import { schemaProjects, schemaStatusConfig } from "@/lib/schemas-data";

export default function SchemaDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const schemaName = params.schemaId as string;

  const project = schemaProjects.find((p) => p.id === projectId);
  const schema = project?.schemas.find((s) => s.name === schemaName);

  if (!project || !schema) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Schema not found</h2>
            <p className="text-sm text-muted-foreground">The schema you are looking for does not exist.</p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const totalColumns = schema.tables.reduce((acc, t) => acc + t.columns.length, 0);
  const totalIndexes = schema.tables.reduce((acc, t) => acc + t.indexes.length, 0);
  const driftCount = schema.tables.filter((t) => t.status === "drift").length;
  const pendingCount = schema.tables.filter((t) => t.status === "pending").length;

  const stats = [
    { title: "Tables", value: schema.tables.length, icon: Table2 },
    { title: "Columns", value: totalColumns, icon: Columns3 },
    { title: "Indexes", value: totalIndexes, icon: Hash },
    { title: "Relations", value: schema.tables.reduce((acc, t) => acc + t.relations.length, 0), icon: Link2 },
  ];

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
              <BreadcrumbItem><BreadcrumbLink href={`/projects/${project.id}`}>{project.name}</BreadcrumbLink></BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem><BreadcrumbPage>{schema.name}</BreadcrumbPage></BreadcrumbItem>
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
              <Link href={`/projects/${project.id}`}>
                <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                  <ArrowLeft className="size-4" />
                </Button>
              </Link>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 flex-wrap">
                  <h1 className="text-2xl font-semibold tracking-tight font-mono">{schema.name}</h1>
                  <Badge variant="outline" className="text-[11px]">{project.name}</Badge>
                  <Badge variant="secondary" className="text-[11px]">{schema.tables.length} tables</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">
                  {driftCount > 0 && <span className="text-red-500 font-medium">{driftCount} with drift · </span>}
                  {pendingCount > 0 && <span className="text-yellow-500 font-medium">{pendingCount} pending review · </span>}
                  Schema structure overview
                </p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Link href={`/projects/${projectId}/schemas/${schema.name}/erd`}>
                  <Button variant="outline" className="h-10 gap-2">
                    <GitBranch className="size-4" />
                    View ERD
                  </Button>
                </Link>
                <Link href={`/projects/${projectId}/schemas/${schema.name}/compare`}>
                  <Button className="h-10 gap-2">
                    <GitCompareArrows className="size-4" />
                    Compare
                  </Button>
                </Link>
              </div>
            </div>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
              <Card key={stat.title}>
                <CardContent className="flex items-center gap-4 p-5">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-lg bg-muted">
                    <stat.icon className="size-5 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-2xl font-semibold tracking-tight">{stat.value}</p>
                    <p className="text-xs text-muted-foreground">{stat.title}</p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Tables */}
          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Tables</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-[30%]">Table</TableHead>
                    <TableHead>Columns</TableHead>
                    <TableHead>Indexes</TableHead>
                    <TableHead>Relations</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead className="w-[50px] text-right"></TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {schema.tables.map((table) => {
                    const status = schemaStatusConfig[table.status];
                    return (
                      <TableRow key={table.id}>
                        <TableCell>
                          <div className="flex items-center gap-2.5">
                            <div className={`size-1.5 rounded-full ${status.dot}`} />
                            <span className="font-mono text-sm">{table.name}</span>
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">{table.columns.length}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{table.indexes.length}</TableCell>
                        <TableCell className="text-sm text-muted-foreground">{table.relations.length}</TableCell>
                        <TableCell>
                          <Badge variant={status.badge} className="text-[10px] px-1.5 py-0">{status.label}</Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                              <Button variant="ghost" size="icon" className="size-8">
                                <MoreHorizontal className="size-4" />
                              </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent align="end">
                              <Link href={`/projects/${projectId}/schemas/${schema.name}/erd`}>
                                <DropdownMenuItem className="gap-2">
                                  <GitBranch className="size-4" />
                                  View in ERD
                                </DropdownMenuItem>
                              </Link>
                              <Link href={`/projects/${projectId}/schemas/${schema.name}/compare`}>
                                <DropdownMenuItem className="gap-2">
                                  <GitCompareArrows className="size-4" />
                                  Compare
                                </DropdownMenuItem>
                              </Link>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </TableCell>
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
