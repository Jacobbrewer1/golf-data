package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	_ "golang.org/x/crypto/x509roots/fallback"

	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/retriever"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/health"
	"github.com/jacobbrewer1/web/logging"
)

const (
	appName = "retriever"
)

type App struct {
	base *web.App

	r repo.Repository
}

func NewApp(l *slog.Logger) (*App, error) {
	webApp, err := web.NewApp(l)
	if err != nil {
		return nil, err
	}

	return &App{
		base: webApp,
	}, nil
}

func (a *App) Start() error {
	if err := a.base.Start(
		web.WithViperConfig(),
		web.WithWorkerPool(),
		web.WithInClusterKubeClient(),
		web.WithLeaderElection(appName),
		web.WithInClusterNatsClient(),
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
		web.WithNatsJetStream("golf-data", jetstream.WorkQueuePolicy, []string{"clubs", "courses", "details", "holes"}),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			a.r = repo.NewRepository(a.base.DBConn())
			return nil
		}),
		web.WithHealthCheck(
			health.NewCheck("kube", func(ctx context.Context) error {
				if _, err := a.base.KubeClient().Discovery().ServerVersion(); err != nil {
					return fmt.Errorf("failed to get server version: %w", err)
				}
				return nil
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(a.base.Logger(), "health-check"))),
			),
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
			health.NewCheck("vault", func(ctx context.Context) error {
				if _, err := a.base.VaultClient().Client().Auth().Token().LookupSelf(); err != nil {
					return err
				}
				return nil
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(a.base.Logger(), "health-check"))),
			),
		),
		web.WithIndefiniteAsyncTask("clubs-fetcher", a.clubsTask(logging.LoggerWithComponent(a.base.Logger(), "clubs-fetcher"))),
		web.WithIndefiniteAsyncTask("courses-fetcher", a.coursesTask(logging.LoggerWithComponent(a.base.Logger(), "courses-fetcher"))),
		web.WithIndefiniteAsyncTask("details-fetcher", a.detailsTask(logging.LoggerWithComponent(a.base.Logger(), "details-fetcher"))),
	); err != nil {
		a.base.Logger().Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	return nil
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
