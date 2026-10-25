-- ============================================
-- RESET: Stock Opname Rita
-- DESTRUCTIVE. Dev/test only.
--
-- Drops every object in schema public (tables, functions, triggers,
-- indexes, sequences, views) and recreates an empty public schema.
-- Does NOT drop the database itself, so schema1.sql can run next
-- on the same connection.
--
-- Usage:
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/reset.sql
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/schema1.sql
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/seed.sql
--
-- Stop the backend first so no connection is holding locks.
-- ============================================

BEGIN;

SET LOCAL lock_timeout = '10s';

DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

-- Restore the usual read access for the app role. The migration role
-- remains the owner and keeps CREATE.
GRANT USAGE ON SCHEMA public TO PUBLIC;

COMMIT;
