package handler

import (
	"context"
	"fmt"
	eavesmodel "github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/olivere/elastic/v7"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

type OwnerPodMapHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *OwnerPodMapParams
	storage       data_access.StorageInterface
}

func (handler *OwnerPodMapHandler) GetOwnerPodMap(debugfrom, key, value string) (int, interface{}, error) {
	sloTraceData := make([]*model.SloTraceData, 0)
	result := []model.SloTraceData{}
	podYamls := make([]*model.PodYaml, 0)
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("GetOwnerPodMap").Observe(cost)
	}()
	if debugfrom == "pod" {
		// get owneref pod with pod key/value
		util := interutils.Util{
			Storage: handler.storage,
		}
		py, err := util.GetPodYaml(podYamls, key, value)
		if err != nil || len(py) == 0 {
			return http.StatusOK, eavesmodel.DataFrame{}, err
		}
		if py[0].Pod == nil {
			return http.StatusOK, eavesmodel.DataFrame{}, err
		}

		if len(py[0].Pod.OwnerReferences) != 0 {
			or := py[0].Pod.OwnerReferences[0]
			value = string(or.UID)
		}
	} else {
		switch key {
		case "name":
			uid, err := findUniqueId(value, handler.storage)
			klog.Info("uid is %s", uid)
			if err != nil {
				klog.Errorf("findUniqueId error, error is %s", err)
				return http.StatusOK, eavesmodel.DataFrame{}, err
			}
			value = uid
		default:
			fmt.Println("currently only supports uid or name")
			return http.StatusOK, eavesmodel.DataFrame{}, nil
		}
	}
	if value == "" {
		return http.StatusOK, eavesmodel.DataFrame{}, nil
	}
	err := handler.storage.QuerySloTraceDataWithOwnerId(&sloTraceData, value,
		model.WithFrom(handler.requestParams.From),
		model.WithTo(handler.requestParams.To),
		model.WithLimit(1000))
	if err != nil {
		return http.StatusOK, eavesmodel.DataFrame{}, fmt.Errorf("QuerySloTraceDataWithOwnerId error, error is %s", err)
	}
	for _, std := range sloTraceData {
		if std.Type == "create" || std.Type == "delete" {
			found := false
			for i, pod := range result {
				if pod.PodUID == std.PodUID {
					if std.Type == "create" {
						result[i].CreatedTime = std.CreatedTime
						result[i].OwnerRefStr = std.OwnerRefStr
						if std.RunningAt.After(std.ReadyAt) {
							std.ReadyAt = std.RunningAt
						}
						result[i].ReadyAt = std.ReadyAt
						result[i].SLOViolationReason = std.SLOViolationReason
					} else {
						result[i].DeletedTime = std.CreatedTime
						result[i].DeleteEndTime = std.DeleteEndTime
						result[i].DeleteResult = std.DeleteResult
					}
					found = true
				}
			}
			if !found {
				if std.RunningAt.After(std.ReadyAt) {
					std.ReadyAt = std.RunningAt
				}
				if std.Type == "delete" {
					std.DeletedTime = std.CreatedTime
				}
				result = append(result, *std)
			}
		}
	}
	return http.StatusOK, service.ConvertSloDataTrace2Graph(result), nil
}

type OwnerPodMapParams struct {
	Key       string
	Value     string
	DebugFrom string

	From time.Time // range query
	To   time.Time // range query

}

func (handler *OwnerPodMapHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *OwnerPodMapHandler) ParseRequest() error {
	params := OwnerPodMapParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		debugfrom := handler.request.URL.Query().Get("debugfrom")
		params.Key = key
		params.Value = value
		params.DebugFrom = debugfrom

		setTPLayout(handler.request.URL.Query(), "from", &params.From)
		setTPLayout(handler.request.URL.Query(), "to", &params.To)
	}

	handler.requestParams = &params
	return nil
}

func (handler *OwnerPodMapHandler) ValidRequest() error {

	return nil
}

func OwnerPodMapFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &OwnerPodMapHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}

func (handler *OwnerPodMapHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("ContainerlifecycleHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.GetOwnerPodMap(handler.requestParams.DebugFrom, handler.requestParams.Key, handler.requestParams.Value)

	return httpStatus, result, err
}

func findUniqueId(workloadName string, storage data_access.StorageInterface) (uid string, err error) {
	esClient, ok := storage.(*data_access.StorageEsImpl)
	if !ok {
		err = fmt.Errorf("parse errror")
		return
	}
	query := elastic.NewBoolQuery().
		Must(
			elastic.NewTermQuery("ExtraProperties.ownerref.name.Value.keyword", workloadName),
			elastic.NewExistsQuery("ExtraProperties.ownerref.uid.Value.keyword"),
			elastic.NewExistsQuery("ExtraProperties.ownerref.name.Value.keyword"),
		)
	aggs := elastic.NewTermsAggregation().
		Field("ExtraProperties.ownerref.name.Value.keyword").
		SubAggregation("group_by_ownerref_uid", elastic.NewTermsAggregation().Field("ExtraProperties.ownerref.uid.Value.keyword"))

	searchResult, err := esClient.DB.Search().
		Index("slo_trace_data_daily").
		Query(query).
		Size(0).
		Aggregation("group_by_ownerref_name", aggs).
		Do(context.Background())
	if err != nil {
		err = fmt.Errorf("failed to execute search query: %v", err)
		klog.Errorf("Failed to execute search query: %v", err)
		return
	}

	if agg, found := searchResult.Aggregations.Terms("group_by_ownerref_name"); found {
		for _, bucket := range agg.Buckets {
			if uidAgg, uidFound := bucket.Aggregations.Terms("group_by_ownerref_uid"); uidFound {
				for _, detail := range uidAgg.Buckets {
					if strKey, ok := detail.Key.(string); ok {
						return strKey, nil
					} else {
						return "", fmt.Errorf("workload uid key is not a string")
					}
				}
			}
			break
		}
	} else {
		klog.Infof("No aggs aggregation found")
	}
	return
}
