package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/vaulty/repositories"
	"github.com/jacobbrewer1/web/cache"
	"github.com/jacobbrewer1/web/health"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/utils"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

const (
	inClusterNatsEndpoint = "nats://nats-headless.nats:4222"
)

var (
	ErrNilVaultClient = errors.New("nil vault client")
	ErrNoHostname     = errors.New("no hostname provided")
)

// AsyncTaskFunc is a function that performs an async task.
type AsyncTaskFunc = func(ctx context.Context)

type StartOption func(*App) error

// WithVaultClient is a StartOption that sets up the vault client.
func WithVaultClient() StartOption {
	return func(a *App) error {
		vc, err := VaultClient(a.baseCtx, logging.LoggerWithComponent(a.l, "vault"), a.vip)
		if err != nil {
			return fmt.Errorf("error getting vault client: %w", err)
		}

		a.vaultClient = vc
		return nil
	}
}

// WithDatabaseFromVault is a StartOption that sets up the database from vault.
func WithDatabaseFromVault() StartOption {
	return func(a *App) error {
		vc := a.vaultClient
		if vc == nil {
			return ErrNilVaultClient
		}

		vs, err := vc.Path(
			a.vip.GetString("vault.database.role"),
			vaulty.WithPrefix(a.vip.GetString("vault.database.path")),
		).GetSecret(a.baseCtx)
		if errors.Is(err, vaulty.ErrSecretNotFound) {
			return fmt.Errorf("secrets not found in vault: %s", a.vip.GetString("vault.database.path"))
		} else if err != nil {
			return fmt.Errorf("error getting secrets from vault: %w", err)
		}

		dbConnector, err := repositories.NewDatabaseConnector(
			repositories.WithContext(a.baseCtx),
			repositories.WithVaultClient(vc),
			repositories.WithCurrentSecrets(vs),
			repositories.WithViper(a.vip),
			repositories.WithConnectorLogger(logging.LoggerWithComponent(a.l, "database_connector")),
		)
		if err != nil {
			return fmt.Errorf("error creating database connector: %w", err)
		}

		db, err := dbConnector.ConnectDB()
		if err != nil {
			return fmt.Errorf("error connecting to database: %w", err)
		}

		a.db = db
		return nil
	}
}

// WithInClusterKubeClient is a StartOption that sets up the in-cluster kube client.
func WithInClusterKubeClient() StartOption {
	return func(a *App) error {
		cfg, err := rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("failed to get in-cluster config: %w", err)
		}

		kubeClient, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		a.kubeClient = kubeClient
		return nil
	}
}

// WithLeaderElection is a StartOption that sets up leader election.
func WithLeaderElection(lockName string) StartOption {
	return func(a *App) error {
		if a.kubeClient == nil {
			return errors.New("must set up kube client before leader election, ensure WithInClusterKubeClient is called")
		} else if utils.PodName == "" {
			return ErrNoHostname
		}

		klog.SetSlogLogger(logging.LoggerWithComponent(a.l, "klog"))

		a.leaderChange = make(chan struct{})

		// Get the deployed namespace
		ns, err := utils.GetDeployedKubernetesNamespace()
		if err != nil {
			return fmt.Errorf("failed to get deployed namespace: %w", err)
		}

		// Create the leader election
		le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
			Lock: &resourcelock.LeaseLock{
				LeaseMeta: v1.ObjectMeta{
					Name:      lockName,
					Namespace: ns,
				},
				Client: a.kubeClient.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: utils.PodName,
				},
			},
			LeaseDuration: 15 * time.Second,
			RenewDeadline: 10 * time.Second,
			RetryPeriod:   2 * time.Second,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					a.l.Info("Started leading")
				},
				OnStoppedLeading: func() {
					a.l.Info("Stopped leading")
				},
				OnNewLeader: func(identity string) {
					a.l.Info("Leader Changed",
						slog.String(logging.KeyIdentity, identity),
					)

					select {
					case a.leaderChange <- struct{}{}:
					default:
						// Prevent blocking
					}
				},
			},
			ReleaseOnCancel: true,
		})
		if err != nil {
			return fmt.Errorf("failed to create leader election: %w", err)
		}

		a.leaderElection = le
		return nil
	}
}

// WithHealthCheck is a StartOption that sets up the health2 check.
func WithHealthCheck(checks ...*health.Check) StartOption {
	return func(a *App) error {
		checkerOpts := make([]health.CheckerOption, 0)
		for _, check := range checks {
			checkerOpts = append(checkerOpts, health.WithCheckerCheck(check))
		}

		checker := health.NewChecker(checkerOpts...)

		a.servers.Store("health", &http.Server{
			Addr:              fmt.Sprintf(":%d", HealthPort),
			Handler:           checker.Handler(),
			ReadHeaderTimeout: 10 * time.Second,
		})

		return nil
	}
}

