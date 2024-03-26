package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	customerrors "github.com/alipay/container-observability-service/pkg/custom-errors"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/klog/v2"
)

type DebuggingPodsHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *model.SloOptions
	storage       data_access.StorageInterface
}

func (handler *DebuggingPodsHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *DebuggingPodsHandler) ParseRequest() error {
	r := handler.request
	params := model.SloOptions{}
	if r.Method == http.MethodGet {
		params.Type = r.URL.Query().Get("type")
		params.Cluster = r.URL.Query().Get("cluster")
		params.BizName = r.URL.Query().Get("bizname")
		params.Result = r.URL.Query().Get("deliveryresult")
		setTPLayout(r.URL.Query(), "from", &params.From)
		setTPLayout(r.URL.Query(), "to", &params.To)
	}
	handler.requestParams = &params
	return nil
}

func (handler *DebuggingPodsHandler) ValidRequest() error {

	return nil
}

func (handler *DebuggingPodsHandler) QueryDebuggingWithType(debugparams model.SloOptions) (int, interface{}, error) {
	var slodatas []*model.Slodata
	var err error
	var httpStatus int
	var deliveryTable interface{}
	if debugparams.Result == "" {
		return httpStatus, nil, err
	}

	if handler.requestParams.Type == "Delete" {
		httpStatus, slodatas, err = handler.queryDeleteSloByResult(handler.requestParams)
	} else if strings.Contains(handler.requestParams.Type, "Upgrade") {
		httpStatus, slodatas, err = handler.queryUpgradeSloByResult(handler.requestParams)
	} else {
		httpStatus, slodatas, err = handler.querySloByResult(handler.requestParams)
	}

	namespaceCount := make(map[string]int)
	nodeCount := make(map[string]int)
	imageCount := make(map[string]int)
	podTypeCount := make(map[string]int)
	clusterCount := make(map[string]int)
	for _, slo := range slodatas {

		if v, ok := namespaceCount[slo.Namespace]; ok {
			namespaceCount[slo.Namespace] = v + 1
		} else {
			namespaceCount[slo.Namespace] = 1
		}
		nodeKey := fmt.Sprintf("%s/%s", slo.NodeName, slo.NodeIP)
		if v, ok := nodeCount[nodeKey]; ok {
			nodeCount[nodeKey] = v + 1
		} else {
			nodeCount[nodeKey] = 1
		}
		if v, ok := podTypeCount[time.Duration(slo.PodSLO).String()]; ok {
			podTypeCount[time.Duration(slo.PodSLO).String()] = v + 1
		} else {
			podTypeCount[time.Duration(slo.PodSLO).String()] = 1
		}
		if v, ok := clusterCount[slo.Cluster]; ok {
			clusterCount[slo.Cluster] = v + 1
		} else {
			clusterCount[slo.Cluster] = 1
		}
		if v, ok := imageCount[slo.PullTimeoutImageName]; ok {
			imageCount[slo.PullTimeoutImageName] = v + 1
		} else {
			imageCount[slo.PullTimeoutImageName] = 1
		}
	}
	switch handler.request.URL.Path {
	case "/clusterdistribute":
		deliveryTable = service.ConvertClusterDistribute2Frame(clusterCount)
	case "/namespacedistribute":
		deliveryTable = service.ConvertNameSpaceDistribute2Frame(namespaceCount)
	case "/nodedistribute":
		deliveryTable = service.ConvertNodeDistribute2Frame(nodeCount)
	case "/podtypedistribute":
		deliveryTable = service.ConvertSloDistribute2Frame(podTypeCount)
	}

	return httpStatus, deliveryTable, err
}

func (handler *DebuggingPodsHandler) queryDeleteSloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {

	res := make([]*model.SloTraceData, 0)
	returnResult := make([]*model.Slodata, 0)

	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, nil, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebuggingWithPodUid").Observe(cost)
	}()

	err := handler.storage.QueryDeleteSloWithResult(&res, requestParams)
	if err != nil {
		return http.StatusOK, returnResult, fmt.Errorf("QueryDebuggingWithPodUid error, error is %s", err)
	}
	for _, v := range res {
		slo := &model.Slodata{}
		by, err := json.Marshal(v)
		if err == nil {
			if er := json.Unmarshal(by, slo); er == nil {
				returnResult = append(returnResult, slo)
			}
		}
	}

	return http.StatusOK, returnResult, nil
}

