package podphase

import (
	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	v1 "k8s.io/api/core/v1"
)

func processPodEventPatch(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodEventPatch")

	if auditEvent == nil || auditEvent.RequestObject == nil {
		return
	}

	//忽略失败的event
	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}

	/*event := &v1.Event{}
	if err := json.Unmarshal(auditEvent.RequestObject.Raw, event); err != nil {
		klog.Infof("Unmarshal failed: %s", auditEvent.AuditID)
		return
	}*/
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	event, ok := auditEvent.ResponseRuntimeObj.(*v1.Event)
	if !ok {
		return
	}

	if event.InvolvedObject.Kind != "Pod" {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	if clusterName == "" {
		clusterName = event.ClusterName
	}
	event.ClusterName = clusterName

	extraInfo := make(map[string]interface{})
	extraInfo["eventObject"] = event
	extraInfo["reason"] = event.Reason
	extraInfo["UserAgent"] = auditEvent.UserAgent

	_ = xsearch.SavePodLifePhase(clusterName, event.InvolvedObject.Namespace, string(event.InvolvedObject.UID),
		event.InvolvedObject.Name, "event", isErrorEvent(event), auditEvent.RequestReceivedTimestamp.Time,
		auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
}
