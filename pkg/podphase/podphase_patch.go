package podphase

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/kube"

	"github.com/oliveagle/jsonpath"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// 处理 Pod 的 Patch
func processPodPatch(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodPatch")

	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}

	if auditEvent.ObjectRef.Subresource == "" {
		processPatchNoSubresource(auditEvent)
		return
	} else if auditEvent.ObjectRef.Subresource == "status" {
		processPatchSubResourceStatus(auditEvent)
		return
	}
}

// 处理非 Status 的字段
// 包括 finalizer，labels，annotation，spec.SchedulerName 等
func processPatchNoSubresource(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPatchNoSubresource ")

	// response pod
	/*responsePod, err := metas.GeneratePodFromEvent(auditEvent)
	if err != nil {
		return
	}*/
	var responsePod *v1.Pod
	if responsePod = auditEvent.TryGetPodFromEvent(); responsePod == nil {
		return
	}

	// request pod
	/*
		pod := v1.Pod{}
		if err := json.Unmarshal(auditEvent.RequestObject.Raw, &pod); err != nil {
			return
		}*/
	var pod *v1.Pod
	ok := false
	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	if pod, ok = auditEvent.RequestRuntimeObj.(*v1.Pod); !ok || pod == nil {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]

	//labels
	for key, value := range pod.Labels {
		opName := "set " + key + " to " + value
		extraInfo := make(map[string]interface{})
		extraInfo["user"] = auditEvent.User
		extraInfo["requestObject"] = auditEvent.RequestObject
		extraInfo["UserAgent"] = auditEvent.UserAgent
		_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
			responsePod.Name, opName, false, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}

	//识别容器升级
	if pod.Annotations["request-action-type"] == "RequestUpgrade" {
		extraInfo := make(map[string]interface{})
		extraInfo["user"] = auditEvent.User
		extraInfo["requestObject"] = auditEvent.RequestObject
		extraInfo["UserAgent"] = auditEvent.UserAgent
		if responsePod.Spec.NodeName != "" {
			nodeStatus := kube.GetNodeSummary(responsePod.Spec.NodeName)
			if nodeStatus != "" {
				extraInfo["NodeStatus"] = nodeStatus
			}
			extraInfo["NodeInfo"] = kube.GetNodeInfo(responsePod.Spec.NodeName)
		}
		var image string
		if len(pod.Spec.Containers) > 0 && pod.Spec.Containers[0].Image != "" {
			image = pod.Spec.Containers[0].Image
		}
		var opName string
		if image != "" {
			opName = "upgrade container " + " to " + image
		} else {
			opName = "upgrade container "
		}
		_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
			responsePod.Name, opName, false, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}

	// 同时更新 slo_pod_info index
	opName := fmt.Sprintf("Enters %s", pod.Spec.SchedulerName)
	extraInfo := make(map[string]interface{})
	extraInfo["user"] = auditEvent.User
	extraInfo["requestObject"] = auditEvent.RequestObject
	extraInfo["UserAgent"] = auditEvent.UserAgent

	_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
		responsePod.Name, opName, false, auditEvent.RequestReceivedTimestamp.Time,
		auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))

	_ = xsearch.SavePodInfoToZSearch(auditEvent.Annotations["cluster"], responsePod, "进行中", auditEvent.StageTimestamp.Time, "", "调度阶段", true)

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("patch").Observe(traceProcessingTime)
}

