package api

import (
	"log/slog"
	"net/http"

	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/pagefilter"
	"github.com/jacobbrewer1/uhttp"
)

func (s *service) GetMarkers(l *slog.Logger, r *http.Request, courseId int64, params *api.GetMarkersParams) (*api.GetMarkersResponse, error) {
	l.Debug("===========")

	details, err := pagefilter.DetailsFromRequest(r)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to parse request into pagefilter details")
	}

	resp, err := s.dom.GetMarkers(details, courseId, params)
	if err != nil {
		return nil, uhttp.NewHTTPError(http.StatusInternalServerError, err, "Failed to get markers")
	}

	return resp, nil
}
