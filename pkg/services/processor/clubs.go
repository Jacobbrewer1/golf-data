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

func (p *processor) Clubs(ctx context.Context) {
	cc, err := p.clubsConsumer.Consume(p.processClubHandler)
	if err != nil {
		p.l.Error("failed to consume clubs", slog.String(logging.KeyError, err.Error()))
		return
	}

	go func() {
		<-ctx.Done()
		cc.Drain()
	}()

	defer cc.Drain()
	<-cc.Closed()
}

func (p *processor) processClubHandler(clubMsg jetstream.Msg) {
	if err := p.processClub(clubMsg); err != nil {
		p.l.Error("failed to process club", slog.String(logging.KeyError, err.Error()))
		if err := clubMsg.Nak(); err != nil {
			p.l.Error("failed to nak club message", slog.String(logging.KeyError, err.Error()))
			return
		}
		return
	}

	if err := clubMsg.Ack(); err != nil { // nolint:revive // Traditional error handling
		p.l.Error("failed to ack club message", slog.String(logging.KeyError, err.Error()))
		return
	}
}

func (p *processor) processClub(clubMsg jetstream.Msg) error {
	club := new(models.Club)
	if err := json.NewDecoder(bytes.NewReader(clubMsg.Data())).Decode(club); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to decode club: %w", err)
	}

	if err := p.dom.SaveClub(club); err != nil {
		return fmt.Errorf("failed to save club: %w", err)
	}

	return nil
}
