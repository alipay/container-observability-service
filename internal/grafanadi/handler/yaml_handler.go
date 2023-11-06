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

type YamlHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *YamlParams
	storage       data_access.StorageInterface
}

func (handler *YamlHandler) GetYamls(params *YamlParams) (int, interface{}, error) {
	podYamls := make([]*model.PodYaml, 0)
	nodeYamls := make([]*model.NodeYaml, 0)
	var err error
	if params == nil {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("GetYamls").Observe(cost)
	}()
	if handler.requestParams.Resource == "pod" && handler.requestParams.Uid != "" {
		err = handler.storage.QueryPodYamlsWithPodUID(&podYamls, handler.requestParams.Uid)
		if err == nil {
			if len(podYamls) > 0 {
				podyaml := &model.PodYaml{
					AuditID:        podYamls[0].AuditID,
					ClusterName:    podYamls[0].ClusterName,
					HostIP:         podYamls[0].HostIP,
					IsBeginDelete:  podYamls[0].IsBeginDelete,
					IsDeleted:      podYamls[0].IsDeleted,
					Pod:            podYamls[0].Pod,
					PodIP:          podYamls[0].PodIP,
					PodName:        podYamls[0].PodName,
					PodUid:         podYamls[0].PodUid,
					StageTimestamp: podYamls[0].StageTimestamp,
				}
				return http.StatusOK, service.TransformYaml2Html(podyaml), nil
			}
		}
	} else if handler.requestParams.Resource == "node" && handler.requestParams.Name != "" {
		err = handler.storage.QueryNodeYamlsWithNodeName(&nodeYamls, handler.requestParams.Name)
		if err != nil {
			klog.Errorf("get nodeyaml failed: %s", err.Error())
		} else {
			if len(nodeYamls) > 0 {
				return http.StatusOK, service.TransformYaml2Html(nodeYamls[0]), nil
			}
		}
	}

	return http.StatusOK, nil, nil
}

type YamlParams struct {
	Resource string
	Uid      string
	Name     string
}

func (handler *YamlHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *YamlHandler) ParseRequest() error {
	params := YamlParams{}
	if handler.request.Method == http.MethodGet {
		resource := handler.request.URL.Query().Get("resource")
		uid := handler.request.URL.Query().Get("uid")
		name := handler.request.URL.Query().Get("name")
		params.Resource = resource
		params.Uid = uid
		params.Name = name
	}

	handler.requestParams = &params
	return nil
}

func (handler *YamlHandler) ValidRequest() error {

	return nil
}

func (handler *YamlHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("ContainerlifecycleHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	if handler.requestParams != nil {
		httpStatus, result, err = handler.GetYamls(handler.requestParams)

	}
	return httpStatus, result, err
}

func YamlFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &YamlHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
