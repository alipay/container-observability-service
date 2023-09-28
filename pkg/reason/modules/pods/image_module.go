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
	"k8s.io/klog"
)

func init() {
	modules.ShareModuleFactory.Register(share.IMAGE, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.IMAGE, ImageReason)
	})
}

func ImageReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_image ")

	//init容器
	klog.V(8).Infof("analyzeContainer init for %s\n", pod.Name)
	failedPull := make(map[string]string)
	for _, hyEvent := range auditEvents {
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

		if reason == "Pulled" {
			imageName := utils.GetImageName(msg)
			if imageName != nil {
				if _, ok := failedPull[*imageName]; ok {
					delete(failedPull, *imageName)
				}
			}
		} else if reason == "Failed" && strings.Contains(msg, "Failed to pull image") {
			imageName := utils.GetImageName(msg)
			if imageName != nil {
				if strings.Contains(msg, "not found") {
					failedPull[*imageName] = "ImageNotFound"
				}
			}
		} else if reason == "InspectFailed" && strings.Contains(msg, "Failed to inspect image") {
			imageName := utils.GetImageName(msg)
			if imageName != nil {
				failedPull[*imageName] = "InspectImageFailed"
			}
		} else if reason == "BackOff" && strings.Contains(msg, "Back-off pulling image") {
			imageName := utils.GetImageName(msg)
			if imageName != nil {
				failedPull[*imageName] = "ImagePullBackOff"
			}
		}

		if strings.Contains(msg, "pull access denied") {
			failedPull["image"] = "ImagePullAccessDenied"
		}
	}

	if len(failedPull) > 0 {
		for _, v := range failedPull {
			if len(v) > 0 {
				return v, true
			}
		}
		return "FailedPullImage", true
	}
	return "", false
}
