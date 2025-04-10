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

func (h *holeMarkerId) Where() (sqlStr string, args []any) {
	return "t.details_id = ?", []any{h.markerId}
}
