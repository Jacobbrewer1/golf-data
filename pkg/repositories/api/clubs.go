package api

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/golf-data/pkg/repositories/api/filters"
	"github.com/jacobbrewer1/pagefilter"
)

var (
	// ErrNoClubs is returned when no clubs are found.
	ErrNoClubs = errors.New("no clubs found")
)

func (r *repository) GetClubs(details *pagefilter.PaginatorDetails, filter *GetClubsFilters) (*pagefilter.PaginatedResponse[models.Club], error) {
	mf := r.getClubFilters(filter)
	pg := pagefilter.NewPaginator(r.db, models.ClubTableName, "id", mf)

	if err := pg.SetDetails(details); err != nil {
		return nil, fmt.Errorf("failed to set paginator details: %w", err)
	}

	pvt, err := pg.Pivot()
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoClubs
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	items := make([]*models.Club, 0)
	if err := pg.Retrieve(pvt, &items); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNoClubs
		default:
			return nil, fmt.Errorf("failed to retrieve clubs: %w", err)
		}
	}

	var total int64 = 0
	if err := pg.Counts(&total); err != nil {
		return nil, fmt.Errorf("failed to count clubs: %w", err)
	}

	return &pagefilter.PaginatedResponse[models.Club]{
		Items: items,
		Total: total,
	}, nil
}

func (r *repository) getClubFilters(filter *GetClubsFilters) *pagefilter.MultiFilter {
	mf := pagefilter.NewMultiFilter()
	if filter == nil {
		return mf
	}

	if filter.Name != nil {
		mf.Add(filters.NewClubNameLike(*filter.Name))
	}

	return mf
}
