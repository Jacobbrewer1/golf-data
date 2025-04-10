package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	api "github.com/jacobbrewer1/golf-data/pkg/apis/specs/dataapi"
	svc "github.com/jacobbrewer1/golf-data/pkg/services/dataapi"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/logging"
)

const (
	appName = "api"
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

	if err := a.Start(); err != nil {
		l.Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	r := mux.NewRouter()
	service := svc.NewService()
	api.RegisterUnauthedHandlers(r, service,
		api.WithLogger(l),
	)

	if err := a.StartServer("api-server", &http.Server{
		Addr:              ":8080",
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}); err != nil {
		l.Error("failed to start api server", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	a.WaitForEnd()
}
