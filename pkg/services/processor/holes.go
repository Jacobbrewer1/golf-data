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

func (p *processor) Holes(ctx context.Context) {
	cc, err := p.holesConsumer.Consume(p.processHoleHandler)
	if err != nil {
		p.l.Error("failed to consume Holes", slog.String(logging.KeyError, err.Error()))
		return
	}

	go func() {
		<-ctx.Done()
		cc.Drain()
	}()

	defer cc.Drain()
	<-cc.Closed()
}

func (p *processor) processHoleHandler(msg jetstream.Msg) {
	if err := p.processDetails(msg); err != nil {
		p.l.Error("failed to process hole", slog.String(logging.KeyError, err.Error()))
		if err := msg.Nak(); err != nil {
			p.l.Error("failed to nak hole message", slog.String(logging.KeyError, err.Error()))
			return
		}
		return
	}

	if err := msg.Ack(); err != nil { // nolint:revive // Traditional error handling
		p.l.Error("failed to ack hole message", slog.String(logging.KeyError, err.Error()))
		return
	}
}

func (p *processor) processHole(msg jetstream.Msg) error {
	hole := new(models.Hole)
	if err := json.NewDecoder(bytes.NewReader(msg.Data())).Decode(hole); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to decode hole: %w", err)
	}

	if err := p.dom.SaveHole(hole); err != nil {
		return fmt.Errorf("failed to save hole: %w", err)
	}

	return nil
}
