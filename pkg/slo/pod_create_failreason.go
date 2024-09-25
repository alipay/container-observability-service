package slo

import (
	"container/list"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/shareutils"

	"github.com/alipay/container-observability-service/pkg/reason"

	"github.com/alipay/container-observability-service/pkg/config"
	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/reason/analyzers"
	"github.com/alipay/container-observability-service/pkg/shares"
	"gopkg.in/yaml.v2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// 分析失败原因
func (data *PodStartupMilestones) analyzeFailedReason() string {
	eventsNormalOrder := make([]*v1.Event, 0)
	v, ok := podEventsMap.Get(data.key)
	if ok && v != nil {
		for item := v.(*list.List).Front(); nil != item; item = item.Next() {
			e := item.Value.(*v1.Event)
			eventsNormalOrder = append(eventsNormalOrder, e)
		}
	}

	if data.PossibleReason == nil {
		reason := ""
		data.PossibleReason = &reason
	}

	//get audit event
	v, ok = podAuditLogMap.Get(data.key)
	auditEvents := make([]*shares.AuditEvent, 0)
	if ok && v != nil {
		for item := v.(*list.List).Front(); nil != item; item = item.Next() {
			auditEvents = append(auditEvents, item.Value.(*shares.AuditEvent))
		}
	}

	if featuregates.IsEnabled(reason.NewReasonFeature) {
		analyzerDAG := analyzers.ShareAnalyzerFactory.GetAnalyzerByType(analyzers.PodCreate)
		if len(auditEvents) > 0 {
			analyzerDAG.AuditEvents = auditEvents
		}

		if spans, callBack := shareutils.GetSpansByUIDAndType(data.latestPod.UID, nil, "PodCreate"); spans != nil {
			defer callBack()
			analyzerDAG.Spans = spans
		}
		analyzerDAG.BeginTime = &data.CreatedTime
		analyzerDAG.EndTime = data.trickTime
		analyzerDAG.Analysis(data.Cluster, data.PodName, data.PodUID)

		return analyzerDAG.GetResult().Result
	}
	return analyzeFailureReasonDetails(data.latestPod, eventsNormalOrder, data.CreatedTime, data.shouldFinishTime, data.IsJob, data.PossibleReason)
}

// 用于分析详细的pod失败原因
func analyzeFailureReasonDetails(podyaml *v1.Pod, events []*v1.Event, cTime time.Time, shouldFinishTime *time.Time, isJob bool, possibleReason *string) string {
	eventsROrder, _ := reverseEvent(events)

	// 调度
	scheduleStats, t := getScheduleStatus(podyaml)
	if scheduleStats == 1 { //已经调度
		if t.Sub(cTime) > 60*time.Second {
			return ScheduleDelay
		}
	} else if scheduleStats == 0 { //调度失败
		return "FailedScheduling"
	} else if scheduleStats == -1 { //未调度
		for _, val := range events {
			if val.Reason == "UnexpectedAdmissionError" {
				if isFileSystemReadOnlyFailure(val) {
					return "FileSystemReadOnly"
				}
				return "UnexpectedAdmissionError"
			}
		}
		return ScheduleDelay
	}

	// 特殊原因
	if len(events) <= 0 { //没有event事件
		return NoEvent
	}

	if isPreemption(events) { //排除抢占调度的场景
		return "Preemption"
	}

	// postStartHookTime failed
	postStartHookTime := calculatePostStartHookTime(events)
	postStartHookTimeout := 180
	if config.GlobalLunettesConfig().PostStartHookTimeout != "" {
		postStartHookTimeoutFromConfigMap, err := time.ParseDuration(config.GlobalLunettesConfig().PostStartHookTimeout)
		if err != nil {
			klog.Errorf("Error parsing user timeout from configmap, %v", err)
		} else {
			// klog.Infof("customized timeout in configmap is %s", postStartHookTimeoutFromConfigMap)
			postStartHookTimeout = int(postStartHookTimeoutFromConfigMap.Seconds())
		}
	}
	if postStartHookTime >= postStartHookTimeout {
		return "PostStartHookTooMuchTime"
	}

	// Runtime Operation failed
	if isRuntimeOperationFailure(events) {
		return "RuntimeOperationFailed"
	}

	// 其他错误，从后向前做
	createdSandbox := false
	mountedVolumes := false
	for _, value := range eventsROrder {
		if strings.EqualFold(value.Reason, "SuccessfulCreatePodSandBox") {
			createdSandbox = true
		}

		if strings.EqualFold(value.Reason, "MountVolume") && strings.Contains(value.Message, "Successfully mounted") {
			mountedVolumes = true
		}

		if value.Reason == "FailedPostStartHook" {
			return value.Reason
		}
		if strings.Contains(value.Message, "hostPath type check failed") {
			return InvalidMountConfig
		}

		if value.Reason == "FailedMount" && !createdSandbox && !mountedVolumes {
			return "FailedMount"
		}

		if value.Reason == "Failed" {
			if strings.Contains(value.Message, "InvalidImageName") {
				return "InvalidImageName"
			}
			if strings.Contains(value.Message, "Failed to pull image") {
				if strings.Contains(value.Message, "not found") || strings.Contains(value.Message, "manifest unknown") {
					return "ImageNotFound"
				}
				return "FailedPullImage"
			}
		}
		if value.Reason == "FailedCreatePodSandBox" && !createdSandbox {
			//网络问题: ip分配超时
			if strings.Contains(value.Message, "timeout to allocate ip for pod") {
				return "AllocateIPTimeout"
			}

			if strings.Contains(value.Message, "failed to setup network for sandbox") {
				//网络问题: mac的nic分配错误
				if strings.Contains(value.Message, "Can not find host nic by mac address") || strings.Contains(value.Message, "no nic found") {
					return "NotFoundNicByMac"
				}
				//网络问题: mac的nic分配错误
				if strings.Contains(value.Message, "fail to allocate ip") {
					return "FailedAllocateIP"
				}

				return "FailedSetNetwork"
			}

			re := regexp.MustCompile("failed to set up sandbox container \"(.+)\" network for pod")
			match := re.FindStringSubmatch(value.Message)
			if match != nil && len(match) >= 1 {
				//网络问题：宿主机中网桥端口满
				if strings.Contains(value.Message, "exchange full") {
					return "BridgeExchangeFull"
				}
				return "FailedSetNetwork"
			}

			//磁盘空间耗尽问题
			if strings.Contains(value.Message, "no space left on device") {
				return "NoDiskSpace"
			}

			*possibleReason = value.Message
			return "FailedCreatePodSandBox"
		}
	}

	// event中给出的Error
	for _, value := range eventsROrder {
		if value.Type != "Normal" {
			if value.Reason == "BackOff" && strings.Contains(value.Message, "restarting failed container") {
				return "CrashLoopBackOff"
			}
			if len(strings.TrimSpace(value.Reason)) > 0 {
				return value.Reason
			}
		}
	}

	// 镜像拉取超时
	imagePullTime := calculateImagePullTime(events, shouldFinishTime)
	for key, value := range imagePullTime {
		if isJob && value > 30 || !isJob && value > 60 {
			*possibleReason = fmt.Sprintf("Image [%s] pull consume too much time", key)
			return PullImageTooMuchTime
		}
	}

	//检测Pod的各个condition
	if podyaml != nil {
		conditions := sortConditionsBytime(podyaml.Status.Conditions)
		for _, value := range conditions {
			klog.Infof("pod[%s], condition type[%s], status[%s]", podyaml.Name, value.Type, value.Status)
			if value.Status == v1.ConditionTrue || (shouldFinishTime != nil && value.LastTransitionTime.After(*shouldFinishTime)) {
				continue
			}
			if value.Type == v1.PodReady && value.Reason != "" {
				return value.Reason
			}
		}
	}

	// kubelet处理延迟（各种原因）
	if isDelay, reason := isKubeletDelay(events); isDelay {
		*possibleReason = reason
		return KubeletDelay
	}

	*possibleReason = getPossibleReason(events)
	klog.Info("reason result timeout_unknown\n")
	if podyaml != nil {
		klog.Infof("-----------------------pod %s timeout_unknown, begin-------------------------", podyaml.Name)
		if pb, err := yaml.Marshal(podyaml); err == nil {
			klog.Infof("PodJson: %s \n", string(pb))
		}

		if events != nil {
			for idx, ev := range events {
				klog.Infof("event_%d: %s\n", idx, ev.String())
			}
		}
		klog.Infof("-----------------------pod %s timeout_unknown, end-------------------------", podyaml.Name)
	}
	return timeoutUnknown
}

// isKubeletDelay 是否是kubelet处理延迟导致的失败
func isKubeletDelay(events []*v1.Event) (bool, string) {
	//1.有Scheduled、但没有Pulled/Pulling/Created/FailedMount
	isScheduled := false
	isPulledOrPulling := false
	for _, event := range events {
		if event.Reason == "Scheduled" {
			isScheduled = true
		} else if event.Reason == "Pulled" || event.Reason == "Pulling" || event.Reason == "Created" ||
			event.Reason == "FailedMount" {
			isPulledOrPulling = true
		}
	}
	if isScheduled && !isPulledOrPulling {
		return true, "Pod has been scheduled, but did not start pull image"
	}

	//2.IpAllocated/Initialized -> Pulled 时间过长 （> 30秒）
	var preEvent *v1.Event = nil
	var currentEvent *v1.Event = nil
	for _, event := range events {
		if event.Reason != "IpAllocated" && event.Reason != "Pulled" && event.Reason != "Initialized" {
			continue
		}
		preEvent = currentEvent
		currentEvent = event
		if preEvent != nil && currentEvent != nil &&
			currentEvent.FirstTimestamp.Time.Sub(preEvent.FirstTimestamp.Time) > 30*time.Second {
			if preEvent.Reason == "IpAllocated" && currentEvent.Reason == "Pulled" &&
				strings.Contains(currentEvent.Message, "already present on machine") {
				return true, "IpAllocated -> Pulled 时间过长 （> 30秒）"
			} else if preEvent.Reason == "Initialized" && currentEvent.Reason == "Pulled" &&
				strings.Contains(currentEvent.Message, "already present on machine") {
				return true, "Initialized -> Pulled 时间过长 （> 30秒）"
			}
		}
	}

	return false, ""
}

// getScheduleStatus, 1:已调度；0：调度失败; -1: 未处理
func getScheduleStatus(pod *v1.Pod) (int, time.Time) {
	var t time.Time
	if pod == nil {
		return -1, t
	}
	if pod.Spec.NodeName != "" {
		return 1, t
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionTrue {
			if pod.Spec.NodeName == "" {
				return 0, t
			}
			return 1, condition.LastTransitionTime.Time
		} else if condition.Type == v1.PodScheduled && condition.Status == v1.ConditionFalse &&
			condition.Reason == v1.PodReasonUnschedulable {
			return 0, condition.LastTransitionTime.Time
		}
	}
	return -1, t
}

func isPreemption(events []*v1.Event) bool {
	for _, value := range events {
		if value.Reason == "PreemptionSuccess" || value.Reason == "ToDoPreemption" {
			return true
		}
	}
	return false
}

// 计算镜像拉取时间
func calculateImagePullTime(events []*v1.Event, shouldFinishTime *time.Time) map[string]float64 {
	result := make(map[string]float64)
	eventMap := make(map[string]*v1.Event)
	var cur *v1.Event

	for _, value := range events {
		if value.Reason != "Pulling" && value.Reason != "Pulled" {
			continue
		}
		cur = value
		if cur != nil && cur.Reason == "Pulled" {
			if strings.Contains(cur.Message, "already present on machine") {
				re := regexp.MustCompile(".*\"(.+)\" already present on machine.*")
				match := re.FindStringSubmatch(cur.Message)
				if match != nil && len(match) >= 2 {
					result[match[1]] = 0
				} else {
					klog.Error(cur)
				}
			} else if strings.Contains(cur.Message, "Successfully pulled image") {
				re := regexp.MustCompile("Successfully pulled image \"(.+)\".*")
				match := re.FindStringSubmatch(cur.Message)
				if match != nil && len(match) >= 2 {
					//匹配上对应的pulling
					if pullingEvent, ok := eventMap[match[1]]; ok {
						result[match[1]] = cur.FirstTimestamp.Time.Sub(pullingEvent.FirstTimestamp.Time).Seconds()
					} else { //如果匹配不上则打印错误
						klog.Errorf("can not match pulling event %v", cur)
					}
				} else {
					klog.Error(cur)
				}
			} else {
				klog.Error(cur)
			}
		} else if cur != nil && cur.Reason == "Pulling" {
			re := regexp.MustCompile("Pulling image \"(.+)\"")
			match := re.FindStringSubmatch(cur.Message)
			if match != nil && len(match) >= 2 {
				//放入map
				eventMap[match[1]] = cur
				// 如果只有pulling事件发生，会因为镜像时间在30或者60s之内，而忽视单pulling事件
				if shouldFinishTime != nil && !cur.FirstTimestamp.IsZero() {
					result[match[1]] = (*shouldFinishTime).Sub(cur.FirstTimestamp.Time).Seconds()
				} else {
					result[match[1]] = 1
				}
			} else {
				klog.Error(cur)
			}
		}
	}

	return result
}

// 计算 PostStartHook 的最大耗时
// 一个 Pod 可能有多个 container，每个 container 都可能有 PostStartHook
func calculatePostStartHookTime(events []*v1.Event) int {

	var start *time.Time
	var end *time.Time

	var maxPostStartHookTime float64 = -1

	for _, value := range events {

		if value.Reason == "Started" {
			start = &value.FirstTimestamp.Time
			continue
		}

		if value.Reason == "SucceedPostStartHook" {
			end = &value.FirstTimestamp.Time
		}

		if start != nil && end != nil {
			currentHookTime := end.Sub(*start).Seconds()
			if currentHookTime > maxPostStartHookTime {
				maxPostStartHookTime = currentHookTime
			}
		}
	}
	return int(maxPostStartHookTime)
}

// 事件中包含 "failed to update resolv file of container" 被分类为 RuntimeOperationFailed
func isRuntimeOperationFailure(events []*v1.Event) bool {
	for _, value := range events {
		if strings.Contains(value.Message, "failed to update resolv file of container") {
			return true
		}
	}
	return false
}

// 事件中包含 "read-only file system, which is unexpected" 被分类为 FileSystemReadOnly
func isFileSystemReadOnlyFailure(event *v1.Event) bool {
	return strings.Contains(event.Message, "read-only file system, which is unexpected")
}

func reverseEvent(events []*v1.Event) ([]*v1.Event, bool) {
	result := make([]*v1.Event, 0)
	hasErr := false
	if len(events) <= 0 {
		return result, false
	}
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type != "Normal" {
			hasErr = true
		}
		result = append(result, events[i])
	}
	return result, hasErr
}

