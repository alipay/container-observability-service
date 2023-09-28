package metrics

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricNamePrefix = "lunettes_"
)

var (
	serviceOperationResultSuccessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "service_operation_successed",
			Help: "service operation success",
		},
		// service/operation
		[]string{"service", "operation"},
	)
	serviceOperationResultFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "service_operation_failed",
			Help: "service operation failed",
		},
		// service/operation
		[]string{"service", "operation"},
	)

	serviceOperationResultSuccessedNamespaced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "service_operation_successed_namespaced",
			Help: "service operation success",
		},
		// service/operation
		[]string{"service", "operation", "namespace"},
	)
	serviceOperationResultFailedNamespaced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "service_operation_failed_namespaced",
			Help: "service operation failed",
		},
		// service/operation
		[]string{"service", "operation", "namespace"},
	)

	SchedulingResultCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "scheduling_count",
			Help: "shceduling count",
		},
		// success/failed
		[]string{"result"},
	)

	ContainerStartingResultCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "container_starting_count",
			Help: "container starting result count",
		},
		// success/failed
		[]string{"result"},
	)

	serviceOperationDurationMilliSeconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MetricNamePrefix + "operation_duration_milliseconds",
			Help:       "how long an operation of service to completed",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
		},
		// service and operation
		[]string{"service", "operation"},
	)

	serviceOperationDurationMilliSecondsNamespaced = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MetricNamePrefix + "operation_duration_milliseconds_namespaced",
			Help:       "how long an operation of service to completed",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
		},
		// service and operation
		[]string{"service", "operation", "namespace"},
	)

	// DebugMethodDurationMilliSeconds 统计方法耗时
	DebugMethodDurationMilliSeconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MetricNamePrefix + "debug_method_duration_milliseconds",
			Help:       "how long a method taken",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// QueryMethodDurationMilliSeconds 统计查询方法耗时
	QueryMethodDurationMilliSeconds = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       MetricNamePrefix + "query_method_duration_milliseconds",
			Help:       "how long a query method taken",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
		},
		[]string{"method"},
	)

	ConsumerDurationMilliSeconds = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Name:       MetricNamePrefix + "consumer_duration_milliseconds",
			Help:       "how long consumer consume one event",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.001, 0.99: 0.001},
		},
	)

	// lunettes_trace_duration_milliseconds_sum
	traceDurationMilliSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    MetricNamePrefix + "trace_duration_milliseconds",
			Help:    "how long a trace from start to finish",
			Buckets: []float64{100, 500, 1000, 2000, 5000, 10000, 60000, 120000, 180000, 300000, 540000},
		},
		// create/delete/...
		[]string{"trace_type"},
	)
	traceDurationMilliSecondsNamespaced = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    MetricNamePrefix + "trace_duration_milliseconds_namespaced",
			Help:    "how long a trace from start to finish",
			Buckets: []float64{100, 500, 1000, 2000, 5000, 10000, 60000, 120000, 180000, 300000, 540000},
		},
		// create/delete/...
		[]string{"trace_type", "namespace"},
	)

	// TraceCreatedCount lunettes_trace_created_count
	TraceCreatedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_created_count",
			Help: "how many trace record have been scheduled so far",
		},
		[]string{"trace_type"},
	)

	TraceNewBeforeFinishCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_new_before_finish_count",
			Help: "how many trace added before older one have finishi",
		},
	)

	SpanCreatedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "span_created_count",
			Help: "how many spans have been created so far",
		},
		[]string{"servive", "operation"},
	)

	TraceTimeoutCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_timeout_count",
			Help: "how many timeout trace record have been scheduled so far",
		},
		[]string{"trace_type"},
	)

	TraceTimeoutCorrectToReadyCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_timeout_corrected_count",
			Help: "how many timeout traces but corrected to normal",
		},
	)

	traceSuccessedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_successed_count",
			Help: "how many successed trace record have been scheduled so far",
		},
		[]string{"trace_type"},
	)

	traceFailedCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_failed_count",
			Help: "how many failed trace record have been scheduled so far",
		},
		[]string{"trace_type", "reason"},
	)

	traceSuccessedCountNamespaced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_successed_count_namespaced",
			Help: "how many successed trace record have been scheduled so far",
		},
		[]string{"trace_type", "namespace"},
	)
	traceFailedCountNamespaced = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "trace_failed_count_namespaced",
			Help: "how many failed trace record have been scheduled so far",
		},
		[]string{"trace_type", "namespace", "reason"},
	)

	PodLifecycleDurationMilliSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: MetricNamePrefix + "pod_phase_duration_milliseconds",
			Help: "how long a trace from start to finish",
		},
		[]string{"phase"},
	)

	LRUCacheCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MetricNamePrefix + "lru_cache_fetch_count",
			Help: "how many times of get element form lru cache",
		},
		[]string{"name", "type"},
	)
)

