package api

import (
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) GetClubs(l *slog.Logger, r *http.Request, params *api.GetClubsParams) (*api.GetClubsResponse, error) {
	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to parse request into pagefilter details")
	}

	resp, err := s.dom.GetClubs(details, params)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to get clubs")
	}

	return resp, nil
}
