package metrics

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Namespace for all Coindistro metrics.
const namespace = "coindistro"

// Metrics holds all Prometheus metric collectors.
type Metrics struct {
	// HTTP metrics
	HTTPRequestCount    *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.SummaryVec
	HTTPResponseSize    *prometheus.SummaryVec

	// Database metrics
	DatabaseLatency  *prometheus.HistogramVec
	DatabaseErrors   *prometheus.CounterVec
	DatabaseQueries  *prometheus.CounterVec
	DatabasePoolSize *prometheus.GaugeVec

	// Redis metrics
	RedisLatency    *prometheus.HistogramVec
	RedisErrors     *prometheus.CounterVec
	RedisOperations *prometheus.CounterVec

	// Worker metrics
	WorkerQueueDepth  *prometheus.GaugeVec
	WorkerJobsTotal   *prometheus.CounterVec
	WorkerJobDuration *prometheus.HistogramVec
	WorkerErrors      *prometheus.CounterVec

	// Scheduler metrics
	SchedulerTasksTotal   *prometheus.CounterVec
	SchedulerTaskDuration *prometheus.HistogramVec
	SchedulerTaskErrors   *prometheus.CounterVec

	// Event bus metrics
	EventsPublished *prometheus.CounterVec
	EventsHandled   *prometheus.CounterVec

	// Cache metrics
	CacheHits   *prometheus.CounterVec
	CacheMisses *prometheus.CounterVec

	// System metrics
	SystemMemoryUsage prometheus.Gauge
	SystemGoroutines  prometheus.Gauge
	SystemCPUUsage    prometheus.Gauge
	SystemOpenFDs     prometheus.Gauge

	// Business metrics
	ActiveUsers       prometheus.Gauge
	TotalTransactions prometheus.Counter
	TotalDeposits     *prometheus.CounterVec
	TotalWithdrawals  *prometheus.CounterVec
	TotalSignals      prometheus.Counter
	ActiveBots        prometheus.Gauge

	// Earn metrics
	EarnActiveProducts      prometheus.Gauge
	EarnActiveParticipants  prometheus.Gauge
	EarnParticipationsTotal prometheus.Counter
	EarnRewardCalculations  prometheus.Counter
	EarnRewardDistributions prometheus.Counter
	EarnProductCapacity     *prometheus.GaugeVec
	EarnAvgAllocation       *prometheus.GaugeVec
	EarnFailedOperations    *prometheus.CounterVec
}

