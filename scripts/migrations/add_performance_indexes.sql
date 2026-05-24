-- Performance Optimization Indexes
-- Created: 2026-02-28
-- Purpose: Fix performance bottlenecks identified in PERFORMANCE_ANALYSIS_REPORT.md

-- ============================================
-- Order Table Indexes
-- ============================================

-- Composite index for order queries by mitra (for /api/v1/bengkels/orders/list/mitra)
CREATE INDEX IF NOT EXISTS idx_orders_bengkel_status_created 
ON orders(bengkel_id, status, created_at DESC);

-- Index for order queries by user (for user order listings)
CREATE INDEX IF NOT EXISTS idx_orders_user_status_created 
ON orders(user_id, status, created_at DESC);

-- Index for pagination queries on orders
CREATE INDEX IF NOT EXISTS idx_orders_created_at 
ON orders(created_at DESC);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_orders_status 
ON orders(status);

-- ============================================
-- Mitra Table Indexes
-- ============================================

-- Index for mitra lookups by email
CREATE INDEX IF NOT EXISTS idx_mitra_email 
ON mitras(email);

-- Index for mitra lookups by ID with related data
CREATE INDEX IF NOT EXISTS idx_mitra_id 
ON mitras(id);

-- ============================================
-- Bengkel Table Indexes
-- ============================================

-- Index for bengkel lookups by mitra_id
CREATE INDEX IF NOT EXISTS idx_bengkel_mitra_id 
ON bengkels(mitra_id);

-- Index for bengkel name searches
CREATE INDEX IF NOT EXISTS idx_bengkel_name 
ON bengkels(bengkel_name);

-- ============================================
-- User Table Indexes
-- ============================================

-- Index for user lookups by email
CREATE INDEX IF NOT EXISTS idx_user_email 
ON users(email);

-- ============================================
-- Foreign Key Indexes (for better JOIN performance)
-- ============================================

-- Order services foreign key index
CREATE INDEX IF NOT EXISTS idx_order_services_order_id 
ON order_services(order_id);

-- Bengkel addresses foreign key index
CREATE INDEX IF NOT EXISTS idx_bengkel_addresses_bengkel_id 
ON bengkel_addresses(bengkel_id);

-- Bengkel photos foreign key index
CREATE INDEX IF NOT EXISTS idx_bengkel_photos_bengkel_id 
ON bengkel_photos(bengkel_id);

-- Bengkel operational foreign key index
CREATE INDEX IF NOT EXISTS idx_bengkel_operational_bengkel_id 
ON bengkel_operationals(bengkel_id);

-- Bengkel services foreign key index
CREATE INDEX IF NOT EXISTS idx_bengkel_services_bengkel_id 
ON bengkel_services(bengkel_id);

-- Vehicles foreign key index for user
CREATE INDEX IF NOT EXISTS idx_vehicles_user_id 
ON vehicles(user_id);

-- ============================================
-- GIN Indexes for JSON/Text Search (optional)
-- ============================================

-- GIN index for full address search in bengkel_addresses (if using PostgreSQL)
-- CREATE INDEX IF NOT EXISTS idx_bengkel_addresses_full_address_gin 
-- ON bengkel_addresses USING gin(to_tsvector('simple', full_address));

-- ============================================
-- Analyze tables after creating indexes
-- ============================================

ANALYZE orders;
ANALYZE mitras;
ANALYZE bengkels;
ANALYZE users;
ANALYZE order_services;
ANALYZE bengkel_addresses;
ANALYZE bengkel_photos;
ANALYZE bengkel_operationals;
ANALYZE bengkel_services;
ANALYZE vehicles;
