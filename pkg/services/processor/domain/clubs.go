package domain

import "github.com/jacobbrewer1/golf-data/pkg/models"

func (d *domain) SaveClub(club *models.Club) error {
	return d.r.SaveClub(club)
}
