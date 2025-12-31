-- =====================================================
-- Bengkelin Service Database Schema
-- Database: MySQL 8.0+ / PostgreSQL 15+
-- Version: 1.0.0
-- =====================================================

-- Set character set and collation for MySQL
-- SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

-- =====================================================
-- CORE USER ENTITIES
-- =====================================================

-- Users table (customers)
CREATE TABLE users (
    id VARCHAR(36) PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    phone_number VARCHAR(255),
    password VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Mitras table (workshop owners/partners)
CREATE TABLE mitras (
    id VARCHAR(36) PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone_number VARCHAR(255),
    bank_name VARCHAR(255),
    bank_number VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- =====================================================
-- BENGKEL (WORKSHOP) ENTITIES
-- =====================================================

-- Bengkels table (workshops)
CREATE TABLE bengkels (
    id VARCHAR(36) PRIMARY KEY,
    mitra_id VARCHAR(36) UNIQUE NOT NULL,
    bengkel_name VARCHAR(255) NOT NULL,
    bengkel_phone VARCHAR(255),
    jumlah_montir INT,
    home_service BOOLEAN DEFAULT FALSE,
    store_service BOOLEAN DEFAULT FALSE,
    is_open BOOLEAN DEFAULT TRUE,
    avatar_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkels_mitra 
        FOREIGN KEY (mitra_id) 
        REFERENCES mitras(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- Bengkel addresses
CREATE TABLE bengkel_addresses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bengkel_id VARCHAR(36) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    label VARCHAR(255),
    full_address TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkel_addresses_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- Bengkel services offered
CREATE TABLE bengkel_services (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bengkel_id VARCHAR(36) NOT NULL,
    nama_service VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkel_services_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- Bengkel operational hours
CREATE TABLE bengkel_operationals (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bengkel_id VARCHAR(36) NOT NULL,
    day VARCHAR(50) NOT NULL,
    opening_time TIME NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkel_operationals_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- Bengkel photo gallery
CREATE TABLE bengkel_photos (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bengkel_id VARCHAR(36) NOT NULL,
    photo_url VARCHAR(500) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkel_photos_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- =====================================================
-- USER-RELATED ENTITIES
-- =====================================================

-- User addresses
CREATE TABLE user_addresses (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    label VARCHAR(255),
    full_address TEXT NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_user_addresses_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- User vehicles
CREATE TABLE vehicles (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    vehicle_type VARCHAR(255) NOT NULL,
    vehicle_number VARCHAR(255) NOT NULL,
    vehicle_color VARCHAR(255),
    vehicle_photo VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_vehicles_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- =====================================================
-- ORDER MANAGEMENT ENTITIES
-- =====================================================

-- Orders table
CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    bengkel_id VARCHAR(36) NOT NULL,
    vehicle_id INT NOT NULL,
    status INT NOT NULL DEFAULT 0,
    is_home_service BOOLEAN,
    total_price FLOAT NOT NULL DEFAULT 0,
    admin_fee FLOAT DEFAULT 0,
    home_service_fee FLOAT DEFAULT 0,
    home_service_schedule VARCHAR(50),
    payment_method VARCHAR(50),
    note TEXT,
    confirmed_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    finished_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_orders_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_orders_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_orders_vehicle 
        FOREIGN KEY (vehicle_id) 
        REFERENCES vehicles(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- Order service items
CREATE TABLE order_services (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    detail TEXT,
    price FLOAT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_order_services_order 
        FOREIGN KEY (order_id) 
        REFERENCES orders(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE
);

-- =====================================================
-- REVIEW AND TESTIMONIAL ENTITIES
-- =====================================================

-- Bengkel testimonials and ratings
CREATE TABLE bengkel_testimonials (
    id INT AUTO_INCREMENT PRIMARY KEY,
    bengkel_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_bengkel_testimonials_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_bengkel_testimonials_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    -- Ensure one review per user per bengkel
    UNIQUE KEY unique_user_bengkel_review (user_id, bengkel_id)
);

-- =====================================================
-- AUTHENTICATION AND SECURITY ENTITIES
-- =====================================================

-- Refresh tokens for JWT authentication
CREATE TABLE refresh_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36),
    mitra_id VARCHAR(36),
    token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT fk_refresh_tokens_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_refresh_tokens_mitra 
        FOREIGN KEY (mitra_id) 
        REFERENCES mitras(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    -- Ensure token belongs to either user or mitra, not both
    CONSTRAINT check_token_owner 
        CHECK ((user_id IS NOT NULL AND mitra_id IS NULL) OR 
               (user_id IS NULL AND mitra_id IS NOT NULL))
);

-- =====================================================
-- COMMUNICATION ENTITIES
-- =====================================================

-- Chat history between users and bengkels
CREATE TABLE chat_histories (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    bengkel_id VARCHAR(36) NOT NULL,
    channel_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_chat_histories_user 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    CONSTRAINT fk_chat_histories_bengkel 
        FOREIGN KEY (bengkel_id) 
        REFERENCES bengkels(id) 
        ON UPDATE CASCADE 
        ON DELETE CASCADE,
    
    -- Ensure unique chat channel per user-bengkel pair
    UNIQUE KEY unique_user_bengkel_chat (user_id, bengkel_id)
);

-- =====================================================
-- SYSTEM CONFIGURATION ENTITIES
-- =====================================================

-- Admin fee configuration
CREATE TABLE admin_fees (
    id INT AUTO_INCREMENT PRIMARY KEY,
    fee_amount FLOAT NOT NULL,
    fee_type VARCHAR(50) NOT NULL DEFAULT 'fixed',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- =====================================================
-- INDEXES FOR PERFORMANCE OPTIMIZATION
-- =====================================================

-- User indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_phone ON users(phone_number);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);

-- Mitra indexes
CREATE INDEX idx_mitras_email ON mitras(email);
CREATE INDEX idx_mitras_phone ON mitras(phone_number);
CREATE INDEX idx_mitras_deleted_at ON mitras(deleted_at);

-- Bengkel indexes
CREATE INDEX idx_bengkels_mitra_id ON bengkels(mitra_id);
CREATE INDEX idx_bengkels_name ON bengkels(bengkel_name);
CREATE INDEX idx_bengkels_home_service ON bengkels(home_service);
CREATE INDEX idx_bengkels_store_service ON bengkels(store_service);
CREATE INDEX idx_bengkels_is_open ON bengkels(is_open);

-- Address indexes for geolocation queries
CREATE INDEX idx_user_addresses_location ON user_addresses(latitude, longitude);
CREATE INDEX idx_user_addresses_user_id ON user_addresses(user_id);
CREATE INDEX idx_user_addresses_primary ON user_addresses(is_primary);
CREATE INDEX idx_bengkel_addresses_location ON bengkel_addresses(latitude, longitude);
CREATE INDEX idx_bengkel_addresses_bengkel_id ON bengkel_addresses(bengkel_id);

-- Vehicle indexes
CREATE INDEX idx_vehicles_user_id ON vehicles(user_id);
CREATE INDEX idx_vehicles_number ON vehicles(vehicle_number);

-- Service indexes
CREATE INDEX idx_bengkel_services_bengkel_id ON bengkel_services(bengkel_id);
CREATE INDEX idx_bengkel_services_name ON bengkel_services(nama_service);

-- Operational hours indexes
CREATE INDEX idx_bengkel_operationals_bengkel_id ON bengkel_operationals(bengkel_id);
CREATE INDEX idx_bengkel_operationals_day ON bengkel_operationals(day);

-- Order indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_bengkel_id ON orders(bengkel_id);
CREATE INDEX idx_orders_vehicle_id ON orders(vehicle_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_home_service ON orders(is_home_service);

-- Order service indexes
CREATE INDEX idx_order_services_order_id ON order_services(order_id);

-- Testimonial indexes
CREATE INDEX idx_bengkel_testimonials_bengkel_id ON bengkel_testimonials(bengkel_id);
CREATE INDEX idx_bengkel_testimonials_user_id ON bengkel_testimonials(user_id);
CREATE INDEX idx_bengkel_testimonials_rating ON bengkel_testimonials(rating);

-- Refresh token indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_mitra_id ON refresh_tokens(mitra_id);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens(is_revoked);
CREATE INDEX idx_refresh_tokens_deleted_at ON refresh_tokens(deleted_at);

-- Chat history indexes
CREATE INDEX idx_chat_histories_user_id ON chat_histories(user_id);
CREATE INDEX idx_chat_histories_bengkel_id ON chat_histories(bengkel_id);

-- =====================================================
-- SAMPLE DATA FOR TESTING
-- =====================================================

-- Sample admin fee configuration
INSERT INTO admin_fees (fee_amount, fee_type, is_active) VALUES 
(5000, 'fixed', TRUE),
(2.5, 'percentage', FALSE);

-- Sample user (password is hashed version of 'password123')
INSERT INTO users (id, first_name, last_name, email, phone_number, password) VALUES 
('550e8400-e29b-41d4-a716-446655440001', 'John', 'Doe', 'john.doe@example.com', '+6281234567890', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VJzQVqKF2');

-- Sample mitra (password is hashed version of 'password123')
INSERT INTO mitras (id, first_name, last_name, email, phone_number, password, bank_name, bank_number) VALUES 
('550e8400-e29b-41d4-a716-446655440002', 'Ahmad', 'Bengkel', 'ahmad@bengkelmakmur.com', '+6281234567891', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VJzQVqKF2', 'BCA', '1234567890');

-- Sample bengkel
INSERT INTO bengkels (id, mitra_id, bengkel_name, bengkel_phone, jumlah_montir, home_service, store_service, is_open) VALUES 
('550e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', 'Bengkel Makmur', '+6281234567892', 5, TRUE, TRUE, TRUE);

-- Sample bengkel address
INSERT INTO bengkel_addresses (bengkel_id, latitude, longitude, label, full_address) VALUES 
('550e8400-e29b-41d4-a716-446655440003', -6.2088, 106.8456, 'Main Workshop', 'Jl. Sudirman No. 123, Jakarta Pusat, DKI Jakarta 10220');

-- Sample bengkel services
INSERT INTO bengkel_services (bengkel_id, nama_service) VALUES 
('550e8400-e29b-41d4-a716-446655440003', 'Ganti Oli'),
('550e8400-e29b-41d4-a716-446655440003', 'Tune Up'),
('550e8400-e29b-41d4-a716-446655440003', 'Service AC'),
('550e8400-e29b-41d4-a716-446655440003', 'Ganti Ban');

-- Sample operational hours
INSERT INTO bengkel_operationals (bengkel_id, day, opening_time) VALUES 
('550e8400-e29b-41d4-a716-446655440003', 'Senin', '08:00:00'),
('550e8400-e29b-41d4-a716-446655440003', 'Selasa', '08:00:00'),
('550e8400-e29b-41d4-a716-446655440003', 'Rabu', '08:00:00'),
('550e8400-e29b-41d4-a716-446655440003', 'Kamis', '08:00:00'),
('550e8400-e29b-41d4-a716-446655440003', 'Jumat', '08:00:00'),
('550e8400-e29b-41d4-a716-446655440003', 'Sabtu', '08:00:00');

-- Sample user address
INSERT INTO user_addresses (user_id, latitude, longitude, label, full_address, is_primary) VALUES 
('550e8400-e29b-41d4-a716-446655440001', -6.2297, 106.8261, 'Home', 'Jl. Thamrin No. 456, Jakarta Pusat, DKI Jakarta 10230', TRUE);

-- Sample vehicle
INSERT INTO vehicles (user_id, vehicle_type, vehicle_number, vehicle_color) VALUES 
('550e8400-e29b-41d4-a716-446655440001', 'Mobil', 'B 1234 ABC', 'Putih');

-- Sample order
INSERT INTO orders (id, user_id, bengkel_id, vehicle_id, status, is_home_service, total_price, admin_fee, home_service_fee, payment_method, note) VALUES 
('550e8400-e29b-41d4-a716-446655440004', '550e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440003', 1, 0, FALSE, 150000, 5000, 0, 'cash', 'Ganti oli dan filter');

-- Sample order services
INSERT INTO order_services (order_id, title, detail, price) VALUES 
('550e8400-e29b-41d4-a716-446655440004', 'Ganti Oli Mesin', 'Oli SAE 10W-40 + Filter', 120000),
('550e8400-e29b-41d4-a716-446655440004', 'Cek Kondisi Umum', 'Pemeriksaan rutin kendaraan', 30000);

-- =====================================================
-- VIEWS FOR COMMON QUERIES
-- =====================================================

-- View for bengkel with complete information
CREATE VIEW bengkel_complete AS
SELECT 
    b.id,
    b.bengkel_name,
    b.bengkel_phone,
    b.jumlah_montir,
    b.home_service,
    b.store_service,
    b.is_open,
    b.avatar_url,
    m.first_name as mitra_first_name,
    m.last_name as mitra_last_name,
    m.email as mitra_email,
    m.phone_number as mitra_phone,
    ba.full_address,
    ba.latitude,
    ba.longitude,
    AVG(bt.rating) as average_rating,
    COUNT(bt.id) as total_reviews,
    b.created_at,
    b.updated_at
FROM bengkels b
LEFT JOIN mitras m ON b.mitra_id = m.id
LEFT JOIN bengkel_addresses ba ON b.id = ba.bengkel_id
LEFT JOIN bengkel_testimonials bt ON b.id = bt.bengkel_id
GROUP BY b.id, ba.id;

-- View for order summary
CREATE VIEW order_summary AS
SELECT 
    o.id,
    o.status,
    o.total_price,
    o.admin_fee,
    o.home_service_fee,
    o.is_home_service,
    o.payment_method,
    o.created_at,
    o.confirmed_at,
    o.finished_at,
    u.first_name as customer_first_name,
    u.last_name as customer_last_name,
    u.email as customer_email,
    b.bengkel_name,
    v.vehicle_type,
    v.vehicle_number,
    COUNT(os.id) as service_count
FROM orders o
LEFT JOIN users u ON o.user_id = u.id
LEFT JOIN bengkels b ON o.bengkel_id = b.id
LEFT JOIN vehicles v ON o.vehicle_id = v.id
LEFT JOIN order_services os ON o.id = os.order_id
GROUP BY o.id;

-- =====================================================
-- STORED PROCEDURES FOR COMMON OPERATIONS
-- =====================================================

-- Procedure to calculate order total with fees
DELIMITER //
CREATE PROCEDURE CalculateOrderTotal(
    IN order_id VARCHAR(36),
    IN is_home_service BOOLEAN,
    OUT total_amount FLOAT
)
BEGIN
    DECLARE service_total FLOAT DEFAULT 0;
    DECLARE admin_fee_amount FLOAT DEFAULT 0;
    DECLARE home_service_fee_amount FLOAT DEFAULT 0;
    
    -- Calculate service total
    SELECT COALESCE(SUM(price), 0) INTO service_total
    FROM order_services 
    WHERE order_services.order_id = order_id;
    
    -- Get admin fee
    SELECT fee_amount INTO admin_fee_amount
    FROM admin_fees 
    WHERE is_active = TRUE 
    LIMIT 1;
    
    -- Calculate home service fee (10% of service total if home service)
    IF is_home_service THEN
        SET home_service_fee_amount = service_total * 0.1;
    END IF;
    
    -- Calculate total
    SET total_amount = service_total + COALESCE(admin_fee_amount, 0) + home_service_fee_amount;
    
    -- Update order with calculated amounts
    UPDATE orders 
    SET 
        total_price = total_amount,
        admin_fee = COALESCE(admin_fee_amount, 0),
        home_service_fee = home_service_fee_amount
    WHERE id = order_id;
END //
DELIMITER ;

-- =====================================================
-- TRIGGERS FOR DATA INTEGRITY
-- =====================================================

-- Trigger to ensure only one primary address per user
DELIMITER //
CREATE TRIGGER ensure_single_primary_address
BEFORE INSERT ON user_addresses
FOR EACH ROW
BEGIN
    IF NEW.is_primary = TRUE THEN
        UPDATE user_addresses 
        SET is_primary = FALSE 
        WHERE user_id = NEW.user_id AND is_primary = TRUE;
    END IF;
END //
DELIMITER ;

-- Trigger to update order timestamps
DELIMITER //
CREATE TRIGGER update_order_timestamps
BEFORE UPDATE ON orders
FOR EACH ROW
BEGIN
    IF NEW.status = 1 AND OLD.status = 0 THEN
        SET NEW.confirmed_at = CURRENT_TIMESTAMP;
    END IF;
    
    IF NEW.status = 3 AND OLD.status != 3 THEN
        SET NEW.finished_at = CURRENT_TIMESTAMP;
    END IF;
END //
DELIMITER ;

-- =====================================================
-- CLEANUP PROCEDURES
-- =====================================================

-- Procedure to clean up expired refresh tokens
DELIMITER //
CREATE PROCEDURE CleanupExpiredTokens()
BEGIN
    UPDATE refresh_tokens 
    SET is_revoked = TRUE, deleted_at = CURRENT_TIMESTAMP
    WHERE expires_at < CURRENT_TIMESTAMP AND is_revoked = FALSE;
END //
DELIMITER ;

-- =====================================================
-- GRANTS AND PERMISSIONS
-- =====================================================

-- Create application user with limited permissions
-- CREATE USER 'bengkelin_app'@'%' IDENTIFIED BY 'secure_password_here';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON bengkelin_db.* TO 'bengkelin_app'@'%';
-- GRANT EXECUTE ON PROCEDURE bengkelin_db.CalculateOrderTotal TO 'bengkelin_app'@'%';
-- GRANT EXECUTE ON PROCEDURE bengkelin_db.CleanupExpiredTokens TO 'bengkelin_app'@'%';
-- FLUSH PRIVILEGES;

-- =====================================================
-- END OF SCHEMA
-- =====================================================