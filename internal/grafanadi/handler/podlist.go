package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	customerrors "github.com/alipay/container-observability-service/pkg/custom-errors"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type PodListHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *model.SloOptions
	storage       data_access.StorageInterface
}

func (handler *PodListHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodListHandler) ParseRequest() error {
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

func (handler *PodListHandler) ValidRequest() error {
	req := handler.requestParams
	if req.Type == "" {
		return fmt.Errorf("type needed")
	}
	return nil
}

func (handler *PodListHandler) QueryDebuggingWithType(debugparams model.SloOptions) (int, interface{}, error) {
	var slodatas []*model.Slodata
	var err error
	var httpStatus int
	var deliveryTable interface{}

	if handler.requestParams.Type == "Delete" {
		httpStatus, slodatas, err = handler.queryDeleteSloByResult(handler.requestParams)
	} else if strings.Contains(handler.requestParams.Type, "Upgrade") {
		httpStatus, slodatas, err = handler.queryUpgradeSloByResult(handler.requestParams)
	} else {
		httpStatus, slodatas, err = handler.querySloByResult(handler.requestParams)
	}

	podList := make([]map[string]string, 0)

	for _, slo := range slodatas {
		dic := transSlodata(slo, handler.requestParams.Env)
		podList = append(podList, dic)

	}

	deliveryTable = service.ConvertPodList2Frame(podList)

	return httpStatus, deliveryTable, err
}

func (handler *PodListHandler) queryDeleteSloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {

	res := make([]*model.SloTraceData, 0)
	returnResult := make([]*model.Slodata, 0)

	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, returnResult, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
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

func (handler *PodListHandler) queryUpgradeSloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {

	res := make([]*model.SloTraceData, 0)
	returnResult := make([]*model.Slodata, 0)

	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, returnResult, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebuggingWithPodUid").Observe(cost)
	}()

	err := handler.storage.QueryUpgradeSloWithResult(&res, requestParams)
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
func (handler *PodListHandler) querySloByResult(requestParams *model.SloOptions) (int, []*model.Slodata, error) {
	// return http.StatusOK, result, nil
	returnResult := make([]*model.Slodata, 0)
	if requestParams == nil || requestParams.Result == "" {
		return http.StatusOK, returnResult, customerrors.Error(customerrors.ErrParams, customerrors.NoDeliveryResult)
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebuggingWithPodUid").Observe(cost)
	}()

	err := handler.storage.QueryCreateSloWithResult(&returnResult, requestParams)
	if err != nil {
		return http.StatusOK, returnResult, fmt.Errorf("QueryDebuggingWithPodUid error, error is %s", err)
	}

	return http.StatusOK, returnResult, nil
}

func (handler *PodListHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("DebuggingHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryDebuggingWithType(*handler.requestParams)

	return httpStatus, result, err
}

func PodlistFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodListHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
