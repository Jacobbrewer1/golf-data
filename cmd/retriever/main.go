package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
	"github.com/nats-io/nats.go"
	_ "golang.org/x/crypto/x509roots/fallback"
)

const (
	appName = "retriever"
)

type App struct {
	base *web.App
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
		web.WithWorkerPool(),
		web.WithInClusterKubeClient(),
		web.WithLeaderElection(appName),
		web.WithInClusterNatsClient(),
		web.WithNatsJetStream("golf-data", []string{"clubs"}),
		web.WithIndefiniteAsyncTask("clubs-fetcher", a.clubsTask()),
		web.WithHealthCheck(map[string]web.HealthCheckFunc{
			"kube": func(ctx context.Context) error {
				if _, err := a.base.KubeClient().Discovery().ServerVersion(); err != nil {
					return fmt.Errorf("failed to get server version: %w", err)
				}
				return nil
			},
			"nats": func(ctx context.Context) error {
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
		}),
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
