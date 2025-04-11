package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

func (d *domain) GetMarkers(paginatorDetails *pagefilter.PaginatorDetails, courseID int64, params *api.GetMarkersParams) (*api.GetMarkersResponse, error) {
	filter := parseGetMarkerFilters(courseID, params)

	clubs, err := d.r.GetMarkers(paginatorDetails, filter)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoMarkers):
			clubs = &pagefilter.PaginatedResponse[models.CourseDetails]{
				Items: make([]*models.CourseDetails, 0),
				Total: 0,
			}
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	respClubs := make([]api.Marker, 0)
	for _, club := range clubs.Items {
		respClubs = append(respClubs, *mapMarkerToAPI(club))
	}

	return &api.GetMarkersResponse{
		Items: respClubs,
		Total: clubs.Total,
	}, nil
}

func parseGetMarkerFilters(courseID int64, params *api.GetMarkersParams) *repo.GetMarkersFilters {
	f := new(repo.GetMarkersFilters)
	f.CourseID = courseID
	return f
}

func mapMarkerToAPI(marker *models.CourseDetails) *api.Marker {
	return &api.Marker{
		CourseRating:    float32(marker.CourseRating),
		Id:              int64(marker.Id),
		Marker:          marker.Marker,
		MetersBackNine:  int64(marker.MetersBackNine),
		MetersFrontNine: int64(marker.MetersFrontNine),
		MetersTotal:     int64(marker.MetersTotal),
		ParBackNine:     int64(marker.ParBackNine),
		ParFrontNine:    int64(marker.ParFrontNine),
		ParTotal:        int64(marker.ParTotal),
		SlopeRating:     int64(marker.SlopeRating),
		YardsBackNine:   int64(marker.YardsBackNine),
		YardsFrontNine:  int64(marker.YardsFrontNine),
		YardsTotal:      int64(marker.YardsTotal),
	}
}
