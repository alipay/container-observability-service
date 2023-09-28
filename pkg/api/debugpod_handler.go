package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	"github.com/iancoleman/orderedmap"
)

const (
	PodLifeStageMaxInterval = time.Duration(25e9)
)

type debugPodHandler struct {
	server        *Server
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *debugPodParam
}

type debugPodParam struct {
	podUID    string
	podName   string
	podIP     string
	hostname  string
	json      string
	diagnosis string
	env       string
}

type podInfo struct {
	ClusterName   string `json:"ClusterName,omitempty"`
	Namespace     string `json:"Namespace,omitempty"`
	PodName       string `json:"PodName,omitempty"`
	PodUID        string `json:"PodUID,omitempty"`
	PodIP         string `json:"PodIP,omitempty"`
	NodeName      string `json:"NodeName,omitempty"`
	NodeIP        string `json:"NodeIP,omitempty"`
	PodPhase      string `json:"PodPhase,omitempty"`
	LastTimeStamp string `json:"LastActiveAt,omitempty"`
	CreateTime    string `json:"CreatedAt,omitempty"`
	State         string `json:"State,omitempty"`
}

func (handler *debugPodHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *debugPodHandler) ParseRequest() error {
	req := debugPodParam{}
	if handler.request.Method == http.MethodGet {
		setSP(handler.request.URL.Query(), "uid", &req.podUID)
		setSP(handler.request.URL.Query(), "name", &req.podName)
		setSP(handler.request.URL.Query(), "podip", &req.podIP)
		setSP(handler.request.URL.Query(), "json", &req.json)
		setSP(handler.request.URL.Query(), "hostname", &req.hostname)
		setSP(handler.request.URL.Query(), "diagnosis", &req.diagnosis)
		setSP(handler.request.URL.Query(), "env", &req.env)
	}
	handler.requestParams = &req
	return nil
}

func (handler *debugPodHandler) ValidRequest() error {
	reqParam := handler.requestParams
	if reqParam.podUID == "" && reqParam.podName == "" && reqParam.podIP == "" && reqParam.hostname == "" {
		return fmt.Errorf("uid, name, podip, hostname are all empty")
	}
	return nil
}

// 全局变量: 保存 goroutine 并发查询的结果
var (
	podYamlRes      []*podyaml
	podPhaseRes     []*podPhase
	spanRes         []*span
	sloPodInfoRes   []*xsearch.SloPodInfo
	sloTraceDataRes []*sloTraceData
)

// 以下函数以闭包的形式调用查询函数, 并将结果保存到全局变量
func queryPodYaml(podUID string) {
	podYamlRes = queryPodYamlWithPodUID(podUID)
}

func queryPodPhase(podUID string) {
	podPhaseRes = queryPodphaseWithPodUID(podUID)
}

func querySpan(podUID string) {
	spanRes = querySpansByUID(podUID)
}

func querySloPodInfo(podUID string) {
	sloPodInfoRes = QuerySloPodInfo(podUID)
}

func querySloTraceData(podUID string) {
	sloTraceDataRes = querySloTraceDataByPodUID(podUID)
}

func (handler *debugPodHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("debugPodHandler.Process")

	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryDebugPod", begin)
	}()

	debugApiCalledCounter("debugPodHandler", handler.request)

	podUID := handler.requestParams.podUID
	if podUID == "" && handler.requestParams.podName != "" {
		uids := getPodUIDListByPodName(handler.requestParams.podName)
		if uids != nil && len(uids) > 0 {
			podUID = uids[0]
		}
	}
	if podUID == "" && handler.requestParams.podIP != "" {
		uids := getPodUIDListByIP(handler.requestParams.podIP)
		if uids != nil && len(uids) > 0 {
			podUID = uids[0]
		}
	}
	if podUID == "" && handler.requestParams.hostname != "" {
		uids := getPodUIDListByHostname(handler.requestParams.hostname)
		if uids != nil && len(uids) > 0 {
			podUID = uids[0]
		}
	}

	var wg sync.WaitGroup

	// 将需要并发的函数加到数组中
	queryFuncSlice := []func(string){queryPodYaml, queryPodPhase, querySpan, querySloPodInfo, querySloTraceData}

	// 并发查询
	for _, f := range queryFuncSlice {
		wg.Add(1)
		go func(f func(string)) {
			defer wg.Done()
			f(podUID)
		}(f)
	}

	result := orderedmap.New()

	podInfoKey := "PodInfo"
	sloDataKey := "SLOData"
	podLifeCycleAllKey := "PodLifeCycle(ALL)"
	podLifeCycleErrorKey := "PodLifeCycle(ERROR)"
	spanKey := "Spans"
	tooMuchTimeSpanKey := "TooMuchTimeSpan"

	// 同步并发查询信息
	wg.Wait()

	//基本信息
	podInfo := getPodInfo(podYamlRes, sloPodInfoRes)
	result.Set(podInfoKey, podInfo)

	if len(sloTraceDataRes) > 0 {
		result.Set(sloDataKey, transSloTraceDataList(sloTraceDataRes, handler.requestParams.env, handler.requestParams.json == "true"))
	}

	// 将 PodPhase 的数组转化成为 trace API 需要的字段，展示在前端
	all := formatPodPhase(podPhaseRes, false)
	err := formatPodPhase(podPhaseRes, true)
	if len(err) == len(all) {
		if len(err) > 0 {
			result.Set(podLifeCycleErrorKey, err)
		}
	} else {
		if len(all) > 0 {
			result.Set(podLifeCycleAllKey, all)
		}
		if len(err) > 0 {
			result.Set(podLifeCycleErrorKey, err)
		}
	}
	result.Set(spanKey, spanRes)

	if len(all) > 0 {
		overtimeStages := make([]string, 0, len(podPhaseRes))
		idx := len(podPhaseRes) - 1
		for idx > 0 {
			stageDiff := podPhaseRes[idx-1].StartTime.Sub(podPhaseRes[idx].StartTime)
			if stageDiff > PodLifeStageMaxInterval {
				overtimeStages = append(overtimeStages, fmt.Sprintf("stage{%d -> %d} diff: %s ", idx, idx-1, stageDiff))
			}
			idx--
		}

		if len(overtimeStages) > 0 {
			result.Set(tooMuchTimeSpanKey, overtimeStages)
		}
	}

	if handler.requestParams.json == "true" {
		bytes, err := json.Marshal(result)
		if err == nil {
			return http.StatusOK, string(bytes), nil
		}
	}

	return http.StatusOK, result, nil
}

