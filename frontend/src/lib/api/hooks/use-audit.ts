"use client";

import { useQuery } from "@tanstack/react-query";

import { auditClient } from "@/lib/api/clients";
import type { AuditEntry } from "@/lib/gen/audit/v1/audit_messages_pb";

export type AuditFilter = {
  eventType?: string;
  actorId?: string;
  resourceType?: string;
  resourceId?: string;
  dateFrom?: string;
  dateTo?: string;
};

export function useAuditEntries(filter: AuditFilter = {}) {
  return useQuery({
    queryKey: [
      "audit",
      filter.eventType,
      filter.actorId,
      filter.resourceType,
      filter.resourceId,
    ],
    queryFn: async () => {
      const res = await auditClient.listAuditEntries({
        eventType: filter.eventType ?? "",
        actorId: filter.actorId ?? "",
        resourceType: filter.resourceType ?? "",
        resourceId: filter.resourceId ?? "",
        dateFrom: filter.dateFrom ?? "",
        dateTo: filter.dateTo ?? "",
      });
      return res.entries;
    },
    staleTime: 30_000,
  });
}

export type { AuditEntry };
