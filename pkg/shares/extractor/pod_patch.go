package extractor

import (
	"encoding/json"
	"fmt"

	"github.com/alipay/container-observability-service/pkg/shares"

	"strings"

	"github.com/oliveagle/jsonpath"

	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	retainKeysStrategy                     = "retainKeys"
	deleteFromPrimitiveListDirectivePrefix = "$deleteFromPrimitiveList"

	finalizerSuffix = "finalizers"
)

type PodPatchProcessor struct {
}

func (p *PodPatchProcessor) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ResponseStatus.Code >= 300 || event.Verb != "patch" {
		return false
	}

	if event.RequestRuntimeObj == nil {
		return false
	}

	return true
}

func (p *PodPatchProcessor) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodCreation")

	_, ok := event.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		klog.Warningf("can not convert event from request runtime obj")
		return nil
	}

	event.Type = shares.AuditTypeOperation

	if event.ObjectRef.Subresource == "" {
		p.processPatch(event)
	} else if event.ObjectRef.Subresource == "status" {
		p.processPatchStatus(event)
	}

	return nil
}

// 处理非 Status 的字段
// 包括 finalizer，labels，annotation，spec.SchedulerName 等
func (p *PodPatchProcessor) processPatch(event *shares.AuditEvent) {
	defer utils.IgnorePanic("processPatch ")

	// request pod
	reqPod, ok := event.RequestRuntimeObj.(*v1.Pod)
	if !ok {
		return
	}

	// Finalizer
	// add finalizer
	if len(reqPod.ObjectMeta.Finalizers) == 1 {
		event.Operation["finalizer:add"] = []string{fmt.Sprintf("finalizer:%s:add", reqPod.ObjectMeta.Finalizers[0])}
	}
	// delete finalizer
	if event.RequestMetaJson != nil {
		for key, value := range event.RequestMetaJson {
			if strings.HasPrefix(key, deleteFromPrimitiveListDirectivePrefix) && strings.HasSuffix(key, finalizerSuffix) {
				operations := make([]string, 0)
				for _, f := range value.([]interface{}) {
					operations = append(operations, fmt.Sprintf("finalizer:%s:delete", f.(string)))
				}
				event.Operation["finalizer:delete"] = operations
			}
		}
	}

	//Labels
	// 必须从原始meta json map中取值，才能识别是add还是delete
	if event.RequestMetaJson != nil {
		if lab, ok := event.RequestMetaJson["labels"]; ok {
			labels := lab.(map[string]interface{})
			toAdd := make([]string, 0)
			toDelete := make([]string, 0)
			for key, value := range labels {
				if isIgnore(key) {
					continue
				}
				//add
				if value != nil {
					toAdd = append(toAdd, fmt.Sprintf("label:%s=%s:add", key, value))
				} else {
					//delete
					toDelete = append(toDelete, fmt.Sprintf("label:%s=%s:delete", key, value))
				}
			}

			//add
			if len(toAdd) > 0 {
				event.Operation["label:add"] = toAdd
			}
			//delete
			if len(toDelete) > 0 {
				event.Operation["label:delete"] = toDelete
			}
		}
	}

	//annotation
	// 必须从原始meta json map中取值，才能识别是add还是delete
	if event.RequestMetaJson != nil {
		if anno, ok := event.RequestMetaJson["annotations"]; ok {
			annotations := anno.(map[string]interface{})
			toAdd := make([]string, 0)
			toDelete := make([]string, 0)
			for key, value := range annotations {
				if isIgnore(key) {
					continue
				}
				//add
				if value != nil {
					toAdd = append(toAdd, fmt.Sprintf("annotation:%s=%s:add", key, value))
				} else {
					//delete
					toDelete = append(toDelete, fmt.Sprintf("annotation:%s=%s:delete", key, value))
				}
			}
			//add
			if len(toAdd) > 0 {
				event.Operation["annotation:add"] = toAdd
			}
			//delete
			if len(toDelete) > 0 {
				event.Operation["annotation:delete"] = toDelete
			}
		}
	}

	//scheduler
	// entry default scheduler。
	if reqPod.Spec.SchedulerName != "" {
		event.Operation[fmt.Sprintf("schedule:%s:entry", reqPod.Spec.SchedulerName)] = []string{}
	}

}

// 处理 Status 的字段的更新
func (p *PodPatchProcessor) processPatchStatus(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPatchStatus ")

	// request pod
	reqPod, ok := auditEvent.RequestRuntimeObj.(*v1.Pod)
	if !ok {
		return
	}

	//condition
	for _, cond := range reqPod.Status.Conditions {
		if string(cond.Type) == "" {
			continue
		}
		auditEvent.Operation[fmt.Sprintf("condition:%s:%s", cond.Type, strings.ToLower(string(cond.Status)))] = []string{}
	}
	//phase
	if string(reqPod.Status.Phase) != "" {
		auditEvent.Operation[fmt.Sprintf("phase:%s:set", reqPod.Status.Phase)] = []string{}
	}

	//podIP
	if string(reqPod.Status.PodIP) != "" {
		auditEvent.Operation[fmt.Sprintf("pod:ip:set")] = []string{}
	}

	//容器状态
	for _, containerStatus := range reqPod.Status.ContainerStatuses {
		auditEvent.Operation["containerReady:set"] = append(auditEvent.Operation["containerReady:set"], fmt.Sprintf("containerReady:%s:%t", containerStatus.Name, containerStatus.Ready))
		auditEvent.Operation["containerRestartCount:set"] = append(auditEvent.Operation["containerRestartCount:set"], fmt.Sprintf("containerRestartCount:%s:%d", containerStatus.Name, containerStatus.RestartCount))
		rs := formatContainerStatus(containerStatus)
		auditEvent.Operation["containerState:set"] = append(auditEvent.Operation["containerState:set"], fmt.Sprintf("containerState:%s:%s", containerStatus.Name, rs))
	}
}

func formatContainerStatus(con v1.ContainerStatus) string {
	result := ""
	if con.State.Running != nil {
		result = "Running"
	} else if con.State.Terminated != nil {
		if con.State.Terminated.Reason != "" {
			result = "terminated:" + con.State.Terminated.Reason + ";exitCode:" + fmt.Sprintf("%d", con.State.Terminated.ExitCode)
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

func isIgnore(key string) bool {
	key = strings.ToLower(key)
	if (strings.Contains(key, "status") && !strings.Contains(key, "update-status")) || strings.Contains(key, "trace") || strings.Contains(key, "spec") {
		return true
	}
	return false
}
