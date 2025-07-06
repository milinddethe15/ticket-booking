package models

import (
	"time"
)

type Event struct {
	ID               int       `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Description      string    `json:"description" db:"description"`
	Venue            string    `json:"venue" db:"venue"`
	StartTime        time.Time `json:"start_time" db:"start_time"`
	EndTime          time.Time `json:"end_time" db:"end_time"`
	TotalTickets     int       `json:"total_tickets" db:"total_tickets"`
	AvailableTickets int       `json:"available_tickets" db:"available_tickets"`
	Price            float64   `json:"price" db:"price"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type Ticket struct {
	ID        int          `json:"id" db:"id"`
	EventID   int          `json:"event_id" db:"event_id"`
	SeatNo    string       `json:"seat_no" db:"seat_no"`
	Status    TicketStatus `json:"status" db:"status"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

type Booking struct {
	ID          int           `json:"id" db:"id"`
	UserID      int           `json:"user_id" db:"user_id"`
	EventID     int           `json:"event_id" db:"event_id"`
	TicketIDs   []int         `json:"ticket_ids" db:"ticket_ids"`
	Quantity    int           `json:"quantity" db:"quantity"`
	TotalAmount float64       `json:"total_amount" db:"total_amount"`
	Status      BookingStatus `json:"status" db:"status"`
	BookingRef  string        `json:"booking_ref" db:"booking_ref"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
	ExpiresAt   time.Time     `json:"expires_at" db:"expires_at"`
}

type User struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Email     string    `json:"email" db:"email"`
	Phone     string    `json:"phone" db:"phone"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type BookingRequest struct {
	UserID   int `json:"user_id" binding:"required"`
	EventID  int `json:"event_id" binding:"required"`
	Quantity int `json:"quantity" binding:"required,min=1,max=10"`
}

type BookingResponse struct {
	Booking *Booking `json:"booking"`
	Message string   `json:"message"`
}

// Enums
type TicketStatus string

const (
	TicketAvailable TicketStatus = "available"
	TicketReserved  TicketStatus = "reserved"
	TicketSold      TicketStatus = "sold"
)

type BookingStatus string

const (
	BookingPending   BookingStatus = "pending"
	BookingConfirmed BookingStatus = "confirmed"
	BookingCancelled BookingStatus = "cancelled"
	BookingExpired   BookingStatus = "expired"
)

// Response types
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
