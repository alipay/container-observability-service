package handler

import (
	"net/http"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/klog/v2"
)

type RawHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *RawdataParams
	storage       data_access.StorageInterface
}

type RawdataParams struct {
	plfId string
}

func (handler *RawHandler) queryPodLifePhaseByID(plfId string) (int, interface{}, error) {
	lifePhase := make([]*model.LifePhase, 0)
	if len(plfId) == 0 {
		return http.StatusOK, nil, nil

	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("PodLifePhaseByID").Observe(cost)
	}()

	err := handler.storage.QueryPodLifePhaseByID(&lifePhase, handler.requestParams.plfId)
	if err != nil {
		klog.Errorf("query lifephase failed: %s", err.Error())
		return http.StatusOK, nil, nil

	}
	if len(lifePhase) > 0 {
		return http.StatusOK, service.ConvertPodphase2Frame(lifePhase), nil
	}

	return http.StatusOK, nil, nil
}

func (handler *RawHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *RawHandler) ParseRequest() error {
	params := RawdataParams{}
	if handler.request.Method == http.MethodGet {

		params.plfId = handler.request.URL.Query().Get("plfid")
	}

	handler.requestParams = &params
	return nil
}

func (handler *RawHandler) ValidRequest() error {
	return nil
}

func (handler *RawHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("RawdataleHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams != nil {
		httpStatus, result, err = handler.queryPodLifePhaseByID(handler.requestParams.plfId)

	}
	return httpStatus, result, err
}

func RawdataFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &RawHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
