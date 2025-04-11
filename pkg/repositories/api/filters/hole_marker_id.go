package filters

import "github.com/jacobbrewer1/pagefilter"

type holeMarkerId struct {
	markerID int64
}

// NewHoleMarkerId creates a new holeMarkerId filter.
func NewHoleMarkerId(markerID int64) pagefilter.Wherer {
	return &holeMarkerId{
		markerID: markerID,
	}
}

func (h *holeMarkerId) Where() (string, []any) {
	return "t.details_id = ?", []any{h.markerID}
}
