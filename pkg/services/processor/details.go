package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/nats-io/nats.go/jetstream"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
)

func (p *processor) Details(ctx context.Context) {
	cc, err := p.detailsConsumer.Consume(p.processDetailsHandler)
	if err != nil {
		p.l.Error("failed to consume details", slog.String(logging.KeyError, err.Error()))
		return
	}

	go func() {
		<-ctx.Done()
		cc.Drain()
	}()

	defer cc.Drain()
	<-cc.Closed()
}

func (p *processor) processDetailsHandler(msg jetstream.Msg) {
	if err := p.processDetails(msg); err != nil {
		p.l.Error("failed to process details", slog.String(logging.KeyError, err.Error()))
		if err := msg.Nak(); err != nil {
			p.l.Error("failed to nak details message", slog.String(logging.KeyError, err.Error()))
			return
		}
		return
	}

	if err := msg.Ack(); err != nil { // nolint:revive // Traditional error handling
		p.l.Error("failed to ack details message", slog.String(logging.KeyError, err.Error()))
		return
	}
}

func (p *processor) processDetails(msg jetstream.Msg) error {
	details := new(models.CourseDetails)
	if err := json.NewDecoder(bytes.NewReader(msg.Data())).Decode(details); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to decode details: %w", err)
	}

	if err := p.dom.SaveDetails(details); err != nil {
		return fmt.Errorf("failed to save details: %w", err)
	}

	return nil
}
