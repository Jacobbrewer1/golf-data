package processor

import "github.com/jacobbrewer1/golf-data/pkg/models"

func (r *repository) SaveClub(club *models.Club) error {
	return club.Save(r.db)
}
