package pods

import (
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	utils2 "github.com/alipay/container-observability-service/pkg/utils"
)

func init() {
	modules.ShareModuleFactory.Register(share.KUBELET_DELAY, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.KUBELET_DELAY, KubeletDelayReason)
	})
}

func KubeletDelayReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	podYaml := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if podYaml == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_kubelet_delay ")

	eventLen := len(auditEvents)
	for idx := eventLen - 1; idx >= 0; idx-- {
		hyEvent := auditEvents[idx]
		if endTime != nil && hyEvent.Event.StageTimestamp.After(*endTime) {
			continue
		}

		if strings.Contains(strings.ToLower(hyEvent.UserAgent), "kubelet") && strings.Contains(strings.ToLower(hyEvent.UserAgent), "kubernetes") {
			return "", false
		}
	}
	return "KubeletDelay", true
}
