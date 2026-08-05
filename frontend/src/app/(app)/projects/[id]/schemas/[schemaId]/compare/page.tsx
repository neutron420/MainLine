"use client";

import { useState } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Columns3, GitCompareArrows, Plus, Minus, PencilLine } from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
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
import {
  useSchema,
  useSchemaVersions,
  useCompareSchemas,
} from "@/lib/api/hooks/use-schemas";
import { getApiErrorMessage } from "@/lib/api/errors";

const diffConfig: Record<string, { label: string; badge: "default" | "destructive" | "secondary"; icon: typeof Plus }> = {
  added: { label: "Added", badge: "default", icon: Plus },
  removed: { label: "Removed", badge: "destructive", icon: Minus },
  modified: { label: "Modified", badge: "secondary", icon: PencilLine },
};

function relativeTime(iso: string | undefined): string {
  if (!iso) return "—";
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "—";
  const mins = Math.floor((Date.now() - then) / 60_000);
  if (mins < 1) return "just now";
  if (mins < 60) return `${mins} min ago`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h ago`;
  return new Date(iso).toLocaleDateString();
}

export default function SchemaComparePage() {
  const params = useParams();
  const projectId = params.id as string;
  const schemaId = params.schemaId as string;

  const { data: schema, isLoading: schemaLoading } = useSchema(projectId, schemaId);
  const { data: versions } = useSchemaVersions(schemaId);

  const sorted = [...(versions ?? [])].sort((a, b) => b.version - a.version);
  const [versionAId, setVersionAId] = useState<string>("");
  const [versionBId, setVersionBId] = useState<string>("");

  const effectiveA = versionAId || sorted[0]?.id || "";
  const effectiveB = versionBId || sorted[1]?.id || "";

  const { data: diff, isLoading: diffLoading, error } = useCompareSchemas(
    effectiveA || undefined,
    effectiveB || undefined,
  );

  const versionById = new Map(sorted.map((v) => [v.id, v]));
  const versionA = versionById.get(effectiveA);
  const versionB = versionById.get(effectiveB);

  const total =
    (diff?.addedObjects.length ?? 0) +
    (diff?.removedObjects.length ?? 0) +
    (diff?.modifiedObjects.length ?? 0);

  if (schemaLoading) {
    return (
                <div className="flex h-full flex-col items-center justify-center gap-4 p-8">
            <GitCompareArrows className="size-12 text-muted-foreground/40" />
            <p className="text-sm text-muted-foreground">Loading schema versions...</p>
          </div>
    );
  }

  if (!schema) {
    return (
      <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
        <AppSidebar />
        <SidebarInset>
          <div className="flex flex-col items-center justify-center h-full gap-4 p-8">
            <GitCompareArrows className="size-12 text-muted-foreground/40" />
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

  const diffSections = [
    { key: "added" as const, label: "Added Objects", items: diff?.addedObjects ?? [], icon: Plus },
    { key: "removed" as const, label: "Removed Objects", items: diff?.removedObjects ?? [], icon: Minus },
    { key: "modified" as const, label: "Modified Objects", items: diff?.modifiedObjects ?? [], icon: PencilLine },
  ];

  return (
    <SidebarProvider style={{ "--sidebar-width": "350px" } as React.CSSProperties}>
      <AppSidebar />
      <SidebarInset>
        <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
          </div>
          {/* Header */}
          <div className="flex items-start gap-4">
            <Link href={`/projects/${projectId}/schemas/${schemaId}`}>
              <Button variant="ghost" size="icon" className="size-10 shrink-0 mt-0.5">
                <ArrowLeft className="size-4" />
              </Button>
            </Link>
            <div>
              <h1 className="text-2xl font-semibold tracking-tight">Schema Compare</h1>
              <p className="text-sm text-muted-foreground mt-1">Compare two versions of {schema.schemaName}</p>
            </div>
          </div>

          {/* Version selectors */}
          <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
            <div className="flex items-center gap-2.5 flex-wrap">
              <Select value={effectiveA} onValueChange={setVersionAId}>
                <SelectTrigger className="h-11 w-64">
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent>
                  {sorted.map((v) => (
                    <SelectItem key={v.id} value={v.id}>
                      v{v.version} · {v.objectCount} objects
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <GitCompareArrows className="size-4 text-muted-foreground" />
              <Select value={effectiveB} onValueChange={setVersionBId}>
                <SelectTrigger className="h-11 w-64">
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent>
                  {sorted.map((v) => (
                    <SelectItem key={v.id} value={v.id}>
                      v{v.version} · {v.objectCount} objects
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>

          {/* Diff summary */}
          <Card>
            <CardHeader className="border-0">
              <div className="flex items-center justify-between gap-4 flex-wrap">
                <CardTitle className="text-base flex items-center gap-2">
                  <GitCompareArrows className="size-4" />
                  Version Diff
                </CardTitle>
                {diff && (
                  <div className="flex items-center gap-2">
                    <Badge variant="default" className="gap-1 text-[10px] px-1.5 py-0">
                      <span className="text-green-300">+</span>{diff.addedObjects.length} added
                    </Badge>
                    <Badge variant="destructive" className="gap-1 text-[10px] px-1.5 py-0">
                      <span>−</span>{diff.removedObjects.length} removed
                    </Badge>
                    <Badge variant="secondary" className="gap-1 text-[10px] px-1.5 py-0">
                      <span className="text-amber-400">~</span>{diff.modifiedObjects.length} modified
                    </Badge>
                  </div>
                )}
              </div>
            </CardHeader>
            <CardContent className="pt-0">
              {!effectiveA || !effectiveB ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  {sorted.length < 2
                    ? "At least two schema versions are needed to compare."
                    : "Select two versions to compare."}
                </p>
              ) : diffLoading ? (
                <p className="text-sm text-muted-foreground py-6 text-center">Comparing versions...</p>
              ) : error ? (
                <p className="text-sm text-red-500 py-6 text-center">{getApiErrorMessage(error)}</p>
              ) : total === 0 ? (
                <p className="text-sm text-muted-foreground py-6 text-center flex items-center justify-center gap-2">
                  <Columns3 className="size-4" />
                  The two versions are identical.
                </p>
              ) : (
                <div className="space-y-6">
                  {versionA && versionB && (
                    <div className="flex items-center gap-2 text-xs text-muted-foreground">
                      <Badge variant="outline" className="font-mono">v{versionA.version}</Badge>
                      <span>· {relativeTime(versionA.createdAt)}</span>
                      <GitCompareArrows className="size-3 mx-1" />
                      <Badge variant="outline" className="font-mono">v{versionB.version}</Badge>
                      <span>· {relativeTime(versionB.createdAt)}</span>
                    </div>
                  )}
                  {diffSections.map((section) =>
                    section.items.length === 0 ? null : (
                      <div key={section.key}>
                        <h3 className="flex items-center gap-2 text-sm font-semibold mb-2">
                          <section.icon className="size-4 text-muted-foreground" />
                          {section.label}
                          <Badge variant={diffConfig[section.key].badge} className="text-[10px] px-1.5 py-0">
                            {section.items.length}
                          </Badge>
                        </h3>
                        <div className="rounded-md border divide-y">
                          {section.items.map((obj, i) => (
                            <div key={`${obj.type}-${obj.name}-${i}`} className="p-3">
                              <div className="flex items-center gap-2 flex-wrap">
                                <span className="font-mono text-sm">{obj.name}</span>
                                <Badge variant="outline" className="text-[10px] px-1.5 py-0">{obj.type}</Badge>
                              </div>
                              {obj.changes.length > 0 && (
                                <div className="mt-2 space-y-1">
                                  {obj.changes.map((change, j) => (
                                    <div key={j} className="flex items-center gap-2 text-xs font-mono">
                                      <span className="text-muted-foreground shrink-0">{change.field}:</span>
                                      {change.before && (
                                        <span className="text-red-500 line-through truncate">{change.before}</span>
                                      )}
                                      {change.after && (
                                        <span className="text-green-500 truncate">{change.after}</span>
                                      )}
                                    </div>
                                  ))}
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      </div>
                    ),
                  )}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </SidebarInset>
    </SidebarProvider>
  );
}
