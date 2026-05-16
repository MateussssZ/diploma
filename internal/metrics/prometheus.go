package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type IMetrics interface {
	GetPrometheusRegistry() *prometheus.Registry

	// gRPC
	GRPCServerRequestsTotalInc(service, method, code string)
	GRPCServerRequestDurationInc(service, method string, duration float64)
	GRPCServerRequestsInFlightInc()
	GRPCServerRequestsInFlightDec()

	// HTTP
	HTTPServerRequestDurationInc(path, method string, duration float64)
	HTTPServerRequestsErrorsTotalInc(path, method string)

	// Generic operations (legacy)
	OperationsTotalInc()
	OperationsErrorsTotalInc()

	// Kafka
	KafkaMessagesConsumedInc(topic string)
	KafkaConsumerErrorsInc(topic string)
	KafkaMessageProcessingDurationInc(topic string, duration float64)

	// WebSocket
	WSConnectionsInc()
	WSConnectionsDec()
	WSBroadcastDroppedInc()
	WSMessagesReceivedInc(action string)
	WSMessagesSentInc(event string)

	// Cache
	CacheHitInc(key string)
	CacheMissInc(key string)
	CacheInvalidationInc(key string)
	CacheErrorInc(operation string)
}

type metricsCollector struct {
	// gRPC
	gRPCServerRequestsTotal    *prometheus.CounterVec
	gRPCServerRequestDuration  *prometheus.HistogramVec
	gRPCServerRequestsInFlight prometheus.Gauge

	// HTTP
	httpServerRequestDuration    *prometheus.HistogramVec
	httpServerRequestErrorsTotal *prometheus.CounterVec

	// Generic
	operationsTotal       prometheus.Counter
	operationsErrorsTotal prometheus.Counter

	// Kafka
	kafkaMessagesConsumed          *prometheus.CounterVec
	kafkaConsumerErrors            *prometheus.CounterVec
	kafkaMessageProcessingDuration *prometheus.HistogramVec

	// WebSocket
	wsConnectionsActive prometheus.Gauge
	wsBroadcastDropped  prometheus.Counter
	wsMessagesReceived  *prometheus.CounterVec
	wsMessagesSent      *prometheus.CounterVec

	// Cache
	cacheHits          *prometheus.CounterVec
	cacheMisses        *prometheus.CounterVec
	cacheInvalidations *prometheus.CounterVec
	cacheErrors        *prometheus.CounterVec
}

type Prometheus struct {
	registry  *prometheus.Registry
	collector metricsCollector
}

func NewPrometheus() *Prometheus {
	registry := prometheus.NewRegistry()

	// gRPC
	gRPCServerRequestsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "grpc_server_requests_total", Help: "Total gRPC requests received"},
		[]string{"service", "method", "code"},
	)
	gRPCServerRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "grpc_server_request_duration_seconds", Help: "gRPC request duration", Buckets: prometheus.DefBuckets},
		[]string{"service", "method"},
	)
	gRPCServerRequestsInFlight := prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "grpc_server_requests_in_flight", Help: "gRPC requests currently processing"},
	)

	// HTTP
	httpServerRequestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_server_request_duration_seconds", Help: "HTTP request duration", Buckets: prometheus.DefBuckets},
		[]string{"path", "method"},
	)
	httpServerRequestErrorsTotal := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_server_requests_errors_total", Help: "Total HTTP requests that resulted in an error"},
		[]string{"path", "method"},
	)

	// Generic
	operationsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "operations_total", Help: "Total number of all operations",
	})
	operationsErrorsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "operations_errors_total", Help: "Total number of failed operations",
	})

	// Kafka
	kafkaMessagesConsumed := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "kafka_messages_consumed_total", Help: "Total Kafka messages consumed"},
		[]string{"topic"},
	)
	kafkaConsumerErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "kafka_consumer_errors_total", Help: "Total Kafka consumer errors"},
		[]string{"topic"},
	)
	kafkaMessageProcessingDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "kafka_message_processing_duration_seconds",
			Help:    "Time to process a single Kafka message",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"topic"},
	)

	// WebSocket
	wsConnectionsActive := prometheus.NewGauge(
		prometheus.GaugeOpts{Name: "ws_connections_active", Help: "Number of active WebSocket connections"},
	)
	wsBroadcastDropped := prometheus.NewCounter(
		prometheus.CounterOpts{Name: "ws_broadcast_dropped_total", Help: "Total WebSocket broadcast events dropped due to full buffers"},
	)
	wsMessagesReceived := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ws_messages_received_total", Help: "Total WebSocket messages received from clients"},
		[]string{"action"},
	)
	wsMessagesSent := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "ws_messages_sent_total", Help: "Total WebSocket events sent to clients"},
		[]string{"event"},
	)

	// Cache
	cacheHits := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_hits_total", Help: "Total cache hits"},
		[]string{"key"},
	)
	cacheMisses := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_misses_total", Help: "Total cache misses"},
		[]string{"key"},
	)
	cacheInvalidations := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_invalidations_total", Help: "Total cache invalidations"},
		[]string{"key"},
	)
	cacheErrors := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "cache_errors_total", Help: "Total cache errors by operation"},
		[]string{"operation"},
	)

	registry.MustRegister(
		gRPCServerRequestsTotal,
		gRPCServerRequestDuration,
		gRPCServerRequestsInFlight,
		httpServerRequestDuration,
		httpServerRequestErrorsTotal,
		operationsTotal,
		operationsErrorsTotal,
		kafkaMessagesConsumed,
		kafkaConsumerErrors,
		kafkaMessageProcessingDuration,
		wsConnectionsActive,
		wsBroadcastDropped,
		wsMessagesReceived,
		wsMessagesSent,
		cacheHits,
		cacheMisses,
		cacheInvalidations,
		cacheErrors,
	)

	return &Prometheus{
		registry: registry,
		collector: metricsCollector{
			gRPCServerRequestsTotal:        gRPCServerRequestsTotal,
			gRPCServerRequestDuration:      gRPCServerRequestDuration,
			gRPCServerRequestsInFlight:     gRPCServerRequestsInFlight,
			httpServerRequestDuration:      httpServerRequestDuration,
			httpServerRequestErrorsTotal:   httpServerRequestErrorsTotal,
			operationsTotal:                operationsTotal,
			operationsErrorsTotal:          operationsErrorsTotal,
			kafkaMessagesConsumed:          kafkaMessagesConsumed,
			kafkaConsumerErrors:            kafkaConsumerErrors,
			kafkaMessageProcessingDuration: kafkaMessageProcessingDuration,
			wsConnectionsActive:            wsConnectionsActive,
			wsBroadcastDropped:             wsBroadcastDropped,
			wsMessagesReceived:             wsMessagesReceived,
			wsMessagesSent:                 wsMessagesSent,
			cacheHits:                      cacheHits,
			cacheMisses:                    cacheMisses,
			cacheInvalidations:             cacheInvalidations,
			cacheErrors:                    cacheErrors,
		},
	}
}

