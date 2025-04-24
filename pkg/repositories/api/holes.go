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
	// ErrNoHoles is returned when no holes are found.
	ErrNoHoles = errors.New("no holes found")
)

func (r *repository) GetHoles(details *pagefilter.PaginatorDetails, filter *GetHolesFilters) (*pagefilter.PaginatedResponse[models.Hole], error) {
	t := prometheus.NewTimer(models.DatabaseLatency.WithLabelValues("get_holes"))
	defer t.ObserveDuration()

	mf := getHolesFilter(filter)
	pg := pagefilter.NewPaginator(r.db, models.HoleTableName, "id", mf)

	if err := pg.SetDetails(details,
		"id",
		"number",
		"par",
		"stroke_index",
		"distance_yards",
		"distance_meters",
	); err != nil {
		return nil, fmt.Errorf("failed to set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoHoles
		default:
			return nil, fmt.Errorf("failed to get holes: %w", err)
		}
	}

	items := make([]*models.Hole, 0)
	if err := pg.Retrieve(pvt, &items); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoHoles
		default:
			return nil, fmt.Errorf("failed to retrieve holes: %w", err)
		}
	}

	var total int64 = 0
	if err := pg.Counts(&total); err != nil {
		return nil, fmt.Errorf("failed to count holes: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.Hole]{
		Items: items,
		Total: total,
	}, nil
}

func getHolesFilter(filter *GetHolesFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if filter == nil {
		return mf
	}
	mf.Add(filters.NewHoleMarkerID(filter.MarkerID))
	return mf
}
