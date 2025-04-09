package retriever

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
)

var (
	// ErrNoClubs is returned when no clubs are found.
	ErrNoClubs = errors.New("no clubs found")
)

func (r *repository) GetClubs() ([]*models.Club, error) {
	clubs, err := models.GetAllClubs(r.db)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoClubs
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	return clubs, nil
}
