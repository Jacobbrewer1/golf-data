package filters

import "github.com/jacobbrewer1/pagefilter"

type clubNameLike struct {
	name string
}

func NewClubNameLike(name string) pagefilter.Wherer {
	return &clubNameLike{
		name: name,
	}
}

func (c *clubNameLike) Where() (sqlStr string, sqlArgs []any) {
	return "t.name LIKE ?", []any{"%" + c.name + "%"}
}
