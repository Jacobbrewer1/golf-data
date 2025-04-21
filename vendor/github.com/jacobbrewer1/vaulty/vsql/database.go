package vsql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

var (
	newDbPingTimeout  = 7 * time.Second
	newDbPingInterval = 500 * time.Millisecond
)

// Database represents a database connection
// It embeds the sqlx.DB struct to provide additional functionality and methods for interacting with the database.
type Database struct {
	*sqlx.DB
}

// NewDatabase establishes a database connection with the given Vault credentials
func NewDatabase(db *sqlx.DB) *Database {
	dbConn := &Database{
		DB: db,
	}

	return dbConn
}

// ReplaceDB replaces the current database connection with a new one.
// It ensures that the new connection is valid and pings it to check its readiness.
// If the new connection is valid, it closes the old connection and replaces it with the new one.
// This is useful for scenarios where the database connection needs to be refreshed or replaced.
func (d *Database) ReplaceDB(ctx context.Context, newDB *sqlx.DB) error {
	// Ensure that the new database connection is valid
	if newDB == nil {
		return errors.New("new database connection is nil")
	}

	// Ping the new database to ensure it's ready
	if err := pingDB(ctx, newDB); err != nil {
		return fmt.Errorf("failed to ping new database: %w", err)
	}

	if err := d.closeAndReplaceDB(newDB); err != nil {
		return fmt.Errorf("failed to close and replace database: %w", err)
	}

	return nil
}

// closeAndReplaceDB closes the old database connection and replaces it with the new one.
func (d *Database) closeAndReplaceDB(newDB *sqlx.DB) error {
	// Close the old database connection
	if err := d.Close(); err != nil {
		return fmt.Errorf("failed to close old database connection: %w", err)
	}

	// Replace the old database connection with the new one
	d.DB = newDB

	return nil
}

// pingDB pings the database to check its readiness.
func pingDB(ctx context.Context, db *sqlx.DB) error {
	ctx, cancel := context.WithTimeout(ctx, newDbPingTimeout)
	defer cancel()

	for {
		err := db.PingContext(ctx)
		if err == nil {
			break
		}
		select {
		case <-time.After(newDbPingInterval):
			continue
		case <-ctx.Done():
			return fmt.Errorf("failed to successfully ping database before context timeout: %w", err)
		}
	}

	return nil
}