// 处理 Status 的字段的更新
func processPatchSubResourceStatus(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPatchSubResourceStatus")

	klog.V(6).Infof("this is patch for subresource status for pod _1: %s", auditEvent.ObjectRef.Name)
	ok := false
	var responsePod *v1.Pod
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	if responsePod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok || responsePod == nil {
		return
	}

	klog.V(6).Infof("this is patch for subresource status for pod _2: %s", auditEvent.ObjectRef.Name)
	var reqPod *v1.Pod
	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	if reqPod, ok = auditEvent.RequestRuntimeObj.(*v1.Pod); !ok || reqPod == nil {
		return
	}
	klog.V(6).Infof("this is patch for subresource status for pod _3: %s", auditEvent.ObjectRef.Name)
	clusterName := auditEvent.Annotations["cluster"]

	extraInfo := make(map[string]interface{})
	extraInfo["requestObject"] = auditEvent.RequestObject
	extraInfo["user"] = auditEvent.User
	extraInfo["hostIP"] = responsePod.Status.HostIP
	extraInfo["UserAgent"] = auditEvent.UserAgent

	//condition
	for _, cond := range reqPod.Status.Conditions {
		if string(cond.Type) == "" {
			continue
		}
		if cond.Status == v1.ConditionTrue {
			opName := "condition:" + string(cond.Type) + ":true"
			_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
				responsePod.Name, opName, false, auditEvent.RequestReceivedTimestamp.Time,
				auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
		} else if cond.Status == v1.ConditionFalse {
			//Pending状态下，ContainersReady、 Ready、Initialized 不记录
			if responsePod.Status.Phase == "Pending" {
				if cond.Type == "Ready" || cond.Type == "ContainersReady" || cond.Type == "Initialized" {
					continue
				}
			}
			tmp := make(map[string]interface{})
			for key, value := range extraInfo {
				tmp[key] = value
			}
			tmp["Message"] = cond.Reason + ": " + cond.Message
			opName := "condition:" + string(cond.Type) + ":false"
			_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
				responsePod.Name, opName, true, auditEvent.RequestReceivedTimestamp.Time,
				auditEvent.RequestReceivedTimestamp.Time, tmp, string(auditEvent.AuditID))
		}
	}

	//phase
	if string(reqPod.Status.Phase) != "" {

		opStr := "pod phase: " + string(responsePod.Status.Phase)
		_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
			responsePod.Name, opStr, false, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}

	//
	//容器状态
	for _, containerStatus := range reqPod.Status.ContainerStatuses {
		if containerStatus.State.Running != nil {
			if auditEvent.StageTimestamp.After(containerStatus.State.Running.StartedAt.Add(5 * time.Second)) {
				continue
			}
		}

		opName := containerStatus.Name + " restartCount:" + fmt.Sprintf("%d", containerStatus.RestartCount) + " "
		opName = opName + "Ready:" + fmt.Sprintf("%t", containerStatus.Ready) + " "
		opName = opName + formatContainerStatus(containerStatus)
		_ = xsearch.SavePodLifePhase(clusterName, responsePod.Namespace, string(responsePod.UID),
			responsePod.Name, opName, false, auditEvent.RequestReceivedTimestamp.Time,
			auditEvent.RequestReceivedTimestamp.Time, extraInfo, string(auditEvent.AuditID))
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("patchSubResource").Observe(traceProcessingTime)
}

func formatContainerStatus(con v1.ContainerStatus) string {
	result := ""
	if con.State.Running != nil {
		result = "Running"
	} else if con.State.Terminated != nil {
		if con.State.Terminated.Reason != "" {
			result = "Terminated:" + con.State.Terminated.Reason + " exitCode:" + fmt.Sprintf("%d", con.State.Terminated.ExitCode)
		}
	} else if con.State.Waiting != nil {
		result = "Waiting:" + con.State.Waiting.Reason
	}
	return result
}

func formatUpdateStatus(s string) string {
	var jsonData interface{}
	err := json.Unmarshal([]byte(s), &jsonData)
	if err != nil {
		return ""
	}
	res, errr := jsonpath.JsonPathLookup(jsonData, "$.statuses")
	if errr != nil {
		return ""
	}
	result := ""
	if _, ok := res.(map[string]interface{}); ok {
		cnameToMap := res.(map[string]interface{})
		for key := range cnameToMap {
			val := cnameToMap[key].(map[string]interface{})
			result = result + key + " " + fmt.Sprintf("%s", val["action"]) +
				":" + fmt.Sprintf("%t", val["success"]) + " " +
				fmt.Sprintf("%s", val["message"]) + ";\n"
		}
	}
	return result
}
