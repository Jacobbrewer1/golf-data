package filters

import "github.com/jacobbrewer1/pagefilter"

type holeMarkerID struct {
	markerID int64
}

// NewHoleMarkerID creates a new holeMarkerID filter.
func NewHoleMarkerID(markerID int64) pagefilter.Wherer {
	return &holeMarkerID{
		markerID: markerID,
	}
}

func (h *holeMarkerID) Where() (sqlStr string, sqlArgs []any) {
	return "t.details_id = ?", []any{h.markerID}
}
