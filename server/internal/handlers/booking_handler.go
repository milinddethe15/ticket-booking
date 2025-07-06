package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/models"
	"github.com/milinddethe15/ticket-booking/internal/repository"
)

type BookingHandler struct {
	bookingRepo *repository.BookingRepository
	eventRepo   *repository.EventRepository
	logger      *logrus.Logger
}

func NewBookingHandler(bookingRepo *repository.BookingRepository, eventRepo *repository.EventRepository, logger *logrus.Logger) *BookingHandler {
	return &BookingHandler{
		bookingRepo: bookingRepo,
		eventRepo:   eventRepo,
		logger:      logger,
	}
}

// BookTickets handles POST /api/bookings
func (h *BookingHandler) BookTickets(c *gin.Context) {
	var request models.BookingRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.WithError(err).Error("Invalid booking request")
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// Validate event exists
	event, err := h.eventRepo.GetEvent(c.Request.Context(), request.EventID)
	if err != nil {
		h.logger.WithError(err).WithField("event_id", request.EventID).Error("Event not found")
		c.JSON(http.StatusNotFound, &models.APIResponse{
			Success: false,
			Error:   "Event not found",
		})
		return
	}

	// Log booking attempt
	h.logger.WithFields(logrus.Fields{
		"user_id":    request.UserID,
		"event_id":   request.EventID,
		"event_name": event.Name,
		"quantity":   request.Quantity,
	}).Info("Booking attempt started")

	// Attempt to book tickets
	booking, err := h.bookingRepo.BookTickets(c.Request.Context(), &request)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":  request.UserID,
			"event_id": request.EventID,
			"quantity": request.Quantity,
		}).Error("Booking failed")

		// Determine appropriate HTTP status code based on error
		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "insufficient tickets") ||
			contains(err.Error(), "not found") {
			statusCode = http.StatusBadRequest
		} else if contains(err.Error(), "already started") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, &models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"booking_id":   booking.ID,
		"booking_ref":  booking.BookingRef,
		"user_id":      request.UserID,
		"event_id":     request.EventID,
		"quantity":     request.Quantity,
		"total_amount": booking.TotalAmount,
	}).Info("Booking successful")

	c.JSON(http.StatusCreated, &models.APIResponse{
		Success: true,
		Data:    booking,
		Message: "Tickets booked successfully. Please complete payment within 15 minutes.",
	})
}

// GetBooking handles GET /api/bookings/:id
func (h *BookingHandler) GetBooking(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid booking ID",
		})
		return
	}

	booking, err := h.bookingRepo.GetBooking(c.Request.Context(), bookingID)
	if err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &models.APIResponse{
				Success: false,
				Error:   "Booking not found",
			})
			return
		}

		h.logger.WithError(err).WithField("booking_id", bookingID).Error("Failed to get booking")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve booking",
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    booking,
	})
}

// ConfirmBooking handles POST /api/bookings/:id/confirm
func (h *BookingHandler) ConfirmBooking(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid booking ID",
		})
		return
	}

	err = h.bookingRepo.ConfirmBooking(c.Request.Context(), bookingID)
	if err != nil {
		h.logger.WithError(err).WithField("booking_id", bookingID).Error("Failed to confirm booking")

		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") ||
			contains(err.Error(), "not in pending status") ||
			contains(err.Error(), "expired") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, &models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.logger.WithField("booking_id", bookingID).Info("Booking confirmed")
	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Message: "Booking confirmed successfully",
	})
}

// CancelBooking handles POST /api/bookings/:id/cancel
func (h *BookingHandler) CancelBooking(c *gin.Context) {
	bookingIDStr := c.Param("id")
	bookingID, err := strconv.Atoi(bookingIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid booking ID",
		})
		return
	}

	err = h.bookingRepo.CancelBooking(c.Request.Context(), bookingID)
	if err != nil {
		h.logger.WithError(err).WithField("booking_id", bookingID).Error("Failed to cancel booking")

		statusCode := http.StatusInternalServerError
		if contains(err.Error(), "not found") ||
			contains(err.Error(), "already cancelled") {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, &models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	h.logger.WithField("booking_id", bookingID).Info("Booking cancelled")
	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Message: "Booking cancelled successfully",
	})
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
