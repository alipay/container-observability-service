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
	modules.ShareModuleFactory.Register(share.ADMISSION, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.ADMISSION, AdmissionReason)
	})
}

func AdmissionReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_admission ")

	for _, hyEvent := range auditEvents {
		//v1.16版本
		reason := ""
		if hyEvent.Reason != "" {
			reason = hyEvent.Reason
		}

		msg := ""
		ev, ok := hyEvent.ResponseRuntimeObj.(*v1.Event)
		if ok && ev != nil {
			msg = ev.Message
		}

		//v1.16版本
		if reason == "SuccessfulCreatePodSandBox" {
			return "", false
		}
		//1.14版本没有明确信息，可以依赖是否开始镜像操作
		if reason == "Pulling" || reason == "Pulled" {
			return "", false
		}

		if reason == "UnexpectedAdmissionError" {
			//FileSystemReadOnly
			if strings.Contains(msg, "read-only file system, which is unexpected") {
				return "FileSystemReadOnly", true
			}
			//no device
			if strings.Contains(msg, "devices unavailable for nvidia.com") {
				return "DeviceUnavailable", true
			}
			//NoDiskSpace
			if strings.Contains(msg, "no space left on device") {
				return "NoDiskSpace", true
			}
		}

		if reason == "Evicted" {
			//evict for condition
			if strings.Contains(msg, "The node had condition") {
				conditon := utils.GetConditionName(msg)
				if conditon != nil {
					return *conditon + "Evicted", true
				}
			}
		}

		if pod != nil && pod.Status.Reason == "Evicted" {
			//evict for condition
			if strings.Contains(pod.Status.Message, "The node had condition") {
				condition := utils.GetConditionName(pod.Status.Message)
				if condition != nil {
					return *condition + "Evicted", true
				}
			}
		}

	}
	return "", false
}
