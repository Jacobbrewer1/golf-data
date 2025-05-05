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

	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/vaulty/vsql"
	"github.com/jacobbrewer1/web/cache"
	"github.com/jacobbrewer1/web/health"
	"github.com/jacobbrewer1/web/k8s"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/workerpool"
)

// Constants used for in-cluster NATS communication and leader election timing configurations.
const (
	// inClusterNatsEndpoint is the default NATS endpoint for in-cluster communication.
	inClusterNatsEndpoint = "nats://nats-headless.nats:4222"

	// leaderElectionLeaseDuration specifies the duration that non-leader candidates
	// will wait to forcefully acquire leadership if the current leader fails to renew.
	leaderElectionLeaseDuration = 15 * time.Second

	// leaderElectionRenewDeadline specifies the duration that the acting leader
	// will attempt to renew its leadership before giving up.
	leaderElectionRenewDeadline = 10 * time.Second

	// leaderElectionRetryPeriod specifies the interval between retries for leader election actions.
	leaderElectionRetryPeriod = 2 * time.Second
)

var (
	// ErrNoHostname is a predefined error that indicates the hostname is not set.
	//
	// This error is returned when a required hostname is missing or not provided.
	ErrNoHostname = errors.New("no hostname provided")
)

// AsyncTaskFunc defines a function type for asynchronous tasks.
//
// This function type represents tasks that run asynchronously and take a context.
// The function should respect the context cancellation and terminate when the context is done.
type AsyncTaskFunc = func(context.Context)

// StartOption defines a function type for configuring the application during startup.
//
// This function type is used to apply configuration options to the application.
// It returns an error if the configuration fails.
//
// Parameters:
//   - *App: A pointer to the application instance to be configured.
//
// Returns:
//   - error: An error if the configuration fails, otherwise nil.
type StartOption = func(*App) error

// WithViperConfig is a StartOption that sets up the viper configuration.
//
// This function initializes a new viper instance, sets the configuration file
// location from the application's base configuration, and reads the configuration
// into the viper instance. If an error occurs while reading the configuration file,
// it returns an error.
//
// Returns:
//   - StartOption: A function that applies the viper configuration to the application.
//
// Behavior:
//   - Sets the configuration file location using the application's base configuration.
//   - Reads the configuration file into the viper instance.
//   - Assigns the initialized viper instance to the application.
//
// Errors:
//   - Returns an error if the configuration file cannot be read.
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
//
// This function sets up a watcher on the application's configuration file using Viper.
// When the configuration file changes, the provided callback functions are executed.
//
// Parameters:
//   - fn: A variadic list of functions to be called when the configuration file changes.
//
// Behavior:
//   - Logs the name of the changed configuration file.
//   - Executes each provided callback function in the order they are passed.
//
// Returns:
//   - StartOption: A function that applies the configuration watcher to the application.
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
//
// This function initializes a Vault client using the application's base context, logger, and Viper configuration.
// The Vault client is then assigned to the application instance.
//
// Returns:
//   - StartOption: A function that applies the Vault client setup to the application.
//
// Errors:
//   - Returns an error if the Vault client cannot be initialized.
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

// WithDatabaseFromVault is a StartOption that sets up the database connection using Vault secrets.
//
// This function retrieves database credentials from Vault and uses them to establish a database connection.
// The retrieved secrets are expected to include the necessary information for connecting to the database.
//
// Returns:
//   - StartOption: A function that applies the database setup to the application.
//
// Behavior:
//   - Retrieves the database role and path from the application's Viper configuration.
//   - Fetches the database secrets from Vault using the specified role and path.
//   - Establishes a database connection using the retrieved secrets and assigns it to the application.
//
// Errors:
//   - Returns an error if the secrets are not found in Vault.
//   - Returns an error if there is an issue retrieving the secrets from Vault.
//   - Returns an error if the database connection cannot be established.
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

		a.db, err = vsql.ConnectDB(
			a.baseCtx,
			logging.LoggerWithComponent(a.l, "database_connector"),
			vc,
			vip,
			vs,
		)
		if err != nil {
			return fmt.Errorf("error creating database connector: %w", err)
		}
		return nil
	}
}

