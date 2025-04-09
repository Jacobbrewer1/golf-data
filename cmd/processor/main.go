package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/processor"
	"github.com/jacobbrewer1/golf-data/pkg/services/processor"
	"github.com/jacobbrewer1/golf-data/pkg/services/processor/domain"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/health"
	"github.com/jacobbrewer1/web/logging"
	"github.com/nats-io/nats.go"
)

const (
	appName = "processor"
)

type App struct {
	base *web.App

	svc processor.Processor
}

func NewApp(l *slog.Logger) (*App, error) {
	base, err := web.NewApp(l)
	if err != nil {
		return nil, err
	}

	return &App{
		base: base,
	}, nil
}

func (a *App) Start() error {
	if err := a.base.Start(
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
		web.WithInClusterNatsClient(),
		web.WithNatsJetStream("golf-data", []string{"clubs", "courses"}),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			serviceRepo := repo.NewRepository(a.base.DBConn())
			serviceDomain := domain.NewDomain(serviceRepo)

			clubsConsumer, err := a.base.CreateNatsJetStreamConsumer(appName+"-clubs", "clubs")
			if err != nil {
				return fmt.Errorf("failed to create nats jetstream consumer: %w", err)
			}

			coursesConsumer, err := a.base.CreateNatsJetStreamConsumer(appName+"-courses", "courses")
			if err != nil {
				return fmt.Errorf("failed to create nats jetstream consumer: %w", err)
			}

			a.svc = processor.NewProcessor(a.base.Logger(), serviceDomain, clubsConsumer, coursesConsumer)
			return nil
		}),
		web.WithHealthCheck(
			health.NewCheck("nats", func(ctx context.Context) error {
				status := a.base.NatsClient().Status()
				switch status {
				case nats.CONNECTED,
					nats.CONNECTING,
					nats.RECONNECTING,
					nats.DRAINING_SUBS,
					nats.DRAINING_PUBS:
					return nil
				default:
					return fmt.Errorf("nats status: %s", status)
				}
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(a.base.Logger(), "health-check"))),
			),
			health.NewCheck("database", func(ctx context.Context) error {
				if err := a.base.DBConn().PingContext(ctx); err != nil {
					return fmt.Errorf("failed to ping database: %w", err)
				}
				return nil
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(a.base.Logger(), "health-check"))),
			),
		),
		web.WithIndefiniteAsyncTask("clubs-processes", a.Clubs),
		web.WithIndefiniteAsyncTask("courses-processes", a.Courses),
	); err != nil {
		a.base.Logger().Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	return nil
}

func (a *App) Clubs(ctx context.Context) {
	a.svc.Clubs(ctx)
}

func (a *App) Courses(ctx context.Context) {
	a.svc.Courses(ctx)
}

func main() {
	l := logging.NewLogger(
		logging.WithDefaultLogger(),
		logging.WithAppName(appName),
	)

	a, err := NewApp(l)
	if err != nil {
		l.Error("failed to create web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	if err := a.Start(); err != nil {
		l.Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	a.base.WaitForEnd(a.base.Shutdown)
}
