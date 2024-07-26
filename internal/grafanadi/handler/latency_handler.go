package handler

import (
	"fmt"
	"net/http"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

type LunettesLatencyHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *LunettesLatencyParams
	storage       data_access.StorageInterface
}

type LunettesLatencyParams struct {
}

func (handler *LunettesLatencyHandler) RequestParams() interface{} {

	return handler.requestParams
}

func (handler *LunettesLatencyHandler) ParseRequest() error {

	return nil
}

func (handler *LunettesLatencyHandler) ValidRequest() error {
	return nil
}

func (handler *LunettesLatencyHandler) QueryLunettesLatency() (int, interface{}, error) {

	eavesdroppingMeta := make([]*model.LunettesMeta, 0)

	err := handler.storage.QueryLunettesLatency(&eavesdroppingMeta, model.WithLimit(200))
	if err != nil {
		return http.StatusOK, eavesdroppingMeta, fmt.Errorf("QueryLunettesLatency error, error is %s", err)
	}

	eavesdroppingMetaTable := service.ConvertMeta2Frame(eavesdroppingMeta)

	return http.StatusOK, eavesdroppingMetaTable, nil
}

func (handler *LunettesLatencyHandler) Process() (int, interface{}, error) {
	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryLunettesLatency()

	return httpStatus, result, err
}

func LunettesLatencyFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &LunettesLatencyHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
