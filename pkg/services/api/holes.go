package api

import (
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) GetHoles(l *slog.Logger, r *http.Request, markerId int64, params *api.GetHolesParams) (*api.GetHolesResponse, error) {
	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to parse request into pagefilter details")
	}

	resp, err := s.dom.GetHoles(details, markerId, params)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to get holes")
	}

	return resp, nil
}
