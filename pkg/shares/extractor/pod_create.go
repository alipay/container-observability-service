/*
*
创建Pod动作
*/
package extractor

import (
	"fmt"

	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodCreateProcessor struct {
}

func (p *PodCreateProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ObjectRef.Subresource != "" || event.Verb != "create" {
		return false
	}

	if event.RequestRuntimeObj == nil {
		return false
	}

	return true
}

// 用于在 tracing API 中添加一个 apicreate 的 phase
func (p *PodCreateProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodCreation")

	resPod, ok := event.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		klog.Warningf("can not convert pod from response runtime obj")
		return nil
	}

	event.Type = shares.AuditTypeOperation
	if event.ResponseStatus.Code == 201 {
		//创建成功
		event.Operation["pod:create:success"] = []string{}
	} else {
		//失败
		event.Operation["pod:create:failed"] = []string{}
	}

	// entry cloud scheduler
	if resPod.Spec.SchedulerName != "" {
		event.Operation[fmt.Sprintf("schedule:%s:entry", resPod.Spec.SchedulerName)] = []string{}
	}

	return nil
}
