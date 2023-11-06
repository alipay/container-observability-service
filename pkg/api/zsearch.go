package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/spans"

	"github.com/oliveagle/jsonpath"

	v1 "k8s.io/api/core/v1"

	"github.com/olivere/elastic/v7"
	"k8s.io/klog/v2"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	k8saudit "k8s.io/apiserver/pkg/apis/audit"
)

const (
	maxResultCount        = 500
	podphaseInexName      = "pod_life_phase"
	podphaseTypeName      = "_doc"
	podYamlIndexName      = "pod_yaml"
	podYamlTypeName       = "_doc"
	sloDataIndexName      = "slo_data"
	sloTraceDataIndexName = "slo_trace_data_daily"
	sloDataTypeName       = "data"

	traceStageAdmission  = "AdmissionStage"
	traceStageDecision   = "DecisionStage"
	traceStageScheduling = "SchedulingStage"
	traceStageRunning    = "RunningStage"
)

var (
	reasonToTraceStageMap = map[string]string{
		"Scheduled": traceStageScheduling,

		"Pulling": traceStageRunning,
		"Pulled":  traceStageRunning,
	}

	operationNameToTraceStageMap = map[string]string{
		"condition:Ready:true": traceStageRunning,

		"apicreate": traceStageAdmission,
		"scheduled": traceStageScheduling,

		"Enters default-scheduler": traceStageScheduling,

		"pod phase: Running": "CompleteDelivery",
	}
)

// podPhase 代表每个 Pod 的 trace 链路中的一个细分步骤
type podPhase struct {
	PlfID string
	// stage 代表 Pod 的 trace 链路中的一个大类
	// OperationName 和 Reason 可以用来决定 stage
	TraceStage    string      `json:"traceStage"`
	DataSourceId  string      `json:"dataSourceId"`
	ClusterName   string      `json:"clusterName"`
	Namespace     string      `json:"namespace"`
	PodName       string      `json:"podName"`
	PodUID        string      `json:"podUID"`
	OperationName string      `json:"operationName"`
	HasErr        bool        `json:"hasErr"`
	StartTime     time.Time   `json:"startTime"`
	ExtraInfo     interface{} `json:"extraInfo,omitempty"`
	Reason        interface{} `json:"reason,omitempty"`
	Message       interface{} `json:"message,omitempty"`
}

type podyaml struct {
	AuditID        string    `json:"auditID"`
	ClusterName    string    `json:"clusterName"`
	IsBeginDelete  string    `json:"isBeginDelete"`
	IsDeleted      string    `json:"isDeleted"`
	Pod            *v1.Pod   `json:"pod"`
	PodName        string    `json:"podName"`
	PodUID         string    `json:"podUID"`
	HostIP         string    `json:"hostIP"`
	PodIP          string    `json:"podIP"`
	StageTimestamp time.Time `json:"stageTimestamp"`
}

type span struct {
	Name       string    `json:"Name"`
	Type       string    `json:"Type"`
	Begin      time.Time `json:"Begin"`
	End        time.Time `json:"End"`
	Elapsed    string    `json:"Elapsed"`
	ActionType string    `json:"ActionType"`
}

type slodata struct {
	PodName                   string
	PodUID                    string
	NodeIP                    string
	NodeName                  string // in some cases, when a pod is scheduled to a specified node, but kubelet doesn't post any pod status to apiserver, nodeIP is empty, but we do want know which node is not working properly.
	PullTimeoutImageName      string
	IsJob                     bool
	Cluster                   string
	Namespace                 string
	StartUpResultFromCreate   string
	StartUpResultFromSchedule string
	DebugUrl                  string
	Created                   time.Time
	CreatedTime               time.Time
	Scheduled                 time.Time
	ContainersReady           time.Time
	RunningAt                 time.Time
	SucceedAt                 time.Time
	FailedAt                  time.Time
	ReadyAt                   time.Time
	DeletedTime               time.Time
	DeleteEndTime             time.Time
	DeleteTimeoutTime         time.Time
	FinishTime                time.Time
	UpgradeEndTime            time.Time
	UpgradeTimeoutTime        time.Time
	UpgradeResult             string
	PossibleReason            string
	DeleteResult              string
	PodSLO                    int64
	DeliverySLO               int64
	SLOViolationReason        string
	DeliveryStatus            string
	DeliveryStatusOrig        string
	SloHint                   string // why current slo class
}

