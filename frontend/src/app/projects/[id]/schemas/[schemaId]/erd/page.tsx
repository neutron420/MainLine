"use client";

import { useMemo } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, KeyRound, Link2, Table2 } from "lucide-react";
import {
  ReactFlow,
  Background,
  BackgroundVariant,
  Controls,
  MiniMap,
  Handle,
  Position,
  MarkerType,
  type Node,
  type Edge,
  type NodeProps,
} from "@xyflow/react";
import "@xyflow/react/dist/style.css";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbLink,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { NotificationsPopover } from "@/components/notifications-popover";
import { Tooltip } from "@heroui/react";
import { schemaProjects, schemaStatusConfig, type SchemaTable } from "@/lib/schemas-data";

function TableNode({ data }: NodeProps<Node<{ table: SchemaTable }>>) {
  const table = data.table;
  return (
    <div className="w-64 rounded-lg border border-border bg-card shadow-md">
      <div className="flex items-center justify-between gap-2 rounded-t-lg border-b bg-muted/50 px-3 py-2">
        <span className="flex items-center gap-2 font-mono text-xs font-semibold truncate">
          <Table2 className="size-3.5 text-muted-foreground" />
          {table.name}
        </span>
        <span className={`size-2 shrink-0 rounded-full ${schemaStatusConfig[table.status].dot}`} />
      </div>
      <div className="py-1">
        {table.columns.map((column) => (
          <div key={column.name} className="flex items-center gap-2 px-3 py-1.5 border-b last:border-b-0">
            {column.key === "PK" ? (
              <KeyRound className="size-3 shrink-0 text-amber-500" />
            ) : column.key === "FK" ? (
              <Link2 className="size-3 shrink-0 text-blue-500" />
            ) : (
              <span className="size-3 shrink-0" />
            )}
            <span className="font-mono text-[11px] truncate">{column.name}</span>
            <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{column.type}</span>
            {column.key === "FK" && (
              <Handle
                type="source"
                position={Position.Right}
                id={`fk-${column.name}`}
                className="!size-2 !border-2 !border-background !bg-blue-500"
              />
            )}
          </div>
        ))}
      </div>
      <Handle
        type="target"
        position={Position.Left}
        id="tgt"
        className="!size-2 !border-2 !border-background !bg-muted-foreground"
      />
    </div>
  );
}

export default function ErdPage() {
  const params = useParams();
  const projectId = params.id as string;
  const schemaName = params.schemaId as string;

  const project = schemaProjects.find((p) => p.id === projectId);
  const schema = project?.schemas.find((s) => s.name === schemaName);

  const nodeTypes = useMemo(() => ({ tableNode: TableNode }), []);

  const nodes = useMemo<Node[]>(() => {
    if (!schema) return [];
    return schema.tables.map((table, i) => ({
      id: table.id,
      type: "tableNode",
      position: { x: (i % 2) * 340, y: Math.floor(i / 2) * 220 },
      data: { table },
    }));
  }, [schema]);

  const edges = useMemo<Edge[]>(() => {
    if (!schema) return [];
    return schema.tables.flatMap((table) =>
      table.relations
        .filter((r) => schema.tables.some((t) => t.name === r.to.split(".")[0]))
        .map((r) => {
          const target = schema.tables.find((t) => t.name === r.to.split(".")[0]);
          return {
            id: `${table.id}-${r.name}`,
            source: table.id,
            sourceHandle: `fk-${r.from.split(".")[1]}`,
            target: target!.id,
            targetHandle: "tgt",
            type: "smoothstep",
            markerEnd: { type: MarkerType.ArrowClosed, color: "#94a3b8" },
            style: { stroke: "#94a3b8" },
          };
        })
    );
  }, [schema]);

  if (!project || !schema) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <Table2 className="size-12 text-muted-foreground/40" />
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
              <BreadcrumbItem><BreadcrumbPage>ERD · {schema.name}</BreadcrumbPage></BreadcrumbItem>
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
            <Link href={`/projects/${project.id}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-3 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight">{schema.name} · ERD</h1>
                <Badge variant="outline" className="text-[11px] gap-1">
                  <Table2 className="size-3" />
                  {schema.tables.length} tables
                </Badge>
                <Badge variant="outline" className="text-[11px] gap-1">
                  <Link2 className="size-3" />
                  {edges.length} relations
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Entity relationship diagram for {project.name}</p>
            </div>
          </div>

          {/* Canvas */}
          <div className="relative h-[620px] overflow-hidden rounded-lg border bg-background">
            <ReactFlow
              nodes={nodes}
              edges={edges}
              nodeTypes={nodeTypes}
              fitView
              fitViewOptions={{ padding: 0.3 }}
              minZoom={0.4}
              maxZoom={1.5}
              proOptions={{ hideAttribution: true }}
            >
              <Background variant={BackgroundVariant.Dots} gap={24} size={1} />
              <Controls position="bottom-right" />
              <MiniMap position="bottom-left" pannable zoomable />
            </ReactFlow>
            <div className="absolute left-3 top-3 z-10 flex flex-wrap items-center gap-x-4 gap-y-1.5 rounded-lg border bg-background/95 px-3 py-2 text-xs text-muted-foreground shadow-sm">
              <span className="flex items-center gap-1.5">
                <KeyRound className="size-3.5 text-amber-500" />
                Primary Key
              </span>
              <span className="flex items-center gap-1.5">
                <Link2 className="size-3.5 text-blue-500" />
                Foreign Key
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full bg-green-500" />
                Verified
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full bg-yellow-500" />
                Pending
              </span>
              <span className="flex items-center gap-1.5">
                <span className="size-2 rounded-full bg-red-500" />
                Drift
              </span>
            </div>
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
