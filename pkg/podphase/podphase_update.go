package podphase

import (
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
)

// 处理 Pod 的 Patch
func processPodUpdate(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodUpdate ")

	// response pod
	responsePod := auditEvent.TryGetPodFromEvent()
	if responsePod == nil {
		return
	}
}