//func transSlodatas(slos []*slodata) []map[string]string {
//	result := make([]map[string]string, 0)
//
//	for _, val := range slos {
//		result = append(result, transSlodata(val))
//	}
//
//	return result
//}

func transSlodata(slo *slodata, env string) map[string]string {
	result := make(map[string]string)

	result["PodName"] = slo.PodName
	result["PodUID"] = slo.PodUID
	//if slo.IsJob {
	//	result["Pod类型"] = "Job类Pod(创建超时90秒)"
	//} else {
	//	result["Pod类型"] = "在线服务类Pod(创建超时10分钟)"
	//}

	result["PodType"] = time.Duration(slo.PodSLO).String()
	result["NodeIP"] = slo.NodeIP
	result["CreatedResult"] = slo.StartUpResultFromCreate
	result["DeliveryStatus"] = slo.DeliveryStatusOrig

	if slo.PossibleReason != "" {
		result["PossibleReason"] = slo.PossibleReason
	}
	result["Cluster"] = slo.Cluster
	//result["创建结果(从调度)"] = slo.StartUpResultFromSchedule
	result["Namespace"] = slo.Namespace
	if !slo.Created.IsZero() {
		result["CreatedTime"] = slo.Created.Format(time.Stamp)
	}
	if !slo.Scheduled.IsZero() {
		result["ScheduledTime"] = slo.Scheduled.Format(time.Stamp)
	}
	if !slo.FinishTime.IsZero() {
		result["FinishTime"] = slo.FinishTime.Format(time.Stamp)
	}
	if !slo.ContainersReady.IsZero() {
		result["ContainersReadyTime"] = slo.ContainersReady.Format(time.Stamp)
	}
	if !slo.RunningAt.IsZero() {
		result["Pod Running Time"] = slo.RunningAt.Format(time.Stamp)
	}
	if !slo.SucceedAt.IsZero() {
		result["SucceedAt"] = slo.SucceedAt.Format(time.Stamp)
	}
	if !slo.FailedAt.IsZero() {
		result["FailedAt"] = slo.FailedAt.Format(time.Stamp)
	}
	if !slo.ReadyAt.IsZero() {
		result["ReadyAt"] = slo.ReadyAt.Format(time.Stamp)
	}

	if !slo.DeleteEndTime.IsZero() {
		result["DeleteEndTime"] = slo.DeleteEndTime.Format(time.Stamp)

		if !slo.CreatedTime.IsZero() {
			result["DeletedTime"] = slo.CreatedTime.Format(time.Stamp)
		}
		if slo.DeleteResult != "" {
			result["DeleteResult"] = slo.DeleteResult
		}
	}

	if !slo.UpgradeEndTime.IsZero() {
		result["UpgradeEndTime"] = slo.UpgradeEndTime.Format(time.Stamp)

		if !slo.CreatedTime.IsZero() {
			result["UpgradeTime"] = slo.CreatedTime.Format(time.Stamp)
		}
		if slo.UpgradeResult != "" {
			result["UpgradeResult"] = slo.UpgradeResult
		}
	}

	return result
}

// 查询 slo_data 索引，用于 debugSLO 这个 URL
// 应该使用 slo_trace_data_daily
func (s *Server) querySloByResult(requestParams *sloReq) []*slodata {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QuerySloByResult", begin)
	}()

	if requestParams == nil || requestParams.Result == "" {
		return make([]*slodata, 0)
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("StartUpResultFromCreate: \"%s\"", requestParams.Result))

	query := elastic.NewBoolQuery().Must(stringQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("Created").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	}

	if requestParams.DeliveryStatus != "" {
		stringQuery4 := elastic.NewQueryStringQuery(fmt.Sprintf("DeliveryStatusOrig: \"%s\"", requestParams.DeliveryStatus))
		query = query.Must(stringQuery4)
	}

	if requestParams.SloTime != "" {
		sloduration, err := time.ParseDuration(requestParams.SloTime)

		if err == nil {
			stringQuery5 := elastic.NewQueryStringQuery(fmt.Sprintf("PodSLO: \"%d\"", int(sloduration)))
			query = query.Must(stringQuery5)
		} else {
			fmt.Printf("Error slotime format %s \n", requestParams.SloTime)
		}
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 500 {
		querySize = 500
	}

	returnResult := make([]*slodata, 0)
	searchReulst, err := s.ESClient.Search().Index(sloTraceDataIndexName).Type(sloDataTypeName).Query(query).Size(querySize).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchReulst.Hits.Hits {
		slo := &slodata{}
		if er := json.Unmarshal(hit.Source, slo); er == nil {
			returnResult = append(returnResult, slo)
		}
	}

	return returnResult
}