func init() {
	if config.GlobalLunettesConfig().ShouldRetainOldMetrics {
		// lunettes_trace_duration_milliseconds
		prometheus.MustRegister(traceDurationMilliSeconds)
		// lunettes_trace_duration_milliseconds_namespaced
		prometheus.MustRegister(traceDurationMilliSecondsNamespaced)
		// lunettes_trace_created_count
		prometheus.MustRegister(TraceCreatedCount)
		// lunettes_trace_timeout_count
		prometheus.MustRegister(TraceTimeoutCount)
		// lunettes_trace_successed_count
		prometheus.MustRegister(traceSuccessedCount)
		// lunettes_trace_failed_count
		prometheus.MustRegister(traceFailedCount)
		// lunettes_trace_successed_count_namespaced
		prometheus.MustRegister(traceSuccessedCountNamespaced)
		// lunettes_trace_failed_count_namespaced
		prometheus.MustRegister(traceFailedCountNamespaced)
	}
	// lunettes_span_created_count
	prometheus.MustRegister(SpanCreatedCount)

	// lunettes_scheduling_count
	prometheus.MustRegister(SchedulingResultCount)
	// lunettes_container_starting_count
	prometheus.MustRegister(ContainerStartingResultCount)

	if config.GlobalLunettesConfig().ShouldRetainOldMetrics {
		// lunettes_service_operation_successed
		prometheus.MustRegister(serviceOperationResultSuccessed)
		// lunettes_service_operation_failed
		prometheus.MustRegister(serviceOperationResultFailed)
		// lunettes_operation_duration_milliseconds
		prometheus.MustRegister(serviceOperationDurationMilliSeconds)
		// lunettes_service_operation_successed_namespaced
		prometheus.MustRegister(serviceOperationResultSuccessedNamespaced)
		// lunettes_service_operation_failed_namespaced
		prometheus.MustRegister(serviceOperationResultFailedNamespaced)
		// lunettes_operation_duration_milliseconds_namespaced
		prometheus.MustRegister(serviceOperationDurationMilliSecondsNamespaced)
	}

	// lunettes_pod_phase_duration_milliseconds
	prometheus.MustRegister(PodLifecycleDurationMilliSeconds)
	// lunettes_trace_new_before_finish_count
	prometheus.MustRegister(TraceNewBeforeFinishCount)

	// lunettes_consumer_duration_milliseconds
	prometheus.MustRegister(ConsumerDurationMilliSeconds)
	// lunettes_debug_method_duration_milliseconds
	prometheus.MustRegister(DebugMethodDurationMilliSeconds)
	// lunettes_query_method_duration_milliseconds
	prometheus.MustRegister(QueryMethodDurationMilliSeconds)
	// lunettes_trace_timeout_corrected_count
	prometheus.MustRegister(TraceTimeoutCorrectToReadyCount)
	// lunettes_lru_cache_fetch_count
	prometheus.MustRegister(LRUCacheCounter)
}

// UpdateServiceOperationDurationMetrics update ServiceOperationDurationMilliSeconds
func UpdateServiceOperationDurationMetrics(service, operation, namespace string, d float64) {
	serviceOperationDurationMilliSecondsNamespaced.WithLabelValues(service, operation, namespace).Observe(d)
	serviceOperationDurationMilliSeconds.WithLabelValues(service, operation).Observe(d)
}

// UpdateServiceOperationResultMetrics update ServiceOperationDurationMilliSeconds
func UpdateServiceOperationResultMetrics(service, operation, namespace string, err bool) {
	if err {
		serviceOperationResultFailed.WithLabelValues(service, operation).Inc()
		serviceOperationResultFailedNamespaced.WithLabelValues(service, operation, namespace).Inc()
	} else {
		serviceOperationResultSuccessed.WithLabelValues(service, operation).Inc()
		serviceOperationResultSuccessedNamespaced.WithLabelValues(service, operation, namespace).Inc()
	}
}

// IncreaseFailedTraceCount increase counter for failed traces
func IncreaseFailedTraceCount(traceType, namespace, reasonName string) {
	traceFailedCount.WithLabelValues(traceType, reasonName).Inc()
	traceFailedCountNamespaced.WithLabelValues(traceType, namespace, reasonName).Inc()
}

// IncreaseSucceedTraceCount increase counter for succeed traces
func IncreaseSucceedTraceCount(traceType, namespace string) {
	traceSuccessedCount.WithLabelValues(traceType).Inc()
	traceSuccessedCountNamespaced.WithLabelValues(traceType, namespace).Inc()
}

// ObserveTraceDuration observe for trace durations
func ObserveTraceDuration(traceType, namespace string, v float64) {
	traceDurationMilliSeconds.WithLabelValues(traceType).Observe(v)
	traceDurationMilliSecondsNamespaced.WithLabelValues(traceType, namespace).Observe(v)
}

func ObserveQueryMethodDuration(method string, begin time.Time) {
	cost := float64(time.Since(begin).Nanoseconds() / time.Millisecond.Nanoseconds())
	QueryMethodDurationMilliSeconds.WithLabelValues(method).Observe(cost)
}
