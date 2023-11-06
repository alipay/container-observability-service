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

type ContainerEventsHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *ContainerlifecycleParams
	storage       data_access.StorageInterface
}

func (handler *ContainerEventsHandler) GetContainerEventsData(key, value string) (int, interface{}, error) {
	podYamls := make([]*model.PodYaml, 0)
	podLifephases := make([]*model.LifePhase, 0)
	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("GetContainerEventsData").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)

	if err := handler.storage.QueryLifePhaseWithPodUid(&podLifephases, value); err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryLifePhaseWithPodUid error, error is %s", err)
	}
	if len(podLifephases) < 1 {
		return http.StatusOK, nil, errors.New("query no data")
	}
	stateDataSlice := service.AddStatusFromPodLifePhase(podLifephases)

	return http.StatusOK, stateDataSlice, nil
}

type ContainerEventsHandlerParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *ContainerEventsHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *ContainerEventsHandler) ParseRequest() error {
	params := ContainerlifecycleParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *ContainerEventsHandler) ValidRequest() error {

	return nil
}

func (handler *ContainerEventsHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("ContainerEventsHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.GetContainerEventsData(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func ContainerEventsFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &ContainerEventsHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
