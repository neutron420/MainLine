"use client";

import { useState, useMemo } from "react";
import {
  Search, Plus, ChevronRight, Database, Table2, FolderKanban,
  KeyRound, Link2, ShieldCheck, Columns3, Copy, Check,
  Loader2, RefreshCw,
} from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb";
import { NotificationsPopover } from "@/components/notifications-popover";
import { useProjects } from "@/lib/api/hooks/use-projects";
import { useSchemas, useSchemaDiagram, useIntrospectSchema } from "@/lib/api/hooks/use-schemas";
import { useConnections } from "@/lib/api/hooks/use-connections";
import type { ColumnInfo } from "@/lib/gen/schema/v1/schema_messages_pb";
import { getApiErrorMessage } from "@/lib/api/errors";

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

function generateTableDDL(schemaName: string, label: string, columns: ColumnInfo[]): string {
  const lines: string[] = [];
  lines.push(`CREATE TABLE ${schemaName}.${label} (`);
  columns.forEach((column, i) => {
    const parts = [`  ${column.name} ${column.type}`];
    if (column.isPk) parts.push("PRIMARY KEY");
    if (!column.nullable) parts.push("NOT NULL");
    if (column.default) parts.push(`DEFAULT ${column.default}`);
    lines.push(parts.join(" ") + (i < columns.length - 1 ? "," : ""));
  });
  lines.push(");");
  return lines.join("\n");
}

