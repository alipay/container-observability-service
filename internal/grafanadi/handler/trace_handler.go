package handler

import (
	"net/http"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type TraceHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *TraceParams
	storage       data_access.StorageInterface
}

type TraceParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *TraceHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *TraceHandler) ParseRequest() error {
	params := TraceParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *TraceHandler) ValidRequest() error {

	return nil
}

func (handler *TraceHandler) QueryTraceWithPodUid(key, value string) (int, interface{}, error) {
	spans := make([]*model.Span, 0)
	podYamls := make([]*model.PodYaml, 0)

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryTrace").Observe(cost)
	}()

	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)

	err := handler.storage.QuerySpanWithPodUid(&spans, value)
	if err != nil {
		return http.StatusOK, nil, err
	}

	datafram := service.ConverSpan2Frame(spans)

	return http.StatusOK, datafram, nil

}

func (handler *TraceHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("TraceHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	httpStatus, result, err = handler.QueryTraceWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)

	return httpStatus, result, err
}

func TraceFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &TraceHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
