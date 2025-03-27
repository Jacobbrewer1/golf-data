package runnables

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
)

type clubToKeyDB struct {
	ctx   context.Context
	l     *slog.Logger
	keydb goredis.Pool
	club  *models.Club
}

func NewClubToKeyDB(
	ctx context.Context,
	l *slog.Logger,
	keydb goredis.Pool,
	club *models.Club,
) workerpool.Runnable {
	return &clubToKeyDB{
		ctx:   ctx,
		l:     l,
		keydb: keydb,
		club:  club,
	}
}

func (c *clubToKeyDB) Run() {
	listItem := bytes.NewBuffer(nil)
	if err := json.NewEncoder(listItem).Encode(c.club); err != nil {
		c.l.Error("failed to encode club", slog.String(logging.KeyError, err.Error()))
		return
	}

	if _, err := c.keydb.DoCtx(c.ctx, "RPUSH", "eg_clubs", listItem.String()); err != nil {
		c.l.Error("failed to insert club to keydb", slog.String(logging.KeyError, err.Error()))
		return
	}

	c.l.Info("club inserted to keydb", slog.Int("club_id", c.club.Id))
}
