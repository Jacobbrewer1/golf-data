package domain

import (
	"errors"
	"fmt"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

func (d *domain) GetCourses(paginatorDetails *pagefilter.PaginatorDetails, clubID int64, params *api.GetCoursesParams) (*api.GetCoursesResponse, error) {
	filter := parseGetCoursesFilters(clubID, params)

	courses, err := d.r.GetCourses(paginatorDetails, filter)
	if err != nil {
		switch {
		case errors.Is(err, repo.ErrNoCourses):
			courses = &pagefilter.PaginatedResponse[models.Course]{
				Items: make([]*models.Course, 0),
				Total: 0,
			}
		default:
			return nil, fmt.Errorf("failed to get clubs: %w", err)
		}
	}

	respClubs := make([]api.Course, 0)
	for _, club := range courses.Items {
		respClubs = append(respClubs, *mapCourseToAPI(club))
	}

	return &api.GetCoursesResponse{
		Items: respClubs,
		Total: courses.Total,
	}, nil
}

func mapCourseToAPI(course *models.Course) *api.Course {
	return &api.Course{
		Id:   int64(course.Id),
		Name: course.Name,
	}
}

func parseGetCoursesFilters(clubID int64, params *api.GetCoursesParams) *repo.GetCoursesFilters {
	f := new(repo.GetCoursesFilters)
	f.ClubID = clubID

	return f
}
