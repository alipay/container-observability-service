package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	TraceProcessingLatencyBuckets = []float64{0.5, 1, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 30, 45, 60, 120, 180, 240, 300, 360,
		480, 600, 900, 1200, 1800, 2700, 3600, 7200, 14400, 43200, 86400}

	// 每个 trace 交付的时间延迟
	TraceProcessingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "trace_processing_latency_seconds",
			Help:    "time used to process one trace",
			Buckets: TraceProcessingLatencyBuckets,
		},

		// trace type: 类型
		[]string{"type"},
	)
)

func init() {
	prometheus.MustRegister(TraceProcessingLatency)
}
