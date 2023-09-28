package api_metrics

import (
	"strconv"
	"time"

	"github.com/alipay/container-observability-service/pkg/utils"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	responseStatusCode = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:      "api_response_status_code",
			Namespace: "lunettes",
			Help:      "api response status code counter",
		},
		[]string{"status_code"},
	)
	apiDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:      "api_latency",
			Namespace: "lunettes",
			Help:      "api latency status code counter",
			Buckets:   prometheus.ExponentialBuckets(50, 2, 14),
		},
		[]string{"path", "code"},
	)
)

func init() {
	prometheus.MustRegister(responseStatusCode)
	prometheus.MustRegister(apiDuration)
}

func IncResponseStatusCode(code int) {
	responseStatusCode.WithLabelValues(strconv.Itoa(code)).Inc()
}

func RecordAPIDuration(path string, code int, start time.Time) {
	apiDuration.WithLabelValues(path, strconv.Itoa(code)).Observe(utils.TimeSinceInMilliSeconds(start))
}
