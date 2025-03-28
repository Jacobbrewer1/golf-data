package processor

import (
	"context"
	"log/slog"

	"github.com/jacobbrewer1/golf-data/pkg/services/processor/domain"
	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/web/cache"
)

type Processor interface {
	Clubs(ctx context.Context) error
}

type processor struct {
	l      *slog.Logger
	bucket cache.HashBucket
	dom    domain.Domain
	keydb  goredis.Pool
}

func NewProcessor(l *slog.Logger, bucket cache.HashBucket, d domain.Domain, keydb goredis.Pool) Processor {
	return &processor{
		l:      l,
		bucket: bucket,
		dom:    d,
		keydb:  keydb,
	}
}
