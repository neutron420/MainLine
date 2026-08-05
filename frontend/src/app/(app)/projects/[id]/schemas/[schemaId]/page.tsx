"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, Table2, Columns3, Hash, Link2, GitCompareArrows, GitBranch, Database, Loader2, RefreshCw } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
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
} from "@/components/ui/sidebar";
import {
  useSchema,
  useSchemaObjects,
  useIntrospectSchema,
} from "@/lib/api/hooks/use-schemas";
import { getApiErrorMessage } from "@/lib/api/errors";

function relativeTime(iso: string | undefined): string {
  if (!iso) return "Never";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "Never";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return `${Math.floor(hours / 24)}d ago`;
}

const typeConfig: Record<string, { label: string; badge: "default" | "secondary" | "destructive" | "outline" }> = {
  table: { label: "Table", badge: "default" },
  view: { label: "View", badge: "secondary" },
  index: { label: "Index", badge: "outline" },
  sequence: { label: "Sequence", badge: "outline" },
  function: { label: "Function", badge: "outline" },
  trigger: { label: "Trigger", badge: "outline" },
  constraint: { label: "Constraint", badge: "outline" },
};

export default function SchemaDetailPage() {
  const params = useParams();
  const projectId = params.id as string;
  const schemaId = params.schemaId as string;

  const { data: schema, isLoading, error } = useSchema(projectId, schemaId);
  const { data: objects } = useSchemaObjects(schema?.currentVersionId);
  const introspect = useIntrospectSchema();
  const [search, setSearch] = useState("");

  const filteredObjects = (objects ?? []).filter((o) => {
    if (search === "") return true;
    const haystack = `${o.objectName} ${o.objectType} ${o.objectSchema ?? ""} ${o.parentObjectId ?? ""}`.toLowerCase();
    return haystack.includes(search.toLowerCase());
  });

  if (isLoading) {
    return (
                <div className="flex h-full flex-col items-center justify-center gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">Loading schema...</p>
          </div>
    );
  }

  if (!schema) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <Database className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Schema not found</h2>
            <p className="text-sm text-muted-foreground">
              {error ? getApiErrorMessage(error) : "The schema you are looking for does not exist."}
            </p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const tables = (objects ?? []).filter((o) => o.objectType === "table");
  const totalObjects = objects?.length ?? 0;

  const stats = [
    { title: "Tables", value: tables.length, icon: Table2 },
    { title: "Objects", value: totalObjects, icon: Columns3 },
    { title: "Version", value: schema.currentVersionId?.slice(0, 8) || "—", icon: Hash },
    { title: "Last Updated", value: relativeTime(schema.updatedAt), icon: Link2 },
  ];

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search objects..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
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
                  <h1 className="text-2xl font-semibold tracking-tight font-mono">{schema.schemaName}</h1>
                  <Badge variant="outline" className="text-[11px]">{schema.connectionId.slice(0, 8)}</Badge>
                  <Badge variant="secondary" className="text-[11px]">{tables.length} tables</Badge>
                </div>
                <p className="text-sm text-muted-foreground mt-1">Schema structure overview</p>
              </div>
              <div className="flex items-center gap-2 shrink-0">
                <Button
                  variant="outline"
                  className="h-10 gap-2"
                  disabled={introspect.isPending || !schema.connectionId}
                  onClick={() => introspect.mutate({ connectionId: schema.connectionId, schemaNames: [schema.schemaName] })}
                >
                  {introspect.isPending ? <Loader2 className="size-4 animate-spin" /> : <RefreshCw className="size-4" />}
                  {introspect.isPending ? "Introspecting…" : "Re-introspect"}
                </Button>
                <Link href={`/projects/${projectId}/schemas/${schemaId}/erd`}>
                  <Button variant="outline" className="h-10 gap-2">
                    <GitBranch className="size-4" />
                    View ERD
                  </Button>
                </Link>
                <Link href={`/projects/${projectId}/schemas/${schemaId}/compare`}>
                  <Button className="h-10 gap-2">
                    <GitCompareArrows className="size-4" />
                    Compare
                  </Button>
                </Link>
              </div>
            </div>
            {introspect.isSuccess && (
              <p className="text-sm text-green-500 flex items-center gap-1.5">
                <RefreshCw className="size-3.5" />
                Schema re-introspected successfully. Refresh the page to see the latest structure.
              </p>
            )}
            {introspect.isError && (
              <p className="text-sm text-red-500">{getApiErrorMessage(introspect.error)}</p>
            )}
          </div>

          {/* Stats */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
              <Card key={stat.title}>
                <CardContent className="flex items-center gap-4 p-5">
                  <div className="flex size-11 shrink-0 items-center justify-center rounded-lg bg-muted">
                    <stat.icon className="size-5 text-muted-foreground" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-2xl font-semibold tracking-tight truncate">{stat.value}</p>
                    <p className="text-xs text-muted-foreground">{stat.title}</p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {/* Objects */}
          <Card>
            <CardHeader className="border-0">
              <CardTitle className="text-base">Schema Objects</CardTitle>
            </CardHeader>
            <CardContent className="pt-0">
              {!objects || objects.length === 0 ? (
                <p className="text-muted-foreground text-sm py-6 text-center">
                  No objects in this schema version yet. Re-introspect the connection to capture the structure.
                </p>
              ) : filteredObjects.length === 0 ? (
                <p className="text-muted-foreground text-sm py-6 text-center">
                  No objects match your search.
                </p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead className="w-[30%]">Name</TableHead>
                      <TableHead>Schema</TableHead>
                      <TableHead>Type</TableHead>
                      <TableHead>Parent</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredObjects.map((obj) => {
                      const type = typeConfig[obj.objectType] ?? { label: obj.objectType, badge: "outline" as const };
                      return (
                        <TableRow key={obj.id}>
                          <TableCell>
                            <div className="flex items-center gap-2.5">
                              <div className="size-1.5 rounded-full bg-primary/70" />
                              <span className="font-mono text-sm">{obj.objectName}</span>
                            </div>
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground font-mono">{obj.objectSchema}</TableCell>
                          <TableCell>
                            <Badge variant={type.badge} className="text-[10px] px-1.5 py-0">{type.label}</Badge>
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground font-mono truncate max-w-[200px]">
                            {obj.parentObjectId ? obj.parentObjectId.slice(0, 8) : "—"}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
