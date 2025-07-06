package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/config"
	"github.com/milinddethe15/ticket-booking/internal/db"
	"github.com/milinddethe15/ticket-booking/internal/models"
)

type BookingRepository struct {
	db     *db.DB
	logger *logrus.Logger
	config *config.Config
}

func NewBookingRepository(database *db.DB, logger *logrus.Logger, cfg *config.Config) *BookingRepository {
	return &BookingRepository{
		db:     database,
		logger: logger,
		config: cfg,
	}
}

// BookTickets implements pessimistic locking for concurrent ticket booking
func (r *BookingRepository) BookTickets(ctx context.Context, request *models.BookingRequest) (*models.Booking, error) {
	var booking *models.Booking

	err := r.db.WithRetry(ctx, 3, 100*time.Millisecond, func() error {
		return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
			var err error
			booking, err = r.bookTicketsWithLock(ctx, tx, request)
			return err
		})
	})

	return booking, err
}

func (r *BookingRepository) bookTicketsWithLock(ctx context.Context, tx *sql.Tx, request *models.BookingRequest) (*models.Booking, error) {
	// Step 1: Lock the event row for update (pessimistic lock)
	var event models.Event
	query := `
		SELECT id, name, available_tickets, price, start_time 
		FROM events 
		WHERE id = $1 
		FOR UPDATE`

	err := tx.QueryRowContext(ctx, query, request.EventID).Scan(
		&event.ID,
		&event.Name,
		&event.AvailableTickets,
		&event.Price,
		&event.StartTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to lock event: %w", err)
	}

	// Step 2: Validate event timing
	if time.Now().After(event.StartTime) {
		return nil, fmt.Errorf("event has already started")
	}

	// Step 3: Check if user has enough locked seats for this booking
	// Note: We'll verify exact count after selecting locked tickets

	// Step 4: Lock and select locked tickets (user's selection)
	ticketQuery := `
		SELECT id, seat_no 
		FROM tickets 
		WHERE event_id = $1 AND status = 'locked' 
		ORDER BY seat_no 
		LIMIT $2 
		FOR UPDATE`

	rows, err := tx.QueryContext(ctx, ticketQuery, request.EventID, request.Quantity)
	if err != nil {
		return nil, fmt.Errorf("failed to select tickets: %w", err)
	}
	defer rows.Close()

	var ticketIDs []int
	var seatNumbers []string

	for rows.Next() {
		var ticketID int
		var seatNo string
		if err := rows.Scan(&ticketID, &seatNo); err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}
		ticketIDs = append(ticketIDs, ticketID)
		seatNumbers = append(seatNumbers, seatNo)
	}

	if len(ticketIDs) < request.Quantity {
		return nil, fmt.Errorf("insufficient locked seats for booking. Found %d locked seats, need %d. Please select seats first", len(ticketIDs), request.Quantity)
	}

	// Step 5: Reserve the tickets
	updateTicketQuery := `
		UPDATE tickets 
		SET status = 'reserved', updated_at = NOW() 
		WHERE id = ANY($1)`

	_, err = tx.ExecContext(ctx, updateTicketQuery, pq.Array(ticketIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to reserve tickets: %w", err)
	}

	// Step 6: Update event available tickets
	updateEventQuery := `
		UPDATE events 
		SET available_tickets = available_tickets - $1, updated_at = NOW() 
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, updateEventQuery, request.Quantity, request.EventID)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	// Step 7: Create booking record
	totalAmount := event.Price * float64(request.Quantity)
	bookingRef := r.generateBookingRef()
	// Use configurable booking expiration duration instead of hardcoded 15 minutes
	expiresAt := time.Now().Add(r.config.App.BookingExpiration)

	insertBookingQuery := `
		INSERT INTO bookings (user_id, event_id, ticket_ids, quantity, total_amount, status, booking_ref, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at`

	var bookingID int
	var createdAt time.Time

	err = tx.QueryRowContext(ctx, insertBookingQuery,
		request.UserID,
		request.EventID,
		pq.Array(ticketIDs),
		request.Quantity,
		totalAmount,
		models.BookingPending,
		bookingRef,
		expiresAt,
	).Scan(&bookingID, &createdAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"booking_id":         bookingID,
		"user_id":            request.UserID,
		"event_id":           request.EventID,
		"quantity":           request.Quantity,
		"ticket_ids":         ticketIDs,
		"seat_numbers":       seatNumbers,
		"total_amount":       totalAmount,
		"booking_expiration": r.config.App.BookingExpiration,
	}).Info("Tickets booked successfully")

	return &models.Booking{
		ID:          bookingID,
		UserID:      request.UserID,
		EventID:     request.EventID,
		TicketIDs:   ticketIDs,
		Quantity:    request.Quantity,
		TotalAmount: totalAmount,
		Status:      models.BookingPending,
		BookingRef:  bookingRef,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
		ExpiresAt:   expiresAt,
	}, nil
}

// ConfirmBooking marks a booking as confirmed and tickets as sold
func (r *BookingRepository) ConfirmBooking(ctx context.Context, bookingID int) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Get booking details with lock
		var booking models.Booking
		query := `
			SELECT id, ticket_ids, status, expires_at 
			FROM bookings 
			WHERE id = $1 
			FOR UPDATE`

		var ticketIDsStr string
		err := tx.QueryRowContext(ctx, query, bookingID).Scan(
			&booking.ID,
			&ticketIDsStr,
			&booking.Status,
			&booking.ExpiresAt,
		)
		if err != nil {
			return fmt.Errorf("booking not found: %w", err)
		}

		// Validate booking status and expiry
		if booking.Status != models.BookingPending {
			return fmt.Errorf("booking is not in pending status")
		}

		if time.Now().After(booking.ExpiresAt) {
			return fmt.Errorf("booking has expired")
		}

		// Parse ticket IDs
		ticketIDs := parseTicketIDs(ticketIDsStr)

		r.logger.WithFields(logrus.Fields{
			"booking_id":        bookingID,
			"ticket_ids_string": ticketIDsStr,
			"parsed_ticket_ids": ticketIDs,
		}).Debug("Confirming booking with ticket IDs")

		// Update tickets to sold
		updateTicketsQuery := `
			UPDATE tickets 
			SET status = 'sold', updated_at = NOW() 
			WHERE id = ANY($1) AND status = 'reserved'`

		result, err := tx.ExecContext(ctx, updateTicketsQuery, pq.Array(ticketIDs))
		if err != nil {
			return fmt.Errorf("failed to confirm tickets: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if int(rowsAffected) != len(ticketIDs) {
			r.logger.WithFields(logrus.Fields{
				"booking_id":     bookingID,
				"ticket_ids":     ticketIDs,
				"expected_count": len(ticketIDs),
				"rows_affected":  rowsAffected,
			}).Error("Mismatch in ticket confirmation count")
			return fmt.Errorf("some tickets could not be confirmed")
		}

		// Update booking status
		updateBookingQuery := `
			UPDATE bookings 
			SET status = 'confirmed', updated_at = NOW() 
			WHERE id = $1`

		_, err = tx.ExecContext(ctx, updateBookingQuery, bookingID)
		if err != nil {
			return fmt.Errorf("failed to confirm booking: %w", err)
		}

		r.logger.WithField("booking_id", bookingID).Info("Booking confirmed successfully")
		return nil
	})
}

// CancelBooking cancels a booking and releases the tickets
func (r *BookingRepository) CancelBooking(ctx context.Context, bookingID int) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Get booking details with lock
		var booking models.Booking
		query := `
			SELECT id, event_id, ticket_ids, quantity, status 
			FROM bookings 
			WHERE id = $1 
			FOR UPDATE`

		var ticketIDsStr string
		err := tx.QueryRowContext(ctx, query, bookingID).Scan(
			&booking.ID,
			&booking.EventID,
			&ticketIDsStr,
			&booking.Quantity,
			&booking.Status,
		)
		if err != nil {
			return fmt.Errorf("booking not found: %w", err)
		}

		if booking.Status == models.BookingCancelled {
			return fmt.Errorf("booking is already cancelled")
		}

		// Parse ticket IDs
		ticketIDs := parseTicketIDs(ticketIDsStr)

		// Release tickets back to available
		updateTicketsQuery := `
			UPDATE tickets 
			SET status = 'available', updated_at = NOW() 
			WHERE id = ANY($1)`

		_, err = tx.ExecContext(ctx, updateTicketsQuery, pq.Array(ticketIDs))
		if err != nil {
			return fmt.Errorf("failed to release tickets: %w", err)
		}

		// Update event available tickets
		updateEventQuery := `
			UPDATE events 
			SET available_tickets = available_tickets + $1, updated_at = NOW() 
			WHERE id = $2`

		_, err = tx.ExecContext(ctx, updateEventQuery, booking.Quantity, booking.EventID)
		if err != nil {
			return fmt.Errorf("failed to update event: %w", err)
		}

		// Update booking status
		updateBookingQuery := `
			UPDATE bookings 
			SET status = 'cancelled', updated_at = NOW() 
			WHERE id = $1`

		_, err = tx.ExecContext(ctx, updateBookingQuery, bookingID)
		if err != nil {
			return fmt.Errorf("failed to cancel booking: %w", err)
		}

		r.logger.WithField("booking_id", bookingID).Info("Booking cancelled successfully")
		return nil
	})
}

// GetBooking retrieves booking details
func (r *BookingRepository) GetBooking(ctx context.Context, bookingID int) (*models.Booking, error) {
	query := `
		SELECT id, user_id, event_id, ticket_ids, quantity, total_amount, 
			   status, booking_ref, created_at, updated_at, expires_at
		FROM bookings 
		WHERE id = $1`

	var booking models.Booking
	var ticketIDsStr string

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.EventID,
		&ticketIDsStr,
		&booking.Quantity,
		&booking.TotalAmount,
		&booking.Status,
		&booking.BookingRef,
		&booking.CreatedAt,
		&booking.UpdatedAt,
		&booking.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, err
	}

	booking.TicketIDs = parseTicketIDs(ticketIDsStr)
	return &booking, nil
}

// Helper functions
func (r *BookingRepository) generateBookingRef() string {
	return fmt.Sprintf("BK%d", time.Now().UnixNano())
}

func joinInts(ints []int, sep string) string {
	if len(ints) == 0 {
		return ""
	}

	result := fmt.Sprintf("%d", ints[0])
	for i := 1; i < len(ints); i++ {
		result += sep + fmt.Sprintf("%d", ints[i])
	}
	return result
}

func parseTicketIDs(ticketIDsStr string) []int {
	// Parser for PostgreSQL array format: {1,2,3} or {10,11,12}
	if len(ticketIDsStr) < 3 {
		return []int{}
	}

	// Remove braces
	ticketIDsStr = ticketIDsStr[1 : len(ticketIDsStr)-1]

	// Handle empty array
	if len(ticketIDsStr) == 0 {
		return []int{}
	}

	var ticketIDs []int
	var currentNumber string

	for _, char := range ticketIDsStr {
		if char >= '0' && char <= '9' {
			currentNumber += string(char)
		} else if char == ',' {
			if currentNumber != "" {
				if id, err := strconv.Atoi(currentNumber); err == nil {
					ticketIDs = append(ticketIDs, id)
				}
				currentNumber = ""
			}
		}
	}

	// Don't forget the last number
	if currentNumber != "" {
		if id, err := strconv.Atoi(currentNumber); err == nil {
			ticketIDs = append(ticketIDs, id)
		}
	}

	return ticketIDs
}
