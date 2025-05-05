package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	kubeCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"

	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/uhttp"
	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/vaulty/vsql"
	"github.com/jacobbrewer1/web/cache"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/version"
	"github.com/jacobbrewer1/workerpool"
)

const (
	// httpReadHeaderTimeout specifies the maximum duration allowed to read HTTP request headers.
	httpReadHeaderTimeout = 10 * time.Second

	// shutdownTimeout specifies the maximum duration allowed for the application to shut down gracefully.
	shutdownTimeout = 15 * time.Second
)

var (
	// MetricsPort defines the port number used by the metrics server.
	MetricsPort = 9090

	// HealthPort defines the port number used by the health server.
	HealthPort = 9091
)

var (
	// ErrNilLogger is an error that indicates a nil logger was provided to the NewApp function.
	// This error is returned to prevent the application from starting without a valid logger.
	ErrNilLogger = errors.New("logger is nil")
)

type (
	// AppConfig is the configuration for the application.
	// It holds the settings required to initialize and run the application.
	AppConfig struct {
		// ConfigLocation specifies the file path to the application's configuration file.
		// This value can be set via the CONFIG_LOCATION environment variable.
		// If not provided, it defaults to "config.json".
		ConfigLocation string `env:"CONFIG_LOCATION" envDefault:"config.json"`
	}

	// App is the application struct.
	// It contains all the components and configurations required to run the application.
	App struct {
		// l is the logger for the application.
		// Used for logging application events and errors.
		l *slog.Logger

		// baseCtx is the base context for the application.
		// It serves as the root context for all operations within the application.
		baseCtx context.Context

		// baseCtxCancel is the base context cancel function.
		// Used to cancel the base context and signal shutdown.
		baseCtxCancel context.CancelFunc

		// baseCfg is the base configuration for the application.
		// Holds the application's configuration settings.
		baseCfg *AppConfig

		// isStartedChan is a channel that is closed when the application is started.
		// Used to signal that the application has completed its startup process.
		isStartedChan chan struct{}

		// startOnce ensures that the start function is only called once.
		// Provides thread-safety for the application's startup process.
		startOnce sync.Once

		// vip is the viper instance for the application.
		// Used for managing configuration settings.
		vip *viper.Viper

		// vaultClient is the vault client for the application.
		// Provides secure access to secrets and sensitive data.
		vaultClient vaulty.Client

		// metricsEnabled is a flag to enable metrics for the application.
		// When true, metrics collection is enabled.
		metricsEnabled bool

		// servers is the list of servers for the application.
		// Stores the HTTP servers managed by the application.
		servers sync.Map

		// shutdownOnce ensures that the shutdown function is only called once.
		// Provides thread-safety for the application's shutdown process.
		shutdownOnce sync.Once

		// shutdownWg is used to wait for all shutdown tasks to complete.
		// Ensures graceful shutdown of the application.
		shutdownWg *sync.WaitGroup

		// db is the database for the application.
		// Represents the application's database connection.
		db *vsql.Database

		// kubeClient interacts with the Kubernetes API server.
		// Used for managing Kubernetes resources.
		kubeClient kubernetes.Interface

		// kubernetesInformerFactory is a factory used for initializing a pod informer.
		// Provides shared informers for Kubernetes resources.
		kubernetesInformerFactory informers.SharedInformerFactory

		// podInformer is an informer for Kubernetes Pod objects.
		// Watches for changes to Pod resources in the cluster.
		podInformer kubeCache.SharedIndexInformer

		// podLister is a lister for Kubernetes Pod objects.
		// Provides read-only access to Pod resources.
		podLister listersv1.PodLister

		// secretInformer is an informer for Kubernetes Secret objects.
		// Watches for changes to Secret resources in the cluster.
		secretInformer kubeCache.SharedIndexInformer

		// secretLister is a lister for Kubernetes Secret objects.
		// Provides read-only access to Secret resources.
		secretLister listersv1.SecretLister

		// leaderElection is the leader election for the application.
		// Manages leader election for distributed systems.
		leaderElection *leaderelection.LeaderElector

		// leaderChange is the channel that is notified when the leader changes.
		// Used to signal changes in leadership status.
		leaderChange chan struct{}

		// redisPool is the redis pool for the application.
		// Manages connections to the Redis database.
		redisPool goredis.Pool

		// workerPool is the worker pool that can execute tasks concurrently.
		// Provides a pool of workers for handling concurrent tasks.
		workerPool workerpool.Pool

		// indefiniteAsyncTasks is the list of indefinite async tasks for the application.
		// Stores tasks that run indefinitely until the application shuts down.
		indefiniteAsyncTasks sync.Map

		// fixedHashBucket is the fixed hash bucket for the application.
		// Used for consistent hashing of data.
		fixedHashBucket *cache.FixedHashBucket

		// serviceEndpointHashBucket is the service endpoint hash bucket for the application.
		// Manages consistent hashing for service endpoints.
		serviceEndpointHashBucket *cache.ServiceEndpointHashBucket

		// natsClient is the NATS client for the application.
		// Provides messaging capabilities using the NATS protocol.
		natsClient *nats.Conn

		// natsJetStream is the NATS JetStream for the application.
		// Manages JetStream streams and consumers.
		natsJetStream jetstream.JetStream

		// natsStream is the NATS stream for the application.
		// Represents a JetStream stream for message storage and retrieval.
		natsStream jetstream.Stream
	}
)

