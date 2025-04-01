package processor

import "github.com/jacobbrewer1/golf-data/pkg/models"

func (r *repository) SaveClub(club *models.Club) error {
	if club.IsPrimaryKeySet() {
		return club.InsertClubWithPK(r.db)
	}
	return club.InsertWithUpdate(r.db)
}

func (r *repository) ClubById(id int) (*models.Club, error) {
	return models.ClubById(r.db, id)
}

func (r *repository) PatchClub(currentClub, newClub *models.Club) error {
	return currentClub.Patch(r.db, newClub)
}
