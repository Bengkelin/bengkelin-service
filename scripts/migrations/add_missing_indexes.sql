-- Missing Performance Indexes Migration
-- Created: 2026-05-08
-- Purpose: Add trigram indexes for ILIKE search, missing FK indexes, and composite indexes

-- ============================================
-- Prerequisites: pg_trgm extension for trigram indexes
-- ============================================
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- ============================================
-- Trigram Indexes for ILIKE Search Queries
-- ============================================

-- Used by SearchBengkelPublic: bengkel_name ILIKE '%query%'
CREATE INDEX IF NOT EXISTS idx_bengkels_bengkel_name_trgm
ON bengkels USING gin (bengkel_name gin_trgm_ops);

-- Used by SearchBengkelPublic: bengkel_services.nama_service ILIKE '%query%'
CREATE INDEX IF NOT EXISTS idx_bengkel_services_nama_service_trgm
ON bengkel_services USING gin (nama_service gin_trgm_ops);

-- Used by SearchBengkelPublic: bengkel_addresses.full_address ILIKE '%query%'
CREATE INDEX IF NOT EXISTS idx_bengkel_addresses_full_address_trgm
ON bengkel_addresses USING gin (full_address gin_trgm_ops);

-- Used by SearchBengkelPublic: bengkel_addresses.city ILIKE '%query%'
CREATE INDEX IF NOT EXISTS idx_bengkel_addresses_city_trgm
ON bengkel_addresses USING gin (city gin_trgm_ops);

-- Used by SearchBengkelPublic: bengkel_addresses.province ILIKE '%query%'
CREATE INDEX IF NOT EXISTS idx_bengkel_addresses_province_trgm
ON bengkel_addresses USING gin (province gin_trgm_ops);

-- ============================================
-- Missing Foreign Key Indexes
-- ============================================

-- Chat histories: queried by sender/receiver in GetAllChatHistoryPaginate
CREATE INDEX IF NOT EXISTS idx_chat_histories_sender_user_id
ON chat_histories(sender_user_id);

CREATE INDEX IF NOT EXISTS idx_chat_histories_receiver_user_id
ON chat_histories(receiver_user_id);

-- Orders: vehicle_id used in JOINs but has no index
CREATE INDEX IF NOT EXISTS idx_orders_vehicle_id
ON orders(vehicle_id);

-- Bengkel testimonials: order_id and user_id used in queries
CREATE INDEX IF NOT EXISTS idx_bengkel_testimonials_order_id
ON bengkel_testimonials(order_id);

CREATE INDEX IF NOT EXISTS idx_bengkel_testimonials_user_id
ON bengkel_testimonials(user_id);

-- ============================================
-- Composite Indexes for Frequent Query Patterns
-- ============================================

-- Chat histories: queried by (sender, receiver) pair with pagination
CREATE INDEX IF NOT EXISTS idx_chat_histories_sender_receiver
ON chat_histories(sender_user_id, receiver_user_id, created_at DESC);

-- Bengkel testimonials: queried by bengkel_id with user preloading
CREATE INDEX IF NOT EXISTS idx_bengkel_testimonials_bengkel_user
ON bengkel_testimonials(bengkel_id, user_id);

-- Orders: queried by vehicle_id with status filtering
CREATE INDEX IF NOT EXISTS idx_orders_vehicle_status
ON orders(vehicle_id, status, created_at DESC);

-- ============================================
-- Analyze tables after creating indexes
-- ============================================

ANALYZE bengkels;
ANALYZE bengkel_services;
ANALYZE bengkel_addresses;
ANALYZE chat_histories;
ANALYZE orders;
ANALYZE bengkel_testimonials;
