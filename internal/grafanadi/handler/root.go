package handler

import (
	"net/http"

	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type RootHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *RootParams
	storage       data_access.StorageInterface
}

type RootParams struct {
}

func (handler *RootHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *RootHandler) ParseRequest() error {
	return nil
}

func (handler *RootHandler) ValidRequest() error {
	return nil
}

func (handler *RootHandler) Root() (int, interface{}, error) {
	return http.StatusOK, "ok", nil

}

func (handler *RootHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("RootHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int
	httpStatus, result, err = handler.Root()
	return httpStatus, result, err
}

func RootFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &RootHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
