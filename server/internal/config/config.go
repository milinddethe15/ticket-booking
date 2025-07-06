package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	App      AppConfig
}

type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type AppConfig struct {
	LogLevel     string
	RateLimitRPS int
	LockTimeout  time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	// Seat and booking configuration
	SeatLockDuration  time.Duration // How long seats remain locked during selection
	BookingExpiration time.Duration // How long users have to complete payment
	CleanupInterval   time.Duration // How often to run expired lock cleanup
}

func Load() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			ReadTimeout:  getDuration("READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDuration("WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "password"),
			DBName:          getEnv("DB_NAME", "ticket_booking"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},

		App: AppConfig{
			LogLevel:     getEnv("LOG_LEVEL", "info"),
			RateLimitRPS: getEnvInt("RATE_LIMIT_RPS", 100),
			LockTimeout:  getDuration("LOCK_TIMEOUT", 30*time.Second),
			MaxRetries:   getEnvInt("MAX_RETRIES", 3),
			RetryDelay:   getDuration("RETRY_DELAY", 100*time.Millisecond),
			// Seat and booking configuration with defaults
			SeatLockDuration:  getDuration("SEAT_LOCK_DURATION", 3*time.Minute),
			BookingExpiration: getDuration("BOOKING_EXPIRATION", 15*time.Minute),
			CleanupInterval:   getDuration("CLEANUP_INTERVAL", 1*time.Minute),
		},
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
