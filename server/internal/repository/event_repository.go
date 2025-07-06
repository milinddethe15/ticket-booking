package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/config"
	"github.com/milinddethe15/ticket-booking/internal/db"
	"github.com/milinddethe15/ticket-booking/internal/models"
)

type EventRepository struct {
	db     *db.DB
	logger *logrus.Logger
	config *config.Config
}

func NewEventRepository(database *db.DB, logger *logrus.Logger, cfg *config.Config) *EventRepository {
	return &EventRepository{
		db:     database,
		logger: logger,
		config: cfg,
	}
}

// GetEvent retrieves an event by ID
func (r *EventRepository) GetEvent(ctx context.Context, eventID int) (*models.Event, error) {
	query := `
		SELECT id, name, description, venue, start_time, end_time, 
			   total_tickets, available_tickets, price, created_at, updated_at
		FROM events 
		WHERE id = $1`

	var event models.Event
	err := r.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.Name,
		&event.Description,
		&event.Venue,
		&event.StartTime,
		&event.EndTime,
		&event.TotalTickets,
		&event.AvailableTickets,
		&event.Price,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("event not found")
		}
		return nil, err
	}

	return &event, nil
}

// GetEvents retrieves all events with pagination
func (r *EventRepository) GetEvents(ctx context.Context, limit, offset int) ([]*models.Event, error) {
	query := `
		SELECT id, name, description, venue, start_time, end_time, 
			   total_tickets, available_tickets, price, created_at, updated_at
		FROM events 
		ORDER BY start_time ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.Description,
			&event.Venue,
			&event.StartTime,
			&event.EndTime,
			&event.TotalTickets,
			&event.AvailableTickets,
			&event.Price,
			&event.CreatedAt,
			&event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, &event)
	}

	return events, nil
}

// CreateEvent creates a new event with tickets
func (r *EventRepository) CreateEvent(ctx context.Context, event *models.Event) (*models.Event, error) {
	var createdEvent *models.Event

	err := r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Insert event
		insertEventQuery := `
			INSERT INTO events (name, description, venue, start_time, end_time, total_tickets, available_tickets, price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
			RETURNING id, created_at, updated_at`

		var eventID int
		err := tx.QueryRowContext(ctx, insertEventQuery,
			event.Name,
			event.Description,
			event.Venue,
			event.StartTime,
			event.EndTime,
			event.TotalTickets,
			event.TotalTickets, // available_tickets = total_tickets initially
			event.Price,
		).Scan(&eventID, &event.CreatedAt, &event.UpdatedAt)

		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}

		// Create tickets for the event
		insertTicketQuery := `
			INSERT INTO tickets (event_id, seat_no, status, created_at, updated_at)
			VALUES ($1, $2, 'available', NOW(), NOW())`

		for i := 1; i <= event.TotalTickets; i++ {
			seatNo := fmt.Sprintf("S%03d", i)
			_, err = tx.ExecContext(ctx, insertTicketQuery, eventID, seatNo)
			if err != nil {
				return fmt.Errorf("failed to create ticket %s: %w", seatNo, err)
			}
		}

		createdEvent = &models.Event{
			ID:               eventID,
			Name:             event.Name,
			Description:      event.Description,
			Venue:            event.Venue,
			StartTime:        event.StartTime,
			EndTime:          event.EndTime,
			TotalTickets:     event.TotalTickets,
			AvailableTickets: event.TotalTickets,
			Price:            event.Price,
			CreatedAt:        event.CreatedAt,
			UpdatedAt:        event.UpdatedAt,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.WithFields(logrus.Fields{
		"event_id":      createdEvent.ID,
		"event_name":    createdEvent.Name,
		"total_tickets": createdEvent.TotalTickets,
	}).Info("Event created successfully")

	return createdEvent, nil
}

// GetAvailableTickets retrieves available tickets for an event
func (r *EventRepository) GetAvailableTickets(ctx context.Context, eventID int, limit int) ([]*models.Ticket, error) {
	query := `
		SELECT id, event_id, seat_no, status, created_at, updated_at
		FROM tickets 
		WHERE event_id = $1 AND status = 'available'
		ORDER BY seat_no
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, eventID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.EventID,
			&ticket.SeatNo,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// GetAllTickets retrieves all tickets for an event (including sold/reserved) for UI display
func (r *EventRepository) GetAllTickets(ctx context.Context, eventID int, limit int) ([]*models.Ticket, error) {
	query := `
		SELECT id, event_id, seat_no, status, created_at, updated_at
		FROM tickets 
		WHERE event_id = $1
		ORDER BY seat_no
		LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, eventID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID,
			&ticket.EventID,
			&ticket.SeatNo,
			&ticket.Status,
			&ticket.CreatedAt,
			&ticket.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, &ticket)
	}

	return tickets, nil
}

// LockSeat temporarily locks a seat for seat selection (3 minutes)
func (r *EventRepository) LockSeat(ctx context.Context, eventID int, seatNo string, userSession string) error {
	return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Check if seat is available
		var currentStatus string
		checkQuery := `SELECT status FROM tickets WHERE event_id = $1 AND seat_no = $2 FOR UPDATE`

		r.logger.WithFields(logrus.Fields{
			"event_id": eventID,
			"seat_no":  seatNo,
			"session":  userSession,
		}).Debug("Attempting to lock seat")

		err := tx.QueryRowContext(ctx, checkQuery, eventID, seatNo).Scan(&currentStatus)
		if err != nil {
			r.logger.WithError(err).WithFields(logrus.Fields{
				"event_id": eventID,
				"seat_no":  seatNo,
			}).Error("Seat not found during lock attempt")
			return fmt.Errorf("seat not found: %w", err)
		}

		r.logger.WithFields(logrus.Fields{
			"event_id":       eventID,
			"seat_no":        seatNo,
			"current_status": currentStatus,
		}).Debug("Current seat status")

		if currentStatus != "available" {
			return fmt.Errorf("seat is no longer available (current status: %s)", currentStatus)
		}

		// Lock the seat temporarily
		lockQuery := `UPDATE tickets SET status = 'locked', updated_at = NOW() WHERE event_id = $1 AND seat_no = $2`
		result, err := tx.ExecContext(ctx, lockQuery, eventID, seatNo)
		if err != nil {
			return fmt.Errorf("failed to lock seat: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			return fmt.Errorf("seat was just taken by another user")
		}

		r.logger.WithFields(logrus.Fields{
			"event_id": eventID,
			"seat_no":  seatNo,
			"session":  userSession,
		}).Info("Seat locked temporarily")

		return nil
	})
}

// UnlockSeat releases a temporarily locked seat
func (r *EventRepository) UnlockSeat(ctx context.Context, eventID int, seatNo string) error {
	query := `UPDATE tickets SET status = 'available', updated_at = NOW() WHERE event_id = $1 AND seat_no = $2 AND status = 'locked'`

	_, err := r.db.ExecContext(ctx, query, eventID, seatNo)
	if err != nil {
		return fmt.Errorf("failed to unlock seat: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"event_id": eventID,
		"seat_no":  seatNo,
	}).Info("Seat unlocked")

	return nil
}

// CleanupExpiredLocks removes locks older than the configured seat lock duration
func (r *EventRepository) CleanupExpiredLocks(ctx context.Context) error {
	// Use the configurable seat lock duration instead of hardcoded '3 minutes'
	lockDurationMinutes := int(r.config.App.SeatLockDuration.Minutes())

	query := fmt.Sprintf(`
		UPDATE tickets 
		SET status = 'available', updated_at = NOW()
		WHERE status = 'locked' 
		AND updated_at < NOW() - INTERVAL '%d minutes'`, lockDurationMinutes)

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired locks: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.WithFields(logrus.Fields{
			"seats_unlocked": rowsAffected,
			"lock_duration":  r.config.App.SeatLockDuration,
		}).Info("Cleaned up expired seat locks")
	}

	return nil
}
