/*
*
创建Pod动作
*/
package podphase

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/kube"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	"k8s.io/klog/v2"
)

// 用于在 tracing API 中添加一个 apicreate 的 phase
func processPodCreation(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodCreation")

	pod := auditEvent.TryGetPodFromEvent()
	if pod == nil {
		return
	}

	if auditEvent.ResponseRuntimeObj == nil {
		return
	}

	/*if err := json.Unmarshal(auditEvent.ResponseObject.Raw, responsePod); err != nil {
		return
	}*/

	clusterName := auditEvent.Annotations["cluster"]

	//
	extraInfo := make(map[string]interface{})
	extraInfo["UserAgent"] = auditEvent.UserAgent
	if pod.Spec.NodeName != "" {
		nodeStatus := kube.GetNodeSummary(pod.Spec.NodeName)
		if nodeStatus != "" {
			extraInfo["NodeStatus"] = nodeStatus
		}
		extraInfo["NodeInfo"] = kube.GetNodeInfo(pod.Spec.NodeName)
	}

	klog.Infof("processing pod creation for id %s", pod.Name)

	if auditEvent.ResponseStatus.Code == 201 {
		//创建成功
		_ = xsearch.SavePodLifePhase(clusterName, pod.Namespace, string(pod.UID), pod.Name, "apicreate",
			false, auditEvent.RequestReceivedTimestamp.Time, auditEvent.RequestReceivedTimestamp.Time,
			extraInfo, string(auditEvent.AuditID))

		// 进入 pod scheduler 阶段
		_ = xsearch.SavePodLifePhase(clusterName, pod.Namespace, string(pod.UID), pod.Name, "Enters "+pod.Spec.SchedulerName,
			false, auditEvent.RequestReceivedTimestamp.Time, auditEvent.RequestReceivedTimestamp.Time,
			extraInfo, string(auditEvent.AuditID))

	} else {
		//失败
		extraInfo["auditEvent.ResponseObject"] = auditEvent.ResponseObject
		_ = xsearch.SavePodLifePhase(clusterName, pod.Namespace, string(pod.UID), pod.Name, "apicreate",
			true, auditEvent.RequestReceivedTimestamp.Time, auditEvent.RequestReceivedTimestamp.Time,
			extraInfo, string(auditEvent.AuditID))
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("phase_create").Observe(traceProcessingTime)
}
