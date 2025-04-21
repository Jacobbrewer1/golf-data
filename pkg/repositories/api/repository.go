package api

import (
	"github.com/jacobbrewer1/vaulty/vsql"
)

type repository struct {
	db *vsql.Database
}

// NewRepository creates a new repository.
func NewRepository(db *vsql.Database) Repository {
	return &repository{
		db: db,
	}
}
