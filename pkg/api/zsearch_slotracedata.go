package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/olivere/elastic"
	"k8s.io/klog/v2"
)

type sloTraceData struct {
	Type                      string //类型：pod创建、pod删除、pod升级等
	Cluster                   string
	Namespace                 string
	PodName                   string
	PodUID                    string
	NodeIP                    string
	DeleteResult              string //删除结果
	PullTimeoutImageName      string
	IsJob                     bool
	StartUpResultFromCreate   string
	StartUpResultFromSchedule string
	DebugUrl                  string
	UpgradeResult             string
	CreatedTime               time.Time
	Scheduled                 time.Time
	ContainersReady           time.Time
	RunningAt                 time.Time
	SucceedAt                 time.Time
	FailedAt                  time.Time
	ReadyAt                   time.Time
	DeletedTime               time.Time
	FinishTime                time.Time
	UpgradeEndTime            time.Time
	PossibleReason            string
	PodSLO                    int64
	DeliverySLO               int64
	SLOViolationReason        string
	DeliveryStatus            string
	DeliveryStatusOrig        string
	SloHint                   string // why current slo class
}

func querySloTraceDataByPodUID(podUID string) []*sloTraceData {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QuerySloTraceData", begin)
	}()

	returnResult := make([]*sloTraceData, 0)
	if podUID == "" {
		return returnResult
	}

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("PodUID: \"%s\"", podUID))
	query := elastic.NewBoolQuery().Must(stringQuery)
	searchResult, err := esClient.Search().Index("slo_trace_data_daily").Type("data").Query(query).
		Size(200).Sort("CreatedTime", false).Do(context.Background())
	if err != nil {
		klog.Error(err)
		return returnResult
	}

	for _, hit := range searchResult.Hits.Hits {
		slo := &sloTraceData{}
		if er := json.Unmarshal(*hit.Source, slo); er == nil {
			returnResult = append(returnResult, slo)
		}
	}

	return returnResult
}
