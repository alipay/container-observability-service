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

type QueryPodListHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *QueryPodListParams
	storage       data_access.StorageInterface
}

type QueryPodListParams struct {
	SearchKey   string
	SearchValue string
	From        time.Time
	To          time.Time
}

func (handler *QueryPodListHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *QueryPodListHandler) ParseRequest() error {
	params := QueryPodListParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.SearchKey = key
		params.SearchValue = value
		setTPLayout(handler.request.URL.Query(), "from", &params.From)
		setTPLayout(handler.request.URL.Query(), "to", &params.To)
	}

	handler.requestParams = &params
	return nil
}

func (handler *QueryPodListHandler) ValidRequest() error {

	reqParam := handler.requestParams
	switch reqParam.SearchKey {
	case "uid", "name", "hostname", "podip":
		return nil
	default:
		return fmt.Errorf("uid, name, podip, hostname are all empty")
	}
}

func (handler *QueryPodListHandler) QueryPodListWithParams(params *QueryPodListParams) (int, interface{}, error) {
	podYamls := make([]*model.PodYaml, 0)
	if params.SearchValue == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodListWithParams").Observe(cost)
	}()

	var err error
	if params.SearchKey == "uid" {
		err = handler.storage.QueryPodYamlWithParams(&podYamls, &model.PodParams{
			Uid:  params.SearchValue,
			From: params.From,
			To:   params.To,
		})
	} else if params.SearchKey == "name" {
		err = handler.storage.QueryPodYamlWithParams(&podYamls, &model.PodParams{
			Name: params.SearchValue,
			From: params.From,
			To:   params.To,
		})
	} else if params.SearchKey == "hostname" {
		err = handler.storage.QueryPodYamlWithParams(&podYamls, &model.PodParams{
			Hostname: params.SearchValue,
			From:     params.From,
			To:       params.To,
		})
	} else if params.SearchKey == "podip" {
		err = handler.storage.QueryPodYamlWithParams(&podYamls, &model.PodParams{
			Podip: params.SearchValue,
			From:  params.From,
			To:    params.To,
		})
	}

	if err != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodListWithParams error, error is %s", err)
	}

	tables := service.ConvertPodyamls2Table(podYamls)
	return http.StatusOK, tables, nil

}

func (handler *QueryPodListHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("QueryPodListHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryPodListWithParams(handler.requestParams)

	return httpStatus, result, err
}

func QueryPodListFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &QueryPodListHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
