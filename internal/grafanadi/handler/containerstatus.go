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
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type ContainerStatusHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *ContainerlifecycleParams
	storage       data_access.StorageInterface
}

func (handler *ContainerStatusHandler) GetContainerStatusData(key, value string) (int, interface{}, error) {
	slolist := make([]*storagemodel.SloTraceData, 0)
	podYamls := make([]*model.PodYaml, 0)
	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("GetContainerStatusData").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)

	if err := handler.storage.QuerySloTraceDataWithPodUID(&slolist, value); err != nil {
		return http.StatusOK, nil, fmt.Errorf("QuerySloTraceDataWithPodUID error, error is %s", err)
	}
	lifephases := make([]*storagemodel.LifePhase, 0)
	if err := handler.storage.QueryLifePhaseWithPodUid(&lifephases, value); err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryLifePhaseWithPodUid error, error is %s", err)
	}
	if len(slolist) < 1 || len(lifephases) < 1 {
		return http.StatusOK, nil, errors.New("query no data")
	}
	stateDataSlice := service.AddStatusFromSloData(slolist, lifephases)

	return http.StatusOK, stateDataSlice, nil
}

type ContainerStatusHandlerParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *ContainerStatusHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *ContainerStatusHandler) ParseRequest() error {
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

func (handler *ContainerStatusHandler) ValidRequest() error {
	return nil
}

func (handler *ContainerStatusHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("ContainerStatusHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.GetContainerStatusData(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func ContainerStatusFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &ContainerStatusHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
