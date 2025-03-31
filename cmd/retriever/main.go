package main

import (
	"log/slog"
	"os"

	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
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
