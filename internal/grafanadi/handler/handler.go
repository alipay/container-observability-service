package handler

import (
	"net/http"

	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
)

type Handler interface {
	ParseRequest() error
	ValidRequest() error
	Process() (int, interface{}, error)
	RequestParams() interface{}
}

type HandlerFunc func(http.ResponseWriter, *http.Request, data_access.StorageInterface) Handler
