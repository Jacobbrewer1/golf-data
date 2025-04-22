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
	"github.com/jacobbrewer1/goredis"
	"github.com/jacobbrewer1/uhttp"
	"github.com/jacobbrewer1/vaulty"
	"github.com/jacobbrewer1/vaulty/vsql"
	"github.com/jacobbrewer1/web/cache"
	"github.com/jacobbrewer1/web/logging"
	"github.com/jacobbrewer1/web/version"
	"github.com/jacobbrewer1/workerpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listersv1 "k8s.io/client-go/listers/core/v1"
	kubeCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/leaderelection"
)

const (
	MetricsPort = 9090
	HealthPort  = 9091

	// httpReadHeaderTimeout is the amount of time allowed to read request headers.
	httpReadHeaderTimeout = 10 * time.Second
	shutdownTimeout       = 15 * time.Second
)

var (
	// ErrNilLogger is returned when a nil logger is passed to NewApp.
	ErrNilLogger = errors.New("logger is nil")
)

type (
	// AppConfig is the configuration for the application.
	AppConfig struct {
		ConfigLocation string `env:"CONFIG_LOCATION" envDefault:"config.json"`
	}

	// App is the application struct.
	App struct {
		// l is the logger for the application.
		l *slog.Logger

		// baseCtx is the base context for the application.
		baseCtx context.Context

		// baseCtxCancel is the base context cancel function.
		baseCtxCancel context.CancelFunc

		// baseCfg is the base configuration for the application.
		baseCfg *AppConfig

		// isStarted a channel that is closed when the application is started.
		isStartedChan chan struct{}

		// startOnce is used to ensure that the start function is only called once.
		startOnce sync.Once

		// vip is the viper instance for the application.
		vip *viper.Viper

		// vaultClient is the vault client for the application.
		vaultClient vaulty.Client

		// metricsEnabled is a flag to enable metrics for the application.
		metricsEnabled bool

		// servers is the list of servers for the application.
		servers sync.Map

		// shutdownOnce is used to ensure that the shutdown function is only called once.
		shutdownOnce sync.Once

		// shutdownWg is used to wait for all shutdown tasks to complete.
		shutdownWg *sync.WaitGroup

		// db is the database for the application.
		db *vsql.Database

		// kubeClient interacts with the Kubernetes API server.
		kubeClient kubernetes.Interface

		// kubernetesInformerFactory is a factory used for initialising a pod informer.
		kubernetesInformerFactory informers.SharedInformerFactory

		// podInformer is an informer for Kubernetes Pod objects.
		podInformer kubeCache.SharedIndexInformer

		// podLister is a lister for Kubernetes Pod objects.
		podLister listersv1.PodLister

		// secretInformer is an informer for Kubernetes Secret objects.
		secretInformer kubeCache.SharedIndexInformer

		// secretLister is a lister for Kubernetes Secret objects.
		secretLister listersv1.SecretLister

		// leaderElection is the leader election for the application.
		leaderElection *leaderelection.LeaderElector

		// leaderChange is the channel that is notified when the leader changes.
		leaderChange chan struct{}

		// redisPool is the redis pool for the application.
		redisPool goredis.Pool

		// workerPool is the worker pool that can execute tasks concurrently.
		workerPool workerpool.Pool

		// indefiniteAsyncTasks is the list of indefinite async tasks for the application.
		indefiniteAsyncTasks sync.Map

		// fixedHashBucket is the fixed hash bucket for the application.
		fixedHashBucket *cache.FixedHashBucket

		// serviceEndpointHashBucket is the service endpoint hash bucket for the application.
		serviceEndpointHashBucket *cache.ServiceEndpointHashBucket

		// natsClient is the nats client for the application.
		natsClient *nats.Conn

		// natsJetStream is the nats jetstream for the application.
		natsJetStream jetstream.JetStream

		// natStream is the nats stream for the application.
		natsStream jetstream.Stream
	}
)

// NewApp creates a new application with the given logger.
func NewApp(l *slog.Logger) (*App, error) {
	if l == nil {
		return nil, ErrNilLogger
	}

	baseCtx, baseCtxCancel := CoreContext()

	baseCfg := new(AppConfig)
	if err := env.Parse(baseCfg); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	return &App{
		l:              l,
		baseCfg:        baseCfg,
		baseCtx:        baseCtx,
		baseCtxCancel:  baseCtxCancel,
		isStartedChan:  make(chan struct{}),
		metricsEnabled: true,
		shutdownWg:     new(sync.WaitGroup),
	}, nil
}

