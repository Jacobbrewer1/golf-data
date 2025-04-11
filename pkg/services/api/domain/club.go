package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

func (d *domain) GetClubs(paginatorDetails *pagefilter.PaginatorDetails, params *api.GetClubsParams) (*api.GetClubsResponse, error) {
	filter := parseGetClubFilters(params)

	clubs, err := d.r.GetClubs(paginatorDetails, filter)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoClubs):
			clubs = &pagefilter.PaginatedResponse[models.Club]{
				Items: make([]*models.Club, 0),
				Total: 0,
			}
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	respClubs := make([]api.Club, 0)
	for _, club := range clubs.Items {
		respClubs = append(respClubs, *mapClubToAPI(club))
	}

	return &api.GetClubsResponse{
		Items: respClubs,
		Total: clubs.Total,
	}, nil
}

func parseGetClubFilters(params *api.GetClubsParams) *repo.GetClubsFilters {
	f := new(repo.GetClubsFilters)

	if params.Name != nil {
		f.Name = params.Name
	}

	return f
}

func mapClubToAPI(club *models.Club) *api.Club {
	return &api.Club{
		Address1: club.Address1,
		Address2: club.Address2,
		Address3: club.Address3,
		Id:       int64(club.Id),
		Name:     club.Name,
		Postcode: club.PostalCode,
	}
}
