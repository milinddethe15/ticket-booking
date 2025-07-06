-- Sample data for testing the ticket booking system
-- This script provides comprehensive test data for development

-- Clean up existing data (for development resets)
TRUNCATE TABLE bookings, tickets, events, users RESTART IDENTITY CASCADE;

-- Insert sample users
INSERT INTO users (name, email, phone) VALUES
('John Doe', 'john.doe@example.com', '+1-555-0101'),
('Jane Smith', 'jane.smith@example.com', '+1-555-0102'),
('Bob Johnson', 'bob.johnson@example.com', '+1-555-0103'),
('Alice Brown', 'alice.brown@example.com', '+1-555-0104'),
('Charlie Wilson', 'charlie.wilson@example.com', '+1-555-0105'),
('David Lee', 'david.lee@example.com', '+1-555-0106'),
('Emma Wilson', 'emma.wilson@example.com', '+1-555-0107'),
('Frank Miller', 'frank.miller@example.com', '+1-555-0108');

-- Insert comprehensive sample events for development testing
INSERT INTO events (name, description, venue, start_time, end_time, total_tickets, available_tickets, price) VALUES
('Tech Conference 2024', 'Annual technology conference with industry experts', 'Convention Center', NOW() + INTERVAL '30 days', NOW() + INTERVAL '31 days', 96, 96, 299.99),
('Rock Concert', 'Amazing rock band live performance', 'Music Hall', NOW() + INTERVAL '45 days', NOW() + INTERVAL '45 days' + INTERVAL '4 hours', 96, 96, 89.99),
('Comedy Night', 'Stand-up comedy show with famous comedians', 'Comedy Club', NOW() + INTERVAL '15 days', NOW() + INTERVAL '15 days' + INTERVAL '3 hours', 96, 96, 45.50),
('Food Festival', 'International food festival with various cuisines', 'City Park', NOW() + INTERVAL '60 days', NOW() + INTERVAL '62 days', 96, 96, 25.00),
('Art Exhibition', 'Modern art exhibition featuring local artists', 'Art Gallery', NOW() + INTERVAL '20 days', NOW() + INTERVAL '50 days', 96, 96, 15.00),
('Jazz Night', 'Smooth jazz performance with renowned artists', 'Blue Note Club', NOW() + INTERVAL '25 days', NOW() + INTERVAL '25 days' + INTERVAL '3 hours', 96, 96, 65.00),
('Theater Musical', 'Broadway-style musical performance', 'Grand Theater', NOW() + INTERVAL '40 days', NOW() + INTERVAL '40 days' + INTERVAL '2.5 hours', 96, 96, 120.00),
('Sports Championship', 'Local sports championship finals', 'Stadium Arena', NOW() + INTERVAL '35 days', NOW() + INTERVAL '35 days' + INTERVAL '3 hours', 96, 96, 75.00),
('Wine Tasting', 'Premium wine tasting event with expert sommeliers', 'Vineyard Estate', NOW() + INTERVAL '50 days', NOW() + INTERVAL '50 days' + INTERVAL '4 hours', 96, 96, 95.00),
('Gaming Convention', 'Annual gaming and esports convention', 'Expo Center', NOW() + INTERVAL '55 days', NOW() + INTERVAL '57 days', 96, 96, 45.00);

-- Generate tickets for all events systematically (96 seats each = 8 rows × 12 seats)
-- All events use the same simple seat pattern: S001 to S096

-- Tech Conference tickets (Event ID 1)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 1, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Rock Concert tickets (Event ID 2)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 2, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Comedy Night tickets (Event ID 3)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 3, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Food Festival tickets (Event ID 4)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 4, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Art Exhibition tickets (Event ID 5)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 5, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Jazz Night tickets (Event ID 6)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 6, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Theater Musical tickets (Event ID 7)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 7, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Sports Championship tickets (Event ID 8)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 8, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Wine Tasting tickets (Event ID 9)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 9, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Gaming Convention tickets (Event ID 10)
INSERT INTO tickets (event_id, seat_no, status) 
SELECT 10, 'S' || LPAD(generate_series(1, 96)::text, 3, '0'), 'available';

-- Sample bookings for testing different scenarios
-- Use subqueries to get correct ticket IDs for each event

-- Confirmed booking (Event 1: Tech Conference - first 2 available tickets)
INSERT INTO bookings (user_id, event_id, ticket_ids, quantity, total_amount, status, booking_ref, expires_at) 
SELECT 1, 1, ARRAY[t1.id, t2.id], 2, 599.98, 'confirmed', 'BK1234567890', NOW() + INTERVAL '1 hour'
FROM (SELECT id FROM tickets WHERE event_id = 1 ORDER BY id LIMIT 1) t1,
     (SELECT id FROM tickets WHERE event_id = 1 ORDER BY id LIMIT 1 OFFSET 1) t2;

-- Pending booking (Event 2: Rock Concert - first 3 available tickets)  
INSERT INTO bookings (user_id, event_id, ticket_ids, quantity, total_amount, status, booking_ref, expires_at)
SELECT 2, 2, ARRAY[t1.id, t2.id, t3.id], 3, 269.97, 'pending', 'BK1234567891', NOW() + INTERVAL '10 minutes'
FROM (SELECT id FROM tickets WHERE event_id = 2 ORDER BY id LIMIT 1) t1,
     (SELECT id FROM tickets WHERE event_id = 2 ORDER BY id LIMIT 1 OFFSET 1) t2,
     (SELECT id FROM tickets WHERE event_id = 2 ORDER BY id LIMIT 1 OFFSET 2) t3;

-- Update ticket statuses based on bookings
-- Confirmed booking - mark first 2 tickets of Event 1 as sold
UPDATE tickets SET status = 'sold' 
WHERE id IN (SELECT id FROM tickets WHERE event_id = 1 ORDER BY id LIMIT 2);
UPDATE events SET available_tickets = available_tickets - 2 WHERE id = 1;

-- Pending booking - mark first 3 tickets of Event 2 as reserved
UPDATE tickets SET status = 'reserved' 
WHERE id IN (SELECT id FROM tickets WHERE event_id = 2 ORDER BY id LIMIT 3);
UPDATE events SET available_tickets = available_tickets - 3 WHERE id = 2;

-- Add some variety - mark some random tickets as sold to simulate real usage
-- Comedy Night (Event 3) - mark 5 random tickets as sold
UPDATE tickets SET status = 'sold' WHERE id IN (
    SELECT id FROM tickets WHERE event_id = 3 AND status = 'available' ORDER BY RANDOM() LIMIT 5
);
UPDATE events SET available_tickets = available_tickets - 5 WHERE id = 3;

-- Food Festival (Event 4) - mark 8 random tickets as sold
UPDATE tickets SET status = 'sold' WHERE id IN (
    SELECT id FROM tickets WHERE event_id = 4 AND status = 'available' ORDER BY RANDOM() LIMIT 8
);
UPDATE events SET available_tickets = available_tickets - 8 WHERE id = 4;

-- Display summary
SELECT 
    e.id, 
    e.name, 
    e.total_tickets, 
    e.available_tickets,
    COUNT(t.id) as actual_tickets_in_db,
    e.price
FROM events e
LEFT JOIN tickets t ON e.id = t.event_id
GROUP BY e.id, e.name, e.total_tickets, e.available_tickets, e.price
ORDER BY e.id; 