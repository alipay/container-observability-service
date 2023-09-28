package extractor

import (
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodUpdateProcessor struct {
}

func (p *PodUpdateProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ResponseStatus.Code >= 300 || event.Verb != "update" {
		return false
	}

	if event.RequestRuntimeObj == nil || event.ResponseRuntimeObj == nil {
		return false
	}

	return true
}

// 处理 Pod 的 Patch
func (p *PodUpdateProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodUpdate ")

	//response pod
	_, ok := event.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		klog.Warningf("can not convert event from response runtime obj")
		return nil
	}

	event.Type = shares.AuditTypeOperation

	return nil
}
