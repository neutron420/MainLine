-- 006_remove_last_introspected_at.sql
-- last_introspected_at was never written by any INSERT/UPDATE; schema
-- versions carry introspection timing via schema_versions.created_at.
ALTER TABLE schemas DROP COLUMN IF EXISTS last_introspected_at;
