-- Add role column to users table
-- Migration: add_role_to_users
-- Created: 2025-10-31

-- Add role column with default value 'user'
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user' NOT NULL;

-- Create index on role for faster queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Update existing users to have 'user' role if NULL
UPDATE users SET role = 'user' WHERE role IS NULL OR role = '';

-- Add comment
COMMENT ON COLUMN users.role IS 'User role: superadmin, admin, user';
