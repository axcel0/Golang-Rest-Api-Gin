package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.SummaryVec
	HTTPResponseSize    *prometheus.SummaryVec
	ActiveConnections   prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	m := &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests by method, endpoint, and status code",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_request_size_bytes",
				Help: "Size of HTTP request in bytes",
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Name: "http_response_size_bytes",
				Help: "Size of HTTP response in bytes",
			},
			[]string{"method", "endpoint"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_active_connections",
				Help: "Number of active HTTP connections",
			},
		),
	}

	return m
}

// Middleware creates a Gin middleware that records metrics for each request
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment active connections
		m.ActiveConnections.Inc()
		defer m.ActiveConnections.Dec()

		// Start timer
		start := time.Now()

		// Get request size
		requestSize := computeApproximateRequestSize(c.Request)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code and endpoint
		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}
		method := c.Request.Method

		// Record metrics
		m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
		m.HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
		m.HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(c.Writer.Size()))
	}
}

// computeApproximateRequestSize computes the approximate size of the request
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s += len(r.URL.String())
	}

	s += len(r.Method)
	s += len(r.Proto)

	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}

	s += len(r.Host)

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}

	return s
}
