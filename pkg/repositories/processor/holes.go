package processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/golf-data/pkg/repositories/processor/filters"
)

var (
	ErrNoHoles = errors.New("no holes found")
)

func (r *repository) SaveHole(hole *models.Hole) error {
	if hole.IsPrimaryKeySet() {
		return hole.InsertHoleWithPK(r.db)
	}
	return hole.InsertWithUpdate(r.db)
}

func (r *repository) HolesByMarkerId(id int) ([]*models.Hole, error) {
	holes, err := models.GetAllHoles(r.db, filters.NewHoleMarkerId(id))
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoHoles
		default:
			return nil, fmt.Errorf("failed to get holes by marker id: %w", err)
		}
	}

	return holes, nil
}

func (r *repository) PatchHole(currentHole, newHole *models.Hole) error {
	if err := currentHole.Patch(r.db, newHole); err != nil {
		return fmt.Errorf("failed to patch hole: %w", err)
	}

	return nil
}
