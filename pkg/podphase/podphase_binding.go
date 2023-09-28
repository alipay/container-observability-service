package podphase

import (
	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	v1 "k8s.io/api/core/v1"
)

func processPodBinding(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodBinding")

	clusterName := auditEvent.Annotations["cluster"]
	extraInfo := make(map[string]interface{})
	extraInfo["requestObject"] = auditEvent.RequestObject
	extraInfo["UserAgent"] = auditEvent.UserAgent

	if auditEvent.ResponseStatus.Code < 300 {
		//创建成功
		//bindObject := &v1.Binding{}
		//err := json.Unmarshal(auditEvent.RequestObject.Raw, bindObject)
		if auditEvent.RequestRuntimeObj != nil {
			bindObject, ok := auditEvent.RequestRuntimeObj.(*v1.Binding)
			if ok {
				extraInfo["target"] = bindObject.Target
			}
		}

		_ = xsearch.SavePodLifePhase(clusterName, auditEvent.ObjectRef.Namespace, string(auditEvent.ObjectRef.UID),
			auditEvent.ObjectRef.Name, "scheduled", false, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	} else {
		//失败
		extraInfo["auditEvent.ResponseObject"] = auditEvent.ResponseObject
		_ = xsearch.SavePodLifePhase(clusterName, auditEvent.ObjectRef.Namespace, string(auditEvent.ObjectRef.UID),
			auditEvent.ObjectRef.Name, "scheduled", true, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}
}
