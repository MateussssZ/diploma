package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type IMetrics interface {
	GetPrometheusRegistry() *prometheus.Registry
	GRPCServerRequestsTotalInc(service, method, code string)
	GRPCServerRequestDurationInc(service, method string, duration float64)
	GRPCServerRequestsInFlightInc()
	GRPCServerRequestsInFlightDec()
	HTTPServerRequestDurationInc(path, method string, duration float64)
	OperationsTotalInc()
	OperationsErrorsTotalInc()
}

type metricsCollector struct {
	gRPCServerRequestsTotal    *prometheus.CounterVec
	gRPCServerRequestDuration  *prometheus.HistogramVec
	gRPCServerRequestsInFlight prometheus.Gauge
	httpServerRequestDuration  *prometheus.HistogramVec
	operationsTotal            prometheus.Counter
	operationsErrorsTotal      prometheus.Counter
}

type Prometheus struct {
	registry  *prometheus.Registry
	collector metricsCollector
}

func NewPrometheus() *Prometheus {
	registry := prometheus.NewRegistry()

	gRPCServerRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_server_requests_total",
			Help: "Total number of gRPC requests received",
		},
		[]string{"service", "method", "code"},
	)

	gRPCServerRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_server_request_duration_seconds",
			Help:    "gRPC request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	gRPCServerRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "grpc_server_requests_in_flight",
			Help: "Current number of gRPC requests being processed",
		},
	)

	httpServerRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_server_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	operationsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "operations_total",
		Help: "Total number of all operations.",
	})

	operationsErrorsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "operations_errors_total",
		Help: "Total number of all failed operations.",
	})

	registry.MustRegister(
		gRPCServerRequestsTotal,
		gRPCServerRequestDuration,
		gRPCServerRequestsInFlight,
		httpServerRequestDuration,
		operationsTotal,
		operationsErrorsTotal,
	)

	return &Prometheus{
		registry: registry,
		collector: metricsCollector{
			gRPCServerRequestsTotal:    gRPCServerRequestsTotal,
			gRPCServerRequestDuration:  gRPCServerRequestDuration,
			gRPCServerRequestsInFlight: gRPCServerRequestsInFlight,
			httpServerRequestDuration:  httpServerRequestDuration,
			operationsTotal:            operationsTotal,
			operationsErrorsTotal:      operationsErrorsTotal,
		},
	}
}

func (p *Prometheus) GetPrometheusRegistry() *prometheus.Registry {
	return p.registry
}

func (p *Prometheus) GRPCServerRequestsTotalInc(service, method, code string) {
	p.collector.gRPCServerRequestsTotal.WithLabelValues(service, method, code).Inc()
}

func (p *Prometheus) GRPCServerRequestDurationInc(service, method string, duration float64) {
	p.collector.gRPCServerRequestDuration.WithLabelValues(service, method).Observe(duration)
}

func (p *Prometheus) GRPCServerRequestsInFlightInc() {
	p.collector.gRPCServerRequestsInFlight.Inc()
}

func (p *Prometheus) GRPCServerRequestsInFlightDec() {
	p.collector.gRPCServerRequestsInFlight.Dec()
}

func (p *Prometheus) HTTPServerRequestDurationInc(path, method string, duration float64) {
	p.collector.httpServerRequestDuration.WithLabelValues(path, method).Observe(duration)
}

func (p *Prometheus) OperationsTotalInc() {
	p.collector.operationsTotal.Inc()
}

func (p *Prometheus) OperationsErrorsTotalInc() {
	p.collector.operationsErrorsTotal.Inc()
}
