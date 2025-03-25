package dataapi

import (
	"log/slog"
	"net/http"

	api "github.com/jacobbrewer1/golf-data/pkg/apis/specs/dataapi"
)

func (s service) GetClubs(l *slog.Logger, r *http.Request, params *api.GetClubsParams) (*api.GetClubsResponse, error) {
	//TODO implement me
	panic("implement me")
}
