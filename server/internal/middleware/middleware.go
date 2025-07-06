package middleware

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/milinddethe15/ticket-booking/internal/models"
)

// RateLimiter creates a rate limiting middleware
func RateLimiter(rps int) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Limit(rps), rps*2) // Allow burst of 2x RPS

	return func(c *gin.Context) {
		if !limiter.AllowN(time.Now(), 1) {
			c.JSON(http.StatusTooManyRequests, &models.APIResponse{
				Success: false,
				Error:   "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Logger creates a structured logging middleware
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get request details
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		// Log request
		entry := logger.WithFields(logrus.Fields{
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"latency":     latency,
			"user_agent":  userAgent,
		})

		if statusCode >= 400 {
			entry.Error("Request completed with error")
		} else {
			entry.Info("Request completed")
		}
	}
}

// CORS middleware for handling cross-origin requests
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Session-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ErrorHandler middleware for centralized error handling
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, &models.APIResponse{
				Success: false,
				Error:   "Internal server error",
				Message: err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, &models.APIResponse{
				Success: false,
				Error:   "Internal server error",
			})
		}
		c.Abort()
	})
}

// RequestTimeout middleware to prevent long-running requests
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		panicChan := make(chan interface{}, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- p
				}
			}()
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed normally
		case p := <-panicChan:
			panic(p)
		case <-ctx.Done():
			c.JSON(http.StatusRequestTimeout, &models.APIResponse{
				Success: false,
				Error:   "Request timeout",
			})
			c.Abort()
		}
	}
}

// Security headers middleware
func Security() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// RequestID middleware to add unique request ID
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Header("X-Request-ID", requestID)
		c.Set("RequestID", requestID)
		c.Next()
	}
}

// Pagination middleware to parse pagination parameters
func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "20")

		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 1 {
			pageInt = 1
		}

		limitInt, err := strconv.Atoi(limit)
		if err != nil || limitInt < 1 || limitInt > 100 {
			limitInt = 20
		}

		offset := (pageInt - 1) * limitInt

		c.Set("page", pageInt)
		c.Set("limit", limitInt)
		c.Set("offset", offset)
		c.Next()
	}
}

// Helper function to generate request ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}
