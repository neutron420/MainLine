"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { driftClient } from "@/lib/api/clients";
import type { DriftEvent } from "@/lib/gen/drift/v1/drift_messages_pb";

export type DriftFilter = {
  connectionId: string | undefined;
  status?: string;
  severity?: string;
  driftType?: string;
};

export function useDriftEvents(filter: DriftFilter) {
  return useQuery({
    queryKey: ["drift", filter.connectionId, filter.status, filter.severity],
    queryFn: async () => {
      if (!filter.connectionId) throw new Error("Missing connection id");
      const res = await driftClient.listDriftEvents({
        connectionId: filter.connectionId,
        status: filter.status ?? "",
        severity: filter.severity ?? "",
        driftType: filter.driftType ?? "",
      });
      return res.events;
    },
    enabled: Boolean(filter.connectionId),
    staleTime: 30_000,
  });
}

export function useDriftEvent(eventId: string | undefined) {
  return useQuery({
    queryKey: ["drift", eventId],
    queryFn: async () => {
      if (!eventId) throw new Error("Missing drift event id");
      const res = await driftClient.getDriftEvent({ id: eventId });
      return res.event ?? null;
    },
    enabled: Boolean(eventId),
    staleTime: 30_000,
  });
}

export function useCheckDrift() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      connectionId: string;
      schemaNames?: string[];
    }) => {
      const res = await driftClient.checkDrift({
        connectionId: input.connectionId,
        schemaNames: input.schemaNames ?? [],
      });
      return {
        events: res.events,
        hasDrift: res.hasDrift,
        totalDrifts: res.totalDrifts,
      };
    },
    onSuccess: (result, input) => {
      queryClient.invalidateQueries({
        queryKey: ["drift", input.connectionId],
      });
      queryClient.invalidateQueries({ queryKey: ["drift"] });
      void result;
    },
  });
}

export function useResolveDriftEvent() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      eventId: string;
      status: string;
    }) => {
      const res = await driftClient.resolveDriftEvent({
        id: input.eventId,
        status: input.status,
      });
      return res.event ?? null;
    },
    onSuccess: (event) => {
      if (event) {
        queryClient.invalidateQueries({ queryKey: ["drift", event.connectionId] });
        queryClient.invalidateQueries({ queryKey: ["drift", event.id] });
      }
    },
  });
}

export type { DriftEvent };
