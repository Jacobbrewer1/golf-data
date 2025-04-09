package processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
)

var (
	// ErrNoCourses is returned when no courses are found.
	ErrNoCourses = errors.New("no courses found")
)

func (r *repository) SaveCourse(course *models.Course) error {
	if course.IsPrimaryKeySet() {
		return course.InsertCourseWithPK(r.db)
	}
	return course.InsertWithUpdate(r.db)
}

func (r *repository) CourseById(id int) (*models.Course, error) {
	course, err := models.CourseById(r.db, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoCourses
		default:
			return nil, fmt.Errorf("failed to get course by id: %w", err)
		}
	}

	return course, nil
}

func (r *repository) PatchCourse(currentCourse, newCourse *models.Course) error {
	if err := currentCourse.Patch(r.db, newCourse); err != nil {
		return fmt.Errorf("failed to patch course: %w", err)
	}

	return nil
}
