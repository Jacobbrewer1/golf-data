package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"

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
		web.WithViperConfig(),
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
		web.WithRedisPool(),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			//svcRepo := repo.NewRepository(a.DBConn())
			//dom := domain.NewDomain(svcRepo)
			//svc := apiSvc.NewService(dom)
			//
			//rateLimiter := uhttp.NewRedisRateLimiter(a.RedisPool(), 10, 25,
			//	uhttp.WithLogger(logging.LoggerWithComponent(l, "rate-limiter")),
			//)

			//api.RegisterUnauthedHandlers(r, svc,
			//	api.WithLogger(logging.LoggerWithComponent(l, "gateway")),
			//	api.WithRateLimiter(func(_ context.Context, r *http.Request) bool {
			//		hostOnly := func(host string) string {
			//			if host == "" {
			//				return ""
			//			}
			//			host, _, err := net.SplitHostPort(host)
			//			if err != nil {
			//				a.Logger().Warn("failed to split host and port", slog.String(logging.KeyError, err.Error()))
			//				return host
			//			}
			//			return host
			//		}
			//
			//		host := hostOnly(r.RemoteAddr)
			//
			//		clientToken := r.Header.Get("X-Client-Token") // Custom header for device/app identification
			//		if clientToken == "" {
			//			clientToken = host // Fallback to remote address if no token is provided
			//		}
			//
			//		agent := r.Header.Get("User-Agent") // Custom header for device identification
			//		if agent == "" {
			//			agent = host // Fallback to remote address if no agent is provided
			//		}
			//
			//		key := fmt.Sprintf("client:%s:agent:%s", clientToken, agent)
			//		key = utils.Sha256([]byte(key))
			//		return rateLimiter.Allow(key)
			//	}),
			//)
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
			health.NewCheck("keydb", func(ctx context.Context) error {
				if _, err := a.RedisPool().DoCtx(ctx, "PING"); err != nil {
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
