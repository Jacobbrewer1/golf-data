package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/processor"
)

func (d *domain) SaveDetails(details *models.CourseDetails) error {
	// Does the course already exist?
	currentCourseDetails, err := d.r.CourseDetailsById(details.Id)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoDetails):
			if err := d.r.SaveDetails(details); err != nil {
				return fmt.Errorf("failed to save course: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("failed to get course by id: %w", err)
		}
	}

	// Update the course
	if err := d.r.PatchDetails(currentCourseDetails, details); err != nil {
		return fmt.Errorf("failed to patch course: %w", err)
	}

	return nil
}