// 查询 slo_data 索引，用于 debugSLO 这个 URL
// 应该使用 slo_trace_data_daily
func (s *Server) queryDeleteSloByResult(requestParams *sloReq) []*slodata {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryDeleteSloByResult", begin)
	}()

	if requestParams == nil || requestParams.Result == "" {
		return make([]*slodata, 0)
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("DeleteResult: \"%s\"", requestParams.Result))
	query := elastic.NewBoolQuery().Must(stringQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	// Type: delete AND NOT DeleteResult: success AND NOT DeleteResult
	if requestParams.Type != "" {
		stringQuery4 := elastic.NewQueryStringQuery(fmt.Sprintf("Type: \"%s\"", requestParams.Type))
		query = query.Must(stringQuery4)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 5000 {
		querySize = 5000
	}

	returnResult := make([]*slodata, 0)
	searchResult, err := s.ESClient.Search().Index(sloTraceDataIndexName).Type(sloDataTypeName).Query(query).Size(querySize).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchResult.Hits.Hits {
		slo := &slodata{}
		if er := json.Unmarshal(hit.Source, slo); er == nil {
			returnResult = append(returnResult, slo)
		}
	}

	return returnResult
}

// 查询 slo_data 索引，用于 debugSLO 这个 URL
// 应该使用 slo_trace_data_daily
func (s *Server) queryUpgradeSloByResult(requestParams *sloReq) []*slodata {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryUpgradeSloByResult", begin)
	}()

	if requestParams == nil || requestParams.Result == "" {
		return make([]*slodata, 0)
	}

	resultQuery := elastic.NewQueryStringQuery(fmt.Sprintf("UpgradeResult: \"%s\"", requestParams.Result))
	query := elastic.NewBoolQuery().Must(resultQuery)

	if requestParams.Cluster != "" {
		query = query.Must(elastic.NewTermQuery("Cluster.keyword", requestParams.Cluster))
	}

	if requestParams.Type != "" {
		typeQuery := elastic.NewQueryStringQuery(fmt.Sprintf("Type: \"%s\"", requestParams.Type))
		query = query.Must(typeQuery)
	}

	// add range query
	if !requestParams.From.IsZero() || !requestParams.To.IsZero() {
		rangeQuery := elastic.NewRangeQuery("CreatedTime").TimeZone("UTC")
		if !requestParams.From.IsZero() {
			rangeQuery = rangeQuery.From(requestParams.From)
		}
		if !requestParams.To.IsZero() {
			rangeQuery = rangeQuery.To(requestParams.To)
		}
		query = query.Must(rangeQuery)
	}

	querySize := 300
	if requestParams.Count != "" {
		count, err := strconv.Atoi(requestParams.Count)
		if err == nil {
			querySize = count
		}
	}
	if querySize > 5000 {
		querySize = 5000
	}

	returnResult := make([]*slodata, 0)
	searchResult, err := s.ESClient.Search().Index(sloTraceDataIndexName).Type(sloDataTypeName).Query(query).Size(querySize).
		Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchResult.Hits.Hits {
		slo := &slodata{}
		if er := json.Unmarshal(hit.Source, slo); er == nil {
			returnResult = append(returnResult, slo)
		}
	}

	return returnResult
}

func (s *Server) queryPodPhaseWithPodUIDOrName(podUID string, podName string) []*podPhase {
	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", podName))
	if podUID != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", podUID))
	}
	query := elastic.NewBoolQuery().Must(stringQuery)

	returnResult := make([]*podPhase, 0)

	searchReulst, err := s.ESClient.Search().Index(podphaseInexName).Type(podphaseTypeName).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchReulst.Hits.Hits {
		podPha := &podPhase{}
		if er := json.Unmarshal(hit.Source, podPha); er == nil {
			returnResult = append(returnResult, podPha)
		}
	}

	return returnResult
}

