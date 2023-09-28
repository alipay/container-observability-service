package replayer

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/slo"
	"github.com/alipay/container-observability-service/pkg/utils"

	"github.com/alipay/container-observability-service/pkg/xsearch"
	corev1 "k8s.io/api/core/v1"
	k8s_audit "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog/v2"
)

type AuditEvent struct {
	Pod   *corev1.Pod
	Event *k8s_audit.Event
}

type AuditProcessor struct {
	// event input queue
	reader *logReader
	// TODO not used yet.
	cluster     string
	enableTrace bool
}

// NewAuditProcessor create new audit log processor
func NewAuditProcessor(
	esConf *xsearch.ElasticSearchConf,
	buffer, interval, traceTimeout time.Duration,
	cluster string, enableTrace bool) (*AuditProcessor, error) {
	auditProcessor := &AuditProcessor{
		cluster:     cluster,
		enableTrace: enableTrace,
	}
	slo.ClusterName = cluster
	xsearch.EsConfig = esConf

	reader, err := newLogReader(auditProcessor,
		esConf, buffer, interval,
		auditProcessor.cluster)

	auditProcessor.reader = reader
	return auditProcessor, err
}

// Start satrt processing traces
func (auditProcessor *AuditProcessor) Start(stopCh <-chan struct{}) {
	start := time.Now()
	defer func() {
		klog.Infof("time for processing: %f minutes", utils.TimeSinceInMinutes(start))
	}()
	auditProcessor.reader.Run(stopCh)
}

func (auditProcessor *AuditProcessor) Stop() {
	auditProcessor.reader.Stop()
}