// Start starts the application and applies the given options.
//
// Note: This function is thread-safe. If the function is called from multiple threads, the function will only be
// executed once; however the function will block all calling threads until the startup is complete.
//
// If the function returns an error, the application will be shut down.
func (a *App) Start(opts ...StartOption) error {
	var startErr error
	a.startOnce.Do(func() {
		defer close(a.isStartedChan)

		a.l.Info("starting application",
			slog.String(logging.KeyGitCommit, version.GitCommit()),
			slog.String(logging.KeyRuntime, fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
			slog.String(logging.KeyBuildDate, version.CommitTimestamp().String()),
		)

		for _, opt := range opts {
			if err := opt(a); err != nil { // nolint:revive // Traditional error handling
				startErr = fmt.Errorf("failed to apply option: %w", err)
				return
			}
		}

		if a.metricsEnabled {
			metricsRouter := mux.NewRouter()
			metricsRouter.Handle("/metrics", promhttp.Handler())
			a.servers.Store("metrics", &http.Server{
				Addr:              fmt.Sprintf(":%d", MetricsPort),
				Handler:           metricsRouter,
				ReadHeaderTimeout: httpReadHeaderTimeout,
			})
		}

		if a.leaderElection != nil {
			go a.leaderElection.Run(a.baseCtx)
		}

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

	a.waitUntilStarted()

	if startErr != nil {
		a.l.Error("error detected in application startup", slog.String(logging.KeyError, startErr.Error()))
		go a.Shutdown()
	}

	return startErr
}

// startServer starts the given server and adds it to the shutdown wait group.
// It logs the server status and handles graceful shutdown.
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

// ChildContext returns a child context of the base context.
func (a *App) ChildContext() (context.Context, context.CancelFunc) {
	return context.WithCancel(a.baseCtx)
}

// TimeoutContext returns a child context of the base context with a timeout.
func (a *App) TimeoutContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(a.baseCtx, timeout)
}

// WaitForEnd waits for the application to complete, either normally or via an interrupt signal.
func (a *App) WaitForEnd(onEnd ...func()) {
	<-a.baseCtx.Done()

	for _, fn := range onEnd {
		fn()
	}
}

// Shutdown stops the application.
//
// Note: This function is not thread-safe. If the function is called from multiple threads, the function is only
// executed once; however the function will block all calling threads until the shutdown is complete.
func (a *App) Shutdown() {
	a.shutdownOnce.Do(func() {
		if a.baseCtxCancel != nil {
			a.baseCtxCancel()
		}

		ctx, cancel := context.WithTimeout(a.baseCtx, shutdownTimeout)
		defer cancel()

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

			if err := server.Shutdown(ctx); err != nil {
				a.l.Error("failed to shutdown server",
					slog.String(logging.KeyServer, nameStr),
					slog.Any(logging.KeyError, err),
				)
			}

			return true
		})

		if a.db != nil {
			if err := a.db.Close(); err != nil {
				a.l.Error("failed to close database", slog.Any(logging.KeyError, err))
			}
		}

		if a.redisPool != nil {
			if err := a.redisPool.Conn().Close(); err != nil {
				a.l.Error("failed to close redis pool", slog.Any(logging.KeyError, err))
			}
		}

		if a.serviceEndpointHashBucket != nil {
			a.serviceEndpointHashBucket.Shutdown()
		}

		if a.workerPool != nil {
			a.workerPool.Stop()
		}

		if a.natsClient != nil {
			a.natsClient.Close()
		}
	})

	a.shutdownWg.Wait()
}

// Logger returns the logger for the application.
func (a *App) Logger() *slog.Logger {
	if a.l == nil {
		panic("logger has not been registered")
	}
	return a.l
}

// VaultClient returns the vault client for the application.
func (a *App) VaultClient() vaulty.Client {
	if a.vaultClient == nil {
		a.l.Error("vault client has not been registered")
		panic("vault client has not been registered")
	}
	return a.vaultClient
}

// Viper returns the viper instance for the application.
func (a *App) Viper() *viper.Viper {
	if a.vip == nil {
		a.l.Error("viper instance has not been registered")
		panic("viper instance has not been registered")
	}
	return a.vip
}

// DBConn returns the database connection for the application.
func (a *App) DBConn() *vsql.Database {
	if a.db == nil {
		a.l.Error("database connection has not been registered")
		panic("database connection has not been registered")
	}
	return a.db
}

// KubeClient returns the Kubernetes client for the application.
func (a *App) KubeClient() kubernetes.Interface {
	if a.kubeClient == nil {
		a.l.Error("kubernetes client has not been registered")
		panic("kubernetes client has not been registered")
	}
	return a.kubeClient
}