// NewApp creates a new application with the given logger.
//
// Parameters:
//   - l: A pointer to a `slog.Logger` instance used for logging application events.
//
// Returns:
//   - A pointer to an `App` instance if successful.
//   - An error if the logger is nil or if the application configuration cannot be parsed.
func NewApp(l *slog.Logger) (*App, error) {
	if l == nil {
		return nil, ErrNilLogger
	}

	// Get the base context and its cancel function.
	baseCtx, baseCtxCancel := CoreContext()

	// Parse the application configuration from environment variables.
	baseCfg := new(AppConfig)
	if err := env.Parse(baseCfg); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	// Return a new App instance with the initialized fields.
	return &App{
		l:              l,
		baseCfg:        baseCfg,
		baseCtx:        baseCtx,
		baseCtxCancel:  baseCtxCancel,
		isStartedChan:  make(chan struct{}), // Channel to signal when the application has started.
		metricsEnabled: true,                // Enable metrics collection by default.
		shutdownWg:     new(sync.WaitGroup), // Wait group for managing shutdown tasks.
	}, nil
}

// Start starts the application and applies the given options.
//
// This function initializes and starts the application by applying the provided options,
// setting up metrics, running leader election (if configured), starting HTTP servers,
// and launching indefinite asynchronous tasks.
//
// Note: This function is thread-safe. If the function is called from multiple threads,
// it will only execute once. However, it will block all calling threads until the startup is complete.
//
// Parameters:
//   - opts: A variadic list of StartOption functions to configure the application.
//
// Returns:
//   - An error if any part of the startup process fails. If an error occurs, the application will be shut down.
func (a *App) Start(opts ...StartOption) error {
	var startErr error
	a.startOnce.Do(func() {
		defer close(a.isStartedChan)

		// Log application startup details.
		a.l.Info("starting application",
			slog.String(logging.KeyGitCommit, version.GitCommit()),
			slog.String(logging.KeyRuntime, fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
			slog.String(logging.KeyCommitTimestamp, version.CommitTimestamp().String()),
		)

		// Apply each provided StartOption to configure the application.
		for _, opt := range opts {
			if err := opt(a); err != nil { // nolint:revive // Traditional error handling
				startErr = fmt.Errorf("failed to apply option: %w", err)
				return
			}
		}

		// Set up metrics server if metrics are enabled.
		if a.metricsEnabled {
			metricsRouter := mux.NewRouter()
			metricsRouter.Handle("/metrics", promhttp.Handler())
			a.servers.Store("metrics", &http.Server{
				Addr:              fmt.Sprintf(":%d", MetricsPort),
				Handler:           metricsRouter,
				ReadHeaderTimeout: httpReadHeaderTimeout,
			})
		}

		// Start leader election if configured.
		if a.leaderElection != nil {
			go a.leaderElection.Run(a.baseCtx)
		}

		// Start all registered HTTP servers.
		var serverErr error
		a.servers.Range(func(name, srv any) bool {
			serverName, ok := name.(string)
			if !ok {
				serverErr = errors.New("failed to cast server name to string")
				return false
			}

			server, ok := srv.(*http.Server)
			if !ok {
				serverErr = fmt.Errorf("failed to cast server %s to http.Server", serverName)
				return false
			}

			a.startServer(serverName, server)
			return true
		})
		if serverErr != nil {
			startErr = fmt.Errorf("server initialization error: %w", serverErr)
			return
		}

		// Start all indefinite asynchronous tasks.
		var taskErr error
		a.indefiniteAsyncTasks.Range(func(name, fn any) bool {
			taskName, ok := name.(string)
			if !ok {
				taskErr = errors.New("failed to cast task name to string")
				return false
			}

			taskFn, ok := fn.(AsyncTaskFunc)
			if !ok {
				taskErr = fmt.Errorf("failed to cast task function %s to AsyncTaskFunc", taskName)
				return false
			}

			a.startAsyncTask(taskName, true, taskFn)
			return true
		})
		if taskErr != nil { // nolint:revive // Traditional error handling
			startErr = fmt.Errorf("async task initialization error: %w", taskErr)
			return
		}
	})

	// Wait for the application to complete its startup sequence.
	a.waitUntilStarted()

	// Handle any startup errors by logging and initiating shutdown.
	if startErr != nil {
		a.l.Error("error detected in application startup", slog.String(logging.KeyError, startErr.Error()))
		go a.Shutdown()
	}

	return startErr
}

// startServer starts the given HTTP server and adds it to the application's shutdown wait group.
//
// Parameters:
//   - name: The name of the server, used for logging purposes.
//   - srv: A pointer to the `http.Server` instance to be started.
//
// This function logs the server's status, starts it in a separate goroutine, and ensures
// graceful shutdown by decrementing the shutdown wait group when the server stops.
func (a *App) startServer(name string, srv *http.Server) {
	l := a.l.With(slog.String(logging.KeyServer, name))

	a.shutdownWg.Add(1)
	go func() {
		defer a.shutdownWg.Done()

		l.Info("server listening")
		if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			l.Info("server shut down gracefully", slog.Any(logging.KeyError, err))
		} else {
			l.Error("server closed", slog.Any(logging.KeyError, err))
		}
	}()
}