// 格式化生命周期事件，展示给 trace API 的前端页面。
func formatPodPhase(podPhases []*podPhase, onlyError bool) []map[string]interface{} {
	defer utils.IgnorePanic("transform")

	allphasess := make([]map[string]interface{}, 0)

	if len(podPhases) <= 0 {
		return allphasess
	}

	currentTraceStage := ""
	for _, val := range podPhases {
		if val.StartTime.IsZero() {
			continue
		}
		if onlyError && !val.HasErr {
			continue
		}

		dic := make(map[string]interface{})
		if val.OperationName != "event" {
			dic["operationName"] = val.OperationName
		}
		dic["AuditID"] = val.PlfID
		if val.HasErr {
			dic["State"] = "ERROR!!"
		}
		dic["startTime"] = val.StartTime
		if val.ExtraInfo != nil {
			agent := val.ExtraInfo.(map[string]interface{})["UserAgent"]
			if agent != "" && agent != nil {
				dic["UserAgent"] = agent
			}
			nodeStatus := val.ExtraInfo.(map[string]interface{})["NodeStatus"]
			if nodeStatus != nil && nodeStatus != "" {
				dic["NodeStatus"] = nodeStatus
			}
			message := val.ExtraInfo.(map[string]interface{})["Message"]
			if message != nil && message != "" {
				dic["message"] = message
			}
			ext := val.ExtraInfo.(map[string]interface{})["eventObject"]
			if ext != nil {
				reason := ext.(map[string]interface{})["reason"]
				if reason != "" && reason != nil {
					dic["reason"] = reason
				}
				msg := ext.(map[string]interface{})["message"]
				if msg != "" && msg != nil {
					dic["message"] = msg
				}
			}
			ext = val.ExtraInfo.(map[string]interface{})["auditEvent.ResponseObject"]
			if ext != nil {
				auditMessage := ext.(map[string]interface{})["message"]
				if auditMessage != nil && auditMessage != "" {
					dic["message"] = auditMessage
				}
				code := ext.(map[string]interface{})["code"]
				if code != nil {
					dic["code"] = code
				}
			}
		}
		// 增加 TraceStage 字段
		dic["TraceStage"] = val.TraceStage

		// 补全 TraceStage，此处是为了后端和前端增加的

		var traceStage string = "abc"
		var ok bool
		if dic["reason"] != nil {
			traceStage, ok = reasonToTraceStageMap[dic["reason"].(string)]
			if !ok {
				if dic["operationName"] != nil {
					traceStage, ok = operationNameToTraceStageMap[dic["operationName"].(string)]
					if !ok {
						traceStage = "ToBeFilled"
					}
				} else {
					traceStage = "ToBeFilled"
				}
			}
		} else if dic["operationName"] != nil {
			traceStage, ok = operationNameToTraceStageMap[dic["operationName"].(string)]
			if !ok {
				traceStage = "ToBeFilled"
			}
		}

		if traceStage != "ToBeFilled" {
			currentTraceStage = traceStage
		}

		dic["TraceStage"] = currentTraceStage

		allphasess = append(allphasess, dic)
	}

	return allphasess
}

