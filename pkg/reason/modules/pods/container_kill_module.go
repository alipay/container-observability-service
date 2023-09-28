package pods

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	utils2 "github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

func init() {
	modules.ShareModuleFactory.Register(share.CONTAINER_KILL, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.CONTAINER_KILL, ContainerKillReason)
	})
}

func ContainerKillReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_container_kill ")

	for _, c := range pod.Spec.Containers {
		rs, hasError := analysisContainerKill(&c, auditEvents, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}
	return "", false
}

func analysisContainerKill(c *v1.Container, auditEvents []*shares.AuditEvent, endTime *time.Time) (string, bool) {
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

		if reason == "SucceedKillingContainer" {
			killInfo := utils.GetKilledContainerName(msg)
			if killInfo != nil {
				killedContainer := killInfo[0]
				if c.Name == killedContainer {
					return "", false
				}
			}
		}
	}
	// todo: kill的更多情况
	return "", false
}
