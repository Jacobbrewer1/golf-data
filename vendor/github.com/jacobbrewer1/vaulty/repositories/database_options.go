package repositories

import (
	"log/slog"
	"sync"
)

type DatabaseOption func(db *Database)

// WithDatabaseLogger sets the logger for the database
func WithDatabaseLogger(l *slog.Logger) DatabaseOption {
	return func(db *Database) {
		db.l = l
	}
}

// WithMutex sets the mutex for the database
func WithMutex(mx *sync.RWMutex) DatabaseOption {
	return func(db *Database) {
		db.mx = mx
	}
}
