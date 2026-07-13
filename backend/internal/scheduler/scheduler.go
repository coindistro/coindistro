package scheduler

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Task represents a scheduled task.
type Task struct {
	ID       string
	Name     string
	Interval time.Duration
	Cron     string // Future: cron expression support
	Handler  TaskHandler
	RunOnce  bool
}

// TaskHandler is a function that executes a scheduled task.
type TaskHandler func(ctx context.Context) error

// TaskStatus represents the current status of a task.
type TaskStatus struct {
	TaskID       string        `json:"task_id"`
	Name         string        `json:"name"`
	LastRunAt    time.Time     `json:"last_run_at"`
	LastDuration time.Duration `json:"last_duration"`
	LastError    string        `json:"last_error,omitempty"`
	RunCount     int64         `json:"run_count"`
	ErrorCount   int64         `json:"error_count"`
	IsRunning    bool          `json:"is_running"`
	NextRunAt    time.Time     `json:"next_run_at"`
}

// Scheduler manages recurring and one-time tasks.
type Scheduler struct {
	mu     sync.RWMutex
	tasks  map[string]*scheduledTask
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

type scheduledTask struct {
	Task
	status TaskStatus
	ticker *time.Ticker
}

// New creates a new Scheduler.
func New(logger *zap.Logger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		tasks:  make(map[string]*scheduledTask),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddTask adds a recurring task to the scheduler.
func (s *Scheduler) AddTask(task Task) {
	s.mu.Lock()
	defer s.mu.Unlock()

	st := &scheduledTask{
		Task: task,
		status: TaskStatus{
			TaskID: task.ID,
			Name:   task.Name,
		},
	}

	s.tasks[task.ID] = st
	s.logger.Info("scheduled task added",
		zap.String("task_id", task.ID),
		zap.String("name", task.Name),
		zap.Duration("interval", task.Interval),
	)
}

// Start begins executing all scheduled tasks.
func (s *Scheduler) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, st := range s.tasks {
		s.startTask(st)
	}

	s.logger.Info("scheduler started", zap.Int("tasks", len(s.tasks)))
}

// Stop gracefully stops all scheduled tasks.
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

// GetStatus returns the status of all tasks.
func (s *Scheduler) GetStatus() []TaskStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	statuses := make([]TaskStatus, 0, len(s.tasks))
	for _, st := range s.tasks {
		statuses = append(statuses, st.status)
	}
	return statuses
}

// GetTaskStatus returns the status of a specific task.
func (s *Scheduler) GetTaskStatus(taskID string) (*TaskStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st, ok := s.tasks[taskID]
	if !ok {
		return nil, false
	}
	status := st.status
	return &status, true
}

// RemoveTask removes a task from the scheduler.
func (s *Scheduler) RemoveTask(taskID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if st, ok := s.tasks[taskID]; ok {
		if st.ticker != nil {
			st.ticker.Stop()
		}
		delete(s.tasks, taskID)
		s.logger.Info("scheduled task removed", zap.String("task_id", taskID))
	}
}

func (s *Scheduler) startTask(st *scheduledTask) {
	if st.RunOnce {
		go s.executeOnce(st)
		return
	}

	st.ticker = time.NewTicker(st.Interval)
	s.wg.Add(1)
	go s.executeRecurring(st)
}

func (s *Scheduler) executeRecurring(st *scheduledTask) {
	defer s.wg.Done()
	defer st.ticker.Stop()

	for {
		select {
		case <-st.ticker.C:
			s.executeTask(st)
		case <-s.ctx.Done():
			return
		}
	}
}

func (s *Scheduler) executeOnce(st *scheduledTask) {
	s.wg.Add(1)
	defer s.wg.Done()

	// Small delay to allow scheduler to start
	time.Sleep(100 * time.Millisecond)
	s.executeTask(st)
}

func (s *Scheduler) executeTask(st *scheduledTask) {
	s.mu.Lock()
	if st.status.IsRunning {
		s.mu.Unlock()
		return
	}
	st.status.IsRunning = true
	st.status.LastRunAt = time.Now()
	s.mu.Unlock()

	start := time.Now()
	err := st.Handler(s.ctx)
	duration := time.Since(start)

	s.mu.Lock()
	st.status.LastDuration = duration
	st.status.RunCount++
	st.status.IsRunning = false
	st.status.NextRunAt = time.Now().Add(st.Interval)

	if err != nil {
		st.status.ErrorCount++
		st.status.LastError = err.Error()
		s.logger.Error("scheduled task failed",
			zap.String("task_id", st.ID),
			zap.String("name", st.Name),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		st.status.LastError = ""
		s.logger.Info("scheduled task completed",
			zap.String("task_id", st.ID),
			zap.String("name", st.Name),
			zap.Duration("duration", duration),
		)
	}
	s.mu.Unlock()
}

// Predefined task IDs for the Coindistro platform.
const (
	TaskMarketSync        = "market_sync"
	TaskDailyReport       = "daily_report"
	TaskLeaderboard       = "leaderboard"
	TaskPortfolioCalc     = "portfolio_calc"
	TaskCertificateGen    = "certificate_gen"
	TaskCleanup           = "cleanup"
	TaskHealthCheck       = "health_check"
	TaskBlockchainSync    = "blockchain_sync"
	TaskPaymentSettlement = "payment_settlement"
	TaskSignalExpiry      = "signal_expiry"
)

// DefaultTaskConfigs returns default configurations for common tasks.
func DefaultTaskConfigs() map[string]time.Duration {
	return map[string]time.Duration{
		TaskMarketSync:        30 * time.Second,
		TaskDailyReport:       24 * time.Hour,
		TaskLeaderboard:       1 * time.Hour,
		TaskPortfolioCalc:     5 * time.Minute,
		TaskCertificateGen:    10 * time.Minute,
		TaskCleanup:           1 * time.Hour,
		TaskHealthCheck:       30 * time.Second,
		TaskBlockchainSync:    1 * time.Minute,
		TaskPaymentSettlement: 1 * time.Minute,
		TaskSignalExpiry:      5 * time.Minute,
	}
}
