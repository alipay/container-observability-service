package pods

import (
	"fmt"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

func init() {
	modules.ShareModuleFactory.Register(share.SCHEDULER, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.SCHEDULER, SchedulerReason)
	})
}

func SchedulerReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	podYaml := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if podYaml == nil {
		return "", false
	}

	schedulerCode, _ := getScheduleStatus(podYaml)
	if schedulerCode == -1 || schedulerCode == 0 { //未处理调度
		klog.V(8).Infof("no schedule: %s\n", podYaml.Name)
		//没有调度，则进行错误扫描
		if podYaml.Spec.SchedulerName == "default-scheduler" {
			result = "ScheduleDelay"
			hasError = true

			for _, cond := range podYaml.Status.Conditions {
				if cond.Type == v1.PodScheduled && cond.Status == v1.ConditionFalse {
					result = "FailedScheduling"
					if strings.Contains(cond.Message, "error getting PVC") {
						result = "FailedSchedulingFindPVC"
					}
					if strings.Contains(cond.Message, "quota not enough") {
						resource := utils.ExtractQuotaResource(cond.Message)
						if len(result) > 0 {
							resource = utils.FirstUpper(resource)
							result = fmt.Sprintf("%sQuotaNotEnough", resource)
						} else {
							result = "QuotaNotEnough"
						}
					}
					break
				}
			}
		}

		return result, hasError
	}
	return "", false
}

// getScheduleStatus, 1:已调度；0：调度失败; -1: 未处理
func getScheduleStatus(pod *v1.Pod) (int, time.Time) {
	var t time.Time
	if pod == nil {
		return -1, t
	}
	if pod.Spec.NodeName != "" {
		return 1, t
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionTrue {
			return 1, condition.LastTransitionTime.Time
		} else if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse &&
			condition.Reason == v1.PodReasonUnschedulable {
			return 0, condition.LastTransitionTime.Time
		}
	}
	return -1, t
}
