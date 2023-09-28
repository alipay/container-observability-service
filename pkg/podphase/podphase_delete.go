/*
*
删除Pod动作
*/
package podphase

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/kube"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	v1 "k8s.io/api/core/v1"
)

func processPodDeletion(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodDeletion")

	/*responsePod := &v1.Pod{}
	if err := json.Unmarshal(auditEvent.ResponseObject.Raw, responsePod); err != nil {
		return
	}*/

	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	responsePod, ok := auditEvent.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	extraInfo := make(map[string]interface{})
	extraInfo["UserAgent"] = auditEvent.UserAgent
	if responsePod.Spec.NodeName != "" {
		nodeStatus := kube.GetNodeSummary(responsePod.Spec.NodeName)
		if nodeStatus != "" {
			extraInfo["NodeStatus"] = nodeStatus
		}
		extraInfo["NodeInfo"] = kube.GetNodeInfo(responsePod.Spec.NodeName)
	}

	if auditEvent.ResponseStatus.Code < 300 { //删除成功
		deleteIm := true
		if responsePod.DeletionGracePeriodSeconds != nil && *responsePod.DeletionGracePeriodSeconds != 0 {
			deleteIm = false
		}
		if len(responsePod.Finalizers) != 0 {
			deleteIm = false
		}
		if deleteIm {
			_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID), responsePod.Name,
				"apideleted", false, auditEvent.RequestReceivedTimestamp.Time, auditEvent.RequestReceivedTimestamp.Time,
				extraInfo, string(auditEvent.AuditID))
		} else {
			_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID), responsePod.Name,
				"apidelete", false, auditEvent.RequestReceivedTimestamp.Time, auditEvent.RequestReceivedTimestamp.Time,
				extraInfo, string(auditEvent.AuditID))
		}

	} else { //失败
		extraInfo["auditEvent.ResponseObject"] = auditEvent.ResponseObject
		_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID), responsePod.Name,
			"apidelete", true, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("delete").Observe(traceProcessingTime)
}
