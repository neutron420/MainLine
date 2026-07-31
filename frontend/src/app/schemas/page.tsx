"use client";

import { useState, useMemo } from "react";
import {
  Search, Plus, ChevronRight, Database, Table2, FolderKanban,
  KeyRound, Link2, ShieldCheck, Clock, TriangleAlert, Columns3, Hash,
  FileCode2, Copy, Check,
} from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuCheckboxItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import { Tooltip } from "@heroui/react";
import { schemaProjects, schemaStatusConfig, projectOptions, type SchemaTable, type SchemaColumn } from "@/lib/schemas-data";

function generateColumnDDL(table: SchemaTable, schema: string, column: SchemaColumn): string {
  const parts = [`ALTER TABLE ${schema}.${table.name} ADD COLUMN ${column.name} ${column.type}`];
  if (!column.nullable) parts.push("NOT NULL");
  if (column.default) parts.push(`DEFAULT ${column.default}`);
  return parts.join(" ") + ";";
}

function generateTableDDL(table: SchemaTable, schema: string): string {
  const lines: string[] = [];
  lines.push(`CREATE TABLE ${schema}.${table.name} (`);
  table.columns.forEach((column, i) => {
    const parts = [`  ${column.name} ${column.type}`];
    if (column.key === "PK") parts.push("PRIMARY KEY");
    if (!column.nullable) parts.push("NOT NULL");
    if (column.default) parts.push(`DEFAULT ${column.default}`);
    lines.push(parts.join(" ") + (i < table.columns.length - 1 ? "," : ""));
  });
  lines.push(");");
  if (table.indexes.length > 0) lines.push("");
  for (const index of table.indexes) {
    lines.push(`CREATE ${index.unique ? "UNIQUE " : ""}INDEX ${index.name} ON ${schema}.${table.name} (${index.columns.join(", ")});`);
  }
  if (table.relations.length > 0) lines.push("");
  for (const relation of table.relations) {
    const fromCol = relation.from.split(".")[1];
    lines.push(`ALTER TABLE ${schema}.${table.name} ADD CONSTRAINT ${relation.name} FOREIGN KEY (${fromCol}) REFERENCES ${relation.to};`);
  }
  return lines.join("\n");
}

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