// ChildContext creates and returns a new child context derived from the application's base context.
//
// This function is useful for creating a cancellable context that is tied to the lifecycle
// of the application's base context.
//
// Returns:
//   - context.Context: The newly created child context.
//   - context.CancelFunc: A function to cancel the child context.
func (a *App) ChildContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(a.baseCtx)
}

// TimeoutContext returns a child context of the application's base context with a specified timeout.
//
// This function is useful for creating a context that automatically cancels after the given duration,
// ensuring that operations do not exceed the specified time limit.
//
// Parameters:
//   - timeout: The duration after which the context will be canceled.
//
// Returns:
//   - context.Context: The newly created child context with the specified timeout.
//   - context.CancelFunc: A function to cancel the child context manually before the timeout.
func (a *App) TimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(a.baseCtx, timeout)
}

// WaitForEnd waits for the application to complete, either normally or via an interrupt signal.
//
// This function blocks until the application's base context is canceled or done.
// Once the application ends, it executes any provided callback functions in the `onEnd` parameter.
//
// Parameters:
//   - onEnd: A variadic list of functions to be executed after the application ends.
func (a *App) WaitForEnd(onEnd ...func()) {
	<-a.baseCtx.Done()

	for _, fn := range onEnd {
		fn()
	}
}

// Shutdown gracefully stops the application by performing the following steps:
// 1. Cancels the base context to signal shutdown to all components.
// 2. Shuts down all registered HTTP servers.
// 3. Closes the database connection, Redis pool, and other resources.
// 4. Waits for all shutdown tasks to complete.
//
// Note: This function is thread-safe and ensures that the shutdown process is executed only once,
// even if called from multiple threads. It blocks all calling threads until the shutdown is complete.
func (a *App) Shutdown() {
	a.shutdownOnce.Do(func() {
		// Cancel the base context to signal shutdown.
		if a.baseCtxCancel != nil {
			a.baseCtxCancel()
		}

		// Create a context with a timeout for graceful shutdown.
		ctx, cancel := context.WithTimeout(a.baseCtx, shutdownTimeout)
		defer cancel()

		// Shut down all registered HTTP servers.
		a.servers.Range(func(name, srv any) bool {
			server, ok := srv.(*http.Server)
			if !ok {
				a.l.Error("failed to cast server to http.Server")
				return false
			}

			nameStr, ok := name.(string)
			if !ok {
				a.l.Warn("failed to cast server name to string")
				nameStr = "unknown"
			}

			// Attempt to gracefully shut down the server.
			if err := server.Shutdown(ctx); err != nil {
				a.l.Error("failed to shutdown server",
					slog.String(logging.KeyServer, nameStr),
					slog.Any(logging.KeyError, err),
				)
			}

			return true
		})

		// Close the database connection if it exists.
		if a.db != nil {
			if err := a.db.Close(); err != nil {
				a.l.Error("failed to close database", slog.Any(logging.KeyError, err))
			}
		}

		// Close the Redis pool if it exists.
		if a.redisPool != nil {
			if err := a.redisPool.Conn().Close(); err != nil {
				a.l.Error("failed to close redis pool", slog.Any(logging.KeyError, err))
			}
		}

		// Shut down the service endpoint hash bucket if it exists.
		if a.serviceEndpointHashBucket != nil {
			a.serviceEndpointHashBucket.Shutdown()
		}

		// Stop the worker pool if it exists.
		if a.workerPool != nil {
			a.workerPool.Stop()
		}

		// Close the NATS client if it exists.
		if a.natsClient != nil {
			a.natsClient.Close()
		}
	})

	// Wait for all shutdown tasks to complete.
	a.shutdownWg.Wait()
}

