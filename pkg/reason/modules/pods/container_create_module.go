package pods

import (
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	utils2 "github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

func init() {
	modules.ShareModuleFactory.Register(share.CONTAINER_CREATE, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.CONTAINER_CREATE, ContainerCreateReason)
	})
}

func ContainerCreateReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_container_start ")

	//init容器
	for _, c := range pod.Spec.InitContainers {
		rs, hasError := analysisContainerCreate(&c, auditEvents, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}

	for _, c := range pod.Spec.Containers {
		rs, hasError := analysisContainerCreate(&c, auditEvents, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}
	return "", false
}

func analysisContainerCreate(c *v1.Container, auditEvents []*shares.AuditEvent, endTime *time.Time) (string, bool) {
	eventLen := len(auditEvents)
	for idx := eventLen - 1; idx >= 0; idx-- {
		hyEvent := auditEvents[idx]
		if endTime != nil && hyEvent.Event.StageTimestamp.After(*endTime) {
			continue
		}

		reason := ""
		if hyEvent.Reason != "" {
			reason = hyEvent.Reason
		}

		msg := ""
		ev, ok := hyEvent.ResponseRuntimeObj.(*v1.Event)
		if ok && ev != nil {
			msg = ev.Message
		}

		if reason == "Created" {
			startInfo := utils.GetCreatedAndStartedContainerName(msg)
			if startInfo != nil {
				startedContainer := startInfo[0]
				if c.Name == startedContainer {
					return "", false
				}
			}
		}

		//FailedCreateContainer
		//msg contains "create container($containerName) on containerd" indicate during the start container stage
		if reason == "Failed" && strings.Contains(msg, "failed to create container") && !strings.Contains(msg, "start container") {
			return "CreateContainerError", true
		}

		reqObj, reqOK := hyEvent.RequestRuntimeObj.(*v1.Pod)
		if reqOK {
			for _, cs := range reqObj.Status.ContainerStatuses {
				if c.Name == cs.Name && cs.State.Waiting != nil && cs.State.Waiting.Reason == "CreateContainerError" {
					return "CreateContainerError", true
				}
			}

		}
	}

	return "", false
}
