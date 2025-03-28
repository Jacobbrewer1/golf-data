package processor

import (
	"github.com/jacobbrewer1/vaulty/repositories"
)

type repository struct {
	db *repositories.Database
}

// NewRepository creates a new repository.
func NewRepository(db *repositories.Database) Repository {
	return &repository{
		db: db,
	}
}
