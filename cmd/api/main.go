package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	apiSvc "github.com/jacobbrewer1/golf-data/pkg/services/api"
	"github.com/jacobbrewer1/golf-data/pkg/services/api/domain"
	"github.com/jacobbrewer1/web"
	"github.com/jacobbrewer1/web/health"
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

	r := mux.NewRouter()
	if err := a.Start(
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			svcRepo := repo.NewRepository(a.DBConn())
			dom := domain.NewDomain(svcRepo)
			svc := apiSvc.NewService(dom)
			api.RegisterUnauthedHandlers(r, svc,
				api.WithLogger(logging.LoggerWithComponent(l, "gateway")),
			)
			return nil
		}),
		web.WithHealthCheck(
			health.NewCheck("database", func(ctx context.Context) error {
				if err := a.DBConn().PingContext(ctx); err != nil {
					return err
				}
				return nil
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(l, "health-check"))),
			),
			health.NewCheck("vault", func(ctx context.Context) error {
				if _, err := a.VaultClient().Client().Auth().Token().LookupSelf(); err != nil {
					return err
				}
				return nil
			},
				health.WithCheckOnStatusChange(health.StandardStatusListener(logging.LoggerWithComponent(l, "health-check"))),
			),
		),
	); err != nil {
		l.Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

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
