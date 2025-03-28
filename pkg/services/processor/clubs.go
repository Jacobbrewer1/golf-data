package processor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
	logKeys "github.com/jacobbrewer1/golf-data/pkg/logging"
	"github.com/jacobbrewer1/golf-data/pkg/models"
	"github.com/jacobbrewer1/web/logging"
)

func (p *processor) Clubs(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			got, err := redis.String(p.keydb.DoCtx(ctx, "BLPOP", "eg:clubs"))
			if err != nil {
				p.l.Error("failed to pop club", slog.String(logging.KeyError, err.Error()))
				continue
			} else if got == "" {
				p.l.Warn("no club to process")
				continue
			}

			clubReader := strings.NewReader(got)
			if err := p.processClub(clubReader); err != nil {
				p.l.Error("failed to process club", slog.String(logging.KeyError, err.Error()))
				continue
			}
		}
	}
}

func (p *processor) processClub(clubReader io.Reader) error {
	club := new(models.Club)
	if err := json.NewDecoder(clubReader).Decode(club); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to decode club: %w", err)
	}

	if !p.bucket.InBucket(strconv.Itoa(club.Id)) {
		p.l.Debug("club not in bucket, republishing to keydb", slog.Int(logKeys.KeyClubId, club.Id))
		if err := p.republishClub(club); err != nil {
			return fmt.Errorf("failed to republish club: %w", err)
		}
		return nil
	}

	if err := p.dom.SaveClub(club); err != nil {
		return fmt.Errorf("failed to save club: %w", err)
	}

	return nil
}

func (p *processor) republishClub(club *models.Club) error {
	clubBytes := bytes.NewBuffer(nil)
	if err := json.NewEncoder(clubBytes).Encode(club); err != nil { // nolint:musttag // This is internal only
		return fmt.Errorf("failed to encode club: %w", err)
	}

	if _, err := p.keydb.Do("RPUSH", "eg:clubs", clubBytes); err != nil {
		return fmt.Errorf("failed to republish club: %w", err)
	}

	return nil
}