// Done returns a channel that is closed when the application is done.
func (a *App) Done() <-chan struct{} {
	return a.baseCtx.Done()
}

// IsLeader returns true if the application is the leader.
func (a *App) IsLeader() bool {
	if a.leaderElection == nil {
		a.l.Debug("leader election not set, assuming leader")
		return true
	}
	return a.leaderElection.IsLeader()
}

// LeaderChange returns a channel that is notified when the leader changes.
func (a *App) LeaderChange() <-chan struct{} {
	return a.leaderChange
}

// StartServer starts a new server with the given name and http.Server.
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

// startAsyncTask starts an async task with the given name and function.
// If indefinite is true, the task is expected to run until the application shuts down.
// The function will trigger application shutdown if an indefinite task ends unexpectedly.
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

// RedisPool returns the redis pool for the application.
func (a *App) RedisPool() goredis.Pool {
	if a.redisPool == nil {
		a.l.Error("redis pool has not been registered")
		panic("redis pool has not been registered")
	}
	return a.redisPool
}

// WorkerPool returns the worker pool for the application.
func (a *App) WorkerPool() workerpool.Pool {
	if a.workerPool == nil {
		a.l.Error("worker pool has not been registered")
		panic("worker pool has not been registered")
	}
	return a.workerPool
}

// NatsClient returns the NATS client for the application.
func (a *App) NatsClient() *nats.Conn {
	if a.natsClient == nil {
		a.l.Error("nats client has not been registered")
		panic("nats client has not been registered")
	}
	return a.natsClient
}

// NatsJetStream returns the JetStream stream for the application.
func (a *App) NatsJetStream() jetstream.JetStream {
	if a.natsJetStream == nil {
		a.l.Error("nats jetstream has not been registered")
		panic("nats jetstream has not been registered")
	}
	return a.natsJetStream
}

// NatsStream returns the NATS stream for the application.
func (a *App) NatsStream() jetstream.Stream {
	if a.natsStream == nil {
		a.l.Error("nats stream has not been registered")
		panic("nats stream has not been registered")
	}
	return a.natsStream
}

// CreateNatsJetStreamConsumer creates a new JetStream consumer with the given name and subject filter.
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
func (a *App) FixedHashBucket() *cache.FixedHashBucket {
	if a.fixedHashBucket == nil {
		a.l.Error("fixed hash bucket has not been registered")
		panic("fixed hash bucket has not been registered")
	}
	return a.fixedHashBucket
}

// ServiceEndpointHashBucket returns the service endpoint hash bucket for the application.
func (a *App) ServiceEndpointHashBucket() *cache.ServiceEndpointHashBucket {
	if a.serviceEndpointHashBucket == nil {
		a.l.Error("service endpoint hash bucket has not been registered")
		panic("service endpoint hash bucket has not been registered")
	}
	return a.serviceEndpointHashBucket
}

// PodLister returns the pod lister for the application.
func (a *App) PodLister() listersv1.PodLister {
	if a.podLister == nil {
		a.l.Error("pod lister has not been registered")
		panic("pod lister has not been registered")
	}
	return a.podLister
}

// PodInformer returns the pod informer for the application.
func (a *App) PodInformer() kubeCache.SharedIndexInformer {
	if a.podInformer == nil {
		a.l.Error("pod informer has not been registered")
		panic("pod informer has not been registered")
	}
	return a.podInformer
}

// KubernetesInformerFactory returns the Kubernetes informer factory for the application.
func (a *App) KubernetesInformerFactory() informers.SharedInformerFactory {
	if a.kubernetesInformerFactory == nil {
		a.l.Error("kubernetes informer factory has not been registered")
		panic("kubernetes informer factory has not been registered")
	}
	return a.kubernetesInformerFactory
}

// SecretLister returns the secret lister for the application.
func (a *App) SecretLister() listersv1.SecretLister {
	if a.secretLister == nil {
		a.l.Error("secret lister has not been registered")
		panic("secret lister has not been registered")
	}
	return a.secretLister
}

// SecretInformer returns the secret informer for the application.
func (a *App) SecretInformer() kubeCache.SharedIndexInformer {
	if a.secretInformer == nil {
		a.l.Error("secret informer has not been registered")
		panic("secret informer has not been registered")
	}
	return a.secretInformer
}

// waitUntilStarted blocks until the application has completed its startup sequence.
// It panics if the isStartedChan is not initialized.
func (a *App) waitUntilStarted() {
	if a.isStartedChan == nil {
		a.l.Error("isStartedChan has not been registered")
		panic("isStartedChan has not been registered")
	}
	<-a.isStartedChan
}
