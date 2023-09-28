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
	modules.ShareModuleFactory.Register(share.CONTAINER_POST_START, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.CONTAINER_POST_START, ContainerPostHookReason)
	})
}

func ContainerPostHookReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_post_hook ")

	//init容器
	for _, c := range pod.Spec.InitContainers {
		cs := utils.GetContainerStatus(c.Name, pod)
		if cs != nil && cs.Ready == true {
			continue
		}

		rs, hasError := analysisPostHookStart(&c, cs, auditEvents, beginTime, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}

	for _, c := range pod.Spec.Containers {
		cs := utils.GetContainerStatus(c.Name, pod)
		if cs != nil && cs.Ready == true {
			continue
		}

		rs, hasError := analysisPostHookStart(&c, cs, auditEvents, beginTime, endTime)
		if hasError && rs != "" {
			return rs, hasError
		}
	}
	return "", false
}

func analysisPostHookStart(c *v1.Container, cs *v1.ContainerStatus, auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (string, bool) {
	hasPostStartHook := false
	if c.Lifecycle != nil && c.Lifecycle.PostStart != nil {
		hasPostStartHook = true
	}

	if !hasPostStartHook {
		return "", false
	}

	if cs != nil && cs.State.Waiting != nil && strings.Contains(cs.State.Waiting.Reason, "PostStartHookError") {
		return "FailedPostStartHook", false
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

		if reason == "WithOutPostStartHook" && strings.Contains(msg, c.Name) {
			return "", false
		}

		if reason == "SucceedPostStartHook" && strings.Contains(msg, c.Name) {
			return "", false
		}

		if reason == "FailedPostStartHook" && strings.Contains(msg, c.Name) {
			return "FailedPostStartHook", false
		}
	}

	return "", false
}
