package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/milinddethe15/ticket-booking/internal/config"
)

type DB struct {
	*sql.DB
	logger *logrus.Logger
}

func NewConnection(cfg *config.DatabaseConfig, logger *logrus.Logger) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return &DB{
		DB:     db,
		logger: logger,
	}, nil
}

func (db *DB) Close() error {
	db.logger.Info("Closing database connection")
	return db.DB.Close()
}

// Transaction wrapper for pessimistic locking
func (db *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				db.logger.WithError(rbErr).Error("Failed to rollback transaction")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	err = fn(tx)
	return err
}

// Retry mechanism for handling deadlocks and temporary failures
func (db *DB) WithRetry(ctx context.Context, maxRetries int, retryDelay time.Duration, fn func() error) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}

		// Check if it's a retryable error (deadlock, connection issue, etc.)
		if !isRetryableError(err) {
			return err
		}

		if i < maxRetries {
			db.logger.WithError(err).Warnf("Operation failed, retrying in %v (attempt %d/%d)", retryDelay, i+1, maxRetries)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}

func isRetryableError(err error) bool {
	// Check for PostgreSQL error codes that indicate retryable errors
	// 40001 = serialization_failure
	// 40P01 = deadlock_detected
	errStr := err.Error()
	return contains(errStr, "40001") ||
		contains(errStr, "40P01") ||
		contains(errStr, "deadlock") ||
		contains(errStr, "serialization failure") ||
		contains(errStr, "connection") ||
		contains(errStr, "timeout")
}

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
