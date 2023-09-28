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

type PodPhaseHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodPhaseParams
	storage       data_access.StorageInterface
}

type PodPhaseParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *PodPhaseHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodPhaseHandler) ParseRequest() error {
	params := PodPhaseParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodPhaseHandler) ValidRequest() error {
	return nil
}

func (handler *PodPhaseHandler) QueryPodPhaseWithPodUid(key, value string) (int, interface{}, error) {
	podPhases := make([]*model.LifePhase, 0)
	podYamls := make([]*model.PodYaml, 0)

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()
	if len(value) != 0 {
		util := interutils.Util{
			Storage: handler.storage,
		}
		util.GetUid(podYamls, key, &value)
		err := handler.storage.QueryLifePhaseWithPodUid(&podPhases, value)

		if err != nil {
			return http.StatusOK, nil, fmt.Errorf("QueryLifePhaseWithPodUid error, error is %s", err)
		}
	}

	datafram := service.ConvertPodPhase2Frame(podPhases)
	return http.StatusOK, datafram, nil

}

func (handler *PodPhaseHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("PodPhaseHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryPodPhaseWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)

	return httpStatus, result, err
}

func PodPhaseFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodPhaseHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
