package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/processor"
)

func (d *domain) SaveHole(hole *models.Hole) error {
	currentHoles, err := d.r.HolesByMarkerId(hole.DetailsId)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoHoles):
			if err := d.r.SaveHole(hole); err != nil {
				return fmt.Errorf("failed to save hole: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("failed to get holes by marker id: %w", err)
		}
	}

	// Find the same hole
	for _, currentHole := range currentHoles {
		if currentHole.DetailsId != hole.DetailsId {
			continue
		}
		if currentHole.Number != hole.Number {
			continue
		}

		// Update the hole
		if err := d.r.PatchHole(currentHole, hole); err != nil {
			return fmt.Errorf("failed to patch hole: %w", err)
		}
		return nil
	}

	// If the hole doesn't exist, insert it
	if err := d.r.SaveHole(hole); err != nil {
		return fmt.Errorf("failed to save hole: %w", err)
	}

	return nil
}
