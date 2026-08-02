"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { projectClient } from "@/lib/api/clients";
import type {
  Connection,
  TestConnectionResponse,
} from "@/lib/gen/project/v1/project_messages_pb";

export type ConnectionInput = {
  projectId: string;
  name: string;
  host: string;
  port: number;
  databaseName: string;
  username: string;
  password: string;
  sslMode?: string;
};

export function useConnections(projectId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "connections"],
    queryFn: async () => {
      if (!projectId) throw new Error("Missing project id");
      const res = await projectClient.listConnections({ projectId });
      return res.connections;
    },
    enabled: Boolean(projectId),
    staleTime: 30_000,
  });
}

export function useConnection(projectId: string | undefined, connectionId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "connections", connectionId],
    queryFn: async () => {
      if (!projectId || !connectionId) throw new Error("Missing connection id");
      const res = await projectClient.getConnection({ id: connectionId });
      return res.connection ?? null;
    },
    enabled: Boolean(projectId && connectionId),
    staleTime: 30_000,
  });
}

export function useCreateConnection() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: ConnectionInput) => {
      const res = await projectClient.createConnection({
        projectId: input.projectId,
        name: input.name,
        host: input.host,
        port: input.port,
        databaseName: input.databaseName,
        username: input.username,
        password: input.password,
        sslMode: input.sslMode ?? "",
      });
      return res.connection ?? null;
    },
    onSuccess: (_conn, input) => {
      queryClient.invalidateQueries({
        queryKey: ["projects", input.projectId, "connections"],
      });
    },
  });
}

export function useUpdateConnection(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      connectionId: string;
      name?: string;
      host?: string;
      port?: number;
      databaseName?: string;
      username?: string;
      password?: string;
      sslMode?: string;
    }) => {
      const res = await projectClient.updateConnection({
        id: input.connectionId,
        name: input.name ?? "",
        host: input.host ?? "",
        port: input.port ?? 0,
        databaseName: input.databaseName ?? "",
        username: input.username ?? "",
        password: input.password ?? "",
        sslMode: input.sslMode ?? "",
      });
      return res.connection ?? null;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["projects", projectId, "connections"],
      });
    },
  });
}

export function useDeleteConnection(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (connectionId: string) => {
      await projectClient.deleteConnection({ id: connectionId });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["projects", projectId, "connections"],
      });
    },
  });
}

export function useTestConnection() {
  return useMutation({
    mutationFn: async (input: {
      projectId: string;
      connectionId: string;
    }): Promise<TestConnectionResponse> => {
      const res = await projectClient.testConnection(input);
      return res;
    },
  });
}

export type { Connection };