// status.conditions按照时间排序
func sortConditionsBytime(conditions []v1.PodCondition) []v1.PodCondition {
	if conditions == nil || len(conditions) == 0 {
		return conditions
	}
	_sortConditiosByTime(conditions, 0, len(conditions)-1)
	return conditions
}
func _sortConditiosByTime(conditions []v1.PodCondition, left int, right int) {
	if left >= right {
		return
	}
	flagCondition := conditions[left]
	l := left
	r := right
	for l < r {
		for r > l && !conditions[r].LastTransitionTime.Time.After(flagCondition.LastTransitionTime.Time) {
			r--
		}
		if l >= r {
			break
		}
		conditions[l] = conditions[r]
		for l < r && !conditions[l].LastTransitionTime.Time.Before(flagCondition.LastTransitionTime.Time) {
			l++
		}
		if l >= r {
			break
		}
		conditions[r] = conditions[l]
	}
	conditions[l] = flagCondition

	_sortConditiosByTime(conditions, left, l-1)
	_sortConditiosByTime(conditions, l+1, right)
}

// 计算耗时最长的阶段
func getPossibleReason(events []*v1.Event) string {
	maxConsume := -0.1
	idx := -1
	i := 1
	for i < len(events) {
		ct := events[i].FirstTimestamp.Time.Sub(events[i-1].FirstTimestamp.Time).Seconds()
		if ct > maxConsume {
			maxConsume = ct
			idx = i
		}
		i++
	}

	if maxConsume >= 0 {
		msgFrom := ""
		msgTo := ""
		if events[idx-1].Reason != "" {
			msgFrom = events[idx-1].Reason
		} else if events[idx-1].Action != "" {
			msgFrom = events[idx-1].Action
		}

		if events[idx].Reason != "" {
			msgTo = events[idx].Reason
		} else if events[idx].Action != "" {
			msgTo = events[idx].Action
		}

		return fmt.Sprintf("The most time consuming stage is from [%s] to [%s]", msgFrom, msgTo)
	} else {
		return "Too little events to judge!"
	}

}

// 分析api_failed原因
func analysisApiFailed(message string) string {
	matchStr := "failed calling webhook"

	if idx := strings.LastIndex(message, matchStr) - 1; idx >= 0 && idx < len(message) {
		message = message[idx:]
		if strings.Contains(message, "failed calling webhook") {
			re := regexp.MustCompile("failed calling webhook.*\"(.*)\": ")
			match := re.FindStringSubmatch(message)
			if match != nil && len(match) >= 1 {
				return match[1]
			}
		}
	} else if strings.Contains(message, "quota") && strings.Contains(message, "admission webhook") {
		return "QuotaWebhookDenied"
	}
	return resultAPIFailed
}
