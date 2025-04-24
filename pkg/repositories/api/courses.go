package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/golf-data/pkg/repositories/api/filters"
	"github.com/jacobbrewer1/pagefilter"
)

var (
	// ErrNoCourses is returned when no courses are found.
	ErrNoCourses = errors.New("no courses found")
)

func (r *repository) GetCourses(details *pagefilter.PaginatorDetails, filter *GetCoursesFilters) (*pagefilter.PaginatedResponse[models.Course], error) {
	t := prometheus.NewTimer(models.DatabaseLatency.WithLabelValues("get_courses"))
	defer t.ObserveDuration()

	mf := r.getCoursesFilter(filter)
	pg := pagefilter.NewPaginator(r.db, models.CourseTableName, "id", mf)

	if err := pg.SetDetails(details,
		"id",
		"name",
	); err != nil {
		return nil, fmt.Errorf("failed to set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoCourses
		default:
			return nil, fmt.Errorf("failed to get courses: %w", err)
		}
	}

	items := make([]*models.Course, 0)
	if err := pg.Retrieve(pvt, &items); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoCourses
		default:
			return nil, fmt.Errorf("failed to retrieve courses: %w", err)
		}
	}

	var total int64 = 0
	if err := pg.Counts(&total); err != nil {
		return nil, fmt.Errorf("failed to count courses: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.Course]{
		Items: items,
		Total: total,
	}, nil
}

func (r *repository) getCoursesFilter(filter *GetCoursesFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if filter == nil {
		return mf
	}
	mf.Add(filters.NewCourseClubID(filter.ClubID))
	return mf
}
