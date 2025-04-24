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
	// ErrNoMarkers is returned when no markers are found.
	ErrNoMarkers = errors.New("no markers found")
)

func (r *repository) GetMarkers(details *pagefilter.PaginatorDetails, filter *GetMarkersFilters) (*pagefilter.PaginatedResponse[models.CourseDetails], error) {
	t := prometheus.NewTimer(models.DatabaseLatency.WithLabelValues("get_markers"))
	defer t.ObserveDuration()

	mf := r.getMarkersFilter(filter)
	pg := pagefilter.NewPaginator(r.db, models.CourseDetailsTableName, "id", mf)

	if err := pg.SetDetails(details,
		"id",
		"marker",
		"slope_rating",
		"course_rating",
		"par_front_nine",
		"par_back_nine",
		"par_total",
		"yards_front_nine",
		"yards_back_nine",
		"yards_total",
		"meters_front_nine",
		"meters_back_nine",
		"meters_total",
	); err != nil {
		return nil, fmt.Errorf("failed to set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoMarkers
		default:
			return nil, fmt.Errorf("failed to get markers: %w", err)
		}
	}

	items := make([]*models.CourseDetails, 0)
	if err := pg.Retrieve(pvt, &items); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoMarkers
		default:
			return nil, fmt.Errorf("failed to retrieve markers: %w", err)
		}
	}

	var total int64 = 0
	if err := pg.Counts(&total); err != nil {
		return nil, fmt.Errorf("failed to count markers: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.CourseDetails]{
		Items: items,
		Total: total,
	}, nil
}

func (r *repository) getMarkersFilter(filter *GetMarkersFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if filter == nil {
		return mf
	}
	mf.Add(filters.NewMarkerCourseID(filter.CourseID))
	return mf
}
