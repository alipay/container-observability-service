package podphase

import (
	"regexp"
	"strings"

	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	v1 "k8s.io/api/core/v1"
)

func processPodEventCreation(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodEventCreation")

	if auditEvent == nil || auditEvent.RequestObject == nil || auditEvent.ResponseStatus.Code >= 300 {
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
	event, ok := auditEvent.RequestRuntimeObj.(*v1.Event)
	if !ok {
		return
	}

	if event.InvolvedObject.Kind != "Pod" {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	event.ClusterName = clusterName

	extraInfo := make(map[string]interface{})
	extraInfo["eventObject"] = event
	extraInfo["reason"] = event.Reason
	extraInfo["UserAgent"] = auditEvent.UserAgent

	// victims are: [xxxxx,xxxxx]
	if event.Reason == "PreemptionSuccess" {
		victims := extractVictims(event.Message)
		if len(victims) > 0 {
			extraInfo["victims"] = victims
		}
	}

	opName := "event"
	_ = xsearch.SavePodLifePhase(clusterName, event.InvolvedObject.Namespace, string(event.InvolvedObject.UID),
		event.InvolvedObject.Name, opName, isErrorEvent(event), auditEvent.RequestReceivedTimestamp.Time,
		auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
}

func isErrorEvent(event *v1.Event) bool {
	if event.Type == "Normal" {
		return false
	}

	return true
}

// extractVictims 提取抢占victims
func extractVictims(str string) []string {
	victims := make([]string, 0)
	re := regexp.MustCompile("victims are.*\\[(.*)\\]")
	match := re.FindStringSubmatch(str)
	if match != nil && len(match) >= 2 {
		for _, v := range strings.Split(match[1], ",") {
			trimed := strings.TrimSpace(v)
			if trimed == "" {
				continue
			}
			victims = append(victims, trimed)
		}
	}
	return victims
}