// WithRedisPool is a StartOption that sets up the redis pool.
func WithRedisPool() StartOption {
	return func(a *App) error {
		vc := a.vaultClient
		if vc == nil {
			return ErrNilVaultClient
		}

		keydbSecret, err := vc.Path(a.vip.GetString("vault.keydb.name")).GetKvSecretV2(a.baseCtx)
		if errors.Is(err, vaulty.ErrSecretNotFound) {
			return fmt.Errorf("secrets not found in vault: %s", a.vip.GetString("vault.keydb.path"))
		} else if err != nil {
			return fmt.Errorf("error getting secrets from vault: %w", err)
		}

		redisPassword, ok := keydbSecret.Data["password"].(string)
		if !ok {
			return errors.New("keydb type assertion failed")
		}

		rp, err := goredis.NewPool(
			goredis.WithLogger(logging.LoggerWithComponent(a.l, "goredis")),
			goredis.WithAddress(a.vip.GetString("keydb.address")),
			goredis.WithNetwork(a.vip.GetString("keydb.network")),
			goredis.WithDialOpts(
				redis.DialPassword(redisPassword),
				redis.DialDatabase(a.vip.GetInt("keydb.database")),
			),
		)
		if err != nil {
			return fmt.Errorf("error creating redis pool: %w", err)
		}

		a.redisPool = rp
		return nil
	}
}

// WithMetricsEnabled is a StartOption that enables metrics.
func WithMetricsEnabled(metricsEnabled bool) StartOption {
	return func(a *App) error {
		a.metricsEnabled = metricsEnabled
		return nil
	}
}

// WithWorkerPool is a StartOption that sets up the worker pool.
func WithWorkerPool() StartOption {
	return func(a *App) error {
		wp := workerpool.New(
			workerpool.WithDelayedStart(),
		)

		a.workerPool = wp
		return nil
	}
}

// WithDependencyBootstrap is a StartOption that bootstraps dependencies.
func WithDependencyBootstrap(fn func(ctx context.Context) error) StartOption {
	return func(a *App) error {
		return fn(a.baseCtx)
	}
}

// WithIndefiniteAsyncTask is a StartOption that sets up an indefinite async task.
func WithIndefiniteAsyncTask(name string, fn AsyncTaskFunc) StartOption {
	return func(a *App) error {
		a.indefiniteAsyncTasks.Store(name, fn)
		return nil
	}
}

// WithServiceEndpointHashBucket is a StartOption that sets up the service endpoint hash bucket.
func WithServiceEndpointHashBucket(appName string) StartOption {
	return func(a *App) error {
		if a.kubeClient == nil {
			return errors.New("must set up kube client before service endpoint hash bucket, ensure WithInClusterKubeClient is called")
		}

		ns, err := utils.GetDeployedKubernetesNamespace()
		if err != nil {
			return fmt.Errorf("failed to get deployed namespace: %w", err)
		}

		sb := cache.NewServiceEndpointHashBucket(
			logging.LoggerWithComponent(a.l, "service_endpoint_hash_bucket"),
			a.kubeClient,
			appName,
			ns,
			utils.PodName,
		)

		a.serviceEndpointHashBucket = sb

		return sb.Start(a.baseCtx)
	}
}

// WithNatsClient is a StartOption that sets up the nats client.
func WithNatsClient(target string) StartOption {
	return func(a *App) error {
		nc, err := nats.Connect(target)
		if err != nil {
			return fmt.Errorf("failed to connect to nats: %w", err)
		}

		a.natsClient = nc
		return nil
	}
}

// WithInClusterNatsClient is a StartOption that sets up the nats client with the in-cluster endpoint.
func WithInClusterNatsClient() StartOption {
	return WithNatsClient(inClusterNatsEndpoint)
}

// WithNatsJetStream is a StartOption that sets up nats jetstream.
func WithNatsJetStream(streamName string, subjects []string) StartOption {
	return func(a *App) error {
		js, err := jetstream.New(a.natsClient)
		if err != nil {
			return fmt.Errorf("failed to create jetstream: %w", err)
		}
		a.natsJetStream = js

		_, err = js.CreateStream(a.baseCtx, jetstream.StreamConfig{
			Name:      streamName,
			Subjects:  subjects,
			Storage:   jetstream.FileStorage,
			Retention: jetstream.LimitsPolicy,
		})
		if err != nil && !errors.Is(err, jetstream.ErrStreamNameAlreadyInUse) {
			return fmt.Errorf("failed to create stream: %w", err)
		}

		a.natsStream, err = js.Stream(a.baseCtx, streamName)
		if err != nil {
			return fmt.Errorf("failed to get stream: %w", err)
		}

		return nil
	}
}
