"use client";

import { useParams } from "next/navigation";
import Link from "next/link";
import { useState } from "react";
import { ArrowLeft, Search, FileText, Download } from "lucide-react";


import {  Badge  } from "@/components/ui/badge";
import {  Button  } from "@/components/ui/button";
import {  Input  } from "@/components/ui/input";
import {  Card, CardContent, CardHeader, CardTitle  } from "@/components/ui/card";















import {  useAuditEntries  } from "@/lib/api/hooks/use-audit";
import {  getApiErrorMessage  } from "@/lib/api/errors";

const auditBadgeConfig: Record<string, "default" | "secondary" | "destructive" | "outline"> = {
  migration: "default",
  drift: "destructive",
  review: "secondary",
  auth: "outline",
  team: "outline",
  schema: "default",
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
  return `${Math.floor(hours / 24)}d ago`;
}

function initialsOf(name: string): string {
  return name
    .split(/[\s@._-]+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("") || "?";
}

export default function AuditPage() {
  const params = useParams();
  const projectId = params.id as string;

  const { data: entries, isLoading, error } = useAuditEntries({
    resourceType: "project",
    resourceId: projectId,
  });

  const [search, setSearch] = useState("");
  const filtered = (entries ?? []).filter((e) => {
    if (search === "") return true;
    const haystack = `${e.eventType} ${e.action} ${e.actorEmail} ${e.actorId} ${e.resourceType} ${e.resourceId}`.toLowerCase();
    return haystack.includes(search.toLowerCase());
  });

  const exportCsv = () => {
    if (filtered.length === 0) return;
    const header = "id,eventType,actorId,actorEmail,action,resourceType,resourceId,createdAt";
    const rows = filtered.map((e) =>
      [e.id, e.eventType, e.actorId, e.actorEmail, e.action, e.resourceType, e.resourceId, e.createdAt]
        .map((v) => `"${String(v).replace(/"/g, '""')}"`)
        .join(","),
    );
    const blob = new Blob([[header, ...rows].join("\n")], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `audit-${projectId.slice(0, 8)}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
            <div className="flex flex-1 flex-col gap-6 p-6">
        <div className="flex flex-wrap items-center gap-3">
            <div className="relative">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
              <Input
                placeholder="Search audit log..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-[180px] lg:w-[220px] h-9 pl-8 text-sm"
              />
            </div>
          </div>
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
                <Badge variant="outline" className="text-[11px]">{filtered.length} entries</Badge>
              </div>
              <p className="text-sm text-muted-foreground mt-1">Every schema-changing action, permanently recorded</p>
            </div>
            <Button
              variant="outline"
              className="h-10 gap-2 shrink-0"
              onClick={exportCsv}
              disabled={filtered.length === 0}
            >
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
              {isLoading ? (
                <p className="text-sm text-muted-foreground py-6 text-center">Loading audit entries...</p>
              ) : error ? (
                <p className="text-sm text-red-500 py-6 text-center">{getApiErrorMessage(error)}</p>
              ) : !entries || entries.length === 0 ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  No audit activity for this project yet.
                </p>
              ) : filtered.length === 0 ? (
                <p className="text-sm text-muted-foreground py-6 text-center">
                  No entries match your search.
                </p>
              ) : (
                <div className="divide-y">
                  {filtered.map((entry) => (
                    <div key={entry.id} className="flex items-start gap-3 py-4">
                      <div className="flex size-8 shrink-0 items-center justify-center rounded-full bg-muted">
                        <span className="text-xs font-medium">{initialsOf(entry.actorEmail || entry.actorId)}</span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <p className="text-sm font-medium">{entry.actorEmail || entry.actorId || "system"}</p>
                          <Badge variant={auditBadgeConfig[entry.eventType] ?? "outline"} className="text-[10px] px-1.5 py-0">
                            {entry.eventType}
                          </Badge>
                          <span className="text-sm text-muted-foreground">{entry.action}</span>
                        </div>
                        <p className="text-sm text-muted-foreground mt-1">
                          {entry.resourceType} · {entry.resourceId}
                        </p>
                      </div>
                      <span className="text-xs text-muted-foreground shrink-0">{relativeTime(entry.createdAt)}</span>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
  );
}
