package metrics

import (
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	spansPrefix = MetricNamePrefix
)

var (
	SpansProcessedPods = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: spansPrefix + "processed_pods_count",
			Help: "how many pods have been processed by spans module",
		},
		[]string{},
	)

	SpansInMemPodsCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: spansPrefix + "in_mem_pods_count",
			Help: "how many pods currently in spans module",
		},
		[]string{},
	)

	/*SpansConsumingResource = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: spansPrefix + "span_consuming_millisecond",
			Help: "Time consuming of span.",
		},
		[]string{"cluster", "namespace", "resource", "type"},
	)*/

	BaseLabels = map[string]bool{
		"cluster": true, "namespace": true, "resource": true, "type": true, "action_type": true,
	}
	SpanConsumingLabels     = map[string]bool{}
	SpansConsumingStatistic = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    spansPrefix + "span_consuming_millisecond_statistic",
			Help:    "Time consuming statistic of span.",
			Buckets: []float64{},
		},
		[]string{"cluster", "namespace", "resource", "type", "action_type"},
	)
)

// clear the metric data of request resource info periodically
func ClearRequestResourceMetric() {

	stopChan := make(chan struct{})
	go wait.Until(func() {
		if SpansConsumingStatistic != nil {
			SpansConsumingStatistic.Reset()
		}
	}, 6*time.Hour, stopChan)
}

// redefine SpansConsumingStatistic if extra-properties chanes
func DefineSpanStatistic(labels []string) {
	//add base label
	newSpanConsumingLabels := map[string]bool{}
	for k, v := range BaseLabels {
		newSpanConsumingLabels[k] = v
	}
	//add other labels
	for idx, _ := range labels {
		newSpanConsumingLabels[labels[idx]] = true
	}

	// define the metrics
	newLabels := make([]string, 0)
	for k, _ := range newSpanConsumingLabels {
		label := strings.TrimSpace(k)
		if len(label) > 0 {
			newLabels = append(newLabels, label)
			SpanConsumingLabels[label] = true
		}
	}

	bucketMap := map[int64]int64{
		60000:  6000,
		600000: 10000,
	}

	buckets := []float64{100, 1100, 2100, 3100, 4100, 5100, 6100, 7100, 8100, 9100, 10000}
	ct := int64(11000)
	for target, step := range bucketMap {
		for ; ct < target; ct += step {
			buckets = append(buckets, float64(ct))
		}
	}

	SpansConsumingStatistic = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    spansPrefix + "span_consuming_millisecond_statistic",
			Help:    "Time consuming statistic of span.",
			Buckets: buckets,
		},
		newLabels,
	)
	prometheus.MustRegister(SpansConsumingStatistic)
	clearSpanMetric()
}

// clear the metric data of request resource info periodically
func clearSpanMetric() {
	go func() {
		for {
			now := time.Now()
			next := now.Add(1 * time.Hour)
			next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), 0, 0, 0, next.Location())

			t := time.NewTicker(next.Sub(now))
			<-t.C

			SpansConsumingStatistic.Reset()
		}
	}()
}
