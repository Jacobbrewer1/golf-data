package domain

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
)

func (d *domain) SaveClub(club *models.Club) error {
	// Does the club already exist?
	currentClub, err := d.r.ClubById(club.Id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if err := d.r.SaveClub(club); err != nil {
				return fmt.Errorf("failed to save club: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("failed to get club by id: %w", err)
		}
	}

	// Update the club
	if err := d.r.PatchClub(currentClub, club); err != nil {
		return fmt.Errorf("failed to patch club: %w", err)
	}

	return nil
}