func (s *Server) queryPodYamlWithPodUIDOrName(req *debugPodParam) []*podyaml {
	result := make([]*podyaml, 0)
	if req == nil {
		return result
	}

	var stringQuery *elastic.QueryStringQuery
	if req.podIP != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podIP: \"%s\"", req.podIP))
	}
	if req.podName != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", req.podName))
	}
	if req.podUID != "" {
		stringQuery = elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", req.podUID))
	}

	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := s.ESClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(100).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		pyaml := &podyaml{}
		if er := json.Unmarshal(hit.Source, pyaml); er == nil {
			if pyaml.Pod != nil {
				result = append(result, pyaml)
			}
		}
	}

	return result
}

func getPodUIDListByHostname(hostname string) []string {
	result := make([]string, 0)
	if hostname == "" {
		return result
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostname.keyword: \"%s\"", hostname))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		var jsonData interface{}
		var err error
		err = json.Unmarshal(hit.Source, &jsonData)
		if err != nil {
			continue
		}
		res, err := jsonpath.JsonPathLookup(jsonData, "$.podUID")
		if err != nil {
			continue
		}
		if str, ok := res.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func getPodUIDListByIP(ip string) []string {
	result := make([]string, 0)
	if ip == "" {
		return result
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podIP.keyword: \"%s\"", ip))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		var jsonData interface{}
		var err error
		err = json.Unmarshal(hit.Source, &jsonData)
		if err != nil {
			continue
		}
		res, err := jsonpath.JsonPathLookup(jsonData, "$.podUID")
		if err != nil {
			continue
		}
		if str, ok := res.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func getPodUIDListByPodName(podName string) []string {
	result := make([]string, 0)
	if podName == "" {
		return result
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", podName))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		var jsonData interface{}
		var err error
		err = json.Unmarshal(hit.Source, &jsonData)
		if err != nil {
			continue
		}
		res, err := jsonpath.JsonPathLookup(jsonData, "$.podUID")
		if err != nil {
			continue
		}
		if str, ok := res.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// 从 zsearch 中查询 pod_life_phase 索引，获取一个 Pod 的所有 trace
func queryPodphaseWithPodUID(podUID string) []*podPhase {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryPodphase", begin)
	}()

	returnResult := make([]*podPhase, 0)
	if podUID == "" {
		return returnResult
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", podUID))
	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := esClient.Search().Index(podphaseInexName).Type(podphaseTypeName).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchResult.Hits.Hits {
		podPha := &podPhase{}
		if er := json.Unmarshal(hit.Source, podPha); er == nil {
			podPha.PlfID = hit.Id
			returnResult = append(returnResult, podPha)
		}
	}

	return returnResult
}

func queryPodphaseWithPodName(podName string) []*podPhase {
	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podName: \"%s\"", podName))
	query := elastic.NewBoolQuery().Must(stringQuery)

	returnResult := make([]*podPhase, 0)
	searchReulst, err := esClient.Search().Index(podphaseInexName).Type(podphaseTypeName).Query(query).Size(200).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchReulst.Hits.Hits {
		podPha := &podPhase{}
		if er := json.Unmarshal(hit.Source, podPha); er == nil {
			podPha.PlfID = hit.Id

			returnResult = append(returnResult, podPha)
		}
	}

	return returnResult
}

func queryPodYamlWithPodUID(podUID string) []*podyaml {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryPodYaml", begin)
	}()

	result := make([]*podyaml, 0)
	if podUID == "" {
		return result
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", podUID))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		pyaml := &podyaml{}
		if er := json.Unmarshal(hit.Source, pyaml); er == nil {
			if pyaml.Pod != nil {
				result = append(result, pyaml)
			}
		}
	}

	return result
}

func QuerySloPodInfo(podUID string) []*xsearch.SloPodInfo {
	defer utils.IgnorePanic("QuerySloPodInfo")

	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QuerySloPodInfo", begin)
	}()

	result := []*xsearch.SloPodInfo{}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("podUID: \"%s\"", podUID))

	query := elastic.NewBoolQuery().Must(stringQuery)

	searchResult, err := esClient.Search().Index("slo_pod_info").Type("_doc").Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchResult.Hits.Hits {
		podInfo := &xsearch.SloPodInfo{}
		if er := json.Unmarshal(hit.Source, podInfo); er == nil {
			result = append(result, podInfo)
		}
	}

	return result
}

