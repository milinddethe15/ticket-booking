-- Drop triggers
DROP TRIGGER IF EXISTS update_bookings_updated_at ON bookings;
DROP TRIGGER IF EXISTS update_tickets_updated_at ON tickets;
DROP TRIGGER IF EXISTS update_events_updated_at ON events;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_bookings_booking_ref;
DROP INDEX IF EXISTS idx_bookings_expires_at;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_event_id;
DROP INDEX IF EXISTS idx_bookings_user_id;
DROP INDEX IF EXISTS idx_tickets_event_id_status;
DROP INDEX IF EXISTS idx_events_available_tickets;
DROP INDEX IF EXISTS idx_events_start_time;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users; 