package extractor

import (
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

type PodBindingProcessor struct {
}

func (p *PodBindingProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ObjectRef.Subresource != "binding" || event.ResponseStatus.Code >= 300 {
		return false
	}
	if event.RequestRuntimeObj == nil {
		return false
	}

	return true
}

func (p *PodBindingProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processBinding")
	//创建成功
	if _, ok := event.RequestRuntimeObj.(*v1.Binding); ok {
		event.Type = shares.AuditTypeOperation
		event.Operation["schedule:binding:success"] = []string{}
	}
	return nil
}
