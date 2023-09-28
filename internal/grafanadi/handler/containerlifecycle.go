package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type ContainerlifecycleHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *ContainerlifecycleParams
	storage       data_access.StorageInterface
}

func (handler *ContainerlifecycleHandler) GetContainerLifecycleData(key, value string) (int, interface{}, error) {
	lifephases := make([]*storagemodel.LifePhase, 0)
	slolist := make([]*storagemodel.SloTraceData, 0)
	if len(value) == 0 {
		return 0, nil, nil
	}
	if err := handler.storage.QueryLifePhaseWithPodUid(&lifephases, value); err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryLifePhaseWithPodUid error, error is %s", err)
	}
	if err := handler.storage.QuerySloTraceDataWithPodUID(&slolist, value); err != nil {
		return http.StatusOK, nil, fmt.Errorf("QuerySloTraceDataWithPodUID error, error is %s", err)
	}
	if len(slolist) < 1 || len(lifephases) < 1 {
		return http.StatusOK, nil, errors.New("query no data")
	}
	dataFrame := service.ConvertLifePhase2State(lifephases, slolist)
	return http.StatusOK, dataFrame, nil
}

type ContainerlifecycleParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *ContainerlifecycleHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *ContainerlifecycleHandler) ParseRequest() error {
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

func (handler *ContainerlifecycleHandler) ValidRequest() error {

	return nil
}

func (handler *ContainerlifecycleHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("ContainerlifecycleHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.GetContainerLifecycleData(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func ContainerlifecycleFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &ContainerlifecycleHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
