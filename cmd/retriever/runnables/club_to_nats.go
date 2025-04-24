package runnables

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
)

type publishClubToNats struct {
	ctx       context.Context
	l         *slog.Logger
	publisher jetstream.JetStream
	club      *models.Club
}

func NewPublishClubToNats(
	ctx context.Context,
	l *slog.Logger,
	publisher jetstream.JetStream,
	club *models.Club,
) workerpool.Runnable {
	return &publishClubToNats{
		ctx:       ctx,
		l:         l,
		publisher: publisher,
		club:      club,
	}
}

func (c *publishClubToNats) Run() {
	listItem := bytes.NewBuffer(nil)
	if err := json.NewEncoder(listItem).Encode(c.club); err != nil { // nolint:musttag // This is internal only
		c.l.Error("failed to encode club", slog.String(logging.KeyError, err.Error()))
		return
	}

	if _, err := c.publisher.Publish(c.ctx, "clubs", listItem.Bytes()); err != nil {
		c.l.Error("failed to publish club to nats", slog.String(logging.KeyError, err.Error()))
		return
	}

	c.l.Debug("club inserted to nats", slog.Int(logKeys.KeyClubId, c.club.Id))
}
