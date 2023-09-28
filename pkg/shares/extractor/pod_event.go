package extractor

import (
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodEventProcessor struct {
}

func (p *PodEventProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "events" || event.ResponseStatus.Code >= 300 {
		return false
	}

	if event.ResponseRuntimeObj == nil {
		return false
	}

	return true
}

// 用于在 tracing API 中添加一个 apicreate 的 phase
func (p *PodEventProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodCreation")

	e, ok := event.ResponseRuntimeObj.(*v1.Event)
	if !ok {
		klog.Warningf("can not convert event from request runtime obj")
		return nil
	}

	if e.InvolvedObject.Kind != "Pod" {
		return nil
	}

	event.Type = shares.AuditTypeEvent
	event.Reason = e.Reason

	return nil
}
