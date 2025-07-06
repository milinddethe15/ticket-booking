package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/models"
	"github.com/milinddethe15/ticket-booking/internal/repository"
)

type EventHandler struct {
	eventRepo *repository.EventRepository
	logger    *logrus.Logger
}

func NewEventHandler(eventRepo *repository.EventRepository, logger *logrus.Logger) *EventHandler {
	return &EventHandler{
		eventRepo: eventRepo,
		logger:    logger,
	}
}

// GetEvents handles GET /api/events
func (h *EventHandler) GetEvents(c *gin.Context) {
	// Get pagination parameters from middleware
	limit := c.GetInt("limit")
	offset := c.GetInt("offset")

	events, err := h.eventRepo.GetEvents(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get events")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve events",
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    events,
	})
}

// GetEvent handles GET /api/events/:id
func (h *EventHandler) GetEvent(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid event ID",
		})
		return
	}

	event, err := h.eventRepo.GetEvent(c.Request.Context(), eventID)
	if err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, &models.APIResponse{
				Success: false,
				Error:   "Event not found",
			})
			return
		}

		h.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get event")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve event",
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    event,
	})
}

// CreateEvent handles POST /api/events
func (h *EventHandler) CreateEvent(c *gin.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.WithError(err).Error("Invalid event request")
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid request format",
			Message: err.Error(),
		})
		return
	}

	// Validate event dates
	if event.StartTime.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Event start time cannot be in the past",
		})
		return
	}

	if event.EndTime.Before(event.StartTime) {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Event end time must be after start time",
		})
		return
	}

	// Validate ticket count
	if event.TotalTickets <= 0 || event.TotalTickets > 10000 {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Total tickets must be between 1 and 10,000",
		})
		return
	}

	// Validate price
	if event.Price < 0 {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Price cannot be negative",
		})
		return
	}

	createdEvent, err := h.eventRepo.CreateEvent(c.Request.Context(), &event)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"event_name":    event.Name,
			"total_tickets": event.TotalTickets,
		}).Error("Failed to create event")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to create event",
		})
		return
	}

	h.logger.WithFields(logrus.Fields{
		"event_id":      createdEvent.ID,
		"event_name":    createdEvent.Name,
		"total_tickets": createdEvent.TotalTickets,
	}).Info("Event created successfully")

	c.JSON(http.StatusCreated, &models.APIResponse{
		Success: true,
		Data:    createdEvent,
		Message: "Event created successfully",
	})
}

// GetAvailableTickets handles GET /api/events/:id/tickets
func (h *EventHandler) GetAvailableTickets(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid event ID",
		})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	tickets, err := h.eventRepo.GetAvailableTickets(c.Request.Context(), eventID, limit)
	if err != nil {
		h.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get available tickets")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve available tickets",
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    tickets,
	})
}

// GetAllTickets handles GET /api/events/:id/tickets/all
func (h *EventHandler) GetAllTickets(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid event ID",
		})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "200")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 500 {
		limit = 200
	}

	tickets, err := h.eventRepo.GetAllTickets(c.Request.Context(), eventID, limit)
	if err != nil {
		h.logger.WithError(err).WithField("event_id", eventID).Error("Failed to get all tickets")
		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   "Failed to retrieve all tickets",
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Data:    tickets,
	})
}

// LockSeat handles POST /api/events/:id/seats/:seatNo/lock
func (h *EventHandler) LockSeat(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid event ID",
		})
		return
	}

	seatNo := c.Param("seatNo")
	userSession := c.GetHeader("X-Session-ID") // You'll need to send this from UI
	if userSession == "" {
		userSession = "anonymous"
	}

	err = h.eventRepo.LockSeat(c.Request.Context(), eventID, seatNo, userSession)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"event_id": eventID,
			"seat_no":  seatNo,
		}).Error("Failed to lock seat")

		statusCode := http.StatusConflict
		if contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, &models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Message: "Seat locked temporarily",
	})
}

// UnlockSeat handles POST /api/events/:id/seats/:seatNo/unlock
func (h *EventHandler) UnlockSeat(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, &models.APIResponse{
			Success: false,
			Error:   "Invalid event ID",
		})
		return
	}

	seatNo := c.Param("seatNo")

	err = h.eventRepo.UnlockSeat(c.Request.Context(), eventID, seatNo)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"event_id": eventID,
			"seat_no":  seatNo,
		}).Error("Failed to unlock seat")

		c.JSON(http.StatusInternalServerError, &models.APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Message: "Seat unlocked",
	})
}
