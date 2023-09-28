package trace

import (
	"os"
	"time"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/shares"
	"k8s.io/klog"
)

var (
	WatcherQueue *queue.BoundedQueue
)

func InitKubeTraceWatcher(cluster string, otlpAddr string) error {
	var err error
	//spanExporter, err = setupOTLP(ctx, otlpAddr, "", false)
	err = initOtlpProcessor(otlpAddr)
	if err != nil {
		klog.Error(err, "unable to set up tracing")
		os.Exit(1)
	}

	processor := NewSpanProcessor(cluster)

	go processor.Compact()
	WatcherQueue = queue.NewBoundedQueue("trace-watcher", 200000, nil)
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

		auditEvent.Wait()
		processor.ProcessEvent(auditEvent)
	})
	return nil
}
