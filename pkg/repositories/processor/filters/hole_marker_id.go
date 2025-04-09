package filters

import "github.com/jacobbrewer1/patcher"

type holeMarkerId struct {
	markerId int
}

func NewHoleMarkerId(markerId int) patcher.Wherer {
	return &holeMarkerId{
		markerId: markerId,
	}
}

func (h *holeMarkerId) Where() (string, []any) {
	return "t.marker_id = ?", []any{h.markerId}
}
