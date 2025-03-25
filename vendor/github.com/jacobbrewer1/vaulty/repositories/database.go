package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type Database struct {
	*sqlx.DB
	mx *sync.RWMutex

	l *slog.Logger
}

// NewDatabase establishes a database connection with the given Vault credentials
func NewDatabase(db *sqlx.DB, opts ...DatabaseOption) *Database {
	dbConn := &Database{
		DB: db,
		mx: new(sync.RWMutex),
		l:  slog.Default(),
	}

	for _, opt := range opts {
		opt(dbConn)
	}

	return dbConn
}

// Reconnect will be called periodically to refresh the database connection
// since the dynamic credentials expire after some time, it will:
//  1. construct a connection string using the given credentials
//  2. establish a database connection
//  3. close & replace the existing connection with the new one behind a mutex
func (d *Database) Reconnect(ctx context.Context, db *sqlx.DB) error {
	ctx, cancelContextFunc := context.WithTimeout(ctx, 7*time.Second)
	defer cancelContextFunc()

	d.l.Debug("Reconnecting to database")

	// wait until the database is ready or timeout expires
	for {
		err := db.PingContext(ctx)
		if err == nil {
			break
		}
		select {
		case <-time.After(500 * time.Millisecond):
			d.l.Debug("Database ping failed, retrying...")
			continue
		case <-ctx.Done():
			return fmt.Errorf("failed to successfully ping database before context timeout: %w", err)
		}
	}

	d.l.Info("New database connection established")

	d.closeReplaceConnection(db)

	return nil
}

func (d *Database) closeReplaceConnection(newDb *sqlx.DB) {
	d.l.Debug("Replacing database connection")

	// close the existing connection, if exists
	if d.DB != nil {
		_ = d.Close()
	}

	d.DB = newDb

	d.l.Debug("Database connection replaced")
}

func (d *Database) Close() error {
	d.l.Debug("Acquiring lock to close database connection")

	d.mx.Lock()
	defer d.mx.Unlock()

	d.l.Debug("Lock acquired to close database connection")

	if d.DB != nil {
		return d.DB.Close()
	}

	return nil
}

func (d *Database) PingContext(ctx context.Context) error {
	d.mx.RLock()
	defer d.mx.RUnlock()

	return d.DB.PingContext(ctx)
}
