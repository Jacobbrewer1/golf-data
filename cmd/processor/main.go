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
	"github.com/jacobbrewer1/web/logging"
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
		web.WithNatsJetStream("golf-data", []string{"clubs"}),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			serviceRepo := repo.NewRepository(a.base.DBConn())
			serviceDomain := domain.NewDomain(serviceRepo)

			clubsConsumer, err := a.base.CreateNatsJetStreamConsumer(appName, "clubs")
			if err != nil {
				return fmt.Errorf("failed to create nats jetstream consumer: %w", err)
			}

			a.svc = processor.NewProcessor(a.base.Logger(), serviceDomain, clubsConsumer)
			return nil
		}),
		web.WithIndefiniteAsyncTask("clubs-processes", a.Clubs),
	); err != nil {
		a.base.Logger().Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	return nil
}

func (a *App) Clubs(ctx context.Context) {
	a.svc.Clubs(ctx)
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