func (p *Prometheus) GetPrometheusRegistry() *prometheus.Registry { return p.registry }

// gRPC
func (p *Prometheus) GRPCServerRequestsTotalInc(service, method, code string) {
	p.collector.gRPCServerRequestsTotal.WithLabelValues(service, method, code).Inc()
}
func (p *Prometheus) GRPCServerRequestDurationInc(service, method string, duration float64) {
	p.collector.gRPCServerRequestDuration.WithLabelValues(service, method).Observe(duration)
}
func (p *Prometheus) GRPCServerRequestsInFlightInc() { p.collector.gRPCServerRequestsInFlight.Inc() }
func (p *Prometheus) GRPCServerRequestsInFlightDec() { p.collector.gRPCServerRequestsInFlight.Dec() }

// HTTP
func (p *Prometheus) HTTPServerRequestDurationInc(path, method string, duration float64) {
	p.collector.httpServerRequestDuration.WithLabelValues(path, method).Observe(duration)
}
func (p *Prometheus) HTTPServerRequestsErrorsTotalInc(path, method string) {
	p.collector.httpServerRequestErrorsTotal.WithLabelValues(path, method).Inc()
}

// Generic
func (p *Prometheus) OperationsTotalInc()       { p.collector.operationsTotal.Inc() }
func (p *Prometheus) OperationsErrorsTotalInc() { p.collector.operationsErrorsTotal.Inc() }

// Kafka
func (p *Prometheus) KafkaMessagesConsumedInc(topic string) {
	p.collector.kafkaMessagesConsumed.WithLabelValues(topic).Inc()
}
func (p *Prometheus) KafkaConsumerErrorsInc(topic string) {
	p.collector.kafkaConsumerErrors.WithLabelValues(topic).Inc()
}
func (p *Prometheus) KafkaMessageProcessingDurationInc(topic string, duration float64) {
	p.collector.kafkaMessageProcessingDuration.WithLabelValues(topic).Observe(duration)
}

// WebSocket
func (p *Prometheus) WSConnectionsInc()      { p.collector.wsConnectionsActive.Inc() }
func (p *Prometheus) WSConnectionsDec()      { p.collector.wsConnectionsActive.Dec() }
func (p *Prometheus) WSBroadcastDroppedInc() { p.collector.wsBroadcastDropped.Inc() }
func (p *Prometheus) WSMessagesReceivedInc(action string) {
	p.collector.wsMessagesReceived.WithLabelValues(action).Inc()
}
func (p *Prometheus) WSMessagesSentInc(event string) {
	p.collector.wsMessagesSent.WithLabelValues(event).Inc()
}

// Cache
func (p *Prometheus) CacheHitInc(key string) {
	p.collector.cacheHits.WithLabelValues(key).Inc()
}
func (p *Prometheus) CacheMissInc(key string) {
	p.collector.cacheMisses.WithLabelValues(key).Inc()
}
func (p *Prometheus) CacheInvalidationInc(key string) {
	p.collector.cacheInvalidations.WithLabelValues(key).Inc()
}
func (p *Prometheus) CacheErrorInc(operation string) {
	p.collector.cacheErrors.WithLabelValues(operation).Inc()
}
