-- Migration script to add Description, Price, and IsAvailable fields to bengkel_services table

-- Add Description column
ALTER TABLE bengkel_services 
ADD COLUMN description TEXT DEFAULT '';

-- Add Price column
ALTER TABLE bengkel_services 
ADD COLUMN price DECIMAL(10,2) DEFAULT 0.00;

-- Add IsAvailable column
ALTER TABLE bengkel_services 
ADD COLUMN is_available BOOLEAN DEFAULT true;

-- Update existing records to have default values
UPDATE bengkel_services 
SET description = '', price = 0.00, is_available = true 
WHERE description IS NULL OR price IS NULL OR is_available IS NULL;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_bengkel_services_is_available ON bengkel_services(is_available);
CREATE INDEX IF NOT EXISTS idx_bengkel_services_price ON bengkel_services(price);
CREATE INDEX IF NOT EXISTS idx_bengkel_services_bengkel_id_available ON bengkel_services(bengkel_id, is_available);