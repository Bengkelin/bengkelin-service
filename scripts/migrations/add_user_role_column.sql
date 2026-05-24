-- Migration: Add role column to users table
-- This migration adds a role column for JWT-based admin authentication,
-- replacing the previous admin secret header pattern.

-- Add role column with default 'user'
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) NOT NULL DEFAULT 'user';

-- Create index on role for efficient role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- To promote a user to admin, run:
-- UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
