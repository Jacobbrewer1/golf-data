package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
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
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

const (
	inClusterNatsEndpoint = "nats://nats-headless.nats:4222"

	leaderElectionLeaseDuration = 15 * time.Second
	leaderElectionRenewDeadline = 10 * time.Second
	leaderElectionRetryPeriod   = 2 * time.Second
)

var (
	// ErrNoHostname is returned when the hostname is not set.
	ErrNoHostname = errors.New("no hostname provided")
)

// AsyncTaskFunc is a function that performs an async task.
type AsyncTaskFunc = func(ctx context.Context)

type StartOption func(*App) error

// WithViperConfig is a StartOption that sets up the viper config.
func WithViperConfig() StartOption {
	return func(a *App) error {
		vip := viper.New()
		vip.SetConfigFile(a.baseCfg.ConfigLocation)
		if err := vip.ReadInConfig(); err != nil {
			return fmt.Errorf("error reading config file into viper: %w", err)
		}
		a.vip = vip
		return nil
	}
}

// WithConfigWatchers is a StartOption that registers functions to be called when the config file changes.
func WithConfigWatchers(fn ...func()) StartOption {
	return func(a *App) error {
		vip := a.Viper()
		vip.OnConfigChange(func(e fsnotify.Event) {
			a.l.Info("Config file changed", slog.String(logging.KeyFile, e.Name))
			for _, f := range fn {
				f()
			}
		})
		vip.WatchConfig()
		return nil
	}
}

// WithVaultClient is a StartOption that sets up the vault client.
func WithVaultClient() StartOption {
	return func(a *App) error {
		vip := a.Viper()
		vc, err := VaultClient(a.baseCtx, logging.LoggerWithComponent(a.l, "vault"), vip)
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
		vc := a.VaultClient()
		vip := a.Viper()

		vs, err := vc.Path(
			vip.GetString("vault.database.role"),
			vaulty.WithPrefix(vip.GetString("vault.database.path")),
		).GetSecret(a.baseCtx)
		if errors.Is(err, vaulty.ErrSecretNotFound) {
			return fmt.Errorf("secrets not found in vault: %s", vip.GetString("vault.database.path"))
		} else if err != nil {
			return fmt.Errorf("error getting secrets from vault: %w", err)
		}

		dbConnector, err := repositories.NewDatabaseConnector(
			repositories.WithContext(a.baseCtx),
			repositories.WithVaultClient(vc),
			repositories.WithCurrentSecrets(vs),
			repositories.WithViper(vip),
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
		switch {
		case utils.PodName() == "":
			return ErrNoHostname
		case lockName == "":
			return errors.New("lock name cannot be empty")
		}

		kubeClient := a.KubeClient()

		klog.SetSlogLogger(logging.LoggerWithComponent(a.l, "klog"))

		a.leaderChange = make(chan struct{})

		// Create the leader election
		le, err := leaderelection.NewLeaderElector(leaderelection.LeaderElectionConfig{
			Lock: &resourcelock.LeaseLock{
				LeaseMeta: v1.ObjectMeta{
					Name:      lockName,
					Namespace: utils.DeployedNamespace(),
				},
				Client: kubeClient.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: utils.PodName(),
				},
			},
			LeaseDuration: leaderElectionLeaseDuration,
			RenewDeadline: leaderElectionRenewDeadline,
			RetryPeriod:   leaderElectionRetryPeriod,
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
		checker, err := health.NewChecker()
		if err != nil {
			return fmt.Errorf("error creating health checker: %w", err)
		}

		for _, check := range checks {
			if err := checker.AddCheck(check); err != nil {
				return fmt.Errorf("error adding health check %s: %w", check.String(), err)
			}
		}

		a.servers.Store("health", &http.Server{
			Addr:              fmt.Sprintf(":%d", HealthPort),
			Handler:           checker.Handler(),
			ReadHeaderTimeout: httpReadHeaderTimeout,
		})

		return nil
	}
}

// WithRedisPool is a StartOption that sets up the redis pool.
func WithRedisPool() StartOption {
	return func(a *App) error {
		vip := a.Viper()
		vc := a.VaultClient()

		keydbPath := vip.GetString("vault.keydb.name")
		keydbSecret, err := vc.Path(keydbPath).GetKvSecretV2(a.baseCtx)
		if errors.Is(err, vaulty.ErrSecretNotFound) {
			return fmt.Errorf("keydb secrets not found in vault path: %s", keydbPath)
		} else if err != nil {
			return fmt.Errorf("error getting keydb secrets from vault: %w", err)
		}

		redisPassword, ok := keydbSecret.Data["password"].(string)
		if !ok {
			return fmt.Errorf("invalid password type in keydb secret at path: %s", keydbPath)
		}

		rp, err := goredis.NewPool(
			goredis.WithLogger(logging.LoggerWithComponent(a.l, "goredis")),
			goredis.WithAddress(vip.GetString("keydb.address")),
			goredis.WithNetwork(vip.GetString("keydb.network")),
			goredis.WithDialOpts(
				redis.DialPassword(redisPassword),
				redis.DialDatabase(vip.GetInt("keydb.database")),
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
		sb := cache.NewServiceEndpointHashBucket(
			logging.LoggerWithComponent(a.l, "service_endpoint_hash_bucket"),
			a.KubeClient(),
			appName,
			utils.DeployedNamespace(),
			utils.PodName(),
		)

		a.serviceEndpointHashBucket = sb

		return sb.Start(a.baseCtx)
	}
}

// WithNatsClient is a StartOption that sets up the nats client.
func WithNatsClient(target string) StartOption {
	return func(a *App) error {
		if target == "" {
			return errors.New("target cannot be empty")
		}

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

// WithNatsJetStream is a StartOption that sets up nats jetstream with the given stream name, retention policy, and subjects.
func WithNatsJetStream(streamName string, retentionPolicy jetstream.RetentionPolicy, subjects []string) StartOption {
	return func(a *App) error {
		natsClient := a.NatsClient()

		js, err := jetstream.New(natsClient)
		if err != nil {
			return fmt.Errorf("failed to create jetstream: %w", err)
		}
		a.natsJetStream = js

		_, err = js.CreateStream(a.baseCtx, jetstream.StreamConfig{
			Name:      streamName,
			Subjects:  subjects,
			Storage:   jetstream.FileStorage,
			Retention: retentionPolicy,
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

// WithKubernetesPodInformer is a StartOption that initialises a Kubernetes SharedInformerFactory and informer for Kubernetes Pod objects.
func WithKubernetesPodInformer(informerOptions ...informers.SharedInformerOption) StartOption {
	return func(a *App) error {
		initKubernetesInformerFactory(a, informerOptions...)

		a.l.Info("creating kubernetes pod informer")
		base := a.kubernetesInformerFactory.Core().V1().Pods()
		a.podInformer = base.Informer()
		a.podLister = base.Lister()
		return nil
	}
}

// WithKubernetesSecretInformer is a StartOption that initialises a Kubernetes SharedInformerFactory and informer for Kubernetes Secret objects.
func WithKubernetesSecretInformer(informerOptions ...informers.SharedInformerOption) StartOption {
	return func(a *App) error {
		initKubernetesInformerFactory(a, informerOptions...)

		a.l.Info("creating kubernetes secret informer")
		base := a.kubernetesInformerFactory.Core().V1().Secrets()
		a.secretInformer = base.Informer()
		a.secretLister = base.Lister()
		return nil
	}
}
