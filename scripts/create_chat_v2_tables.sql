-- Chat V2 Database Migration
-- Creates all necessary tables for Chat V2 feature
-- Run this script to set up Chat V2 functionality

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
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_chat_rooms_v2_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_chat_rooms_v2_bengkel_id FOREIGN KEY (bengkel_id) REFERENCES bengkels(id) ON DELETE CASCADE
);

-- Create chat_messages_v2 table
CREATE TABLE IF NOT EXISTS chat_messages_v2 (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    sender_id VARCHAR(36) NOT NULL,
    sender_type VARCHAR(10) NOT NULL CHECK (sender_type IN ('user', 'mitra')),
    message_type VARCHAR(10) NOT NULL DEFAULT 'text' CHECK (message_type IN ('text', 'image', 'file', 'system')),
    content TEXT NOT NULL,
    file_url VARCHAR(500),
    file_name VARCHAR(255),
    file_size BIGINT,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP NULL,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP NULL,
    reply_to_id VARCHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraints
    CONSTRAINT fk_chat_messages_v2_room_id FOREIGN KEY (room_id) REFERENCES chat_rooms_v2(id) ON DELETE CASCADE,
    CONSTRAINT fk_chat_messages_v2_reply_to_id FOREIGN KEY (reply_to_id) REFERENCES chat_messages_v2(id) ON DELETE SET NULL
);

-- Create chat_participants_v2 table (for future group chat support)
CREATE TABLE IF NOT EXISTS chat_participants_v2 (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(36) NOT NULL,
    participant_id VARCHAR(36) NOT NULL,
    participant_type VARCHAR(10) NOT NULL CHECK (participant_type IN ('user', 'mitra')),
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_seen_at TIMESTAMP NULL,
    unread_count INTEGER DEFAULT 0,
    
    -- Foreign key constraints
    CONSTRAINT fk_chat_participants_v2_room_id FOREIGN KEY (room_id) REFERENCES chat_rooms_v2(id) ON DELETE CASCADE,
    
    -- Unique constraint to prevent duplicate participants
    CONSTRAINT uk_chat_participants_v2_room_participant UNIQUE (room_id, participant_id, participant_type)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_chat_rooms_v2_user_id ON chat_rooms_v2(user_id);
CREATE INDEX IF NOT EXISTS idx_chat_rooms_v2_bengkel_id ON chat_rooms_v2(bengkel_id);
CREATE INDEX IF NOT EXISTS idx_chat_rooms_v2_is_active ON chat_rooms_v2(is_active);
CREATE INDEX IF NOT EXISTS idx_chat_rooms_v2_last_message_at ON chat_rooms_v2(last_message_at);

CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_room_id ON chat_messages_v2(room_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_sender_id ON chat_messages_v2(sender_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_sender_type ON chat_messages_v2(sender_type);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_message_type ON chat_messages_v2(message_type);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_is_read ON chat_messages_v2(is_read);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_created_at ON chat_messages_v2(created_at);
CREATE INDEX IF NOT EXISTS idx_chat_messages_v2_reply_to_id ON chat_messages_v2(reply_to_id);

CREATE INDEX IF NOT EXISTS idx_chat_participants_v2_room_id ON chat_participants_v2(room_id);
CREATE INDEX IF NOT EXISTS idx_chat_participants_v2_participant_id ON chat_participants_v2(participant_id);
CREATE INDEX IF NOT EXISTS idx_chat_participants_v2_participant_type ON chat_participants_v2(participant_type);
CREATE INDEX IF NOT EXISTS idx_chat_participants_v2_is_active ON chat_participants_v2(is_active);

-- Create triggers for updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply triggers to tables
DROP TRIGGER IF EXISTS update_chat_rooms_v2_updated_at ON chat_rooms_v2;
CREATE TRIGGER update_chat_rooms_v2_updated_at
    BEFORE UPDATE ON chat_rooms_v2
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_chat_messages_v2_updated_at ON chat_messages_v2;
CREATE TRIGGER update_chat_messages_v2_updated_at
    BEFORE UPDATE ON chat_messages_v2
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE chat_rooms_v2 IS 'Chat rooms between users and bengkels for Chat V2 feature';
COMMENT ON TABLE chat_messages_v2 IS 'Messages within chat rooms for Chat V2 feature';
COMMENT ON TABLE chat_participants_v2 IS 'Participants in chat rooms (for future group chat support)';

COMMENT ON COLUMN chat_rooms_v2.room_name IS 'Unique room identifier in format: chat_{user_id}_{bengkel_id}';
COMMENT ON COLUMN chat_messages_v2.sender_type IS 'Type of sender: user or mitra';
COMMENT ON COLUMN chat_messages_v2.message_type IS 'Type of message: text, image, file, or system';
COMMENT ON COLUMN chat_messages_v2.file_size IS 'File size in bytes for file messages';
COMMENT ON COLUMN chat_participants_v2.unread_count IS 'Number of unread messages for this participant';

-- Insert initial data or setup (if needed)
-- This section can be used for any initial configuration

COMMIT;