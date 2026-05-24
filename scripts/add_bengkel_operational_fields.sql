-- Migration script to add JamTutup and IsActive fields to bengkel_operationals table

-- Add JamTutup column
ALTER TABLE bengkel_operationals 
ADD COLUMN jam_tutup VARCHAR(20) DEFAULT '17:00';

-- Add IsActive column
ALTER TABLE bengkel_operationals 
ADD COLUMN is_active BOOLEAN DEFAULT true;

-- Update existing records to have default closing time and active status
UPDATE bengkel_operationals 
SET jam_tutup = '17:00', is_active = true 
WHERE jam_tutup IS NULL OR is_active IS NULL;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_bengkel_operationals_is_active ON bengkel_operationals(is_active);
CREATE INDEX IF NOT EXISTS idx_bengkel_operationals_bengkel_id_hari ON bengkel_operationals(bengkel_id, hari);