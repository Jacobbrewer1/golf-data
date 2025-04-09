package retriever

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
)

var (
	// ErrNoCourses is returned when no courses are found
	ErrNoCourses = errors.New("no courses found")
)

func (r *repository) GetCourses() ([]*models.Course, error) {
	courses, err := models.GetAllCourses(r.db)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoCourses
		default:
			return nil, fmt.Errorf("failed to get courses: %w", err)
		}
	}
	return courses, nil
}