func getPodInfo(podYamls []*podyaml, sloPodInfos []*xsearch.SloPodInfo) *podInfo {
	result := &podInfo{}

	if len(podYamls) != 0 {
		if podYamls[0].ClusterName != "" {
			result.ClusterName = podYamls[0].ClusterName
		}
		if podYamls[0].Pod.Namespace != "" {
			result.Namespace = podYamls[0].Pod.Namespace
		}
		if podYamls[0].Pod.Name != "" {
			result.PodName = podYamls[0].Pod.Name
		}
		if podYamls[0].Pod.UID != "" {
			result.PodUID = string(podYamls[0].Pod.UID)
		}
		result.CreateTime = podYamls[0].Pod.CreationTimestamp.Format(time.RFC3339Nano)
		result.PodIP = podYamls[0].Pod.Status.PodIP
		result.NodeIP = podYamls[0].Pod.Status.HostIP
		result.PodPhase = string(podYamls[0].Pod.Status.Phase)
		result.NodeName = podYamls[0].Pod.Spec.NodeName
		result.LastTimeStamp = podYamls[0].StageTimestamp.Format(time.RFC3339Nano)
		if podYamls[0].IsBeginDelete == "true" {
			result.State = "IsBeginDelete"
		}
		if podYamls[0].IsDeleted == "true" {
			result.State = "IsDeleted"
		}

	}

	return result
}

// SLO数据格式化
func transSloTraceDataList(sloList []*sloTraceData, env string, apiCall bool) map[string]interface{} {
	result := make(map[string]interface{})

	podTypeKey := "PodType"
	createResultKey := "CreationResult"
	deleteResultKey := "DeletionResult"
	createdAtKey := "CreatedAt"
	scheduledAtKey := "ScheduledAt"
	containersReadyAtKey := "ContainersReadyAt"
	podRunningAtKey := "PodRunningAt"
	deletedAtKey := "DeletedAt"
	possibleReasonKey := "PossibleReason"
	upgradeKey := "Upgrade"

	for _, slo := range sloList {
		result[podTypeKey] = time.Duration(slo.PodSLO).String()
		result["DeliveryStatus"] = slo.DeliveryStatusOrig

		result["SloHint"] = slo.SloHint

		if slo.Type == "create" {
			result[createResultKey] = slo.StartUpResultFromCreate
		}
		if slo.Type == "delete" {
			result[deleteResultKey] = slo.DeleteResult
		}
		if !slo.CreatedTime.IsZero() && slo.Type == "create" {
			result[createdAtKey] = slo.CreatedTime.Format(time.RFC3339Nano)
		}
		if !slo.Scheduled.IsZero() && slo.Type == "create" {
			result[scheduledAtKey] = slo.Scheduled.Format(time.RFC3339Nano)
		}
		if !slo.ContainersReady.IsZero() && slo.Type == "create" {
			result[containersReadyAtKey] = slo.ContainersReady.Format(time.RFC3339Nano)
		}
		if !slo.RunningAt.IsZero() && slo.Type == "create" {
			result[podRunningAtKey] = slo.RunningAt.Format(time.RFC3339Nano)
		}
		if !slo.SucceedAt.IsZero() && slo.Type == "create" {
			result["SucceedAt"] = slo.SucceedAt.Format(time.RFC3339Nano)
		}
		if !slo.FailedAt.IsZero() && slo.Type == "create" {
			result["FailedAt"] = slo.FailedAt.Format(time.RFC3339Nano)
		}
		if !slo.ReadyAt.IsZero() && slo.Type == "create" {
			result["ReadyAt"] = slo.ReadyAt.Format(time.RFC3339Nano)
		}
		if !slo.CreatedTime.IsZero() && slo.Type == "delete" {
			result[deletedAtKey] = slo.CreatedTime.Format(time.RFC3339Nano)
		}

		if slo.PossibleReason != "" {
			result[possibleReasonKey] = slo.PossibleReason
		}
	}

	upgradeList := getUpgradeOpList(sloList, apiCall)
	if len(upgradeList) > 0 {
		result[upgradeKey] = upgradeList
	}

	return result
}

func getUpgradeOpList(sloList []*sloTraceData, apiCall bool) []map[string]string {
	result := make([]map[string]string, 0)

	count := 0
	for _, slo := range sloList {
		if slo.Type == "pod_upgrade" {
			count = count + 1
			if count > 5 {
				break
			}
			m := make(map[string]string)
			if apiCall {
				m["UpgradedAt"] = slo.CreatedTime.Format(time.RFC3339Nano)
				m["UpgradeResult"] = slo.UpgradeResult
				m["UpgradeFinishAt"] = slo.UpgradeEndTime.Format(time.RFC3339Nano)
			} else {
				m["UpgradeStarted"] = slo.CreatedTime.Format(time.RFC3339Nano)
				m["UpgradeResult"] = slo.UpgradeResult
				m["UpgradeEnd"] = slo.UpgradeEndTime.Format(time.RFC3339Nano)
			}

			result = append(result, m)
		}
	}

	return result
}

func debugPodFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &debugPodHandler{
		server:  s,
		request: r,
		writer:  w,
	}
}
