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
	"github.com/jacobbrewer1/web/cache"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/utils"
)

const (
	appName = "processor"
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

	var service processor.Processor

	if err := a.Start(
		web.WithVaultClient(),
		web.WithDatabaseFromVault(),
		web.WithInClusterKubeClient(),
		web.WithRedisPool(),
		web.WithWorkerPool(),
		web.WithLeaderElection(appName),
		web.WithDependencyBootstrap(func(ctx context.Context) error {
			namespace, err := utils.GetDeployedKubernetesNamespace()
			if err != nil {
				return fmt.Errorf("failed to get deployed namespace: %w", err)
			}

			hashBucket := cache.NewServiceEndpointHashBucket(
				logging.LoggerWithComponent(l, "hash-bucket"),
				a.KubeClient(),
				appName,
				namespace,
				utils.PodName,
			)

			if err := hashBucket.Start(ctx); err != nil {
				return fmt.Errorf("failed to start hash bucket: %w", err)
			}

			serviceRepo := repo.NewRepository(a.DBConn())
			serviceDomain := domain.NewDomain(serviceRepo)
			service = processor.NewProcessor(l, hashBucket, serviceDomain, a.RedisPool())

			return nil
		}),
		web.WithIndefiniteAsyncTask("clubs-processes", service.Clubs),
	); err != nil {
		l.Error("failed to start web app", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	a.WaitForEnd()
}
