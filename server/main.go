package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/config"
	"github.com/milinddethe15/ticket-booking/internal/db"
	"github.com/milinddethe15/ticket-booking/internal/handlers"
	"github.com/milinddethe15/ticket-booking/internal/middleware"
	"github.com/milinddethe15/ticket-booking/internal/repository"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to load configuration")
	}

	// Setup logger
	logger := setupLogger(cfg.App.LogLevel)
	logger.Info("Starting ticket booking service")

	// Connect to database
	database, err := db.NewConnection(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()

	// Initialize repositories with configuration
	bookingRepo := repository.NewBookingRepository(database, logger, cfg)
	eventRepo := repository.NewEventRepository(database, logger, cfg)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler()
	eventHandler := handlers.NewEventHandler(eventRepo, logger)
	bookingHandler := handlers.NewBookingHandler(bookingRepo, eventRepo, logger)

	// Start background cleanup routine for expired seat locks with configurable interval
	go startSeatLockCleanup(eventRepo, logger, cfg.App.CleanupInterval)

	// Setup HTTP server
	router := setupRouter(cfg, logger, healthHandler, eventHandler, bookingHandler)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Server exited")
}

func setupLogger(logLevel string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	return logger
}

func setupRouter(cfg *config.Config, logger *logrus.Logger, healthHandler *handlers.HealthHandler, eventHandler *handlers.EventHandler, bookingHandler *handlers.BookingHandler) *gin.Engine {
	// Set Gin mode
	if cfg.App.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Apply global middleware
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Security())
	router.Use(middleware.RequestID())
	router.Use(middleware.RequestTimeout(30 * time.Second))
	router.Use(middleware.RateLimiter(cfg.App.RateLimitRPS))

	// Health check routes (no rate limiting)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API routes
	v1 := router.Group("/api/v1")
	{
		// Event routes
		events := v1.Group("/events")
		events.Use(middleware.Pagination())
		{
			events.GET("", eventHandler.GetEvents)
			events.GET("/:id", eventHandler.GetEvent)
			events.POST("", eventHandler.CreateEvent)
			events.GET("/:id/tickets", eventHandler.GetAvailableTickets)
			events.GET("/:id/tickets/all", eventHandler.GetAllTickets)
			events.POST("/:id/seats/:seatNo/lock", eventHandler.LockSeat)
			events.POST("/:id/seats/:seatNo/unlock", eventHandler.UnlockSeat)
		}

		// Booking routes
		bookings := v1.Group("/bookings")
		{
			bookings.POST("", bookingHandler.BookTickets)
			bookings.GET("/:id", bookingHandler.GetBooking)
			bookings.POST("/:id/confirm", bookingHandler.ConfirmBooking)
			bookings.POST("/:id/cancel", bookingHandler.CancelBooking)
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Endpoint not found",
		})
	})

	return router
}

// startSeatLockCleanup runs a background routine to cleanup expired seat locks with configurable interval
func startSeatLockCleanup(eventRepo *repository.EventRepository, logger *logrus.Logger, cleanupInterval time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	logger.WithField("cleanup_interval", cleanupInterval).Info("Started seat lock cleanup routine")

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := eventRepo.CleanupExpiredLocks(ctx); err != nil {
				logger.WithError(err).Error("Failed to cleanup expired seat locks")
			}
			cancel()
		}
	}
}
