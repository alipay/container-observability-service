package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type DebugPodListHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *DebugPodListParams
	storage       data_access.StorageInterface
}

type DebugPodListParams struct {
	SearchKey   string
	SearchValue string
}

func (handler *DebugPodListHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *DebugPodListHandler) ParseRequest() error {
	params := DebugPodListParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.SearchKey = key
		params.SearchValue = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *DebugPodListHandler) ValidRequest() error {

	reqParam := handler.requestParams
	switch reqParam.SearchKey {
	case "uid", "name", "hostname", "podip":
		return nil
	default:
		return fmt.Errorf("uid, name, podip, hostname are all empty")
	}
}

func (handler *DebugPodListHandler) QueryDebugPodListWithPodUid(key, value string) (int, interface{}, error) {
	podYamls := make([]*model.PodYaml, 0)
	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryDebugPodListWithPodUid").Observe(cost)
	}()

	var err error
	if key == "uid" {
		err = handler.storage.QueryPodYamlsWithPodUID(&podYamls, value)
	} else if key == "name" {
		err = handler.storage.QueryPodYamlsWithPodName(&podYamls, value)
	} else if key == "hostname" {
		err = handler.storage.QueryPodYamlsWithHostName(&podYamls, value)
	} else if key == "podip" {
		err = handler.storage.QueryPodYamlsWithPodIp(&podYamls, value)
	}

	if err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodYamlsWithPodUID error, error is %s", err)
	}

	tables := service.ConvertPodYamls2Table(podYamls)
	return http.StatusOK, tables, nil

}

func (handler *DebugPodListHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("DebugPodListHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryDebugPodListWithPodUid(handler.requestParams.SearchKey, handler.requestParams.SearchValue)

	return httpStatus, result, err
}

func DebugPodListFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &DebugPodListHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
