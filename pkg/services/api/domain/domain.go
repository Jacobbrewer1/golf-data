package domain

import (
	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

type Domain interface {
	GetClubs(paginatorDetails *pagefilter.PaginatorDetails, params *api.GetClubsParams) (*api.GetClubsResponse, error)
	GetCourses(paginatorDetails *pagefilter.PaginatorDetails, clubID int64, params *api.GetCoursesParams) (*api.GetCoursesResponse, error)
	GetMarkers(paginatorDetails *pagefilter.PaginatorDetails, courseID int64, params *api.GetMarkersParams) (*api.GetMarkersResponse, error)
	GetHoles(paginatorDetails *pagefilter.PaginatorDetails, markerId int64, params *api.GetHolesParams) (*api.GetHolesResponse, error)
}

type domain struct {
	r repo.Repository
}

func NewDomain(r repo.Repository) Domain {
	return &domain{
		r: r,
	}
}