// Logger returns the logger for the application.
//
// This function ensures that the application's logger is properly initialized
// before returning it. If the logger is not registered, it will panic.
//
// Returns:
//   - *slog.Logger: The logger instance for the application.
//
// Panics:
//   - If the logger has not been registered.
func (a *App) Logger() *slog.Logger {
	if a.l == nil {
		panic("logger has not been registered")
	}
	return a.l
}

// VaultClient returns the vault client for the application.
//
// This function ensures that the application's vault client is properly initialized
// before returning it. If the vault client is not registered, it logs an error and panics.
//
// Returns:
//   - vaulty.Client: The vault client instance for the application.
//
// Panics:
//   - If the vault client has not been registered.
func (a *App) VaultClient() vaulty.Client {
	if a.vaultClient == nil {
		a.l.Error("vault client has not been registered")
		panic("vault client has not been registered")
	}
	return a.vaultClient
}

// Viper returns the viper instance for the application.
//
// This function ensures that the application's viper instance is properly initialized
// before returning it. If the viper instance is not registered, it logs an error and panics.
//
// Returns:
//   - *viper.Viper: The viper instance for the application.
//
// Panics:
//   - If the viper instance has not been registered.
func (a *App) Viper() *viper.Viper {
	if a.vip == nil {
		a.l.Error("viper instance has not been registered")
		panic("viper instance has not been registered")
	}
	return a.vip
}

// DBConn returns the database connection for the application.
//
// This function ensures that the application's database connection is properly initialized
// before returning it. If the database connection is not registered, it logs an error and panics.
//
// Returns:
//   - *vsql.Database: The database connection for the application.
//
// Panics:
//   - If the database connection has not been registered.
func (a *App) DBConn() *vsql.Database {
	if a.db == nil {
		a.l.Error("database connection has not been registered")
		panic("database connection has not been registered")
	}
	return a.db
}

// KubeClient returns the Kubernetes client for the application.
//
// This function ensures that the application's Kubernetes client is properly initialized
// before returning it. If the Kubernetes client is not registered, it logs an error and panics.
//
// Returns:
//   - kubernetes.Interface: The Kubernetes client for the application.
//
// Panics:
//   - If the Kubernetes client has not been registered.
func (a *App) KubeClient() kubernetes.Interface {
	if a.kubeClient == nil {
		a.l.Error("kubernetes client has not been registered")
		panic("kubernetes client has not been registered")
	}
	return a.kubeClient
}

// Done returns a channel that is closed when the application's base context is done.
//
// This function provides a way to listen for the completion or cancellation
// of the application's base context, which signals the end of the application's lifecycle.
//
// Returns:
//   - <-chan struct{}: A read-only channel that is closed when the application's base context is done.
func (a *App) Done() <-chan struct{} {
	return a.baseCtx.Done()
}

