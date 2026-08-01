"use client";

import { useMemo, useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Search, Columns3, Hash, Link2, GitCompareArrows, KeyRound } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
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
import { schemaProjects, type SchemaTable, type SchemaColumn, type SchemaIndex, type SchemaRelation } from "@/lib/schemas-data";

type DiffState = "same" | "added" | "removed" | "modified";

const diffRowClass: Record<DiffState, string> = {
  same: "",
  added: "bg-green-500/10",
  removed: "bg-red-500/10",
  modified: "bg-amber-500/10",
};

const diffSignClass: Record<DiffState, string> = {
  same: "",
  added: "text-green-600",
  removed: "text-red-600",
  modified: "text-amber-600",
};

const diffSign: Record<DiffState, string> = {
  same: "",
  added: "+",
  removed: "−",
  modified: "~",
};

interface ColumnDiffRow {
  base?: SchemaColumn;
  other?: SchemaColumn;
  state: DiffState;
}

interface IndexDiffRow {
  base?: SchemaIndex;
  other?: SchemaIndex;
  state: DiffState;
}

interface RelationDiffRow {
  base?: SchemaRelation;
  other?: SchemaRelation;
  state: DiffState;
}

function diffColumns(base: SchemaTable, other: SchemaTable): ColumnDiffRow[] {
  const rows: ColumnDiffRow[] = [];
  const otherByName = new Map(other.columns.map((c) => [c.name, c]));
  for (const column of base.columns) {
    const oc = otherByName.get(column.name);
    if (!oc) {
      rows.push({ base: column, state: "removed" });
    } else if (oc.type !== column.type || oc.nullable !== column.nullable || oc.default !== column.default || (oc.key ?? null) !== (column.key ?? null)) {
      rows.push({ base: column, other: oc, state: "modified" });
    } else {
      rows.push({ base: column, other: oc, state: "same" });
    }
    otherByName.delete(column.name);
  }
  for (const [, oc] of otherByName) rows.push({ other: oc, state: "added" });
  return rows;
}

function diffIndexes(base: SchemaTable, other: SchemaTable): IndexDiffRow[] {
  const rows: IndexDiffRow[] = [];
  const otherByName = new Map(other.indexes.map((i) => [i.name, i]));
  for (const index of base.indexes) {
    const oi = otherByName.get(index.name);
    if (!oi) {
      rows.push({ base: index, state: "removed" });
    } else if (oi.unique !== index.unique || oi.type !== index.type || oi.columns.join() !== index.columns.join()) {
      rows.push({ base: index, other: oi, state: "modified" });
    } else {
      rows.push({ base: index, other: oi, state: "same" });
    }
    otherByName.delete(index.name);
  }
  for (const [, oi] of otherByName) rows.push({ other: oi, state: "added" });
  return rows;
}

function diffRelations(base: SchemaTable, other: SchemaTable): RelationDiffRow[] {
  const rows: RelationDiffRow[] = [];
  const otherByName = new Map(other.relations.map((r) => [r.name, r]));
  for (const relation of base.relations) {
    const or = otherByName.get(relation.name);
    if (!or) {
      rows.push({ base: relation, state: "removed" });
    } else if (or.from !== relation.from || or.to !== relation.to) {
      rows.push({ base: relation, other: or, state: "modified" });
    } else {
      rows.push({ base: relation, other: or, state: "same" });
    }
    otherByName.delete(relation.name);
  }
  for (const [, or] of otherByName) rows.push({ other: or, state: "added" });
  return rows;
}

function countDiff(rows: { state: DiffState }[]) {
  let added = 0;
  let removed = 0;
  let modified = 0;
  for (const row of rows) {
    if (row.state === "added") added++;
    else if (row.state === "removed") removed++;
    else if (row.state === "modified") modified++;
  }
  return { added, removed, modified };
}

