package domain

import (
	api "github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	"github.com/jacobbrewer1/pagefilter"
)

type Domain interface {
	GetClubs(paginatorDetails *pagefilter.PaginatorDetails, params *api.GetClubsParams) (*api.GetClubsResponse, error)
}

type domain struct {
	r repo.Repository
}

func NewDomain(r repo.Repository) Domain {
	return &domain{
		r: r,
	}
}
