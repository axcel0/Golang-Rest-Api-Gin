-- Rollback role column from users table
-- Migration: add_role_to_users (down)
-- Created: 2025-10-31

-- Drop index
DROP INDEX IF EXISTS idx_users_role;

-- Drop role column
ALTER TABLE users DROP COLUMN IF EXISTS role;
