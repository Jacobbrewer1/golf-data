package filters

import "github.com/jacobbrewer1/pagefilter"

type courseClubID struct {
	clubID int64
}

func NewCourseClubID(clubID int64) pagefilter.Wherer {
	return &courseClubID{
		clubID: clubID,
	}
}

func (c *courseClubID) Where() (string, []any) {
	return "t.club_id = ?", []any{c.clubID}
}
