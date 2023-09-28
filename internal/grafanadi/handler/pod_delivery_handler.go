package handler

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type PodDeliveryHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodDeliveryParams
	storage       data_access.StorageInterface
}

type PodDeliveryParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *PodDeliveryHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodDeliveryHandler) ParseRequest() error {
	params := PodDeliveryParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodDeliveryHandler) ValidRequest() error {

	return nil
}

func (handler *PodDeliveryHandler) QueryPodDeliveryWithPodUid(key, value string) (int, interface{}, error) {
	sloTraces := make([]*model.SloTraceData, 0)
	podYamls := make([]*model.PodYaml, 0)

	var deliveryTable interface{}

	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)

	err := handler.storage.QuerySloTraceDataWithPodUID(&sloTraces, value)
	if err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodInfoWithPodUid error, error is %s", err)
	}
	if len(sloTraces) == 0 {
		return http.StatusOK, nil, errors.New("query no data")
	}
	switch handler.request.URL.Path {
	case "/deliverypodcreatetable":
		deliveryTable = service.ConverPodCreate2Frame(sloTraces)
	case "/deliverypoddeletetable":
		deliveryTable = service.ConvertPodDelete2Frame(sloTraces)
	}

	return http.StatusOK, deliveryTable, nil

}

func (handler *PodDeliveryHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("PodDeliveryHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryPodDeliveryWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func PodDeliveryFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodDeliveryHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
