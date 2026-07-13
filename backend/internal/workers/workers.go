package workers

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/coindistro/backend/internal/queue"
)

// Job represents a unit of work to be executed by a worker.
type Job struct {
	ID         string
	Type       string
	Payload    map[string]interface{}
	Retries    int
	MaxRetries int
}

// Result represents the outcome of a job execution.
type Result struct {
	JobID   string
	Error   error
	Success bool
}

// Worker processes jobs asynchronously.
type Worker struct {
	id      int
	jobs    <-chan Job
	results chan<- Result
	logger  *zap.Logger
	quit    chan struct{}
}

// NewWorker creates a new worker.
func NewWorker(id int, jobs <-chan Job, results chan<- Result, logger *zap.Logger) *Worker {
	return &Worker{
		id:      id,
		jobs:    jobs,
		results: results,
		logger:  logger,
		quit:    make(chan struct{}),
	}
}

// Start begins processing jobs in a goroutine.
func (w *Worker) Start(ctx context.Context, handler func(ctx context.Context, job Job) error) {
	go func() {
		w.logger.Info("worker started", zap.Int("worker_id", w.id))
		for {
			select {
			case job, ok := <-w.jobs:
				if !ok {
					return
				}
				result := w.processJob(ctx, job, handler)
				w.results <- result
			case <-w.quit:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop signals the worker to stop.
func (w *Worker) Stop() {
	close(w.quit)
}

func (w *Worker) processJob(ctx context.Context, job Job, handler func(ctx context.Context, job Job) error) Result {
	w.logger.Info("processing job",
		zap.Int("worker_id", w.id),
		zap.String("job_id", job.ID),
		zap.String("job_type", job.Type),
	)

	err := handler(ctx, job)
	if err != nil {
		w.logger.Error("job failed",
			zap.Int("worker_id", w.id),
			zap.String("job_id", job.ID),
			zap.Error(err),
		)
		return Result{JobID: job.ID, Error: err, Success: false}
	}

	return Result{JobID: job.ID, Success: true}
}

// Pool manages a pool of workers.
type Pool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	workers    []*Worker
	logger     *zap.Logger
	wg         sync.WaitGroup
	quit       chan struct{}
}

// PoolConfig holds configuration for a worker pool.
type PoolConfig struct {
	NumWorkers int
	QueueSize  int
	Logger     *zap.Logger
}

// NewPool creates a new worker pool.
func NewPool(cfg PoolConfig) *Pool {
	if cfg.NumWorkers <= 0 {
		cfg.NumWorkers = 5
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = 100
	}

	pool := &Pool{
		numWorkers: cfg.NumWorkers,
		jobs:       make(chan Job, cfg.QueueSize),
		results:    make(chan Result, cfg.QueueSize),
		logger:     cfg.Logger,
		quit:       make(chan struct{}),
	}

	return pool
}

// Start initializes the worker pool and starts all workers.
func (p *Pool) Start(ctx context.Context, handler func(ctx context.Context, job Job) error) {
	for i := 0; i < p.numWorkers; i++ {
		worker := NewWorker(i+1, p.jobs, p.results, p.logger)
		p.workers = append(p.workers, worker)
		worker.Start(ctx, handler)
	}

	p.wg.Add(1)
	go p.collectResults(ctx)
}

// Submit adds a job to the pool.
func (p *Pool) Submit(job Job) {
	select {
	case p.jobs <- job:
	case <-p.quit:
	}
}

// SubmitWithRetry submits a job with retry configuration.
func (p *Pool) SubmitWithRetry(job Job, maxRetries int) {
	job.MaxRetries = maxRetries
	p.Submit(job)
}

// Results returns the results channel.
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop gracefully shuts down the worker pool.
func (p *Pool) Stop() {
	close(p.quit)
	for _, worker := range p.workers {
		worker.Stop()
	}
	close(p.jobs)
	p.wg.Wait()
	close(p.results)
}

func (p *Pool) collectResults(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case result, ok := <-p.results:
			if !ok {
				return
			}
			if !result.Success && result.Error != nil {
				p.logger.Error("job result error",
					zap.String("job_id", result.JobID),
					zap.Error(result.Error),
				)
			}
		case <-p.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}

// QueueWorker is a worker that processes messages from a message queue.
type QueueWorker struct {
	queue   queue.Queue
	handler queue.Handler
	logger  *zap.Logger
	quit    chan struct{}
}

// NewQueueWorker creates a new queue-based worker.
func NewQueueWorker(q queue.Queue, handler queue.Handler, logger *zap.Logger) *QueueWorker {
	return &QueueWorker{
		queue:   q,
		handler: handler,
		logger:  logger,
		quit:    make(chan struct{}),
	}
}

// Start begins consuming messages from a queue topic.
func (w *QueueWorker) Start(ctx context.Context, topics ...string) {
	for _, topic := range topics {
		go w.consume(ctx, topic)
	}
}

// Stop signals the worker to stop.
func (w *QueueWorker) Stop() {
	close(w.quit)
}

func (w *QueueWorker) consume(ctx context.Context, topic string) {
	msgChan, err := w.queue.Subscribe(ctx, topic)
	if err != nil {
		w.logger.Error("failed to subscribe to queue topic",
			zap.String("topic", topic),
			zap.Error(err),
		)
		return
	}

	w.logger.Info("queue worker started", zap.String("topic", topic))
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			if err := w.handler.Handle(ctx, msg); err != nil {
				w.logger.Error("queue message handling failed",
					zap.String("topic", topic),
					zap.String("message_id", msg.ID),
					zap.Error(err),
				)
				_ = w.queue.Nack(ctx, topic, msg.ID, true)
			} else {
				_ = w.queue.Ack(ctx, topic, msg.ID)
			}
		case <-w.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}

// JobHandler is a function type for processing jobs.
type JobHandler func(ctx context.Context, job Job) error

// Registry holds registered job handlers.
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]JobHandler
}

// NewRegistry creates a new job handler registry.
func NewRegistry() *Registry {
	return &Registry{
		handlers: make(map[string]JobHandler),
	}
}

// Register registers a handler for a job type.
func (r *Registry) Register(jobType string, handler JobHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[jobType] = handler
}

// Get returns the handler for a job type.
func (r *Registry) Get(jobType string) (JobHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	handler, ok := r.handlers[jobType]
	return handler, ok
}

// Run executes the registered handler for a job.
func (r *Registry) Run(ctx context.Context, job Job) error {
	handler, ok := r.Get(job.Type)
	if !ok {
		return nil // No handler registered - silently skip
	}
	return handler(ctx, job)
}

// Predefined job types for the Coindistro platform.
const (
	// Email jobs
	JobSendEmail         = "email.send"
	JobSendVerification  = "email.verification"
	JobSendPasswordReset = "email.password_reset"

	// Notification jobs
	JobSendPushNotification = "notification.push"
	JobSendSMS              = "notification.sms"

	// Signal jobs
	JobBroadcastSignal = "signal.broadcast"

	// Payment jobs
	JobProcessPayment = "payment.process"
	JobSettlePayment  = "payment.settle"

	// Blockchain jobs
	JobSyncBlockchain     = "blockchain.sync"
	JobConfirmTransaction = "blockchain.confirm"

	// Report jobs
	JobGenerateReport  = "report.generate"
	JobSendDailyReport = "report.daily"

	// System jobs
	JobCleanup     = "system.cleanup"
	JobHealthCheck = "system.health_check"
)
