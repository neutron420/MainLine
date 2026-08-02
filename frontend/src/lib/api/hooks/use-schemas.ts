"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { schemaClient } from "@/lib/api/clients";
import type {
  Schema,
  SchemaVersion,
  DiagramEdge,
  DiagramNode,
  SchemaDiff,
} from "@/lib/gen/schema/v1/schema_messages_pb";

export function useSchemas(projectId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "schemas"],
    queryFn: async () => {
      if (!projectId) throw new Error("Missing project id");
      const res = await schemaClient.listSchemas({ projectId });
      return res.schemas;
    },
    enabled: Boolean(projectId),
    staleTime: 30_000,
  });
}

export function useSchema(projectId: string | undefined, schemaId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "schemas", schemaId],
    queryFn: async () => {
      if (!projectId || !schemaId) throw new Error("Missing schema id");
      const res = await schemaClient.getSchema({ id: schemaId });
      return res.schema ?? null;
    },
    enabled: Boolean(projectId && schemaId),
    staleTime: 30_000,
  });
}

export function useSchemaVersions(schemaId: string | undefined) {
  return useQuery({
    queryKey: ["schemas", schemaId, "versions"],
    queryFn: async () => {
      if (!schemaId) throw new Error("Missing schema id");
      const res = await schemaClient.listSchemaVersions({ schemaId });
      return res.versions;
    },
    enabled: Boolean(schemaId),
    staleTime: 30_000,
  });
}

export function useIntrospectSchema() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      connectionId: string;
      schemaNames?: string[];
    }): Promise<{ schema: Schema; version: SchemaVersion }> => {
      const res = await schemaClient.introspectSchema({
        connectionId: input.connectionId,
        schemaNames: input.schemaNames ?? [],
      });
      return { schema: res.schema!, version: res.schemaVersion! };
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useSchemaDiagram(schemaVersionId: string | undefined) {
  return useQuery({
    queryKey: ["schemas", schemaVersionId, "diagram"],
    queryFn: async () => {
      if (!schemaVersionId) throw new Error("Missing schema version id");
      const res = await schemaClient.getSchemaDiagram({
        schemaVersionId,
        includeDetails: true,
      });
      return {
        nodes: res.nodes,
        edges: res.edges,
      } as { nodes: DiagramNode[]; edges: DiagramEdge[] };
    },
    enabled: Boolean(schemaVersionId),
    staleTime: 60_000,
  });
}

export function useCompareSchemas(
  schemaVersionId: string | undefined,
  otherVersionId: string | undefined,
) {
  return useQuery({
    queryKey: ["schemas", "compare", schemaVersionId, otherVersionId],
    queryFn: async () => {
      if (!schemaVersionId || !otherVersionId) throw new Error("Missing version ids");
      const res = await schemaClient.compareSchemaVersions({
        versionAId: schemaVersionId,
        versionBId: otherVersionId,
      });
      return res.diff as SchemaDiff | null;
    },
    enabled: Boolean(schemaVersionId && otherVersionId),
    staleTime: 60_000,
  });
}

export function useSchemaObjects(schemaVersionId: string | undefined) {
  return useQuery({
    queryKey: ["schemas", schemaVersionId, "objects"],
    queryFn: async () => {
      if (!schemaVersionId) throw new Error("Missing schema version id");
      const res = await schemaClient.listSchemaObjects({ schemaVersionId });
      return res.objects;
    },
    enabled: Boolean(schemaVersionId),
    staleTime: 60_000,
  });
}

export type { Schema, SchemaVersion, DiagramNode, DiagramEdge, SchemaDiff };
