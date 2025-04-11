package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

func (d *domain) GetHoles(paginatorDetails *pagefilter.PaginatorDetails, markerId int64, params *api.GetHolesParams) (*api.GetHolesResponse, error) {
	filter := parseGetHolesFilters(markerId, params)

	holes, err := d.r.GetHoles(paginatorDetails, filter)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoHoles):
			holes = &pagefilter.PaginatedResponse[models.Hole]{
				Items: make([]*models.Hole, 0),
				Total: 0,
			}
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	respClubs := make([]api.Hole, 0)
	for _, club := range holes.Items {
		respClubs = append(respClubs, *mapHoleToAPI(club))
	}

	return &api.GetHolesResponse{
		Items: respClubs,
		Total: holes.Total,
	}, nil
}

func parseGetHolesFilters(markerId int64, params *api.GetHolesParams) *repo.GetHolesFilters {
	f := new(repo.GetHolesFilters)
	f.MarkerID = markerId
	return f
}

func mapHoleToAPI(hole *models.Hole) *api.Hole {
	return &api.Hole{
		DistanceMeters: int64(hole.DistanceMeters),
		DistanceYards:  int64(hole.DistanceYards),
		Number:         int64(hole.Number),
		Id:             int64(hole.Id),
		Par:            int64(hole.Par),
		StrokeIndex:    int64(hole.StrokeIndex),
	}
}
