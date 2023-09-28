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
	modules.ShareModuleFactory.Register(share.CONTAINER_START, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.CONTAINER_START, ContainerStartReason)
	})
}

func ContainerStartReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_container_start ")

	//init容器
	for _, c := range pod.Spec.InitContainers {
		cs := utils.GetContainerStatus(c.Name, pod)
		if cs != nil && cs.Ready == true {
			continue
		}

		rs, hasError := analysisContainerStart(&c, cs, auditEvents, beginTime, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}

	for _, c := range pod.Spec.Containers {
		cs := utils.GetContainerStatus(c.Name, pod)
		if cs != nil && cs.Ready == true {
			continue
		}

		rs, hasError := analysisContainerStart(&c, cs, auditEvents, beginTime, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}
	return "", false
}

func analysisContainerStart(c *v1.Container, cs *v1.ContainerStatus, auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (string, bool) {

	// status中的更新错误
	if cs != nil && (*cs).State.Waiting != nil && strings.Contains((*cs).State.Waiting.Reason, "CrashLoopBackOff") {
		return "CrashLoopBackOff", true
	}
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

		if cs != nil && cs.Started != nil && *cs.Started == true && (*cs).State.Running.StartedAt.After(*beginTime) {
			return "", false
		}

		if reason == "Started" {
			startInfo := utils.GetCreatedAndStartedContainerName(msg)
			if startInfo != nil {
				startedContainer := startInfo[0]
				if c.Name == startedContainer {
					return "", false
				}
			}
		}

		//容器启动crash backoff
		if reason == "BackOff" && (strings.Contains(msg, "Back-off restarting") || strings.Contains(msg, "Back-off failed container")) {
			return "CrashLoopBackOff", true
		}

		//判断event中的信息
		if strings.EqualFold(reason, "Failed") && strings.Contains(msg, c.Name) {
			if strings.Contains(msg, "failed to start container") || strings.Contains(msg, "Error") {
				return "RunContainerError", true
			}
		}
	}

	//判断启动异常退出
	if cs != nil && (*cs).State.Terminated != nil && strings.Contains((*cs).State.Terminated.Reason, "Error") {
		return "RunContainerError", true
	}

	return "", false
}
