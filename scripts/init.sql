-- Initialize database for Bengkelin API
-- This script runs when the MySQL container starts for the first time

-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS bengkelin_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Use the database
USE bengkelin_db;

-- Create user if it doesn't exist (MySQL 8.0+ syntax)
CREATE USER IF NOT EXISTS 'bengkelin_user'@'%' IDENTIFIED BY 'bengkelin_password';

-- Grant privileges
GRANT ALL PRIVILEGES ON bengkelin_db.* TO 'bengkelin_user'@'%';

-- Flush privileges
FLUSH PRIVILEGES;

-- Set timezone
SET time_zone = '+07:00';

-- Create basic tables structure (GORM will handle migrations)
-- These are just basic structures, GORM AutoMigrate will handle the full schema

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires DATETIME(3),
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_email (email),
    INDEX idx_users_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Mitras table
CREATE TABLE IF NOT EXISTS mitras (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone VARCHAR(20),
    avatar VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    reset_token VARCHAR(255),
    reset_token_expires DATETIME(3),
    INDEX idx_mitras_deleted_at (deleted_at),
    INDEX idx_mitras_email (email),
    INDEX idx_mitras_phone (phone)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    token VARCHAR(500) UNIQUE NOT NULL,
    user_id BIGINT UNSIGNED,
    mitra_id BIGINT UNSIGNED,
    user_type ENUM('user', 'mitra') NOT NULL,
    expires_at DATETIME(3) NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    device_info VARCHAR(255),
    ip_address VARCHAR(45),
    INDEX idx_refresh_tokens_deleted_at (deleted_at),
    INDEX idx_refresh_tokens_token (token),
    INDEX idx_refresh_tokens_user_id (user_id),
    INDEX idx_refresh_tokens_mitra_id (mitra_id),
    INDEX idx_refresh_tokens_expires_at (expires_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (mitra_id) REFERENCES mitras(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Bengkels table
CREATE TABLE IF NOT EXISTS bengkels (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    mitra_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    phone VARCHAR(20),
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    image VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    rating DECIMAL(3, 2) DEFAULT 0.00,
    total_reviews INT DEFAULT 0,
    INDEX idx_bengkels_deleted_at (deleted_at),
    INDEX idx_bengkels_mitra_id (mitra_id),
    INDEX idx_bengkels_is_active (is_active),
    INDEX idx_bengkels_rating (rating),
    FOREIGN KEY (mitra_id) REFERENCES mitras(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user (password: admin123)
INSERT IGNORE INTO users (name, email, password, is_verified, created_at, updated_at) 
VALUES (
    'Admin User', 
    'admin@bengkelin.com', 
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL.hl.hl.', 
    TRUE, 
    NOW(), 
    NOW()
);

-- Insert sample mitra (password: mitra123)
INSERT IGNORE INTO mitras (name, email, password, phone, is_verified, created_at, updated_at) 
VALUES (
    'Sample Mitra', 
    'mitra@bengkelin.com', 
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL.hl.hl.', 
    '081234567890',
    TRUE, 
    NOW(), 
    NOW()
);

-- Show completion message
SELECT 'Database initialization completed successfully!' as message;