// IsLeader checks if the application is the leader in a distributed system.
//
// This function determines the leadership status of the application. If leader election
// is not configured, it assumes the application is the leader and logs a debug message.
//
// Returns:
//   - bool: True if the application is the leader, false otherwise.
func (a *App) IsLeader() bool {
	if a.leaderElection == nil {
		a.l.Debug("leader election not set, assuming leader")
		return true
	}
	return a.leaderElection.IsLeader()
}

// LeaderChange returns a channel that is notified when the leader changes.
//
// This function provides a mechanism to listen for changes in leadership status
// in a distributed system. The returned channel is closed or notified whenever
// the application's leadership state changes.
//
// Returns:
//   - <-chan struct{}: A read-only channel that signals leadership changes.
func (a *App) LeaderChange() <-chan struct{} {
	return a.leaderChange
}

// StartServer starts a new server with the given name and http.Server.
//
// This function registers and starts an HTTP server for the application.
// It ensures that the server name is unique and applies default handlers
// for "not found" and "method not allowed" responses if they are not already set.
//
// Parameters:
//   - name: A string representing the name of the server.
//   - srv: A pointer to an `http.Server` instance to be started.
//
// Returns:
//   - error: An error if the server with the given name already exists.
//
// Behavior:
//   - If a server with the same name is already registered, it returns an error.
//   - If the server handler is a Gorilla Mux router, it ensures that default
//     "not found" and "method not allowed" handlers are applied if not set.
//   - The server is stored in the application's server map and started.
func (a *App) StartServer(name string, srv *http.Server) error {
	if _, found := a.servers.Load(name); found {
		return fmt.Errorf("server %s already exists", name)
	}

	// If the server handler is gorilla mux, check the not found handler and method not allowed handler
	if muxRouter, ok := srv.Handler.(*mux.Router); ok {
		if muxRouter.NotFoundHandler == nil {
			a.l.Info("not found handler not set for server, applying default handler")
			muxRouter.NotFoundHandler = uhttp.NotFoundHandler()
		}
		if muxRouter.MethodNotAllowedHandler == nil {
			a.l.Info("method not allowed handler not set for server, applying default handler")
			muxRouter.MethodNotAllowedHandler = uhttp.MethodNotAllowedHandler()
		}
	}

	a.servers.Store(name, srv)
	a.startServer(name, srv)
	return nil
}

// startAsyncTask starts an asynchronous task with the given name and function.
//
// This function launches a goroutine to execute the provided task function. If the task
// is marked as indefinite, it is expected to run until the application shuts down. If
// an indefinite task ends unexpectedly, the application shutdown is triggered.
//
// Parameters:
//   - name: A string representing the name of the task, used for logging purposes.
//   - indefinite: A boolean indicating whether the task is expected to run indefinitely.
//   - fn: An AsyncTaskFunc representing the task function to be executed.
//
// Behavior:
//   - Logs the start of the task.
//   - Adds the task to the application's shutdown wait group.
//   - Executes the task function in a separate goroutine.
//   - If the task is indefinite and ends unexpectedly (without context cancellation),
//     it logs an error and triggers application shutdown.
func (a *App) startAsyncTask(name string, indefinite bool, fn AsyncTaskFunc) {
	a.l.Info("starting async task", slog.String(logging.KeyName, name))
	a.shutdownWg.Add(1)
	go func() {
		defer a.shutdownWg.Done()
		fn(a.baseCtx)

		// If task is configured as indefinite and the task stops before we stop running the entire app, close the app down
		// with an error.
		if indefinite && !errors.Is(a.baseCtx.Err(), context.Canceled) { // nolint:revive // Traditional error handling
			a.l.Error("indefinite async task closed before app shutdown",
				slog.String(logging.KeyName, name),
			)
			a.baseCtxCancel()
		}
	}()
}

// RedisPool returns the Redis connection pool for the application.
//
// This function ensures that the Redis pool is properly initialized before returning it.
// If the Redis pool is not registered, it logs an error and panics.
//
// Returns:
//   - goredis.Pool: The Redis connection pool for the application.
//
// Panics:
//   - If the Redis pool has not been registered.
func (a *App) RedisPool() goredis.Pool {
	if a.redisPool == nil {
		a.l.Error("redis pool has not been registered")
		panic("redis pool has not been registered")
	}
	return a.redisPool
}

