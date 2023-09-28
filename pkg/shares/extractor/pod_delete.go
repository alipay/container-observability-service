/*
*
删除Pod动作
*/
package extractor

import (
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type PodDeleteProcessor struct {
}

func (p *PodDeleteProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ObjectRef.Subresource != "" || event.Verb != "delete" {
		return false
	}

	if event.ResponseRuntimeObj == nil {
		return false
	}

	return true
}

func (p *PodDeleteProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodDeletion ")

	responsePod, ok := event.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		klog.Warningf("can not convert pod from request runtime obj")
		return nil
	}

	event.Type = shares.AuditTypeOperation
	if event.ResponseStatus.Code < 300 {
		//创建成功
		deleteIm := true
		if responsePod.DeletionGracePeriodSeconds != nil && *responsePod.DeletionGracePeriodSeconds != 0 {
			deleteIm = false
		}
		if len(responsePod.Finalizers) != 0 {
			deleteIm = false
		}

		if deleteIm {
			event.Operation["pod:deleted:success"] = []string{}
		} else {
			event.Operation["pod:delete:success"] = []string{}
		}
	} else {
		//失败
		event.Operation["pod:delete:failed"] = []string{}
	}

	return nil
}
