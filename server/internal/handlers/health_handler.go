package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/milinddethe15/ticket-booking/internal/models"
)

const AppVersion = "1.0.0"

// HealthHandler handles health check endpoints
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Health handles GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, &models.HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   AppVersion,
	})
}

// Ready handles GET /ready for readiness probe
func (h *HealthHandler) Ready(c *gin.Context) {
	// In a real application, you would check database connectivity,
	// external services, etc. here
	c.JSON(http.StatusOK, &models.APIResponse{
		Success: true,
		Message: "Service is ready",
	})
}
