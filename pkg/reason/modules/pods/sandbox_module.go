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
	modules.ShareModuleFactory.Register(share.SANDBOX, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.SANDBOX, SandboxReason)
	})
}

func SandboxReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	podYaml := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if podYaml == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_sandbox ")

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

		//v1.16版本
		if reason == "SuccessfulCreatePodSandBox" {
			return "", false
		}
		//1.14版本没有明确信息，可以依赖是否开始镜像操作
		if reason == "Pulling" || reason == "Pulled" {
			return "", false
		}

		if reason == "FailedCreatePodSandBox" {
			msg := ""
			ev, ok := hyEvent.ResponseRuntimeObj.(*v1.Event)
			if ok && ev != nil {
				msg = ev.Message
			}

			//创建sandbox cri rpc超时问题
			if strings.Contains(msg, "context deadline exceeded") {
				return "CreatePodSandBoxTimeout", true
			}
			return "FailedCreatePodSandBox", true
		}
	}

	return "", false
}
