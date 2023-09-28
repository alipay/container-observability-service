package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alipay/container-observability-service/pkg/utils"
)

type debugNodeYamlHandler struct {
	server        *Server
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *debugNodeYamlParams
}

type debugNodeYamlParams struct {
	NodeUID  string
	NodeName string
	Json     string
}

func (handler *debugNodeYamlHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *debugNodeYamlHandler) ParseRequest() error {
	params := debugNodeYamlParams{}
	if handler.request.Method == http.MethodGet {
		setSP(handler.request.URL.Query(), "name", &params.NodeName)
		setSP(handler.request.URL.Query(), "uid", &params.NodeUID)
		setSP(handler.request.URL.Query(), "json", &params.Json)
	}
	handler.requestParams = &params
	return nil
}

func (handler *debugNodeYamlHandler) ValidRequest() error {
	req := handler.requestParams
	if req.NodeUID == "" && req.NodeName == "" {
		return fmt.Errorf("uid or name or podip needed")
	}

	return nil
}

func (handler *debugNodeYamlHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("debugNodeYamlHandler.Process")
	debugApiCalledCounter("debugNodeYamlHandler", handler.request)

	var result interface{}
	if handler.requestParams.NodeName != "" {
		result = queryNodeYamlWithNodeName(handler.requestParams.NodeName)
	}

	if handler.requestParams.Json == "true" {
		bytes, err := json.Marshal(result)
		if err == nil {
			return http.StatusOK, string(bytes), nil
		}
	}

	return http.StatusOK, result, nil
}

func debugNodeYamlFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &debugNodeYamlHandler{
		server:  s,
		request: r,
		writer:  w,
	}
}