// WorkerPool returns the worker pool for the application.
//
// This function ensures that the worker pool is properly initialized before returning it.
// If the worker pool is not registered, it logs an error and panics.
//
// Returns:
//   - workerpool.Pool: The worker pool for the application.
//
// Panics:
//   - If the worker pool has not been registered.
func (a *App) WorkerPool() workerpool.Pool {
	if a.workerPool == nil {
		a.l.Error("worker pool has not been registered")
		panic("worker pool has not been registered")
	}
	return a.workerPool
}

// NatsClient returns the NATS client for the application.
//
// This function ensures that the NATS client is properly initialized before returning it.
// If the NATS client is not registered, it logs an error and panics.
//
// Returns:
//   - *nats.Conn: The NATS client for the application.
//
// Panics:
//   - If the NATS client has not been registered.
func (a *App) NatsClient() *nats.Conn {
	if a.natsClient == nil {
		a.l.Error("nats client has not been registered")
		panic("nats client has not been registered")
	}
	return a.natsClient
}

// NatsJetStream returns the JetStream instance for the application.
//
// This function ensures that the JetStream instance is properly initialized before returning it.
// If the JetStream instance is not registered, it logs an error and panics.
//
// Returns:
//   - jetstream.JetStream: The JetStream instance for the application.
//
// Panics:
//   - If the JetStream instance has not been registered.
func (a *App) NatsJetStream() jetstream.JetStream {
	if a.natsJetStream == nil {
		a.l.Error("nats jetstream has not been registered")
		panic("nats jetstream has not been registered")
	}
	return a.natsJetStream
}

// NatsStream returns the NATS stream for the application.
//
// This function ensures that the NATS stream is properly initialized before returning it.
// If the NATS stream is not registered, it logs an error and panics.
//
// Returns:
//   - jetstream.Stream: The NATS stream for the application.
//
// Panics:
//   - If the NATS stream has not been registered.
func (a *App) NatsStream() jetstream.Stream {
	if a.natsStream == nil {
		a.l.Error("nats stream has not been registered")
		panic("nats stream has not been registered")
	}
	return a.natsStream
}

