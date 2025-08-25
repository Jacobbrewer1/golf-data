package sync

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type AsyncTaskFunc = func(ctx context.Context)

const (
	// workerPoolTimeout is the maximum time a worker will wait for a task before exiting.
	workerPoolTimeout = 10 * time.Second
)

var (
	workerPoolLabels = []string{
		"name",
	}

	// metricWorkerPoolBacklogSize represents the maximum number of tasks that can be in the backlog before blocking.
	metricWorkerPoolBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_web_worker_pool_backlog_size",
		Help: "The maximum number of tasks that can be in the backlog.",
	}, workerPoolLabels)

	// metricWorkerPoolActiveWorkers represents the number of workers currently processing a task.
	metricWorkerPoolActiveWorkers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_web_worker_pool_active_workers",
		Help: "The number of workers processing a task.",
	}, workerPoolLabels)

	// metricWorkerPoolActiveBacklogSize represents the number of tasks currently in the backlog.
	metricWorkerPoolActiveBacklogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_web_worker_pool_active_backlog_size",
		Help: "The number of tasks in the backlog yet to be processed.",
	}, workerPoolLabels)

	// metricWorkerPoolMaxWorkers represents the configured number of workers which process tasks.
	metricWorkerPoolMaxWorkers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "app_web_worker_pool_max_workers",
		Help: "The number of workers configured to process tasks.",
	}, workerPoolLabels)

	// metricWorkerPoolTaskDurationNano represents a histogram of task execution durations.
	metricWorkerPoolTaskDurationNano = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "app_web_worker_pool_task_duration_nanoseconds",
		Help: "The duration of task run by this pool, in nanoseconds.",
	}, workerPoolLabels)
)

type WorkerPool interface {
	// SubmitBlocking submits a task to be processed. If no workers are available to process the work, the call blocks until a worker is available.
	SubmitBlocking(task AsyncTaskFunc)

	// Close closes the worker pool.
	Close()
}

// workerPool implements a fixed-size worker pool of n workers.
type workerPool struct {
	// ctx is the context used to cancel the worker pool.
	ctx context.Context

	// mut is a mutex to protect the workQueue and limiter.
	mut sync.RWMutex

	// size is the number of workers in the pool.
	size uint

	// limiter is a channel used to limit the number of concurrent workers.
	limiter chan struct{}

	// workQueue is a channel used to queue tasks for processing.
	workQueue chan AsyncTaskFunc

	// workerWait is a wait group used to wait for all workers to finish processing.
	workerWait sync.WaitGroup

	// labels are the Prometheus labels used to identify the worker pool.
	labels prometheus.Labels
}

// NewWorkerPool creates a new instance of workerPool of specified size and backlog. The backlog is increased by the
// specified size parameter to ensure that at least size number of tasks can be pushed in parallel.
func NewWorkerPool(ctx context.Context, name string, size, backlog uint) WorkerPool {
	labels := prometheus.Labels{
		"name": name,
	}
	metricWorkerPoolMaxWorkers.With(labels).Set(float64(size))
	metricWorkerPoolBacklogSize.With(labels).Set(float64(size + backlog))
	return &workerPool{
		size:      size,
		limiter:   make(chan struct{}, size),
		workQueue: make(chan AsyncTaskFunc, size+backlog),
		ctx:       ctx,
		labels:    labels,
	}
}

// SubmitBlocking submits a task to be processed. If no workers are available to process the work, the call blocks
// until a worker is available.
func (w *workerPool) SubmitBlocking(task AsyncTaskFunc) {
	w.mut.RLock()
	defer w.mut.RUnlock()

	// Pool is closing, drop any newly submitted messages.
	if w.workQueue == nil {
		return
	}
	if w.limiter == nil {
		return
	}

	// Push a task to the work queue. This will result in a blocking call if the backlog of the work queue
	// has been reached.
	w.workQueue <- task
	metricWorkerPoolActiveBacklogSize.With(w.labels).Inc()

	// Limit based on the configured pool size. If we have been limited, return early to prevent spawning another worker.
	select {
	case w.limiter <- struct{}{}:
	default:
		return
	}

	w.workerWait.Add(1)
	go func() {
		w.mut.RLock()
		defer w.mut.RUnlock()
		defer func() {
			w.workerWait.Done()

			// Worker has exited, so allow another in the future.
			<-w.limiter
			metricWorkerPoolActiveWorkers.With(w.labels).Dec()
		}()

		metricWorkerPoolActiveWorkers.With(w.labels).Inc()

		// Initialise a timer which will stop the worker after a period of inactivity.
		timer := time.NewTimer(workerPoolTimeout)
		defer timer.Stop()

		// Attempt to pull a task to execute. To prevent thrashing goroutines, the worker will try to execute as much
		// work as possible before exiting.
		for {
			select {
			case <-w.ctx.Done():
				// If the parent context is closed, stop the worker.
				return
			case <-timer.C:
				// Close down the worker if no work has been received recently.
				return
			case t, ok := <-w.workQueue:
				if !ok {
					return
				}

				metricWorkerPoolActiveBacklogSize.With(w.labels).Dec()

				start := time.Now()

				t(w.ctx)

				metricWorkerPoolTaskDurationNano.With(w.labels).Observe(float64(time.Since(start).Nanoseconds()))

				// Reset processing timer.
				timer.Reset(workerPoolTimeout)
			}
		}
	}()
}

func (w *workerPool) Close() {
	w.mut.Lock()
	defer w.mut.Unlock()

	if w.workQueue == nil {
		return
	}

	// Allow remaining work to complete.
	close(w.workQueue)
	close(w.limiter)

	// Wait for remaining work to complete.
	w.workerWait.Wait()

	w.workQueue = nil
	w.limiter = nil
}
