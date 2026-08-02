"use client";

import { useState } from "react";
import Link from "next/link";
import { Database, Trash2, Loader2, Plug, ExternalLink } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  useConnections,
  useDeleteConnection,
  useTestConnection,
} from "@/lib/api/hooks/use-connections";
import type { Project } from "@/lib/api/hooks/use-projects";

const statusConfig: Record<string, { label: string; badge: "default" | "destructive" | "outline"; dot: string }> = {
  connected: { label: "Connected", badge: "default", dot: "bg-emerald-500" },
  error: { label: "Error", badge: "destructive", dot: "bg-red-500" },
  unknown: { label: "Unknown", badge: "outline", dot: "bg-muted-foreground" },
};

export function ProjectConnectionsCard({ project }: { project: Project }) {
  const { data: connections, isLoading } = useConnections(project.id);
  const deleteConnection = useDeleteConnection(project.id);
  const testConnection = useTestConnection();
  const [testResult, setTestResult] = useState<Record<string, string>>({});
  const [testing, setTesting] = useState<string | null>(null);

  const runTest = async (connectionId: string) => {
    setTesting(connectionId);
    setTestResult((prev) => ({ ...prev, [connectionId]: "" }));
    const res = await testConnection.mutateAsync({ projectId: project.id, connectionId });
    setTestResult((prev) => ({
      ...prev,
      [connectionId]: res.success
        ? `OK — ${res.serverVersion || "connected"} (${res.latencyMs} ms)`
        : `Failed — ${res.error || "connection error"}`,
    }));
    setTesting(null);
  };

  return (
    <div className="space-y-3">
      <div className="flex items-center justify-between">
        <p className="text-sm font-medium flex items-center gap-2">
          {project.name}
          <Badge variant="outline" className="text-[10px] px-1.5 py-0">{project.slug}</Badge>
        </p>
        <Link href={`/projects/${project.id}/connections`} className="text-xs text-primary hover:underline inline-flex items-center gap-1">
          Manage <ExternalLink className="size-3" />
        </Link>
      </div>
      {isLoading && <p className="text-xs text-muted-foreground">Loading connections...</p>}
      {!isLoading && (connections?.length ?? 0) === 0 && (
        <p className="text-xs text-muted-foreground">No connections in this project.</p>
      )}
      {(connections ?? []).map((conn) => {
        const status = statusConfig[conn.connectionStatus] ?? statusConfig.unknown;
        return (
          <div key={conn.id} className="flex items-center gap-4 py-3.5 border rounded-lg px-4">
            <div className="flex size-9 shrink-0 items-center justify-center rounded-lg bg-muted">
              <Database className="size-4 text-muted-foreground" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2">
                <p className="text-sm font-medium">{conn.name}</p>
                <Badge variant={status.badge} className="text-[10px] px-1.5 py-0 gap-1">
                  <span className={`size-1.5 rounded-full ${status.dot}`} />
                  {status.label}
                </Badge>
              </div>
              <p className="text-xs text-muted-foreground truncate font-mono mt-0.5">
                {conn.host}:{conn.port} · {conn.databaseName}
              </p>
              {testResult[conn.id] && (
                <p className={`text-xs mt-1 font-mono ${testResult[conn.id].startsWith("OK") ? "text-emerald-600" : "text-destructive"}`}>
                  {testResult[conn.id]}
                </p>
              )}
            </div>
            <Button
              variant="outline"
              size="sm"
              className="h-8 gap-1.5"
              disabled={testing === conn.id}
              onClick={() => runTest(conn.id)}
            >
              {testing === conn.id ? <Loader2 className="size-3.5 animate-spin" /> : <Plug className="size-3.5" />}
              Test
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-8 text-destructive"
              onClick={() => deleteConnection.mutate(conn.id)}
            >
              <Trash2 className="size-4" />
            </Button>
          </div>
        );
      })}
    </div>
  );
}