// New creates and registers all Prometheus metrics.
func New() *Metrics {
	m := &Metrics{
		// HTTP metrics
		HTTPRequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "path"},
		),
		HTTPRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_request_size_bytes",
				Help:      "HTTP request size in bytes",
			},
			[]string{"method"},
		),
		HTTPResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: namespace,
				Name:      "http_response_size_bytes",
				Help:      "HTTP response size in bytes",
			},
			[]string{"method"},
		),

		// Database metrics
		DatabaseLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "database",
				Name:      "query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
			},
			[]string{"operation"},
		),
		DatabaseErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "database",
				Name:      "errors_total",
				Help:      "Total number of database errors",
			},
			[]string{"operation"},
		),
		DatabaseQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "database",
				Name:      "queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation"},
		),
		DatabasePoolSize: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "database",
				Name:      "pool_size",
				Help:      "Database connection pool size",
			},
			[]string{"state"},
		),

		// Redis metrics
		RedisLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "command_duration_seconds",
				Help:      "Redis command duration in seconds",
				Buckets:   []float64{.0005, .001, .0025, .005, .01, .025, .05, .1, .25, .5},
			},
			[]string{"command"},
		),
		RedisErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "errors_total",
				Help:      "Total number of Redis errors",
			},
			[]string{"command"},
		),
		RedisOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "redis",
				Name:      "operations_total",
				Help:      "Total number of Redis operations",
			},
			[]string{"command"},
		),

		// Worker metrics
		WorkerQueueDepth: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "worker",
				Name:      "queue_depth",
				Help:      "Current worker queue depth",
			},
			[]string{"pool"},
		),
		WorkerJobsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "worker",
				Name:      "jobs_total",
				Help:      "Total number of jobs processed",
			},
			[]string{"type", "status"},
		),
		WorkerJobDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "worker",
				Name:      "job_duration_seconds",
				Help:      "Worker job duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"type"},
		),
		WorkerErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "worker",
				Name:      "errors_total",
				Help:      "Total number of worker errors",
			},
			[]string{"type"},
		),

		// Scheduler metrics
		SchedulerTasksTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "scheduler",
				Name:      "tasks_total",
				Help:      "Total number of scheduled tasks executed",
			},
			[]string{"task", "status"},
		),
		SchedulerTaskDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "scheduler",
				Name:      "task_duration_seconds",
				Help:      "Scheduled task duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"task"},
		),
		SchedulerTaskErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "scheduler",
				Name:      "task_errors_total",
				Help:      "Total number of scheduled task errors",
			},
			[]string{"task"},
		),

		// Event bus metrics
		EventsPublished: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "events",
				Name:      "published_total",
				Help:      "Total number of events published",
			},
			[]string{"type"},
		),
		EventsHandled: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "events",
				Name:      "handled_total",
				Help:      "Total number of events handled",
			},
			[]string{"type"},
		),

		// Cache metrics
		CacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "cache",
				Name:      "hits_total",
				Help:      "Total number of cache hits",
			},
			[]string{"cache"},
		),
		CacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "cache",
				Name:      "misses_total",
				Help:      "Total number of cache misses",
			},
			[]string{"cache"},
		),

		// System metrics
		SystemMemoryUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		}),
		SystemGoroutines: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "goroutines_total",
			Help:      "Current number of goroutines",
		}),
		SystemCPUUsage: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "cpu_usage_percent",
			Help:      "Current CPU usage percentage",
		}),
		SystemOpenFDs: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "system",
			Name:      "open_fds",
			Help:      "Current number of open file descriptors",
		}),

		// Business metrics
		ActiveUsers: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "business",
			Name:      "active_users",
			Help:      "Current number of active users",
		}),
		TotalTransactions: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "business",
			Name:      "transactions_total",
			Help:      "Total number of transactions",
		}),
		TotalDeposits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "business",
				Name:      "deposits_total",
				Help:      "Total number of deposits",
			},
			[]string{"currency"},
		),
		TotalWithdrawals: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "business",
				Name:      "withdrawals_total",
				Help:      "Total number of withdrawals",
			},
			[]string{"currency"},
		),
		TotalSignals: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "business",
			Name:      "signals_total",
			Help:      "Total number of trading signals published",
		}),
		ActiveBots: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "business",
			Name:      "active_bots",
			Help:      "Current number of active trading bots",
		}),

		// Earn metrics
		EarnActiveProducts: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "earn",
			Name:      "active_products",
			Help:      "Number of active earn products",
		}),
		EarnActiveParticipants: promauto.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "earn",
			Name:      "active_participants",
			Help:      "Distinct users with active earn participations",
		}),
		EarnParticipationsTotal: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "earn",
			Name:      "participations_total",
			Help:      "Total earn participations created",
		}),
		EarnRewardCalculations: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "earn",
			Name:      "reward_calculations_total",
			Help:      "Total reward calculations performed",
		}),
		EarnRewardDistributions: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "earn",
			Name:      "reward_distributions_total",
			Help:      "Total reward distributions recorded",
		}),
		EarnProductCapacity: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "earn",
				Name:      "product_capacity_used",
				Help:      "Capacity used per earn product",
			},
			[]string{"product_id"},
		),
		EarnAvgAllocation: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "earn",
				Name:      "avg_allocation",
				Help:      "Average allocation per earn product",
			},
			[]string{"product_id"},
		),
		EarnFailedOperations: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "earn",
				Name:      "failed_operations_total",
				Help:      "Failed earn operations by type",
			},
			[]string{"operation"},
		),
	}

	return m
}

// RecordSystemMetrics periodically records system-level metrics.
func (m *Metrics) RecordSystemMetrics(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			m.SystemMemoryUsage.Set(float64(memStats.Alloc))
			m.SystemGoroutines.Set(float64(runtime.NumGoroutine()))
		case <-ctx.Done():
			return
		}
	}
}

// Handler returns an HTTP handler for the /metrics endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware returns a Gin middleware that records HTTP metrics.
func Middleware(metrics *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		metrics.HTTPRequestCount.WithLabelValues(method, path, http.StatusText(status)).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration.Seconds())
	}
}

// RecordDBLatency records database query latency.
func (m *Metrics) RecordDBLatency(operation string, duration time.Duration) {
	m.DatabaseLatency.WithLabelValues(operation).Observe(duration.Seconds())
	m.DatabaseQueries.WithLabelValues(operation).Inc()
}

// RecordDBError records a database error.
func (m *Metrics) RecordDBError(operation string) {
	m.DatabaseErrors.WithLabelValues(operation).Inc()
}