// CreateNatsJetStreamConsumer creates a new JetStream consumer with the specified name and subject filter.
//
// This function ensures that the NATS stream is properly initialized before attempting to create or update
// a JetStream consumer. If the NATS stream is not registered, it returns an error.
//
// Parameters:
//   - consumerName: A string representing the durable name of the consumer. This ensures state retention.
//   - subjectFilter: A string representing the subject filter for the consumer. Only messages matching this
//     subject will be delivered to the consumer.
//
// Returns:
//   - jetstream.Consumer: The created or updated JetStream consumer instance.
//   - error: An error if the NATS stream is not registered or if the consumer creation fails.
//
// Behavior:
//   - If the NATS stream is nil, it returns an error indicating that the stream is not registered.
//   - Creates or updates a JetStream consumer with the provided configuration, including durable name,
//     explicit acknowledgment policy, subject filter, and delivery policy.
//
// Example:
//
//	cons, err := app.CreateNatsJetStreamConsumer("my-consumer", "my.subject")
//	if err != nil {
//	    log.Fatalf("Failed to create consumer: %v", err)
//	}
func (a *App) CreateNatsJetStreamConsumer(consumerName, subjectFilter string) (jetstream.Consumer, error) {
	if a.natsStream == nil {
		return nil, errors.New("nats stream is nil, ensure WithNatsJetStream is called")
	}

	cons, err := a.natsStream.CreateOrUpdateConsumer(a.baseCtx, jetstream.ConsumerConfig{
		Durable:       consumerName,                // Durable name ensures state retention
		AckPolicy:     jetstream.AckExplicitPolicy, // Acknowledge messages explicitly
		FilterSubject: subjectFilter,               // Only accept messages from this subject
		DeliverPolicy: jetstream.DeliverAllPolicy,  // Deliver all messages
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get consumer: %w", err)
	}

	return cons, nil
}

// FixedHashBucket returns the fixed hash bucket for the application.
//
// This function ensures that the fixed hash bucket is properly initialized before returning it.
// If the fixed hash bucket is not registered, it logs an error and panics.
//
// Returns:
//   - *cache.FixedHashBucket: The fixed hash bucket instance for the application.
//
// Panics:
//   - If the fixed hash bucket has not been registered.
func (a *App) FixedHashBucket() *cache.FixedHashBucket {
	if a.fixedHashBucket == nil {
		a.l.Error("fixed hash bucket has not been registered")
		panic("fixed hash bucket has not been registered")
	}
	return a.fixedHashBucket
}

// ServiceEndpointHashBucket returns the service endpoint hash bucket for the application.
//
// This function ensures that the service endpoint hash bucket is properly initialized before returning it.
// If the service endpoint hash bucket is not registered, it logs an error and panics.
//
// Returns:
//   - *cache.ServiceEndpointHashBucket: The service endpoint hash bucket instance for the application.
//
// Panics:
//   - If the service endpoint hash bucket has not been registered.
func (a *App) ServiceEndpointHashBucket() *cache.ServiceEndpointHashBucket {
	if a.serviceEndpointHashBucket == nil {
		a.l.Error("service endpoint hash bucket has not been registered")
		panic("service endpoint hash bucket has not been registered")
	}
	return a.serviceEndpointHashBucket
}

// PodLister returns the pod lister for the application.
//
// This function ensures that the pod lister is properly initialized before returning it.
// If the pod lister is not registered, it logs an error and panics.
//
// Returns:
//   - listersv1.PodLister: The pod lister instance for the application.
//
// Panics:
//   - If the pod lister has not been registered.
func (a *App) PodLister() listersv1.PodLister {
	if a.podLister == nil {
		a.l.Error("pod lister has not been registered")
		panic("pod lister has not been registered")
	}
	return a.podLister
}

// PodInformer returns the pod informer for the application.
//
// This function ensures that the pod informer is properly initialized before returning it.
// If the pod informer is not registered, it logs an error and panics.
//
// Returns:
//   - kubeCache.SharedIndexInformer: The pod informer instance for the application.
//
// Panics:
//   - If the pod informer has not been registered.
func (a *App) PodInformer() kubeCache.SharedIndexInformer {
	if a.podInformer == nil {
		a.l.Error("pod informer has not been registered")
		panic("pod informer has not been registered")
	}
	return a.podInformer
}

// KubernetesInformerFactory returns the Kubernetes informer factory for the application.
//
// This function ensures that the Kubernetes informer factory is properly initialized before returning it.
// If the Kubernetes informer factory is not registered, it logs an error and panics.
//
// Returns:
//   - informers.SharedInformerFactory: The Kubernetes informer factory instance for the application.
//
// Panics:
//   - If the Kubernetes informer factory has not been registered.
func (a *App) KubernetesInformerFactory() informers.SharedInformerFactory {
	if a.kubernetesInformerFactory == nil {
		a.l.Error("kubernetes informer factory has not been registered")
		panic("kubernetes informer factory has not been registered")
	}
	return a.kubernetesInformerFactory
}

// SecretLister returns the secret lister for the application.
//
// This function ensures that the secret lister is properly initialized before returning it.
// If the secret lister is not registered, it logs an error and panics.
//
// Returns:
//   - listersv1.SecretLister: The secret lister instance for the application.
//
// Panics:
//   - If the secret lister has not been registered.
func (a *App) SecretLister() listersv1.SecretLister {
	if a.secretLister == nil {
		a.l.Error("secret lister has not been registered")
		panic("secret lister has not been registered")
	}
	return a.secretLister
}

// SecretInformer returns the secret informer for the application.
//
// This function ensures that the secret informer is properly initialized before returning it.
// If the secret informer is not registered, it logs an error and panics.
//
// Returns:
//   - kubeCache.SharedIndexInformer: The secret informer instance for the application.
//
// Panics:
//   - If the secret informer has not been registered.
func (a *App) SecretInformer() kubeCache.SharedIndexInformer {
	if a.secretInformer == nil {
		a.l.Error("secret informer has not been registered")
		panic("secret informer has not been registered")
	}
	return a.secretInformer
}

// waitUntilStarted blocks until the application has completed its startup sequence.
//
// This function ensures that the application has fully started before proceeding.
// It waits for the `isStartedChan` channel to be closed, which signals the completion
// of the startup process. If the `isStartedChan` channel is not initialized, it logs
// an error and panics.
//
// Panics:
//   - If `isStartedChan` is nil, indicating that the channel has not been properly initialized.
func (a *App) waitUntilStarted() {
	if a.isStartedChan == nil {
		a.l.Error("isStartedChan has not been registered")
		panic("isStartedChan has not been registered")
	}
	<-a.isStartedChan
}
