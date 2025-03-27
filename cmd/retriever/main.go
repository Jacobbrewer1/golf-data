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

func main() {
	l := logging.NewLogger(
		logging.WithDefaultLogger(),
		logging.WithAppName(appName),
	)

	a, err := web.NewApp(l)
	if err != nil {
		l.Error("failed to create web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	if err := a.Start(
		web.WithVaultClient(),
		web.WithRedisPool(),
		web.WithWorkerPool(),
		web.WithLeaderElection(appName),
		web.WithIndefiniteAsyncTask("clubs-fetcher", clubsTask(l, a.RedisPool, a.WorkerPool, a.IsLeader, a.LeaderChange())),
	); err != nil {
		l.Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	a.WaitForEnd()
}
