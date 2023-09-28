package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	StartupLatencyBuckets = []float64{0.5, 1, 2, 4, 6, 10, 20, 30, 45, 60, 120, 180, 240, 300, 360,
		480, 600, 900, 1200, 1800, 2700, 3600, 7200, 14400, 43200, 86400}

	// PodStartupLatencyExcludingShceduling is a prometheus metric for monitoring pod startup latency.
	PodStartupLatencyExcludingShceduling = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_pod_startup_latency_excluding_scheduing_second",
			Help:    "Pod startup latencies in seconds, without scheduling times",
			Buckets: StartupLatencyBuckets,
		},
		[]string{"cluster", "namespace", "ownerref", "milestone", "is_job"},
	)

	// PodStartupLatency is a prometheus metric for monitoring pod startup latency including scheduling times.
	PodStartupLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_pod_startup_latency_second",
			Help:    "Pod startup latencies in seconds, with image pull times",
			Buckets: StartupLatencyBuckets,
		},
		[]string{"cluster", "namespace", "ownerref", "milestone", "is_job", "scheduling_strategy", "cores"},
	)

	// PodStartupSLOLatency is a prometheus metric for monitoring pod startup latency including scheduling times.
	PodStartupSLOLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_pod_startup_slo_latency_second",
			Help:    "Pod startup slo latencies in seconds, with image pull times",
			Buckets: StartupLatencyBuckets,
		},
		[]string{"cluster", "namespace", "ownerref", "milestone", "is_job", "slo_time", "cores"},
	)

	// PodStartupK8sSLOLatency is a prometheus metric for monitoring pod startup latency including scheduling times.
	PodStartupK8sSLOLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_pod_startup_k8s_slo_latency_second",
			Help:    "Pod startup slo latencies in seconds, with image pull times",
			Buckets: StartupLatencyBuckets,
		},
		[]string{"cluster", "namespace", "slo_time", "cores"},
	)

	// PodStartupResultExcludingScheduling Pod startup result without scheduling
	PodStartupResultExcludingScheduling = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_startup_result_excluding_scheduling_count",
			Help: "Pod startup result, succeed or failed or timeout",
		},
		[]string{"cluster", "namespace", "ownerref", "result", "is_job", "delivery_status"},
	)

	// PodStartupResult is a prometheus metric for monitoring pod startup result
	PodStartupResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_startup_result_count",
			Help: "Pod startup result, succeed or failed or timeout",
		},
		[]string{"cluster", "namespace", "ownerref", "result", "scheduling_strategy", "cores", "is_job", "node_ip", "delivery_status", "podslo"},
	)

	// PodCreateTotal is a prometheus metric for monitoring pod create total count
	PodCreateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_create_total_count",
			Help: "Pod create total",
		},
		[]string{"cluster", "namespace", "cores", "is_job"},
	)

	// PodStartupSLOResult is a prometheus metric for monitoring pod startup result
	PodStartupSLOResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_startup_slo_result_count",
			Help: "Pod startup slo result, succeed or failed or timeout",
		},
		[]string{"cluster", "namespace", "ownerref", "result", "slo_time", "cores", "is_job", "priority", "delivery_status", "slo_reason", "slotime_adjusted"},
	)
	// PodStartupK8sSLOResult is a prometheus metric for monitoring pod startup result
	PodStartupK8sSLOResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_startup_k8s_slo_result_count",
			Help: "Pod startup slo result, succeed or failed or timeout",
		},
		[]string{"cluster", "namespace", "result", "slo_time", "cores", "delivery_status", "slo_reason"},
	)

	// PodCreateAPIResult 记录pod create API的返回结果
	PodCreateAPIResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_create_api_result_count",
			Help: "pod create api return result",
		},
		[]string{"cluster", "namespace", "resultCode"},
	)

	//EventConsumedCount is a prometheus metric for monitoring speed of event consumed
	EventConsumedCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "lunettes_events_consumed_create_slo",
			Help: "events processed by consumer",
		},
	)

	MethodDurationMilliSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "lunettes_method_duration_milliseconds",
			Help: "how long an method operation to completed",
		},
		[]string{"phase"},
	)

	// PodDeleteResult delete
	PodDeleteResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_delete_result_count",
			Help: "Pod delete result, succeed or failed or reason",
		},
		[]string{"cluster", "namespace", "node_ip", "result"},
	)

	PodDeleteResultInDay = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_delete_result_count_day",
			Help: "Pod delete result, succeed or failed or reason",
		},
		[]string{"cluster", "namespace", "result"},
	)

	PodDeleteResultInWeek = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_delete_result_count_week",
			Help: "Pod delete result, succeed or failed or reason",
		},
		[]string{"cluster", "namespace", "result"},
	)

	PodDeleteApiCode = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_delete_api_code",
			Help: "Pod delete api response code",
		},
		[]string{"cluster", "namespace", "apicode"},
	)
	PodDeleteLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "slo_pod_delete_latency",
			Help:    "pod delete latency",
			Buckets: []float64{0.5, 1, 2, 4, 6, 10, 20, 30, 45, 60, 120, 180, 240, 300, 360, 480, 600, 900, 1200, 1800, 2700, 3600, 7200, 14400, 43200, 86400, 86400 * 2, 86400 * 3, 86400 * 4, 86400 * 5, 86400 * 6, 86400 * 7},
		},
		[]string{"cluster", "namespace", "phase"},
	)

	PodDeleteLatencyQuantiles = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "slo_pod_delete_latency_quantiles_in_seconds",
			Help:       "pod delete latency in seconds with quantiles",
			MaxAge:     time.Hour,
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"pod_type"},
	)

	//PodUpgradeResultCounter update
	PodUpgradeResultCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "slo_pod_upgrade_result_count",
			Help: "",
		},
		[]string{"cluster", "namespace", "node_ip", "result"},
	)

	// SloAnalysisResultGauge slo analysis result of the configmap
	SloAnalysisResultGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_analysis_result",
			Help: "slo analysis result of the configmap.",
		},
		[]string{"is_custom", "result", "description", "type"},
	)
)

