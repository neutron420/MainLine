"use client";

import { useEffect, useRef, useState } from "react";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { migrationClient } from "@/lib/api/clients";
import type {
  Migration,
  MigrationLogEntry,
  MigrationRun,
  MigrationStatusMessage,
} from "@/lib/gen/migration/v1/migration_messages_pb";

export type CreateMigrationInput = {
  projectId: string;
  title: string;
  version: string;
  upSql: string;
  downSql?: string;
  description?: string;
};

export function useMigrations(projectId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "migrations"],
    queryFn: async () => {
      if (!projectId) throw new Error("Missing project id");
      const res = await migrationClient.listMigrations({ projectId });
      return res.migrations;
    },
    enabled: Boolean(projectId),
    staleTime: 30_000,
  });
}

export function useMigration(projectId: string | undefined, migrationId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "migrations", migrationId],
    queryFn: async () => {
      if (!projectId || !migrationId) throw new Error("Missing migration id");
      const res = await migrationClient.getMigration({ id: migrationId });
      return res.migration ?? null;
    },
    enabled: Boolean(projectId && migrationId),
    staleTime: 30_000,
  });
}

export function useCreateMigration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: CreateMigrationInput) => {
      const res = await migrationClient.createMigration({
        projectId: input.projectId,
        title: input.title,
        version: input.version,
        upSql: input.upSql,
        downSql: input.downSql ?? "",
        description: input.description ?? "",
      });
      return res.migration ?? null;
    },
    onSuccess: (_migration, input) => {
      queryClient.invalidateQueries({
        queryKey: ["projects", input.projectId, "migrations"],
      });
    },
  });
}

export function useDeleteMigration(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (migrationId: string) => {
      await migrationClient.deleteMigration({ id: migrationId });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["projects", projectId, "migrations"],
      });
    },
  });
}

export function useMigrationRuns(migrationId: string | undefined) {
  return useQuery({
    queryKey: ["migrations", migrationId, "runs"],
    queryFn: async () => {
      if (!migrationId) throw new Error("Missing migration id");
      const res = await migrationClient.listMigrationRuns({ migrationId });
      return res.runs;
    },
    enabled: Boolean(migrationId),
    staleTime: 15_000,
  });
}

export function useExecuteMigration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { migrationId: string; connectionId: string }) => {
      const res = await migrationClient.executeMigration(input);
      return res.run ?? null;
    },
    onSuccess: (run, input) => {
      if (run) {
        queryClient.invalidateQueries({
          queryKey: ["migrations", input.migrationId, "runs"],
        });
        queryClient.invalidateQueries({
          queryKey: ["projects", undefined, "migrations"],
        });
      }
    },
  });
}

export function useRollbackMigration() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { migrationId: string; runId: string }) => {
      const res = await migrationClient.rollbackMigration({ runId: input.runId });
      return res.run ?? null;
    },
    onSuccess: (run, input) => {
      if (run) {
        queryClient.invalidateQueries({
          queryKey: ["migrations", input.migrationId, "runs"],
        });
      }
    },
  });
}

/**
 * Real-time migration execution progress via the WatchMigration
 * server-streaming RPC. Returns the latest status message plus a live
 * event buffer for the timeline UI.
 */
export function useWatchMigration(runId: string | undefined, enabled = true) {
  const [status, setStatus] = useState<MigrationStatusMessage | null>(null);
  const [logs, setLogs] = useState<MigrationLogEntry[]>([]);
  const [connected, setConnected] = useState(false);
  const logsRef = useRef<MigrationLogEntry[]>([]);

  useEffect(() => {
    if (!runId || !enabled) return;

    let cancelled = false;
    const abort: AbortController | null = new AbortController();

    const stream = async () => {
      try {
        const iterable = migrationClient.watchMigration(
          { runId },
          { signal: abort!.signal },
        );
        for await (const msg of iterable) {
          if (cancelled) return;
          setStatus(msg);
          if (msg.lastLog) {
            logsRef.current = [...logsRef.current, msg.lastLog];
            setLogs(logsRef.current);
          }
          if (msg.state === "completed" || msg.state === "failed" || msg.state === "rolled_back") {
            break;
          }
        }
      } catch {
        // stream closed (abort or transient error) - surface disconnect state
      } finally {
        if (!cancelled) setConnected(false);
      }
    };

    setConnected(true);
    void stream();

    return () => {
      cancelled = true;
      abort?.abort();
      setConnected(false);
    };
  }, [runId, enabled]);

  return { status, logs, connected };
}

/**
 * Real-time rollback progress via the WatchRollback server-streaming RPC.
 * Shares the same message shape as WatchMigration.
 */
export function useWatchRollback(runId: string | undefined, enabled = true) {
  const [status, setStatus] = useState<MigrationStatusMessage | null>(null);
  const [logs, setLogs] = useState<MigrationLogEntry[]>([]);
  const [connected, setConnected] = useState(false);
  const logsRef = useRef<MigrationLogEntry[]>([]);

  useEffect(() => {
    if (!runId || !enabled) return;

    let cancelled = false;
    const abort: AbortController | null = new AbortController();

    const stream = async () => {
      try {
        const iterable = migrationClient.watchRollback(
          { runId },
          { signal: abort!.signal },
        );
        for await (const msg of iterable) {
          if (cancelled) return;
          setStatus(msg);
          if (msg.lastLog) {
            logsRef.current = [...logsRef.current, msg.lastLog];
            setLogs(logsRef.current);
          }
          if (msg.state === "completed" || msg.state === "failed" || msg.state === "rolled_back") {
            break;
          }
        }
      } catch {
        // stream closed (abort or transient error) - surface disconnect state
      } finally {
        if (!cancelled) setConnected(false);
      }
    };

    setConnected(true);
    void stream();

    return () => {
      cancelled = true;
      abort?.abort();
      setConnected(false);
    };
  }, [runId, enabled]);

  return { status, logs, connected };
}

export type { Migration, MigrationRun, MigrationStatusMessage, MigrationLogEntry };