function ProjectSchemas({
  projectId,
  projectName,
  onSelectSchema,
  selectedSchemaId,
}: {
  projectId: string;
  projectName: string;
  onSelectSchema: (projectId: string, schemaId: string) => void;
  selectedSchemaId: string | null;
}) {
  const { data: schemas } = useSchemas(projectId);
  const [open, setOpen] = useState(true);

  return (
    <div key={projectId}>
      <button
        onClick={() => setOpen((o) => !o)}
        className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 hover:bg-muted text-left"
      >
        <ChevronRight className={`size-3.5 text-muted-foreground transition-transform ${open ? "rotate-90" : ""}`} />
        <FolderKanban className="size-4 text-muted-foreground" />
        <span className="font-medium truncate">{projectName}</span>
        <span className="ml-auto text-xs text-muted-foreground">{schemas?.length ?? 0}</span>
      </button>
      {open && (
        <div className="ml-4 border-l pl-2">
          {!schemas || schemas.length === 0 ? (
            <p className="px-2 py-1 text-xs text-muted-foreground">No schemas</p>
          ) : (
            schemas.map((schema) => (
              <button
                key={schema.id}
                onClick={() => onSelectSchema(projectId, schema.id)}
                className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left ${
                  selectedSchemaId === schema.id ? "bg-primary/10 text-primary" : "hover:bg-muted"
                }`}
              >
                <Database className="size-4 text-muted-foreground" />
                <span className="font-mono text-sm truncate">{schema.schemaName}</span>
                <span className="ml-auto text-xs text-muted-foreground">{schema.currentVersionId.slice(0, 6)}</span>
              </button>
            ))
          )}
        </div>
      )}
    </div>
  );
}

export default function SchemasPage() {
  const { data: projects, isLoading } = useProjects();
  const introspect = useIntrospectSchema();

  const [selection, setSelection] = useState<{ projectId: string; schemaId: string } | null>(null);
  const [search, setSearch] = useState("");

  const { data: projectSchemas } = useSchemas(selection?.projectId);
  const selectedSchema =
    projectSchemas?.find((s) => s.id === selection?.schemaId) ?? null;

  const { data: diagram, isLoading: diagramLoading } = useSchemaDiagram(
    selectedSchema?.currentVersionId,
  );

  const [introspectOpen, setIntrospectOpen] = useState(false);
  const [projectId, setProjectId] = useState("");
  const [connectionId, setConnectionId] = useState("");
  const [schemaNames, setSchemaNames] = useState("");
  const { data: connections } = useConnections(projectId || undefined);

  const doIntrospect = () => {
    if (!connectionId || !schemaNames.trim() || introspect.isPending) return;
    introspect.mutate(
      {
        connectionId,
        schemaNames: schemaNames.split(",").map((s) => s.trim()).filter(Boolean),
      },
      {
        onSuccess: (result) => {
          setIntrospectOpen(false);
          setSchemaNames("");
          setSelection({ projectId, schemaId: result.schema.id });
        },
      },
    );
  };

  const tables = useMemo(
    () =>
      (diagram?.nodes ?? []).map((node) => ({
        id: node.id,
        name: (node.data as { label?: string } | undefined)?.label ?? node.id,
        columns: (node.data as { columns?: ColumnInfo[] } | undefined)?.columns ?? [],
      })),
    [diagram],
  );

  const q = search.trim().toLowerCase();
  const filteredTables = q ? tables.filter((t) => t.name.toLowerCase().includes(q)) : tables;

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">
          <SidebarTrigger className="-ml-1 size-9" />
          <Separator orientation="vertical" className="mr-2 data-[orientation=vertical]:h-5" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem><BreadcrumbPage>Schemas</BreadcrumbPage></BreadcrumbItem>
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
          {/* Header + Introspect */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Schemas</h1>
              <p className="text-sm text-muted-foreground mt-1">Explore database schemas across projects</p>
            </div>
            <Dialog open={introspectOpen} onOpenChange={setIntrospectOpen}>
              <DialogTrigger asChild>
                <Button className="h-11 gap-2">
                  <Plus className="size-4" />
                  Track Schema
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <DialogHeader>
                  <DialogTitle>Track a Schema</DialogTitle>
                  <DialogDescription>
                    Introspect a connection to create a tracked schema version.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="proj">Project</Label>
                    <Select value={projectId} onValueChange={setProjectId}>
                      <SelectTrigger id="proj" className="h-11">
                        <SelectValue placeholder="Select project" />
                      </SelectTrigger>
                      <SelectContent>
                        {(projects ?? []).map((p) => (
                          <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="conn">Connection</Label>
                    <Select value={connectionId} onValueChange={setConnectionId} disabled={!projectId}>
                      <SelectTrigger id="conn" className="h-11">
                        <SelectValue placeholder={projectId ? "Select connection" : "Select a project first"} />
                      </SelectTrigger>
                      <SelectContent>
                        {(connections ?? []).map((c) => (
                          <SelectItem key={c.id} value={c.id}>
                            {c.name} ({c.databaseName})
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="sn">Schema names (comma separated)</Label>
                    <Input
                      id="sn"
                      placeholder="public, auth"
                      className="h-11 font-mono"
                      value={schemaNames}
                      onChange={(e) => setSchemaNames(e.target.value)}
                    />
                  </div>
                  {introspect.isError && (
                    <p className="text-sm text-red-500">{getApiErrorMessage(introspect.error)}</p>
                  )}
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setIntrospectOpen(false)}>Cancel</Button>
                  <Button
                    onClick={doIntrospect}
                    disabled={!connectionId || !schemaNames.trim() || introspect.isPending}
                  >
                    {introspect.isPending ? (
                      <>
                        <Loader2 className="size-4 animate-spin" />
                        Introspecting…
                      </>
                    ) : (
                      "Introspect"
                    )}
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Search */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input
                placeholder="Search tables..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 h-11"
              />
            </div>
          </div>

          {/* Explorer */}
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[340px_1fr]">
            {/* Tree */}
            <Card className="h-fit">
              <CardHeader className="border-0 pb-2">
                <CardTitle className="text-base">Explorer</CardTitle>
              </CardHeader>
              <CardContent className="pt-0">
                <div className="max-h-[520px] overflow-y-auto pr-1 flex flex-col text-sm">
                  {isLoading ? (
                    <p className="px-2 py-2 text-xs text-muted-foreground">Loading projects...</p>
                  ) : !projects || projects.length === 0 ? (
                    <p className="px-2 py-2 text-xs text-muted-foreground">No projects yet.</p>
                  ) : (
                    projects.map((project) => (
                      <ProjectSchemas
                        key={project.id}
                        projectId={project.id}
                        projectName={project.name}
                        selectedSchemaId={selection?.schemaId ?? null}
                        onSelectSchema={(pid, sid) => setSelection({ projectId: pid, schemaId: sid })}
                      />
                    ))
                  )}
                </div>
                <Separator className="my-4" />
                <div className="flex items-center gap-4 px-1 pb-1">
                  <span className="flex items-center gap-1.5 text-xs text-green-500 font-medium">
                    <ShieldCheck className="size-3.5" />
                    Schemas tracked per project
                  </span>
                </div>
              </CardContent>
            </Card>

            {/* Detail */}
            {selection && selectedSchema ? (
              <Card className="h-fit">
                <CardHeader className="border-0">
                  <div className="flex flex-wrap items-center gap-3">
                    <CardTitle className="text-lg flex items-center gap-2">
                      <Database className="size-4 text-muted-foreground" />
                      <span className="font-mono">{selectedSchema.schemaName}</span>
                    </CardTitle>
                    <Badge variant="outline" className="text-[10px] px-1.5 py-0">
                      {selectedSchema.currentVersionId.slice(0, 8)}
                    </Badge>
                    <Button
                      variant="outline"
                      size="sm"
                      className="h-7 gap-1.5 text-xs ml-auto"
                      disabled={introspect.isPending}
                      onClick={() => {
                        if (selectedSchema.connectionId) {
                          introspect.mutate({
                            connectionId: selectedSchema.connectionId,
                            schemaNames: [selectedSchema.schemaName],
                          });
                        }
                      }}
                    >
                      <RefreshCw className={`size-3 ${introspect.isPending ? "animate-spin" : ""}`} />
                      Re-introspect
                    </Button>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    Connection {selectedSchema.connectionId.slice(0, 8)} · last introspected{" "}
                    {selectedSchema.updatedAt
                      ? new Date(selectedSchema.updatedAt).toLocaleString()
                      : "—"}
                  </p>
                  {introspect.isSuccess && (
                    <p className="text-sm text-green-500">Schema re-introspected successfully.</p>
                  )}
                  {introspect.isError && (
                    <p className="text-sm text-red-500">{getApiErrorMessage(introspect.error)}</p>
                  )}
                </CardHeader>
                <CardContent className="pt-0">
                  {diagramLoading ? (
                    <p className="text-sm text-muted-foreground py-8 text-center">Loading schema structure...</p>
                  ) : filteredTables.length === 0 ? (
                    <p className="text-sm text-muted-foreground py-8 text-center">
                      No tables {q ? "matching the search" : "in this schema version"}.
                    </p>
                  ) : (
                    <Tabs defaultValue="tables">
                      <TabsList variant="line" className="w-full justify-start gap-4">
                        <TabsTrigger value="tables" className="gap-1.5">
                          <Columns3 className="size-4" />
                          Tables
                          <span className="text-xs text-muted-foreground">{filteredTables.length}</span>
                        </TabsTrigger>
                        <TabsTrigger value="relations" className="gap-1.5">
                          <Link2 className="size-4" />
                          Relations
                          <span className="text-xs text-muted-foreground">{diagram?.edges.length ?? 0}</span>
                        </TabsTrigger>
                      </TabsList>

                      <TabsContent value="tables" className="mt-4 space-y-6">
                        {filteredTables.map((table) => (
                          <div key={table.id} className="rounded-md border">
                            <div className="flex items-center justify-between border-b bg-muted/50 px-3 py-2">
                              <span className="flex items-center gap-2 font-mono text-xs font-semibold">
                                <Table2 className="size-3.5 text-muted-foreground" />
                                {table.name}
                              </span>
                              <div className="flex items-center gap-2">
                                <span className="text-[10px] text-muted-foreground">{table.columns.length} columns</span>
                                <CopyButton text={generateTableDDL(selectedSchema.schemaName, table.name, table.columns)} />
                              </div>
                            </div>
                            <Table>
                              <TableHeader>
                                <TableRow>
                                  <TableHead className="w-[30%]">Column</TableHead>
                                  <TableHead className="w-[25%]">Type</TableHead>
                                  <TableHead>Key</TableHead>
                                  <TableHead>Nullable</TableHead>
                                  <TableHead className="text-right">Default</TableHead>
                                </TableRow>
                              </TableHeader>
                              <TableBody>
                                {table.columns.length === 0 ? (
                                  <TableRow>
                                    <TableCell colSpan={5} className="text-center text-xs text-muted-foreground py-4">
                                      No columns
                                    </TableCell>
                                  </TableRow>
                                ) : (
                                  table.columns.map((column) => (
                                    <TableRow key={column.name}>
                                      <TableCell>
                                        <div className="flex items-center gap-2">
                                          {column.isPk && <KeyRound className="size-3 text-amber-500 shrink-0" />}
                                          {column.isFk && <Link2 className="size-3 text-blue-500 shrink-0" />}
                                          <span className="font-mono text-sm">{column.name}</span>
                                        </div>
                                      </TableCell>
                                      <TableCell>
                                        <span className="font-mono text-xs text-muted-foreground">{column.type}</span>
                                      </TableCell>
                                      <TableCell>
                                        {column.isPk && (
                                          <Badge variant="outline" className="text-[10px] px-1.5 py-0">PK</Badge>
                                        )}
                                        {column.isFk && (
                                          <Badge variant="outline" className="text-[10px] px-1.5 py-0">FK</Badge>
                                        )}
                                      </TableCell>
                                      <TableCell className="text-xs text-muted-foreground">
                                        {column.nullable ? "YES" : "NO"}
                                      </TableCell>
                                      <TableCell className="text-right font-mono text-xs text-muted-foreground">
                                        {column.default ?? "—"}
                                      </TableCell>
                                    </TableRow>
                                  ))
                                )}
                              </TableBody>
                            </Table>
                          </div>
                        ))}
                      </TabsContent>

                      <TabsContent value="relations" className="mt-4">
                        {!diagram?.edges || diagram.edges.length === 0 ? (
                          <div className="flex flex-col items-center justify-center py-12 text-center">
                            <Link2 className="size-10 text-muted-foreground/40 mb-3" />
                            <h3 className="text-sm font-medium">No relations</h3>
                            <p className="text-xs text-muted-foreground mt-1">This schema has no foreign keys</p>
                          </div>
                        ) : (
                          <Table>
                            <TableHeader>
                              <TableRow>
                                <TableHead className="w-[35%]">From</TableHead>
                                <TableHead className="w-[15%]"></TableHead>
                                <TableHead className="w-[35%]">To</TableHead>
                                <TableHead className="text-right">Label</TableHead>
                              </TableRow>
                            </TableHeader>
                            <TableBody>
                              {diagram.edges.map((edge) => (
                                <TableRow key={edge.id}>
                                  <TableCell>
                                    <span className="font-mono text-xs text-blue-500">{edge.source}</span>
                                    <span className="text-xs text-muted-foreground">.{edge.sourceHandle}</span>
                                  </TableCell>
                                  <TableCell className="text-center text-muted-foreground">→</TableCell>
                                  <TableCell>
                                    <span className="font-mono text-xs">{edge.target}</span>
                                  </TableCell>
                                  <TableCell className="text-right">
                                    <span className="font-mono text-[10px] text-muted-foreground">{edge.label}</span>
                                  </TableCell>
                                </TableRow>
                              ))}
                            </TableBody>
                          </Table>
                        )}
                      </TabsContent>
                    </Tabs>
                  )}
                </CardContent>
              </Card>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-20 text-center">
                  <Database className="size-12 text-muted-foreground/40 mb-4" />
                  <h3 className="text-lg font-medium">Select a schema</h3>
                  <p className="text-sm text-muted-foreground mt-1">Choose a schema from the explorer to see its structure</p>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
