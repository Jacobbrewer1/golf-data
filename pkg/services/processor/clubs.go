package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
	"github.com/nats-io/nats.go/jetstream"
)

func (p *processor) Clubs(ctx context.Context) {
	cc, err := p.clubsConsumer.Consume(p.processClub)
	if err != nil {
		p.l.Error("failed to consume clubs", slog.String(logging.KeyError, err.Error()))
		return
	}

	defer cc.Stop()

	<-ctx.Done()
}

func (p *processor) processClub(clubMsg jetstream.Msg) {
	defer func() {
		if err := clubMsg.Ack(); err != nil { // nolint:revive // Traditional error handling
			p.l.Error("failed to ack club message", slog.String(logging.KeyError, err.Error()))
			return
		}
	}()

	clubReader := bytes.NewReader(clubMsg.Data())

	club := new(models.Club)
	if err := json.NewDecoder(clubReader).Decode(club); err != nil { // nolint:musttag // This is internal only
		p.l.Error("failed to decode club", slog.String(logging.KeyError, err.Error()))
		return
	}

	if err := p.dom.SaveClub(club); err != nil { // nolint:revive // Traditional error handling
		p.l.Error("failed to save club", slog.String(logging.KeyError, err.Error()))
		return
	}
}