func queryPodYamlsWithNodeIP(nodeIP string, withDeleted bool) []*podyaml {
	result := make([]*podyaml, 0)
	if nodeIP == "" {
		return result
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("hostIP.keyword: \"%s\"", nodeIP))
	deleteFalse := elastic.NewQueryStringQuery(fmt.Sprintf("isDeleted.keyword: \"%t\"", withDeleted))
	query := elastic.NewBoolQuery().Must(stringQuery).Must(deleteFalse)
	searchResult, err := esClient.Search().Index(podYamlIndexName).Type(podYamlTypeName).Query(query).Size(300).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	dedup := make(map[string]string)
	for _, hit := range searchResult.Hits.Hits {
		pyaml := &podyaml{}
		if er := json.Unmarshal(hit.Source, pyaml); er == nil {
			if pyaml.Pod != nil {
				if _, ok := dedup[pyaml.PodUID]; !ok {
					result = append(result, pyaml)
					dedup[pyaml.PodUID] = "true"
				}
			}
		}
	}

	return result
}

func queryPreemptionMasterByVictimPodNam(podName string) *podPhase {
	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery("operationName: \"event\""))
	query = query.Must(elastic.NewQueryStringQuery("extraInfo.reason: \"PreemptionSuccess\""))
	query = query.Must(elastic.NewQueryStringQuery(fmt.Sprintf("extraInfo.victims: \"%s\"", podName)))

	searchReulst, err := esClient.Search().Index(podphaseInexName).Type(podphaseTypeName).Query(query).Size(1).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return nil
	}

	for _, hit := range searchReulst.Hits.Hits {
		podPha := &podPhase{}
		if er := json.Unmarshal(hit.Source, podPha); er == nil {
			return podPha
		}
	}

	return nil
}

func queryPodLifePhaseByID(docId string) *podPhase {
	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery(fmt.Sprintf("_id: \"%s\"", docId)))
	searchReulst, err := esClient.Search().Index(podphaseInexName).Type(podphaseTypeName).Query(query).Size(1).
		Sort("startTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return nil
	}

	for _, hit := range searchReulst.Hits.Hits {
		podPha := &podPhase{}
		if er := json.Unmarshal(hit.Source, podPha); er == nil {
			podPha.PlfID = hit.Id
			return podPha
		}
	}

	return nil
}

func queryAuditLogByID(auditID string) *k8saudit.Event {
	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery(fmt.Sprintf("auditID: \"%s\"", auditID)))
	searchReulst, err := esClient.Search().Index("audit_*").Type("doc").Query(query).Size(1).
		Sort("stageTimestamp", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return nil
	}

	for _, hit := range searchReulst.Hits.Hits {
		auditLog := &k8saudit.Event{}
		if er := json.Unmarshal(hit.Source, auditLog); er == nil {
			return auditLog
		}
	}

	return nil
}

func querySpansByUID(uid string) []*span {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QuerySpans", begin)
	}()

	result := make([]*span, 0)
	if uid == "" {
		return result
	}

	query := elastic.NewBoolQuery().Must(elastic.NewQueryStringQuery(fmt.Sprintf("OwnerRef.UID: \"%s\" AND OwnerRef.Resource: pods", uid)))
	searchReulst, err := esClient.Search().Index(spans.SpanIndex).Type(spans.SpanDocType).Query(query).Size(80).
		Sort("Elapsed", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return result
	}

	for _, hit := range searchReulst.Hits.Hits {
		originalSpan := &spans.Span{}
		if er := json.Unmarshal(hit.Source, originalSpan); er == nil {
			result = append(result, &span{
				Name:       originalSpan.Name,
				Type:       originalSpan.Type,
				Begin:      originalSpan.Begin,
				End:        originalSpan.End,
				Elapsed:    (time.Duration(originalSpan.Elapsed) * time.Millisecond).String(),
				ActionType: originalSpan.ActionType,
			})
		}
	}

	return result
}
