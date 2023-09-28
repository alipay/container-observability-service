package utils

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/shares"
	corev1 "k8s.io/api/core/v1"
)

// PodConditionExists() whether the condition is exists.
func PodConditionExists(conditions []corev1.PodCondition, conditionKey string) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodConditionType(conditionKey) {
			return true
		}
	}
	return false
}

// 获取镜像名称
func GetImageName(msg string) *string {
	if strings.Contains(msg, "already present on machine") {
		re := regexp.MustCompile(".*\"(.+)\" already present on machine.*")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Successfully pulled image") {
		re := regexp.MustCompile("Successfully pulled image \"(.+)\".*")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			//匹配上对应的pulling
			return &match[1]
		}
	} else if strings.Contains(msg, "Pulling image") {
		re := regexp.MustCompile("Pulling image \"(.+)\"")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Failed to pull image") {
		re := regexp.MustCompile("Failed to pull image \"(.+)\":")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Back-off pulling image") {
		re := regexp.MustCompile("Back-off pulling image \"(.+)\"")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Failed to inspect image") {
		re := regexp.MustCompile("Failed to inspect image \"(.*)\":")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	}
	return nil
}

// 从Created reason 获取container name
// msg："Created container task, elapsedTime 65.729783ms"
// msg: "Started container task, elapsedTime 261.205298ms"
// msg: ""
func GetCreatedAndStartedContainerName(msg string) []string {
	re := regexp.MustCompile("Created container (.+), elapsedTime (.+)")
	match := re.FindStringSubmatch(msg)
	if match != nil && len(match) >= 2 {
		return match[1:3]
	}

	re = regexp.MustCompile("Started container (.+), elapsedTime (.+)")
	match = re.FindStringSubmatch(msg)
	if match != nil && len(match) >= 2 {
		return match[1:3]
	}
	//兼容1.14版本
	if !strings.Contains(strings.ToLower(msg), "elapsedTime") {
		re = regexp.MustCompile("Created container (.+)")
		match = re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 1 {
			return append(append([]string{}, match[1]), "")
		}
	}

	if !strings.Contains(strings.ToLower(msg), "elapsedTime") {
		re = regexp.MustCompile("Started container (.+)")
		match = re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 1 {
			return append(append([]string{}, match[1]), "")
		}
	}

	return nil
}

// 从Killed message 获取container name
func GetKilledContainerName(msg string) []string {
	re := regexp.MustCompile("Stopping container (.+), elapsedTime (.+)")
	match := re.FindStringSubmatch(msg)
	if len(match) >= 2 {
		return match[1:3]
	}
	return nil
}

// 根据镜像找到容器的名字
func GetContainerNameByImageName(image string, pod *corev1.Pod) string {
	for _, co := range pod.Spec.InitContainers {
		if co.Image == image {
			return co.Name
		}
	}
	for _, co := range pod.Spec.Containers {
		if co.Image == image {
			return co.Name
		}
	}
	return ""
}

// 获取容器名称
func GetContainerName(msg string) *string {
	if strings.Contains(msg, "Starting to execute poststart hook for container") {
		re := regexp.MustCompile("Starting to execute poststart hook for container (.*) with")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Successfully execute poststart hook for container") {
		re := regexp.MustCompile("Successfully execute poststart hook for container (.*),")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Waiting:CreateContainerError") {
		re := regexp.MustCompile("(.*) restartCount:.* Ready:.* Waiting:CreateContainerError")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "Terminated:Error") {
		re := regexp.MustCompile("(.*) restartCount:.* Ready:.* Terminated:Error exitCode:.*")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	} else if strings.Contains(msg, "restartCount") && strings.Contains(msg, "Ready") {
		re := regexp.MustCompile("(.*) restartCount:.* Ready:.*")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	}
	return nil
}

// 获取poststarthook超时时间
func GetPostStartHookTimeout(msg string) *string {
	if strings.Contains(msg, "Starting to execute poststart hook for container") {
		re := regexp.MustCompile("Starting to execute poststart hook for container .* with timeout: (.*)")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	}
	return nil
}

// 获取poststarthook耗时时间
func GetPostStartHookElapsed(msg string) *string {
	if strings.Contains(msg, "Successfully execute poststart hook for container") {
		re := regexp.MustCompile("Successfully execute poststart hook for container .*, elapsedTime (.*)")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	}
	return nil
}

// 从pod的containerstatuses中获取某个container壮体啊
func GetContainerStatus(containerName string, pod *corev1.Pod) *corev1.ContainerStatus {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name == containerName {
			return &cs
		}
	}
	return nil
}

// 判断pod是否ready
func IsPodReady(pod *corev1.Pod, startTime *time.Time) bool {
	if pod != nil {
		for _, cd := range pod.Status.Conditions {
			if cd.Type == "Ready" && strings.EqualFold(string(cd.Status), "true") {
				if startTime != nil {
					if cd.LastTransitionTime.After(*startTime) {
						return true
					}
					return false
				}
				return true
			}
		}
	}
	return false
}

// 判断pod是否ready
func IsPodSucceed(pod *corev1.Pod) bool {
	if pod != nil && pod.Status.Phase == corev1.PodSucceeded {
		return true
	}
	return false
}

func DiagnosisToPhase(result *share.ReasonResult) map[string]interface{} {
	rs := make(map[string]interface{})
	mostTimeSpan := ""
	podSpanAnalysis := result.Diagnosis["pod_span_analysis"]
	if podSpanAnalysis != nil {
		span := podSpanAnalysis.(string)
		mostTimeSpan = span[strings.Index(span, "[")+1 : strings.Index(span, "]")]
	}

	rs["TraceStage"] = "运行阶段"
	rs["UserAgent"] = "lunettes/analyzer"
	rs["operationName"] = fmt.Sprintf("create_result:%s  most_time_consuming:%s", result.Result, mostTimeSpan)
	rs["message"] = fmt.Sprintf("（请忽略）临时加入诊断结果，不是pod的真正生命周期，后续会去除")
	rs["startTime"] = result.Diagnosis["pod_start_time"]

	return rs
}

// 获取被node的condition驱逐的原因
func GetConditionName(msg string) *string {
	if strings.Contains(msg, "The node had condition") {
		re := regexp.MustCompile("The node had condition: \\[(.*)\\]")
		match := re.FindStringSubmatch(msg)
		if match != nil && len(match) >= 2 {
			return &match[1]
		}
	}
	return nil
}

// 从hyper event数组中获取最新pod yaml
func GetPodYamlFromHyperEvents(auditEvents []*shares.AuditEvent, endTime *time.Time) *corev1.Pod {
	if auditEvents == nil {
		return nil
	}
	for i := len(auditEvents) - 1; i >= 0; i-- {
		if endTime != nil && auditEvents[i].StageTimestamp.After(*endTime) {
			continue
		}

		if auditEvents[i].ResponseRuntimeObj.GetObjectKind().GroupVersionKind().Kind == "Pod" {
			return auditEvents[i].ResponseRuntimeObj.(*corev1.Pod)
		}
	}
	return nil
}

// 从msg中提取quota不足的资源
func ExtractQuotaResource(msg string) string {
	regResource := regexp.MustCompile(".* resource \"(.*)\" not enough")
	resourceMatches := regResource.FindStringSubmatch(msg)
	if len(resourceMatches) >= 2 {
		return resourceMatches[1]
	}

	return ""
}

// 首字母大写
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