function ColumnCell({ column, state }: { column?: SchemaColumn; state: DiffState }) {
  return (
    <div className={`flex h-9 items-center gap-2 px-3 border-b last:border-b-0 ${diffRowClass[state]}`}>
      <span className={`w-3 shrink-0 font-mono text-xs ${diffSignClass[state]}`}>{diffSign[state]}</span>
      {column ? (
        <>
          {column.key === "PK" ? (
            <KeyRound className="size-3 shrink-0 text-amber-500" />
          ) : column.key === "FK" ? (
            <Link2 className="size-3 shrink-0 text-blue-500" />
          ) : (
            <span className="size-3 shrink-0" />
          )}
          <span className="font-mono text-xs truncate">{column.name}</span>
          <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{column.type}</span>
        </>
      ) : (
        <span className="sr-only">empty</span>
      )}
    </div>
  );
}

export default function SchemaComparePage() {
  const params = useParams();
  const projectId = params.id as string;
  const schemaName = params.schemaId as string;

  const project = schemaProjects.find((p) => p.id === projectId);
  const schema = project?.schemas.find((s) => s.name === schemaName);

  const [baseId, setBaseId] = useState(schema?.tables[0]?.id ?? "");
  const [otherId, setOtherId] = useState(schema?.tables[1]?.id ?? "");

  const base = schema?.tables.find((t) => t.id === baseId);
  const other = schema?.tables.find((t) => t.id === otherId);

  const columnRows = useMemo(() => (base && other ? diffColumns(base, other) : []), [base, other]);
  const indexRows = useMemo(() => (base && other ? diffIndexes(base, other) : []), [base, other]);
  const relationRows = useMemo(() => (base && other ? diffRelations(base, other) : []), [base, other]);

  const columnCounts = countDiff(columnRows);
  const indexCounts = countDiff(indexRows);
  const relationCounts = countDiff(relationRows);

  if (!project || !schema || (schema.tables.length < 2) || !base || !other) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <GitCompareArrows className="size-12 text-muted-foreground/40" />
            <h2 className="text-xl font-semibold">Cannot compare</h2>
            <p className="text-sm text-muted-foreground">This schema needs at least two tables to compare.</p>
            <Link href={`/projects/${projectId}`}>
              <Button variant="outline">Back to Project</Button>
            </Link>
          </div>
        </SidebarInset>
      </SidebarProvider>
    );
  }

  const tabsConfig = [
    { value: "columns", label: "Columns", icon: Columns3, counts: columnCounts },
    { value: "indexes", label: "Indexes", icon: Hash, counts: indexCounts },
    { value: "relations", label: "Relations", icon: Link2, counts: relationCounts },
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
              <BreadcrumbItem><BreadcrumbPage>Compare · {schema.name}</BreadcrumbPage></BreadcrumbItem>
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
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Schema Compare</h1>
              <p className="text-sm text-muted-foreground mt-1">Compare two tables in {schema.name} · {project.name}</p>
            </div>
          </div>

          {/* Table selectors */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="flex items-center gap-2.5 flex-wrap">
              <Select value={baseId} onValueChange={setBaseId}>
                <SelectTrigger className="h-11 w-64">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {schema.tables.map((t) => (
                    <SelectItem key={t.id} value={t.id}>{schema.name}.{t.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <GitCompareArrows className="size-4 text-muted-foreground" />
              <Select value={otherId} onValueChange={setOtherId}>
                <SelectTrigger className="h-11 w-64">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {schema.tables.map((t) => (
                    <SelectItem key={t.id} value={t.id}>{schema.name}.{t.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Diff */}
          <Card>
            <CardHeader className="border-0">
              <div className="flex items-center justify-between gap-4 flex-wrap">
                <CardTitle className="text-base">Diff</CardTitle>
                <div className="flex items-center gap-2">
                  <Badge variant="default" className="gap-1 text-[10px] px-1.5 py-0">
                    <span className="text-green-300">+</span>{columnCounts.added + indexCounts.added + relationCounts.added} added
                  </Badge>
                  <Badge variant="destructive" className="gap-1 text-[10px] px-1.5 py-0">
                    <span>−</span>{columnCounts.removed + indexCounts.removed + relationCounts.removed} removed
                  </Badge>
                  <Badge variant="secondary" className="gap-1 text-[10px] px-1.5 py-0">
                    <span className="text-amber-400">~</span>{columnCounts.modified + indexCounts.modified + relationCounts.modified} modified
                  </Badge>
                </div>
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              <Tabs defaultValue="columns">
                <TabsList variant="line" className="w-full justify-start gap-4">
                  {tabsConfig.map((tab) => (
                    <TabsTrigger key={tab.value} value={tab.value} className="gap-1.5">
                      <tab.icon className="size-4" />
                      {tab.label}
                    </TabsTrigger>
                  ))}
                </TabsList>

                <TabsContent value="columns" className="mt-4">
                  <div className="grid grid-cols-2 divide-x rounded-md border overflow-hidden">
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{base.name}</span>
                        <span className="ml-auto font-mono text-[10px] text-muted-foreground">{base.columns.length} columns</span>
                      </div>
                      {columnRows.map((row, i) => (
                        <ColumnCell key={i} column={row.base} state={row.state} />
                      ))}
                    </div>
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{other.name}</span>
                        <span className="ml-auto font-mono text-[10px] text-muted-foreground">{other.columns.length} columns</span>
                      </div>
                      {columnRows.map((row, i) => (
                        <ColumnCell key={i} column={row.other} state={row.state} />
                      ))}
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="indexes" className="mt-4">
                  <div className="grid grid-cols-2 divide-x rounded-md border overflow-hidden">
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{base.name}</span>
                      </div>
                      {indexRows.map((row, i) => (
                        <div key={i} className={`flex h-9 items-center gap-2 px-3 border-b last:border-b-0 ${diffRowClass[row.state]}`}>
                          <span className={`w-3 shrink-0 font-mono text-xs ${diffSignClass[row.state]}`}>{diffSign[row.state]}</span>
                          {row.base && (
                            <>
                              <Hash className="size-3 shrink-0 text-muted-foreground" />
                              <span className="font-mono text-xs truncate">{row.base.name}</span>
                              <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{row.base.columns.join(", ")}</span>
                            </>
                          )}
                        </div>
                      ))}
                    </div>
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{other.name}</span>
                      </div>
                      {indexRows.map((row, i) => (
                        <div key={i} className={`flex h-9 items-center gap-2 px-3 border-b last:border-b-0 ${diffRowClass[row.state]}`}>
                          <span className={`w-3 shrink-0 font-mono text-xs ${diffSignClass[row.state]}`}>{diffSign[row.state]}</span>
                          {row.other && (
                            <>
                              <Hash className="size-3 shrink-0 text-muted-foreground" />
                              <span className="font-mono text-xs truncate">{row.other.name}</span>
                              <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{row.other.columns.join(", ")}</span>
                            </>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                </TabsContent>

                <TabsContent value="relations" className="mt-4">
                  <div className="grid grid-cols-2 divide-x rounded-md border overflow-hidden">
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{base.name}</span>
                      </div>
                      {relationRows.map((row, i) => (
                        <div key={i} className={`flex h-9 items-center gap-2 px-3 border-b last:border-b-0 ${diffRowClass[row.state]}`}>
                          <span className={`w-3 shrink-0 font-mono text-xs ${diffSignClass[row.state]}`}>{diffSign[row.state]}</span>
                          {row.base && (
                            <>
                              <Link2 className="size-3 shrink-0 text-muted-foreground" />
                              <span className="font-mono text-xs truncate">{row.base.name}</span>
                              <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{row.base.from} → {row.base.to}</span>
                            </>
                          )}
                        </div>
                      ))}
                    </div>
                    <div className="min-w-0">
                      <div className="flex h-9 items-center gap-2 bg-muted/50 px-3 border-b">
                        <span className="font-mono text-xs font-semibold truncate">{other.name}</span>
                      </div>
                      {relationRows.map((row, i) => (
                        <div key={i} className={`flex h-9 items-center gap-2 px-3 border-b last:border-b-0 ${diffRowClass[row.state]}`}>
                          <span className={`w-3 shrink-0 font-mono text-xs ${diffSignClass[row.state]}`}>{diffSign[row.state]}</span>
                          {row.other && (
                            <>
                              <Link2 className="size-3 shrink-0 text-muted-foreground" />
                              <span className="font-mono text-xs truncate">{row.other.name}</span>
                              <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{row.other.from} → {row.other.to}</span>
                            </>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                </TabsContent>
              </Tabs>
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
