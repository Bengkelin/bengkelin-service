-- Add cancellation fields to orders table
-- Run this migration to support order cancellation tracking

ALTER TABLE orders 
ADD COLUMN IF NOT EXISTS cancelled_by VARCHAR(36),
ADD COLUMN IF NOT EXISTS cancelled_reason VARCHAR(50),
ADD COLUMN IF NOT EXISTS cancelled_at TIMESTAMP NULL;

-- Add index for better query performance
CREATE INDEX IF NOT EXISTS idx_orders_cancelled_by ON orders(cancelled_by);
CREATE INDEX IF NOT EXISTS idx_orders_cancelled_at ON orders(cancelled_at);

-- Add comment for documentation
COMMENT ON COLUMN orders.cancelled_by IS 'User ID or Mitra ID who cancelled the order';
COMMENT ON COLUMN orders.cancelled_reason IS 'Reason for cancellation (enum: cancelled_by_user, cancelled_by_mitra, etc.)';
COMMENT ON COLUMN orders.cancelled_at IS 'Timestamp when order was cancelled';