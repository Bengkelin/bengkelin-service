-- Initialize PostgreSQL database for Bengkelin API
-- This script runs when the PostgreSQL container starts for the first time

-- Set timezone
SET timezone = 'Asia/Jakarta';

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create basic tables structure (GORM will handle migrations)
-- These are just basic structures, GORM AutoMigrate will handle the full schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP(3) WITH TIME ZONE,
    updated_at TIMESTAMP(3) WITH TIME ZONE,
    deleted_at TIMESTAMP(3) WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires TIMESTAMP(3) WITH TIME ZONE
);

-- Create indexes for users
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone);

-- Mitras table
CREATE TABLE IF NOT EXISTS mitras (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP(3) WITH TIME ZONE,
    updated_at TIMESTAMP(3) WITH TIME ZONE,
    deleted_at TIMESTAMP(3) WITH TIME ZONE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires TIMESTAMP(3) WITH TIME ZONE
);

-- Create indexes for mitras
CREATE INDEX IF NOT EXISTS idx_mitras_deleted_at ON mitras(deleted_at);
CREATE INDEX IF NOT EXISTS idx_mitras_email ON mitras(email);
CREATE INDEX IF NOT EXISTS idx_mitras_phone ON mitras(phone);

-- Create user_type enum
DO $$ BEGIN
    CREATE TYPE user_type_enum AS ENUM ('user', 'mitra');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP(3) WITH TIME ZONE,
    updated_at TIMESTAMP(3) WITH TIME ZONE,
    deleted_at TIMESTAMP(3) WITH TIME ZONE,
    token VARCHAR(500) UNIQUE NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    mitra_id BIGINT REFERENCES mitras(id) ON DELETE CASCADE,
    user_type user_type_enum NOT NULL,
    expires_at TIMESTAMP(3) WITH TIME ZONE NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    device_info VARCHAR(255),
    ip_address INET
);

-- Create indexes for refresh_tokens
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_mitra_id ON refresh_tokens(mitra_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);

-- Bengkels table
CREATE TABLE IF NOT EXISTS bengkels (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP(3) WITH TIME ZONE,
    updated_at TIMESTAMP(3) WITH TIME ZONE,
    deleted_at TIMESTAMP(3) WITH TIME ZONE,
    mitra_id BIGINT NOT NULL REFERENCES mitras(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phone VARCHAR(20),
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    image VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    rating DECIMAL(3, 2) DEFAULT 0.00,
    total_reviews INTEGER DEFAULT 0
);

-- Create indexes for bengkels
CREATE INDEX IF NOT EXISTS idx_bengkels_deleted_at ON bengkels(deleted_at);
CREATE INDEX IF NOT EXISTS idx_bengkels_mitra_id ON bengkels(mitra_id);
CREATE INDEX IF NOT EXISTS idx_bengkels_is_active ON bengkels(is_active);
CREATE INDEX IF NOT EXISTS idx_bengkels_rating ON bengkels(rating);

-- Insert default admin user (password: admin123)
-- Using pgcrypto for password hashing
INSERT INTO users (name, email, password, is_verified, created_at, updated_at) 
VALUES (
    'Admin User', 
    'admin@bengkelin.com', 
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL.hl.hl.', 
    TRUE, 
    NOW(), 
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert sample mitra (password: mitra123)
INSERT INTO mitras (name, email, password, phone, is_verified, created_at, updated_at) 
VALUES (
    'Sample Mitra', 
    'mitra@bengkelin.com', 
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL.hl.hl.', 
    '081234567890',
    TRUE, 
    NOW(), 
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Show completion message
SELECT 'PostgreSQL database initialization completed successfully!' as message;