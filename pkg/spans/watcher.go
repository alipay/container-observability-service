package spans

import (
	"os"
	"time"

	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/klog/v2"
)

var (
	WatcherQueue          *queue.BoundedQueue
	DeliverySpanProcessor *SpanProcessor
)

// actually cluster is not necessary
func InitKubeSpanWatcher(cluster string, otlpAddr string) error {
	prometheus.MustRegister(metrics.SpansProcessedPods)
	prometheus.MustRegister(metrics.SpansInMemPodsCount)
	metrics.ClearRequestResourceMetric()

	var err error
	//spanExporter, err = setupOTLP(ctx, otlpAddr, "", false)
	if featuregates.IsEnabled(JaegerFeature) {
		err = initOtlpProcessor(otlpAddr)
		if err != nil {
			klog.Error(err, "unable to set up tracing")
			os.Exit(1)
		}
	}

	DeliverySpanProcessor = NewSpanProcessor(cluster)

	go DeliverySpanProcessor.Compact()
	WatcherQueue = queue.NewBoundedQueue("spans-watcher", 200000, nil)
	WatcherQueue.StartLengthReporting(10 * time.Second)
	WatcherQueue.IsDropEventOnFull = false
	go WatcherQueue.StartConsumers(1, func(item interface{}) {
		if item == nil {
			return
		}

		auditEvent, ok := item.(*shares.AuditEvent)
		if !ok || auditEvent == nil {
			return
		}

		auditEvent.CanProcess(shares.SpanProcessNode)
		DeliverySpanProcessor.ProcessEvent(auditEvent)
	})
	return nil
}
