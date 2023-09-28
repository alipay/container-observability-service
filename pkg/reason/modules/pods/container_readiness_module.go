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
	modules.ShareModuleFactory.Register(share.CONTAINER_READINESS, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.CONTAINER_READINESS, ContainerReadinessReason)
	})
}

func ContainerReadinessReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_container_readiness ")

	isNotReady := false

	//init容器
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.Ready == false {
			isNotReady = true
			break
		}
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready == false {
			isNotReady = true
			break
		}
	}

	//condition
	if !isNotReady {
		for _, cd := range pod.Status.Conditions {
			if cd.Type == v1.ContainersReady {
				if cd.Status == v1.ConditionFalse {
					isNotReady = true
				} else {
					if beginTime != nil && cd.LastTransitionTime.Time.Before(*beginTime) {
						isNotReady = true
					}
				}
			}
		}
	}

	//container not ready，但是没有错误报出
	if isNotReady {
		//container health check
		rs, hasError := analysisContainerReadiness(auditEvents, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}

		return "ContainerNotReady", false
	}
	return "", false
}

func analysisContainerReadiness(auditEvents []*shares.AuditEvent, endTime *time.Time) (string, bool) {
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

		readinessError := "readiness probe errored"
		readinessFailed := "readiness probe failed"

		if reason == "Unhealthy" && (strings.Contains(strings.ToLower(msg), readinessFailed) || strings.Contains(strings.ToLower(msg), readinessError)) {
			return "ContainerReadinessFailed", true
		}
	}

	return "", false
}