export default function SchemasPage() {
  const [search, setSearch] = useState("");
  const [projectFilter, setProjectFilter] = useState("all");
  const [expandedProjects, setExpandedProjects] = useState<Set<string>>(new Set(["p1"]));
  const [expandedSchemas, setExpandedSchemas] = useState<Set<string>>(new Set(["s1"]));
  const [expandedTables, setExpandedTables] = useState<Set<string>>(new Set(["t-users"]));
  const [selectedTableId, setSelectedTableId] = useState<string>("t-users");
  const [selectedColumn, setSelectedColumn] = useState<string | null>(null);
  const [createOpen, setCreateOpen] = useState(false);

  const toggle = (setter: React.Dispatch<React.SetStateAction<Set<string>>>, key: string) => {
    setter((prev) => {
      const next = new Set(prev);
      if (next.has(key)) next.delete(key);
      else next.add(key);
      return next;
    });
  };

  const filteredProjects = useMemo(() => {
    const q = search.trim().toLowerCase();
    return schemaProjects
      .filter((p) => projectFilter === "all" || p.name === projectFilter)
      .map((p) => ({
        ...p,
        schemas: p.schemas
          .map((s) => ({
            ...s,
            tables: s.tables.filter((t) =>
              q === "" ||
              t.name.toLowerCase().includes(q) ||
              t.columns.some((c) => c.name.toLowerCase().includes(q))
            ),
          }))
          .filter((s) => s.tables.length > 0),
      }))
      .filter((p) => p.schemas.length > 0);
  }, [search, projectFilter]);

  const isSearching = search.trim().length > 0;

  const driftCount = useMemo(
    () => schemaProjects.flatMap((p) => p.schemas).flatMap((s) => s.tables).filter((t) => t.status === "drift").length,
    []
  );
  const pendingCount = useMemo(
    () => schemaProjects.flatMap((p) => p.schemas).flatMap((s) => s.tables).filter((t) => t.status === "pending").length,
    []
  );

  const selectedTable = useMemo(() => {
    for (const p of schemaProjects) {
      for (const s of p.schemas) {
        const t = s.tables.find((t) => t.id === selectedTableId);
        if (t) return { table: t, project: p.name, schema: s.name };
      }
    }
    return null;
  }, [selectedTableId]);

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
          {/* Header + New Schema */}
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Schemas</h1>
              <p className="text-sm text-muted-foreground mt-1">Explore database schemas across projects</p>
            </div>
            <Dialog open={createOpen} onOpenChange={setCreateOpen}>
              <DialogTrigger asChild>
                <Tooltip delay={0}>
                  <Button className="h-11 gap-2">
                    <Plus className="size-4" />
                    New Schema
                  </Button>
                  <Tooltip.Content>
                    <p>Create a new schema</p>
                  </Tooltip.Content>
                </Tooltip>
              </DialogTrigger>
              <DialogContent className="sm:max-w-[480px]">
                <DialogHeader>
                  <DialogTitle>New Schema</DialogTitle>
                  <DialogDescription>
                    Create a new schema in one of your projects.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-5 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="sname">Schema Name</Label>
                    <Input id="sname" placeholder="e.g. auth" className="h-11 font-mono" />
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="sproject">Project</Label>
                    <Select>
                      <SelectTrigger id="sproject" className="h-11">
                        <SelectValue placeholder="Select project" />
                      </SelectTrigger>
                      <SelectContent>
                        {projectOptions.map((p) => (
                          <SelectItem key={p} value={p.toLowerCase().replace(/ /g, "-")}>{p}</SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>
                  <div className="grid gap-2">
                    <Label htmlFor="sdesc">Description (optional)</Label>
                    <Textarea id="sdesc" placeholder="Purpose of this schema..." className="min-h-[80px]" />
                  </div>
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setCreateOpen(false)}>Cancel</Button>
                  <Button onClick={() => setCreateOpen(false)}>Create Schema</Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          {/* Search + Project filter */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="relative flex-1 max-w-md">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 size-4 text-muted-foreground" />
              <Input
                placeholder="Search tables or columns..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9 h-11"
              />
            </div>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" className="h-11">
                  {projectFilter === "all" ? "Project" : projectFilter}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuLabel>Filter by Project</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuCheckboxItem checked={projectFilter === "all"} onCheckedChange={() => setProjectFilter("all")}>All Projects</DropdownMenuCheckboxItem>
                {projectOptions.map((p) => (
                  <DropdownMenuCheckboxItem key={p} checked={projectFilter === p} onCheckedChange={() => setProjectFilter(p)}>{p}</DropdownMenuCheckboxItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
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
                  {filteredProjects.map((project) => {
                    const projectOpen = expandedProjects.has(project.id) || isSearching;
                    const tableCount = project.schemas.reduce((acc, s) => acc + s.tables.length, 0);
                    return (
                      <div key={project.id}>
                        <button
                          onClick={() => toggle(setExpandedProjects, project.id)}
                          className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 hover:bg-muted text-left"
                        >
                          <ChevronRight className={`size-3.5 text-muted-foreground transition-transform ${projectOpen ? "rotate-90" : ""}`} />
                          <FolderKanban className="size-4 text-muted-foreground" />
                          <span className="font-medium">{project.name}</span>
                          <span className="ml-auto text-xs text-muted-foreground">{tableCount}</span>
                        </button>
                        {projectOpen && (
                          <div className="ml-4 border-l pl-2">
                            {project.schemas.map((schema) => {
                              const schemaOpen = expandedSchemas.has(schema.id) || isSearching;
                              return (
                                <div key={schema.id}>
                                  <button
                                    onClick={() => toggle(setExpandedSchemas, schema.id)}
                                    className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 hover:bg-muted text-left"
                                  >
                                    <ChevronRight className={`size-3.5 text-muted-foreground transition-transform ${schemaOpen ? "rotate-90" : ""}`} />
                                    <Database className="size-4 text-muted-foreground" />
                                    <span>{schema.name}</span>
                                    <span className="ml-auto text-xs text-muted-foreground">{schema.tables.length}</span>
                                  </button>
                                  {schemaOpen && (
                                    <div className="ml-4 border-l pl-2">
                                      {schema.tables.map((table) => {
                                        const tableOpen = expandedTables.has(table.id) || isSearching;
                                        const selected = selectedTableId === table.id;
                                        return (
                                          <div key={table.id}>
                                            <button
                                              onClick={() => {
                                                setSelectedTableId(table.id);
                                                setSelectedColumn(null);
                                                toggle(setExpandedTables, table.id);
                                              }}
                                              className={`flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-left ${
                                                selected ? "bg-primary/10 text-primary" : "hover:bg-muted"
                                              }`}
                                            >
                                              <ChevronRight className={`size-3.5 text-muted-foreground transition-transform ${tableOpen ? "rotate-90" : ""}`} />
                                              <Table2 className={`size-4 ${selected ? "text-primary" : "text-muted-foreground"}`} />
                                              <span className="font-medium">{table.name}</span>
                                              <span className={`ml-auto size-1.5 rounded-full ${schemaStatusConfig[table.status].dot}`} />
                                            </button>
                                            {tableOpen && (
                                              <div className="ml-4 border-l pl-2">
                                                {table.columns.map((column) => (
                                                  <button
                                                    key={column.name}
                                                    onClick={() => {
                                                      setSelectedTableId(table.id);
                                                      setSelectedColumn(column.name);
                                                      toggle(setExpandedTables, table.id);
                                                    }}
                                                    className={`flex w-full items-center gap-2 rounded-md px-2 py-1 text-xs text-left ${
                                                      selectedTableId === table.id && selectedColumn === column.name
                                                        ? "bg-primary/10 text-primary"
                                                        : "hover:bg-muted"
                                                    }`}
                                                  >
                                                    {column.key === "PK" ? (
                                                      <KeyRound className="size-3 text-amber-500 shrink-0" />
                                                    ) : column.key === "FK" ? (
                                                      <Link2 className="size-3 text-blue-500 shrink-0" />
                                                    ) : (
                                                      <span className="size-1 shrink-0 rounded-full bg-muted-foreground/40" />
                                                    )}
                                                    <span className="font-mono truncate">{column.name}</span>
                                                    <span className="ml-auto font-mono text-[10px] text-muted-foreground truncate">{column.type}</span>
                                                  </button>
                                                ))}
                                              </div>
                                            )}
                                          </div>
                                        );
                                      })}
                                    </div>
                                  )}
                                </div>
                              );
                            })}
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
                {filteredProjects.length === 0 && (
                  <div className="flex flex-col items-center justify-center py-12 text-center">
                    <Database className="size-10 text-muted-foreground/40 mb-3" />
                    <h3 className="text-sm font-medium">No schemas found</h3>
                    <p className="text-xs text-muted-foreground mt-1">Try adjusting your search or filters</p>
                  </div>
                )}
                <Separator className="my-4" />
                <div className="flex items-center gap-4 px-1 pb-1">
                  {driftCount > 0 && (
                    <span className="flex items-center gap-1.5 text-xs text-red-500 font-medium">
                      <TriangleAlert className="size-3.5" />
                      {driftCount} with drift
                    </span>
                  )}
                  {pendingCount > 0 && (
                    <span className="flex items-center gap-1.5 text-xs text-yellow-500 font-medium">
                      <Clock className="size-3.5" />
                      {pendingCount} pending review
                    </span>
                  )}
                  {driftCount === 0 && pendingCount === 0 && (
                    <span className="flex items-center gap-1.5 text-xs text-green-500 font-medium">
                      <ShieldCheck className="size-3.5" />
                      All schemas healthy
                    </span>
                  )}
                </div>
              </CardContent>
            </Card>

            {/* Detail */}
            {selectedTable ? (
              <Card className="h-fit">
                <CardHeader className="border-0">
                  <div className="flex flex-wrap items-center gap-3">
                    <CardTitle className="text-lg flex items-center gap-2">
                      <Table2 className="size-4 text-muted-foreground" />
                      {selectedTable.table.name}
                    </CardTitle>
                    <Badge variant={schemaStatusConfig[selectedTable.table.status].badge} className="text-[10px] px-1.5 py-0">
                      {schemaStatusConfig[selectedTable.table.status].label}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {selectedTable.project} · {selectedTable.schema} · {selectedTable.table.columns.length} columns · {selectedTable.table.indexes.length} indexes
                  </p>
                </CardHeader>
                <CardContent className="pt-0">
                  <Tabs defaultValue="columns" key={selectedTable.table.id}>
                    <TabsList variant="line" className="w-full justify-start gap-4">
                      <TabsTrigger value="columns" className="gap-1.5">
                        <Columns3 className="size-4" />
                        Columns
                        <span className="text-xs text-muted-foreground">{selectedTable.table.columns.length}</span>
                      </TabsTrigger>
                      <TabsTrigger value="indexes" className="gap-1.5">
                        <Hash className="size-4" />
                        Indexes
                        <span className="text-xs text-muted-foreground">{selectedTable.table.indexes.length}</span>
                      </TabsTrigger>
                      <TabsTrigger value="relations" className="gap-1.5">
                        <Link2 className="size-4" />
                        Relations
                        <span className="text-xs text-muted-foreground">{selectedTable.table.relations.length}</span>
                      </TabsTrigger>
                      <TabsTrigger value="sql" className="gap-1.5">
                        <FileCode2 className="size-4" />
                        SQL
                      </TabsTrigger>
                    </TabsList>

                    <TabsContent value="columns" className="mt-4 space-y-4">
                      {selectedColumn && (
                        <div className="rounded-md border">
                          <div className="flex items-center justify-between border-b px-3 py-2">
                            <span className="text-xs font-medium text-muted-foreground">
                              Column DDL · <span className="font-mono">{selectedColumn}</span>
                            </span>
                            <CopyButton text={generateColumnDDL(selectedTable.table, selectedTable.schema, selectedTable.table.columns.find((c) => c.name === selectedColumn)!)} />
                          </div>
                          <pre className="bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed">
                            {generateColumnDDL(selectedTable.table, selectedTable.schema, selectedTable.table.columns.find((c) => c.name === selectedColumn)!)}
                          </pre>
                        </div>
                      )}
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
                          {selectedTable.table.columns.map((column) => (
                            <TableRow
                              key={column.name}
                              onClick={() => setSelectedColumn(column.name)}
                              className={`cursor-pointer ${selectedColumn === column.name ? "bg-primary/5" : ""}`}
                            >
                              <TableCell>
                                <div className="flex items-center gap-2">
                                  {column.key === "PK" && <KeyRound className="size-3 text-amber-500 shrink-0" />}
                                  {column.key === "FK" && <Link2 className="size-3 text-blue-500 shrink-0" />}
                                  <span className="font-mono text-sm">{column.name}</span>
                                </div>
                              </TableCell>
                              <TableCell><span className="font-mono text-xs text-muted-foreground">{column.type}</span></TableCell>
                              <TableCell>
                                {column.key && (
                                  <Badge variant="outline" className="text-[10px] px-1.5 py-0">{column.key}</Badge>
                                )}
                              </TableCell>
                              <TableCell className="text-xs text-muted-foreground">{column.nullable ? "YES" : "NO"}</TableCell>
                              <TableCell className="text-right font-mono text-xs text-muted-foreground">
                                {column.default ?? "—"}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </TabsContent>

                    <TabsContent value="indexes" className="mt-4">
                      <Table>
                        <TableHeader>
                          <TableRow>
                            <TableHead className="w-[35%]">Index</TableHead>
                            <TableHead>Columns</TableHead>
                            <TableHead>Type</TableHead>
                            <TableHead className="text-right">Unique</TableHead>
                          </TableRow>
                        </TableHeader>
                        <TableBody>
                          {selectedTable.table.indexes.map((index) => (
                            <TableRow key={index.name}>
                              <TableCell><span className="font-mono text-sm">{index.name}</span></TableCell>
                              <TableCell>
                                <span className="font-mono text-xs text-muted-foreground">{index.columns.join(", ")}</span>
                              </TableCell>
                              <TableCell>
                                <Badge variant="outline" className="text-[10px] px-1.5 py-0 font-mono">{index.type}</Badge>
                              </TableCell>
                              <TableCell className="text-right">
                                {index.unique && <Badge variant="default" className="text-[10px] px-1.5 py-0">Unique</Badge>}
                              </TableCell>
                            </TableRow>
                          ))}
                        </TableBody>
                      </Table>
                    </TabsContent>

                    <TabsContent value="relations" className="mt-4">
                      {selectedTable.table.relations.length > 0 ? (
                        <Table>
                          <TableHeader>
                            <TableRow>
                              <TableHead className="w-[40%]">Constraint</TableHead>
                              <TableHead>From</TableHead>
                              <TableHead>To</TableHead>
                            </TableRow>
                          </TableHeader>
                          <TableBody>
                            {selectedTable.table.relations.map((relation) => (
                              <TableRow key={relation.name}>
                                <TableCell><span className="font-mono text-sm">{relation.name}</span></TableCell>
                                <TableCell><span className="font-mono text-xs text-blue-500">{relation.from}</span></TableCell>
                                <TableCell><span className="font-mono text-xs text-muted-foreground">{relation.to}</span></TableCell>
                              </TableRow>
                            ))}
                          </TableBody>
                        </Table>
                      ) : (
                        <div className="flex flex-col items-center justify-center py-12 text-center">
                          <Link2 className="size-10 text-muted-foreground/40 mb-3" />
                          <h3 className="text-sm font-medium">No relations</h3>
                          <p className="text-xs text-muted-foreground mt-1">This table has no foreign keys</p>
                        </div>
                      )}
                    </TabsContent>

                    <TabsContent value="sql" className="mt-4">
                      <div className="rounded-md border">
                        <div className="flex items-center justify-between border-b px-3 py-2">
                          <span className="font-mono text-xs font-medium text-muted-foreground">
                            CREATE TABLE {selectedTable.schema}.{selectedTable.table.name}
                          </span>
                          <CopyButton text={generateTableDDL(selectedTable.table, selectedTable.schema)} />
                        </div>
                        <pre className="bg-muted/50 px-3 py-2.5 overflow-x-auto font-mono text-xs leading-relaxed">
                          {generateTableDDL(selectedTable.table, selectedTable.schema)}
                        </pre>
                      </div>
                    </TabsContent>
                  </Tabs>
                </CardContent>
              </Card>
            ) : (
              <Card>
                <CardContent className="flex flex-col items-center justify-center py-20 text-center">
                  <Database className="size-12 text-muted-foreground/40 mb-4" />
                  <h3 className="text-lg font-medium">Select a table</h3>
                  <p className="text-sm text-muted-foreground mt-1">Choose a table from the explorer to see its details</p>
                </CardContent>
              </Card>
            )}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
