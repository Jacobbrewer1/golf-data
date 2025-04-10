package domain

import (
	api "github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/pagefilter"
)

type Domain interface {
	GetClubs(paginatorDetails pagefilter.PaginatorDetails) (api.GetClubsResponse, error)
}

type domain struct {
}
