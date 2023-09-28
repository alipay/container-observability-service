package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alipay/container-observability-service/pkg/utils"
)

type debugPodYamlHandler struct {
	server        *Server
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *debugPodYamlParams
}

type debugPodYamlParams struct {
	PodUID  string
	PodName string
	PodIP   string
	NodeIP  string
	Json    string
}

func (handler *debugPodYamlHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *debugPodYamlHandler) ParseRequest() error {
	params := debugPodYamlParams{}
	if handler.request.Method == http.MethodGet {
		setSP(handler.request.URL.Query(), "name", &params.PodName)
		setSP(handler.request.URL.Query(), "uid", &params.PodUID)
		setSP(handler.request.URL.Query(), "podip", &params.PodIP)
		setSP(handler.request.URL.Query(), "json", &params.Json)
		setSP(handler.request.URL.Query(), "nodeip", &params.NodeIP)
	}
	handler.requestParams = &params
	return nil
}

func (handler *debugPodYamlHandler) ValidRequest() error {
	req := handler.requestParams
	if req.PodUID == "" && req.PodName == "" && req.PodIP == "" && req.NodeIP == "" {
		return fmt.Errorf("uid or name or podip or nodeip needed")
	}

	return nil
}

func (handler *debugPodYamlHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("debugPodYamlHandler.Process")
	debugApiCalledCounter("debugPodYamlHandler", handler.request)

	var result interface{}
	if handler.requestParams.PodUID != "" {
		yamls := queryPodYamlWithPodUID(handler.requestParams.PodUID)
		if len(yamls) > 0 {
			result = yamls[0]
		}
	}

	if handler.requestParams.NodeIP != "" {
		yamls := queryPodYamlsWithNodeIP(handler.requestParams.NodeIP, false)
		result = yamls
	}

	if handler.requestParams.Json == "true" {
		bytes, err := json.Marshal(result)
		if err == nil {
			return http.StatusOK, string(bytes), nil
		}
	}

	return http.StatusOK, result, nil
}

func debugPodYamlFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &debugPodYamlHandler{
		server:  s,
		request: r,
		writer:  w,
	}
}