func (handler *DebuggingPodsHandler) queryUpgradeSloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {

	res := make([]*model.SloTraceData, 0)
	returnResult := make([]*model.Slodata, 0)

	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, nil, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebuggingWithPodUid").Observe(cost)
	}()

	err := handler.storage.QueryUpgradeSloWithResult(&res, requestParams)
	if err != nil {
		return http.StatusBadRequest, returnResult, fmt.Errorf("QueryDebuggingWithPodUid error, error is %s", err)
	}
	for _, v := range res {
		slo := &model.Slodata{}
		by, err := json.Marshal(v)
		if err == nil {
			if er := json.Unmarshal(by, slo); er == nil {
				returnResult = append(returnResult, slo)
			}
		}
	}

	return http.StatusOK, returnResult, nil
}
func (handler *DebuggingPodsHandler) querySloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {
	// return http.StatusOK, result, nil
	returnResult := make([]*model.Slodata, 0)
	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, nil, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebuggingWithPodUid").Observe(cost)
	}()

	err := handler.storage.QueryCreateSloWithResult(&returnResult, requestParams)
	if err != nil {
		return http.StatusBadRequest, returnResult, fmt.Errorf("QueryDebuggingWithPodUid error, error is %s", err)
	}

	return http.StatusOK, returnResult, nil
}

func (handler *DebuggingPodsHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("DebuggingHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryDebuggingWithType(*handler.requestParams)

	return httpStatus, result, err
}

func DebuggingPodsFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &DebuggingPodsHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
func setTPLayout(values url.Values, name string, f *time.Time) {
	s := values.Get(name)
	layOut := "2006-01-02T15:04:05"
	if s != "" {
		data, err := strconv.ParseInt(s, 10, 64)
		nowTime := time.Unix(data/1000, 0).Format("2006-01-02T15:04:05")
		t, err := time.ParseInLocation(layOut, nowTime, time.Local)
		if err != nil {
			klog.Errorf("failed to parse time %s for %s: %s", s, name, err.Error())
		} else {
			*f = t
		}
	}
}
func transSlodata(slo *model.Slodata, env string) map[string]string {
	result := make(map[string]string)

	result["PodName"] = slo.PodName
	result["PodUID"] = slo.PodUID
	result["PodDebugUrl"] = "诊断"

	result["SLO"] = time.Duration(slo.PodSLO).String()
	result["NodeIP"] = slo.NodeIP
	result["CreatedResult"] = slo.StartUpResultFromCreate
	result["DeliveryStatus"] = slo.DeliveryStatusOrig

	if slo.SLOViolationReason != "" {
		result["SLOViolationReason"] = slo.SLOViolationReason
	}
	result["Cluster"] = slo.Cluster
	//result["创建结果(从调度)"] = slo.StartUpResultFromSchedule
	result["Namespace"] = slo.Namespace
	if !slo.Created.IsZero() {
		result["CreatedTime"] = getTime(slo.Created.Format(time.RFC3339Nano))
	}
	if !slo.Scheduled.IsZero() {
		result["ScheduledTime"] = getTime(slo.Scheduled.Format(time.RFC3339Nano))
	}
	if !slo.FinishTime.IsZero() {
		result["FinishTime"] = getTime(slo.FinishTime.Format(time.RFC3339Nano))

	}
	if !slo.ContainersReady.IsZero() {
		result["ContainersReadyTime"] = getTime(slo.ContainersReady.Format(time.RFC3339Nano))
	}
	if !slo.RunningAt.IsZero() {
		result["PodRunningTime"] = getTime(slo.RunningAt.Format(time.RFC3339Nano))
	}
	if !slo.SucceedAt.IsZero() {
		result["SucceedAt"] = getTime(slo.SucceedAt.Format(time.RFC3339Nano))
	}
	if !slo.FailedAt.IsZero() {
		result["FailedAt"] = getTime(slo.FailedAt.Format(time.RFC3339Nano))
	}
	if !slo.ReadyAt.IsZero() {
		result["ReadyAt"] = getTime(slo.ReadyAt.Format(time.RFC3339Nano))
	}

	if !slo.DeleteEndTime.IsZero() {
		result["DeleteEndTime"] = getTime(slo.DeleteEndTime.Format(time.RFC3339Nano))

		if !slo.CreatedTime.IsZero() {
			result["DeleteTime"] = getTime(slo.CreatedTime.Format(time.RFC3339Nano))
		}
		if slo.DeleteResult != "" {
			result["DeleteResult"] = slo.DeleteResult
		}
	}

	if !slo.UpgradeEndTime.IsZero() {
		result["UpgradeEndTime"] = getTime(slo.UpgradeEndTime.Format(time.RFC3339Nano))

		if !slo.CreatedTime.IsZero() {
			result["UpgradeTime"] = getTime(slo.CreatedTime.Format(time.RFC3339Nano))
		}
		if slo.UpgradeResult != "" {
			result["UpgradeResult"] = slo.UpgradeResult
		}
	}

	return result
}
func getTime(str string) string {
	tt, _ := time.Parse("2006-01-02T15:04:05Z07:00", str)
	return tt.Format("2006-01-02 15:04:05")
}
