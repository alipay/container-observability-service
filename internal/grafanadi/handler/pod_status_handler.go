package handler

import (
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

type PodStatusHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodStatusParams
	storage       data_access.StorageInterface
}

type PodStatusParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *PodStatusHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodStatusHandler) ParseRequest() error {
	params := PodStatusParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodStatusHandler) ValidRequest() error {
	return nil
}

func (handler *PodStatusHandler) QueryPodStatusWithPodUid(key, value string) (int, interface{}, error) {
	podInfos := make([]*model.PodInfo, 0)
	podYamls := make([]*model.PodYaml, 0)

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

	errInfo := handler.storage.QueryPodInfoWithPodUid(&podInfos, value)
	errYaml := handler.storage.QueryPodYamlsWithPodUID(&podYamls, value)

	if errInfo != nil && errYaml != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodInfoWithPodUid error, error is %s", errInfo)
	}
	if len(podInfos) < 1 || len(podYamls) < 1 {
		return http.StatusOK, nil, nil
	}
	dataFrame := service.ConvertPodStatus2Frame(*podInfos[0], *podYamls[0])
	return http.StatusOK, dataFrame, nil

}

func (handler *PodStatusHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("PodStatusHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryPodStatusWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func PodStatusFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodStatusHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
