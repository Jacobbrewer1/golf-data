package api

import (
	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/pagefilter"
)

type Repository interface {
	GetClubs(details *pagefilter.PaginatorDetails, filters *GetClubsFilters) (*pagefilter.PaginatedResponse[models.Club], error)
}

type GetClubsFilters struct {
	Name *string
}
