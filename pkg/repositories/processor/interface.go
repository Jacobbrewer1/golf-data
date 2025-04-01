package processor

import "github.com/jacobbrewer1/golf-data/pkg/models"

type Repository interface {
	SaveClub(club *models.Club) error
	ClubById(id int) (*models.Club, error)
	PatchClub(currentClub, newClub *models.Club) error
}
