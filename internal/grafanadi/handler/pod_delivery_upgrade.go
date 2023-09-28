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

type PodDeliveryUpgradeHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodDeliveryUpgradeParams
	storage       data_access.StorageInterface
}

type PodDeliveryUpgradeParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *PodDeliveryUpgradeHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodDeliveryUpgradeHandler) ParseRequest() error {
	params := PodDeliveryUpgradeParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodDeliveryUpgradeHandler) ValidRequest() error {
	return nil
}

func (handler *PodDeliveryUpgradeHandler) QueryPodDeliveryUpgradeWithPodUid(key, value string) (int, interface{}, error) {
	sloTraces := make([]*model.SloTraceData, 0)
	podYamls := make([]*model.PodYaml, 0)

	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodeDelivery").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)

	err := handler.storage.QuerySloTraceDataWithPodUID(&sloTraces, value)
	if err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodInfoWithPodUid error, error is %s", err)
	}

	datafram := service.ConvertPodUpgrade2Frame(sloTraces)
	return http.StatusOK, datafram, nil

}

func (handler *PodDeliveryUpgradeHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("PodDeliveryUpgradeHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryPodDeliveryUpgradeWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func PodDeliveryUpgradeFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodDeliveryUpgradeHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
