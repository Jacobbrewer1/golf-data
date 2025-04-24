package pagefilter

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/jacobbrewer1/pagefilter/common"
)

var (
	ErrInvalidPaginatorDetails = errors.New("invalid paginator details")
)

// PaginatorDetails contains pagination details
type PaginatorDetails struct {
	Limit          int
	LastVal        string
	LastID         string
	Offset         int
	SortBy         string
	SortDir        string
	sortComparator string
	sortOperator   string
}

// DetailsFromRequest retrieves the paginator details from the request.
func DetailsFromRequest(req *http.Request) (*PaginatorDetails, error) {
	q := req.URL.Query()

	limit, err := getLimit(q)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, ErrInvalidPaginatorDetails)
	}

	offset, err := getOffset(q)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, ErrInvalidPaginatorDetails)
	}

	return &PaginatorDetails{
		Limit:   limit,
		LastVal: q.Get(QueryLastVal),
		LastID:  q.Get(QueryLastID),
		Offset:  offset,
		SortBy:  q.Get(QuerySortBy),
		SortDir: q.Get(QuerySortDir),
	}, nil
}

func getLimit(q url.Values) (int, error) {
	limit := queryLimitDefault
	limitStr := q.Get(QueryLimit)
	if limitStr == "" {
		return limit, nil
	}
	var err error
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("invalid limit: %w", err)
	}
	if limit > queryLimitMax {
		return queryLimitMax, nil
	}
	if limit < queryLimitMin {
		return queryLimitMin, nil
	}
	return limit, nil
}

func getOffset(q url.Values) (int, error) {
	offsetStr := q.Get(QueryOffset)
	if offsetStr == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return 0, fmt.Errorf("invalid offset: %w", err)
	}
	return offset, nil
}

// GetPaginatorDetails loads paginator details from a request. Requests have each pagination detail determined
// separately by codegen.
func GetPaginatorDetails(
	limit *common.LimitParam,
	lastVal *common.LastValue,
	lastID *common.LastId,
	offset *common.Offset,
	sortBy *common.SortBy,
	sortDir *common.SortDirection,
) *PaginatorDetails {
	d := new(PaginatorDetails)
	if lastID != nil {
		d.LastID = *lastID
	}
	if lastVal != nil {
		d.LastVal = *lastVal
	}
	if limit != nil {
		d.Limit = *limit
	}
	if offset != nil {
		d.Offset = *offset
	}
	if sortBy != nil {
		d.SortBy = *sortBy
	}
	if sortDir != nil {
		d.SortDir = string(*sortDir)
	}
	if d.Limit <= 0 {
		d.Limit = queryLimitDefault
	}
	if d.Limit > queryLimitMax {
		d.Limit = queryLimitMax
	}
	return d
}

// RemoveLimit removes the limit from the paginator details.
func (p *PaginatorDetails) RemoveLimit() {
	p.Limit = -1
}
