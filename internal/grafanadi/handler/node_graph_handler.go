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

type NodeGraphHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *NodeGraphParams
	storage       data_access.StorageInterface
}

type NodeGraphParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *NodeGraphHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *NodeGraphHandler) ParseRequest() error {
	params := NodeGraphParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *NodeGraphHandler) ValidRequest() error {

	return nil
}

func (handler *NodeGraphHandler) QueryPodDeliveryWithPodUid(key, value string) (int, interface{}, error) {
	podYamls := make([]*model.PodYaml, 0)
	var deliveryTable interface{}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)
	err := handler.storage.QueryPodYamlsWithPodUID(&podYamls, value)
	if err != nil {
		return http.StatusOK, nil, err
	}

	switch handler.request.URL.Path {
	case "/podyamlgraphnodes":
		nodes, _ := service.ConvertPodGraph2Frame(podYamls)
		deliveryTable = service.ConvertPodYamlGraphNodes2Frame(nodes)
	case "/podyamlgraphedges":
		_, edges := service.ConvertPodGraph2Frame(podYamls)
		deliveryTable = service.ConvertPodYamlGraphEdges2Frame(edges)
	}

	return http.StatusOK, deliveryTable, nil

}

func (handler *NodeGraphHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("NodeGraphHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryPodDeliveryWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)

	return httpStatus, result, err
}

func NodeGraphParamsFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &NodeGraphHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
