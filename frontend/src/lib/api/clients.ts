import { createClient, type Client } from "@connectrpc/connect";

import { AuditService } from "@/lib/gen/audit/v1/audit_service_pb";
import { AuthService } from "@/lib/gen/auth/v1/auth_service_pb";
import { DriftService } from "@/lib/gen/drift/v1/drift_service_pb";
import { EventService } from "@/lib/gen/event/v1/event_service_pb";
import { MigrationService } from "@/lib/gen/migration/v1/migration_service_pb";
import { ProjectService } from "@/lib/gen/project/v1/project_service_pb";
import { SchemaService } from "@/lib/gen/schema/v1/schema_service_pb";

import { transport } from "./transport";

export const authClient = createClient(AuthService, transport);
export const projectClient = createClient(ProjectService, transport);
export const schemaClient = createClient(SchemaService, transport);
export const migrationClient = createClient(MigrationService, transport);
export const driftClient = createClient(DriftService, transport);
export const auditClient = createClient(AuditService, transport);
export const eventClient = createClient(EventService, transport);

export type { Client };
