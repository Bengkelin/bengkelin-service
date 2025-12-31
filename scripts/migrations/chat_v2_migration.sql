-- Chat V2 Database Migration
-- This script creates the necessary tables for the Chat V2 system

-- Create chat_rooms_v2 table
CREATE TABLE IF NOT EXISTS chat_rooms_v2 (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    bengkel_id VARCHAR(36) NOT NULL,
    room_name VARCHAR(255) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT TRUE,
    last_message TEXT,
    last_message_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_chat_rooms_v2_user_id (user_id),
    INDEX idx_chat_rooms_v2_bengkel_id (bengkel_id),
    INDEX idx_chat_rooms_v2_last_message_at (last_message_at),
    INDEX idx_chat_rooms_v2_active (is_active),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (bengkel_id) REFERENCES bengkels(id) ON DELETE CASCADE
);

-- Create chat_messages_v2 table
CREATE TABLE IF NOT EXISTS chat_messages_v2 (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    sender_id VARCHAR(36) NOT NULL,
    sender_type ENUM('user', 'mitra') NOT NULL,
    message_type ENUM('text', 'image', 'file', 'system') DEFAULT 'text',
    content TEXT NOT NULL,
    file_url VARCHAR(500) NULL,
    file_name VARCHAR(255) NULL,
    file_size BIGINT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP NULL,
    reply_to_id VARCHAR(36) NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_chat_messages_v2_room_id (room_id),
    INDEX idx_chat_messages_v2_sender_id (sender_id),
    INDEX idx_chat_messages_v2_sender_type (sender_type),
    INDEX idx_chat_messages_v2_created_at (created_at),
    INDEX idx_chat_messages_v2_is_read (is_read),
    INDEX idx_chat_messages_v2_message_type (message_type),
    
    FOREIGN KEY (room_id) REFERENCES chat_rooms_v2(id) ON DELETE CASCADE,
    FOREIGN KEY (reply_to_id) REFERENCES chat_messages_v2(id) ON DELETE SET NULL
);

-- Create chat_participants_v2 table
CREATE TABLE IF NOT EXISTS chat_participants_v2 (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    participant_id VARCHAR(36) NOT NULL,
    participant_type ENUM('user', 'mitra') NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_seen_at TIMESTAMP NULL,
    unread_count INT DEFAULT 0,
    
    INDEX idx_chat_participants_v2_room_id (room_id),
    INDEX idx_chat_participants_v2_participant_id (participant_id),
    INDEX idx_chat_participants_v2_participant_type (participant_type),
    INDEX idx_chat_participants_v2_active (is_active),
    
    UNIQUE KEY unique_room_participant (room_id, participant_id, participant_type),
    FOREIGN KEY (room_id) REFERENCES chat_rooms_v2(id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_chat_rooms_v2_composite ON chat_rooms_v2(user_id, bengkel_id, is_active);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_room_sender ON chat_messages_v2(room_id, sender_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_unread ON chat_messages_v2(room_id, is_read, sender_id);

-- Add triggers to update updated_at timestamps
DELIMITER $$

CREATE TRIGGER IF NOT EXISTS chat_rooms_v2_updated_at
    BEFORE UPDATE ON chat_rooms_v2
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

CREATE TRIGGER IF NOT EXISTS chat_messages_v2_updated_at
    BEFORE UPDATE ON chat_messages_v2
    FOR EACH ROW
BEGIN
    SET NEW.updated_at = CURRENT_TIMESTAMP;
END$$

DELIMITER ;

-- Insert sample data for testing (optional)
-- Uncomment the following lines if you want to insert test data

/*
-- Sample chat room
INSERT INTO chat_rooms_v2 (id, user_id, bengkel_id, room_name, is_active, created_at, updated_at) 
VALUES (
    'sample-room-id-1', 
    'sample-user-id-1', 
    'sample-bengkel-id-1', 
    'chat_sample-user-id-1_sample-bengkel-id-1', 
    TRUE, 
    NOW(), 
    NOW()
) ON DUPLICATE KEY UPDATE id=id;

-- Sample participants
INSERT INTO chat_participants_v2 (id, room_id, participant_id, participant_type, joined_at, is_active, unread_count) 
VALUES 
    ('sample-participant-1', 'sample-room-id-1', 'sample-user-id-1', 'user', NOW(), TRUE, 0),
    ('sample-participant-2', 'sample-room-id-1', 'sample-mitra-id-1', 'mitra', NOW(), TRUE, 0)
ON DUPLICATE KEY UPDATE id=id;

-- Sample messages
INSERT INTO chat_messages_v2 (id, room_id, sender_id, sender_type, message_type, content, is_read, created_at, updated_at) 
VALUES 
    ('sample-message-1', 'sample-room-id-1', 'sample-user-id-1', 'user', 'text', 'Hello, I need help with my car!', FALSE, NOW(), NOW()),
    ('sample-message-2', 'sample-room-id-1', 'sample-mitra-id-1', 'mitra', 'text', 'Hi! I would be happy to help. What seems to be the problem?', FALSE, NOW(), NOW())
ON DUPLICATE KEY UPDATE id=id;
*/

-- Update last_message in chat room based on latest message
UPDATE chat_rooms_v2 cr
SET 
    last_message = (
        SELECT content 
        FROM chat_messages_v2 cm 
        WHERE cm.room_id = cr.id 
        ORDER BY cm.created_at DESC 
        LIMIT 1
    ),
    last_message_at = (
        SELECT created_at 
        FROM chat_messages_v2 cm 
        WHERE cm.room_id = cr.id 
        ORDER BY cm.created_at DESC 
        LIMIT 1
    )
WHERE EXISTS (
    SELECT 1 
    FROM chat_messages_v2 cm 
    WHERE cm.room_id = cr.id
);

-- Create a view for easy querying of room information with participant details
CREATE OR REPLACE VIEW chat_rooms_v2_with_participants AS
SELECT 
    cr.*,
    GROUP_CONCAT(
        CONCAT(cp.participant_id, ':', cp.participant_type, ':', cp.is_active)
        SEPARATOR ','
    ) as participants
FROM chat_rooms_v2 cr
LEFT JOIN chat_participants_v2 cp ON cr.id = cp.room_id
GROUP BY cr.id;

-- Create a view for message statistics
CREATE OR REPLACE VIEW chat_message_stats_v2 AS
SELECT 
    room_id,
    COUNT(*) as total_messages,
    COUNT(CASE WHEN is_read = FALSE THEN 1 END) as unread_messages,
    COUNT(CASE WHEN message_type = 'text' THEN 1 END) as text_messages,
    COUNT(CASE WHEN message_type = 'image' THEN 1 END) as image_messages,
    COUNT(CASE WHEN message_type = 'file' THEN 1 END) as file_messages,
    MAX(created_at) as last_message_time,
    MIN(created_at) as first_message_time
FROM chat_messages_v2
GROUP BY room_id;

COMMIT;