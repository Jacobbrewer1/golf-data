package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/api"
	apiSvc "github.com/jacobbrewer1/golf-data/pkg/services/api"
	"github.com/jacobbrewer1/golf-data/pkg/services/api/domain"
	"github.com/jacobbrewer1/uhttp"
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

			rateLimiter := uhttp.NewRateLimiter(10, 50,
				uhttp.WithLogger(logging.LoggerWithComponent(l, "rate-limiter")),
			)

			api.RegisterUnauthedHandlers(r, svc,
				api.WithLogger(logging.LoggerWithComponent(l, "gateway")),
				api.WithRateLimiter(func(_ context.Context, r *http.Request) bool {
					clientToken := r.Header.Get("X-Client-Token") // Custom header for device/app identification
					if clientToken == "" {
						clientToken = r.RemoteAddr // Fallback to remote address if no token is provided

						// Remove the port from the remote address as this can change
						clientToken, _, err = net.SplitHostPort(clientToken)
						if err != nil {
							clientToken = r.RemoteAddr // Fallback to remote address if split fails
							a.Logger().Warn("failed to split host and port", slog.String(logging.KeyError, err.Error()))
						}
					}

					method := r.Method
					key := fmt.Sprintf("client:%s:method:%s", clientToken, method)
					return rateLimiter.Allow(key)
				}),
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
