package pods

import (
	"regexp"
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
	modules.ShareModuleFactory.Register(share.NETWORK, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.NETWORK, NetworkReason)
	})
}

func NetworkReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	podYaml := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if podYaml == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_network ")
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

			//网络问题: ip分配超时
			if strings.Contains(msg, "timeout to allocate ip for pod") {
				return "AllocateIPTimeout", true
			}

			if strings.Contains(msg, "failed to setup network for sandbox") {
				//网络问题: mac的nic分配错误
				if strings.Contains(msg, "Can not find host nic by mac address") || strings.Contains(msg, "no nic found") {
					return "NotFoundNicByMac", true
				}
				//网络问题: mac的nic分配错误
				if strings.Contains(msg, "fail to allocate ip") {
					return "FailedAllocateIP", true
				}

				return "FailedSetNetwork", true
			}

			re := regexp.MustCompile("failed to set sloup sandbox container \"(.+)\" network for pod")
			match := re.FindStringSubmatch(msg)
			if match != nil && len(match) >= 1 {
				//网络问题：宿主机中网桥端口满
				if strings.Contains(msg, "exchange full") {
					return "BridgeExchangeFull", true
				}
				return "FailedSetNetwork", true
			}
		}
	}
	return "", false
}
