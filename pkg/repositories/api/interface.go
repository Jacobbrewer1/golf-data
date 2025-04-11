package api

import (
	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/pagefilter"
)

type Repository interface {
	GetClubs(details *pagefilter.PaginatorDetails, filter *GetClubsFilters) (*pagefilter.PaginatedResponse[models.Club], error)
	GetCourses(details *pagefilter.PaginatorDetails, filter *GetCoursesFilters) (*pagefilter.PaginatedResponse[models.Course], error)
	GetMarkers(details *pagefilter.PaginatorDetails, filter *GetMarkersFilters) (*pagefilter.PaginatedResponse[models.CourseDetails], error)
	GetHoles(details *pagefilter.PaginatorDetails, filter *GetHolesFilters) (*pagefilter.PaginatedResponse[models.Hole], error)
}

type GetClubsFilters struct {
	Name *string
}

type GetCoursesFilters struct {
	ClubID int64
}

type GetMarkersFilters struct {
	CourseID int64
}

type GetHolesFilters struct {
	MarkerID int64
}
