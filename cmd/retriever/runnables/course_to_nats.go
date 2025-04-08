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

type publishCourseToNats struct {
	ctx       context.Context
	l         *slog.Logger
	publisher jetstream.JetStream
	course    *models.Course
}

func NewPublishCourseToNats(
	ctx context.Context,
	l *slog.Logger,
	publisher jetstream.JetStream,
	course *models.Course,
) workerpool.Runnable {
	return &publishCourseToNats{
		ctx:       ctx,
		l:         l,
		publisher: publisher,
		course:    course,
	}
}

func (c *publishCourseToNats) Run() {
	listItem := bytes.NewBuffer(nil)
	if err := json.NewEncoder(listItem).Encode(c.course); err != nil { // nolint:musttag // This is internal only
		c.l.Error("failed to encode course", slog.String(logging.KeyError, err.Error()))
		return
	}

	if _, err := c.publisher.Publish(c.ctx, "courses", listItem.Bytes()); err != nil {
		c.l.Error("failed to publish course to nats", slog.String(logging.KeyError, err.Error()))
		return
	}

	c.l.Debug("course inserted to nats", slog.Int(logKeys.KeyCourseId, c.course.Id))
}
