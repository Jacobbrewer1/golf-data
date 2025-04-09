package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
	"github.com/nats-io/nats.go/jetstream"
)

func (p *processor) Courses(ctx context.Context) {
	cc, err := p.coursesConsumer.Consume(p.processCourseHandler)
	if err != nil {
		p.l.Error("failed to consume courses", slog.String(logging.KeyError, err.Error()))
		return
	}

	go func() {
		<-ctx.Done()
		cc.Drain()
	}()

	defer cc.Drain()
	<-cc.Closed()
}

func (p *processor) processCourseHandler(courseMsg jetstream.Msg) {
	if err := p.processCourse(courseMsg); err != nil {
		p.l.Error("failed to process course", slog.String(logging.KeyError, err.Error()))
		if err := courseMsg.Nak(); err != nil {
			p.l.Error("failed to nak course message", slog.String(logging.KeyError, err.Error()))
			return
		}
		return
	}

	if err := courseMsg.Ack(); err != nil { // nolint:revive // Traditional error handling
		p.l.Error("failed to ack club message", slog.String(logging.KeyError, err.Error()))
		return
	}
}

func (p *processor) processCourse(courseMsg jetstream.Msg) error {
	course := new(models.Course)
	if err := json.NewDecoder(bytes.NewReader(courseMsg.Data())).Decode(course); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to decode course: %w", err)
	}

	if err := p.dom.SaveCourse(course); err != nil {
		return fmt.Errorf("failed to save course: %w", err)
	}

	return nil
}
