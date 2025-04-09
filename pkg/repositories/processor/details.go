package processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
)

var (
	// ErrNoDetails is returned when no course details are found
	ErrNoDetails = errors.New("no course details found")
)

func (r *repository) SaveDetails(details *models.CourseDetails) error {
	if details.IsPrimaryKeySet() {
		return details.InsertCourseDetailsWithPK(r.db)
	}
	return details.InsertWithUpdate(r.db)
}

func (r *repository) CourseDetailsById(id int) (*models.CourseDetails, error) {
	course, err := models.CourseDetailsById(r.db, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoDetails
		default:
			return nil, fmt.Errorf("failed to get details by id: %w", err)
		}
	}

	return course, nil
}

func (r *repository) PatchDetails(currentDetails, newDetails *models.CourseDetails) error {
	if err := currentDetails.Patch(r.db, newDetails); err != nil {
		return fmt.Errorf("failed to patch details: %w", err)
	}

	return nil
}
