package metrics

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	RequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests in flight",
		},
		[]string{"method"},
	)

	CacheOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_operations_total",
			Help: "Total number of Redis cache get operations by prefix and result",
		},
		[]string{"prefix", "result"},
	)

	RateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of requests rejected with HTTP 429 by route rule",
		},
		[]string{"route"},
	)

	AuthLoginAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_login_attempts_total",
			Help: "Total number of login attempts by outcome",
		},
		[]string{"result"},
	)

	DailyRunsStartedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "game_daily_runs_started_total",
			Help: "Total number of daily runs started by tier",
		},
		[]string{"tier"},
	)

	DailyRunsCompletedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "game_daily_runs_completed_total",
			Help: "Total number of daily runs finished by tier and result (victory or defeat)",
		},
		[]string{"tier", "result"},
	)

	CombatActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "game_combat_actions_total",
			Help: "Total number of daily run combat actions by action type and outcome",
		},
		[]string{"action", "outcome"},
	)

	CharacterPrestigesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "game_character_prestiges_total",
			Help: "Total number of character prestige ascensions completed",
		},
	)
)

func RegisterDBPoolMetrics(pool *pgxpool.Pool) {
	prometheus.MustRegister(&pgxPoolCollector{
		pool: pool,
		acquired: prometheus.NewDesc(
			"db_pool_acquired_conns",
			"Number of currently acquired connections in the PostgreSQL pool",
			nil, nil,
		),
		idle: prometheus.NewDesc(
			"db_pool_idle_conns",
			"Number of currently idle connections in the PostgreSQL pool",
			nil, nil,
		),
		total: prometheus.NewDesc(
			"db_pool_total_conns",
			"Total number of connections currently in the PostgreSQL pool",
			nil, nil,
		),
		max: prometheus.NewDesc(
			"db_pool_max_conns",
			"Maximum configured size of the PostgreSQL connection pool",
			nil, nil,
		),
	})
}

type pgxPoolCollector struct {
	pool     *pgxpool.Pool
	acquired *prometheus.Desc
	idle     *prometheus.Desc
	total    *prometheus.Desc
	max      *prometheus.Desc
}

func (c *pgxPoolCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.acquired
	ch <- c.idle
	ch <- c.total
	ch <- c.max
}

func (c *pgxPoolCollector) Collect(ch chan<- prometheus.Metric) {
	stat := c.pool.Stat()
	ch <- prometheus.MustNewConstMetric(c.acquired, prometheus.GaugeValue, float64(stat.AcquiredConns()))
	ch <- prometheus.MustNewConstMetric(c.idle, prometheus.GaugeValue, float64(stat.IdleConns()))
	ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, float64(stat.TotalConns()))
	ch <- prometheus.MustNewConstMetric(c.max, prometheus.GaugeValue, float64(stat.MaxConns()))
}

func Handler() http.Handler {
	return promhttp.Handler()
}
