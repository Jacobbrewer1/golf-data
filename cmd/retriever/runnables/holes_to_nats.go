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

type publishHolesToNats struct {
	ctx       context.Context
	l         *slog.Logger
	publisher jetstream.JetStream
	holes     []*models.Hole
}

func NewPublishHolesToNats(
	ctx context.Context,
	l *slog.Logger,
	publisher jetstream.JetStream,
	holes []*models.Hole,
) workerpool.Runnable {
	return &publishHolesToNats{
		ctx:       ctx,
		l:         l,
		publisher: publisher,
		holes:     holes,
	}
}

func (c *publishHolesToNats) Run() {
	for _, hole := range c.holes {
		listItem := bytes.NewBuffer(nil)
		if err := json.NewEncoder(listItem).Encode(hole); err != nil { // nolint:musttag // This is internal only
			c.l.Error("failed to encode hole", slog.String(logging.KeyError, err.Error()))
			return
		}

		if _, err := c.publisher.Publish(c.ctx, "holes", listItem.Bytes()); err != nil {
			c.l.Error("failed to publish hole to nats", slog.String(logging.KeyError, err.Error()))
			return
		}

		c.l.Debug("hole inserted to nats", slog.Int(logKeys.KeyDetailsId, hole.DetailsId))
	}
}
