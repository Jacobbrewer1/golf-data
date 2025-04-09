package runnables

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go/jetstream"
)

type publishDetailsToNats struct {
	ctx       context.Context
	l         *slog.Logger
	publisher jetstream.JetStream
	details   *models.CourseDetails
}

func NewPublishDetailsToNats(
	ctx context.Context,
	l *slog.Logger,
	publisher jetstream.JetStream,
	details *models.CourseDetails,
) workerpool.Runnable {
	return &publishDetailsToNats{
		ctx:       ctx,
		l:         l,
		publisher: publisher,
		details:   details,
	}
}

func (c *publishDetailsToNats) Run() {
	listItem := bytes.NewBuffer(nil)
	if err := json.NewEncoder(listItem).Encode(c.details); err != nil { // nolint:musttag // This is internal only
		c.l.Error("failed to encode details", slog.String(logging.KeyError, err.Error()))
		return
	}

	if _, err := c.publisher.Publish(c.ctx, "details", listItem.Bytes()); err != nil {
		c.l.Error("failed to publish details to nats", slog.String(logging.KeyError, err.Error()))
		return
	}

	c.l.Debug("details inserted to nats", slog.Int(logKeys.KeyCourseId, c.details.CourseId))
}
