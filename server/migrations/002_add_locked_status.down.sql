-- Remove 'locked' status from tickets table
ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_status_check;
ALTER TABLE tickets ADD CONSTRAINT tickets_status_check CHECK (status IN ('available', 'reserved', 'sold')); 