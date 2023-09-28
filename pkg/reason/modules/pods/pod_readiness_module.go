package pods

import (
	"fmt"
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
	modules.ShareModuleFactory.Register(share.POD_READINESS, func() modules.DeliveryModule {
		return modules.NewDAGDeliveryModule(share.POD_READINESS, PodReadinessReason)
	})
}

func PodReadinessReason(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	pod := utils.GetPodYamlFromHyperEvents(auditEvents, endTime)
	if pod == nil {
		return "", false
	}
	defer utils2.IgnorePanic("analyze_pod_readiness ")

	conditionMap := make(map[string]string)
	rs := "PodNotReady"
	for _, ct := range pod.Status.Conditions {
		conditionMap[string(ct.Type)] = string(ct.Status)
	}

	if cr, ok := conditionMap[string(v1.ContainersReady)]; ok && cr != string(v1.ConditionTrue) {
		return "", false
	}

	//如果有readinessGates
	for _, rg := range pod.Spec.ReadinessGates {
		if _, ok := conditionMap[string(rg.ConditionType)]; !ok || strings.EqualFold(conditionMap[string(rg.ConditionType)], "false") {
			rs = fmt.Sprintf("%s_NotReady", string(rg.ConditionType))
		}
	}

	if rs != "PodNotReady" {
		return rs, true
	}
	return "", false
}
