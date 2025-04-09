package processor

import (
	"context"
	"log/slog"

	"github.com/jacobbrewer1/golf-data/pkg/services/processor/domain"
	"github.com/nats-io/nats.go/jetstream"
)

type Processor interface {
	Clubs(ctx context.Context)
	Courses(ctx context.Context)
	Details(ctx context.Context)
	Holes(ctx context.Context)
}

type processor struct {
	l               *slog.Logger
	dom             domain.Domain
	clubsConsumer   jetstream.Consumer
	coursesConsumer jetstream.Consumer
	detailsConsumer jetstream.Consumer
	holesConsumer   jetstream.Consumer
}

func NewProcessor(l *slog.Logger, d domain.Domain, clubsConsumer, coursesConsumer, detailsConsumer, holesConsumer jetstream.Consumer) Processor {
	return &processor{
		l:               l,
		dom:             d,
		clubsConsumer:   clubsConsumer,
		coursesConsumer: coursesConsumer,
		detailsConsumer: detailsConsumer,
		holesConsumer:   holesConsumer,
	}
}
