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

type BizInfoHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *BizInfoParams
	storage       data_access.StorageInterface
}

type BizInfoParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *BizInfoHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *BizInfoHandler) ParseRequest() error {
	params := BizInfoParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *BizInfoHandler) ValidRequest() error {
	return nil
}

func (handler *BizInfoHandler) QueryBizInfoWithPodUid(key, value string) (int, interface{}, error) {
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

	if errInfo != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodInfoWithPodUid error, error is %s", errInfo)
	}

	dataFrame := service.ConvertBizInfo2Frame(podInfos)
	return http.StatusOK, dataFrame, nil

}

func (handler *BizInfoHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("BizInfoHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryBizInfoWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}
	return httpStatus, result, err
}

func BizInfoFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &BizInfoHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
