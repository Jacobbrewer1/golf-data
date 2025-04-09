package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/processor"
)

func (d *domain) SaveCourse(course *models.Course) error {
	// Does the course already exist?
	currentCourse, err := d.r.CourseById(course.Id)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoCourses):
			if err := d.r.SaveCourse(course); err != nil {
				return fmt.Errorf("failed to save course: %w", err)
			}
			return nil
		default:
			return fmt.Errorf("failed to get course by id: %w", err)
		}
	}

	// Update the course
	if err := d.r.PatchCourse(currentCourse, course); err != nil {
		return fmt.Errorf("failed to patch course: %w", err)
	}

	return nil
}