func init() {
	// create
	prometheus.MustRegister(PodStartupLatencyExcludingShceduling)
	prometheus.MustRegister(PodStartupLatency)
	prometheus.MustRegister(PodStartupSLOLatency)
	prometheus.MustRegister(PodStartupK8sSLOLatency)
	prometheus.MustRegister(PodStartupResultExcludingScheduling)
	prometheus.MustRegister(PodStartupResult)
	prometheus.MustRegister(PodStartupSLOResult)
	prometheus.MustRegister(PodStartupK8sSLOResult)
	prometheus.MustRegister(PodCreateTotal)
	prometheus.MustRegister(PodCreateAPIResult)
	prometheus.MustRegister(EventConsumedCount)
	prometheus.MustRegister(MethodDurationMilliSeconds)

	//delete
	prometheus.MustRegister(PodDeleteResult)
	prometheus.MustRegister(PodDeleteResultInDay)
	prometheus.MustRegister(PodDeleteResultInWeek)
	prometheus.MustRegister(PodDeleteApiCode)
	prometheus.MustRegister(PodDeleteLatency)
	prometheus.MustRegister(PodDeleteLatencyQuantiles)

	// update
	prometheus.MustRegister(PodUpgradeResultCounter)

	//slo analysis result of the configmap
	prometheus.MustRegister(SloAnalysisResultGauge)

	clearSLOMetric()
}

// clear the metric data of request resource info periodically
func clearSLOMetric() {
	go func() {
		for {
			now := time.Now()
			next := now.Add(72 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())

			t := time.NewTicker(next.Sub(now))
			<-t.C

			//create
			PodStartupLatencyExcludingShceduling.Reset()
			PodStartupLatency.Reset()
			PodStartupSLOLatency.Reset()
			PodStartupResultExcludingScheduling.Reset()
			PodStartupResult.Reset()
			PodStartupSLOResult.Reset()
			PodCreateTotal.Reset()
			PodCreateAPIResult.Reset()
			MethodDurationMilliSeconds.Reset()
			//delete
			PodDeleteResult.Reset()
			PodDeleteResultInDay.Reset()
			PodDeleteResultInWeek.Reset()
			PodDeleteApiCode.Reset()
			PodDeleteLatency.Reset()
			PodDeleteLatencyQuantiles.Reset()
			//update
			PodUpgradeResultCounter.Reset()
		}
	}()
}