// WithInClusterKubeClient is a StartOption that sets up the in-cluster Kubernetes client.
//
// This function configures the application to use the Kubernetes client for in-cluster communication.
// It retrieves the in-cluster configuration and initializes a Kubernetes client instance.
//
// Returns:
//   - StartOption: A function that applies the in-cluster Kubernetes client setup to the application.
//
// Behavior:
//   - Retrieves the in-cluster configuration using Kubernetes' `rest.InClusterConfig`.
//   - Creates a Kubernetes client using the retrieved configuration.
//   - Assigns the created Kubernetes client to the application instance.
//
// Errors:
//   - Returns an error if the in-cluster configuration cannot be retrieved.
//   - Returns an error if the Kubernetes client cannot be created.
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

// WithLeaderElection is a StartOption that sets up leader election using Kubernetes lease locks.
//
// This function configures leader election for the application using Kubernetes' lease lock mechanism.
// It ensures that only one instance of the application acts as the leader at any given time.
//
// Parameters:
//   - lockName: The name of the lease lock resource to be created. This is typically the name of the application.
//
// Returns:
//   - StartOption: A function that applies the leader election setup to the application.
//
// Behavior:
//   - Validates that the pod name and lock name are set.
//   - Configures Kubernetes client for leader election.
//   - Sets up leader election callbacks for handling leadership changes.
//
// Errors:
//   - Returns ErrNoHostname if the pod name is not set.
//   - Returns an error if the lock name is empty.
//   - Returns an error if the leader election setup fails.
func WithLeaderElection(lockName string) StartOption {
	return func(a *App) error {
		switch {
		case k8s.PodName() == "":
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
					Namespace: k8s.DeployedNamespace(),
				},
				Client: kubeClient.CoordinationV1(),
				LockConfig: resourcelock.ResourceLockConfig{
					Identity: k8s.PodName(),
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

// WithHealthCheck is a StartOption that sets up the health check server.
//
// This function initializes a health check server for the application. It ensures
// that only one health check server is registered. The server uses the provided
// health checks to monitor the application's health status.
//
// Parameters:
//   - checks: A variadic list of health.Check pointers representing the health checks
//     to be added to the health checker.
//
// Returns:
//   - StartOption: A function that applies the health check server setup to the application.
//
// Behavior:
//   - Ensures that a health check server is not already registered.
//   - Creates a new health checker and adds the provided health checks to it.
//   - Registers the health check server with the application.
//
// Errors:
//   - Returns an error if a health check server is already registered.
//   - Returns an error if the health checker cannot be created.
//   - Returns an error if any of the provided health checks cannot be added.
func WithHealthCheck(checks ...*health.Check) StartOption {
	return func(a *App) error {
		if _, exists := a.servers.Load("health"); exists {
			return errors.New("health check server already registered")
		}

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

// WithRedisPool is a StartOption that sets up the Redis connection pool.
//
// This function initializes a Redis connection pool using the application's Viper configuration
// and Vault client. It retrieves the Redis credentials from Vault and uses them to configure
// the connection pool.
//
// Returns:
//   - StartOption: A function that applies the Redis pool setup to the application.
//
// Behavior:
//   - Retrieves the Redis credentials (e.g., password) from Vault using the configured path.
//   - Configures the Redis connection pool with the retrieved credentials and other settings
//     from the application's Viper configuration.
//   - Assigns the initialized Redis pool to the application instance.
//
// Errors:
//   - Returns an error if the Redis credentials cannot be retrieved from Vault.
//   - Returns an error if the password type in the Vault secret is invalid.
//   - Returns an error if the Redis connection pool cannot be created.
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

// WithMetricsEnabled is a StartOption that enables or disables metrics for the application.
//
// This function sets the `metricsEnabled` flag in the application instance, which determines
// whether metrics collection is enabled.
//
// Parameters:
//   - metricsEnabled: A boolean value indicating whether metrics should be enabled.
//
// Returns:
//   - StartOption: A function that applies the metrics configuration to the application.
//
// Behavior:
//   - Updates the `metricsEnabled` field in the application instance.
func WithMetricsEnabled(metricsEnabled bool) StartOption {
	return func(a *App) error {
		a.metricsEnabled = metricsEnabled
		return nil
	}
}

// WithWorkerPool is a StartOption that sets up the worker pool for the application.
//
// This function initializes a new worker pool with delayed start enabled and assigns it
// to the application instance.
//
// Returns:
//   - StartOption: A function that applies the worker pool setup to the application.
//
// Behavior:
//   - Creates a new worker pool with delayed start enabled.
//   - Assigns the created worker pool to the application instance.
func WithWorkerPool() StartOption {
	return func(a *App) error {
		wp := workerpool.New(
			workerpool.WithDelayedStart(),
		)

		a.workerPool = wp
		return nil
	}
}

// WithDependencyBootstrap is a StartOption that bootstraps application dependencies.
//
// This function allows the user to provide a custom dependency bootstrap function, which
// is executed during the application startup process.
//
// Parameters:
//   - fn: A function that takes a context and performs dependency initialization.
//
// Returns:
//   - StartOption: A function that applies the dependency bootstrap to the application.
//
// Behavior:
//   - Executes the provided dependency bootstrap function with the application's base context.
//
// Errors:
//   - Returns an error if the provided bootstrap function fails.
func WithDependencyBootstrap(fn func(ctx context.Context) error) StartOption {
	return func(a *App) error {
		return fn(a.baseCtx)
	}
}

// WithIndefiniteAsyncTask is a StartOption that sets up an indefinite asynchronous task.
//
// This function registers an asynchronous task with the application that runs indefinitely.
// The task is stored in the application's `indefiniteAsyncTasks` map and can be executed
// during the application's lifecycle.
//
// Parameters:
//   - name: A string representing the name of the asynchronous task.
//   - fn: An AsyncTaskFunc representing the function to be executed asynchronously.
//
// Returns:
//   - StartOption: A function that applies the indefinite async task setup to the application.
//
// Behavior:
//   - Stores the provided asynchronous task function in the application's `indefiniteAsyncTasks` map.
func WithIndefiniteAsyncTask(name string, fn AsyncTaskFunc) StartOption {
	return func(a *App) error {
		a.indefiniteAsyncTasks.Store(name, fn)
		return nil
	}
}

// WithFixedHashBucket is a StartOption that sets up a fixed-size hash bucket.
//
// This function initializes a fixed-size hash bucket using the provided size and assigns
// it to the application instance. The hash bucket is used for caching or other purposes
// requiring a fixed-size data structure.
//
// Parameters:
//   - size: An unsigned integer representing the size of the hash bucket.
//
// Returns:
//   - StartOption: A function that applies the fixed hash bucket setup to the application.
//
// Behavior:
//   - Creates a new fixed-size hash bucket using the provided size.
//   - Assigns the created hash bucket to the application's `fixedHashBucket` field.
func WithFixedHashBucket(size uint) StartOption {
	return func(a *App) error {
		hb := cache.NewFixedHashBucket(size)
		a.fixedHashBucket = hb
		return nil
	}
}

// WithServiceEndpointHashBucket is a StartOption that sets up the service endpoint hash bucket.
//
// This function initializes a service endpoint hash bucket for the application. The hash bucket
// is used to manage service endpoints in a Kubernetes environment.
//
// Parameters:
//   - appName: A string representing the name of the application.
//
// Returns:
//   - StartOption: A function that applies the service endpoint hash bucket setup to the application.
//
// Behavior:
//   - Creates a new service endpoint hash bucket using the provided application name and Kubernetes client.
//   - Assigns the created hash bucket to the application's `serviceEndpointHashBucket` field.
//   - Starts the service endpoint hash bucket with the application's base context.
//
// Errors:
//   - Returns an error if the service endpoint hash bucket fails to start.
func WithServiceEndpointHashBucket(appName string) StartOption {
	return func(a *App) error {
		sb := cache.NewServiceEndpointHashBucket(
			logging.LoggerWithComponent(a.l, "service_endpoint_hash_bucket"),
			a.KubeClient(),
			appName,
			k8s.DeployedNamespace(),
			k8s.PodName(),
		)

		a.serviceEndpointHashBucket = sb

		return sb.Start(a.baseCtx)
	}
}

// WithNatsClient is a StartOption that sets up the NATS client.
//
// This function initializes a NATS client using the provided target endpoint.
// The created client is assigned to the application's `natsClient` field.
//
// Parameters:
//   - target: A string representing the NATS server endpoint.
//
// Returns:
//   - StartOption: A function that applies the NATS client setup to the application.
//
// Behavior:
//   - Validates that the target endpoint is not empty.
//   - Connects to the NATS server using the provided target endpoint.
//   - Assigns the created NATS client to the application.
//
// Errors:
//   - Returns an error if the target endpoint is empty.
//   - Returns an error if the connection to the NATS server fails.
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

// WithInClusterNatsClient is a StartOption that sets up the NATS client with the in-cluster endpoint.
//
// This function is a wrapper around `WithNatsClient` that uses the default in-cluster NATS endpoint.
//
// Returns:
//   - StartOption: A function that applies the in-cluster NATS client setup to the application.
//
// Behavior:
//   - Calls `WithNatsClient` with the predefined in-cluster NATS endpoint.
func WithInClusterNatsClient() StartOption {
	return WithNatsClient(inClusterNatsEndpoint)
}

// WithNatsJetStream is a StartOption that sets up NATS JetStream with the given stream name, retention policy, and subjects.
//
// This function initializes a JetStream instance using the application's NATS client. It creates or updates
// a stream with the specified configuration and assigns the JetStream instance and stream to the application.
//
// Parameters:
//   - streamName: A string representing the name of the JetStream stream to be created or updated.
//   - retentionPolicy: A jetstream.RetentionPolicy specifying the retention policy for the stream.
//   - subjects: A slice of strings representing the subjects to be associated with the stream.
//
// Returns:
//   - StartOption: A function that applies the JetStream setup to the application.
//
// Behavior:
//   - Retrieves the NATS client from the application.
//   - Initializes a JetStream instance using the NATS client.
//   - Creates or updates a stream with the specified name, retention policy, and subjects.
//   - Assigns the JetStream instance and the created stream to the application.
//
// Errors:
//   - Returns an error if the JetStream instance cannot be created.
//   - Returns an error if the stream cannot be created or updated.
//   - Returns an error if the stream cannot be retrieved.
func WithNatsJetStream(streamName string, retentionPolicy jetstream.RetentionPolicy, subjects []string) StartOption {
	return func(a *App) error {
		natsClient := a.NatsClient()

		js, err := jetstream.New(natsClient)
		if err != nil {
			return fmt.Errorf("failed to create jetstream: %w", err)
		}
		a.natsJetStream = js

		_, err = js.CreateOrUpdateStream(a.baseCtx, jetstream.StreamConfig{
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

// WithKubernetesPodInformer is a StartOption that sets up a Kubernetes Pod informer.
//
// This function initializes a Kubernetes SharedInformerFactory and creates an informer
// for Kubernetes Pod objects. The informer is used to watch and cache Pod resources
// in the Kubernetes cluster.
//
// Parameters:
//   - informerOptions: A variadic list of SharedInformerOption values to configure the SharedInformerFactory.
//
// Returns:
//   - StartOption: A function that applies the Kubernetes Pod informer setup to the application.
//
// Behavior:
//   - Initializes the Kubernetes SharedInformerFactory with the provided options.
//   - Creates an informer and lister for Kubernetes Pod objects.
//   - Logs the creation of the Kubernetes Pod informer.
//
// Errors:
//   - Returns an error if the Kubernetes SharedInformerFactory cannot be initialized.
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

// WithKubernetesSecretInformer is a StartOption that sets up a Kubernetes Secret informer.
//
// This function initializes a Kubernetes SharedInformerFactory and creates an informer
// for Kubernetes Secret objects. The informer is used to watch and cache Secret resources
// in the Kubernetes cluster.
//
// Parameters:
//   - informerOptions: A variadic list of SharedInformerOption values to configure the SharedInformerFactory.
//
// Returns:
//   - StartOption: A function that applies the Kubernetes Secret informer setup to the application.
//
// Behavior:
//   - Initializes the Kubernetes SharedInformerFactory with the provided options.
//   - Creates an informer and lister for Kubernetes Secret objects.
//   - Logs the creation of the Kubernetes Secret informer.
//
// Errors:
//   - Returns an error if the Kubernetes SharedInformerFactory cannot be initialized.
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
