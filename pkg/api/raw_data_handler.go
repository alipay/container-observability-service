package api

import (
	"fmt"
	"net/http"

	"github.com/alipay/container-observability-service/pkg/utils"
)

type rawDataHandler struct {
	server        *Server
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *rawDataReq
}

type rawDataReq struct {
	auditid string
	plfid   string
}

func (handler *rawDataHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *rawDataHandler) ParseRequest() error {
	r := handler.request
	podReasonReq := rawDataReq{}
	if r.Method == http.MethodGet {
		setSP(r.URL.Query(), "auditid", &podReasonReq.auditid)
		setSP(r.URL.Query(), "plfid", &podReasonReq.plfid)
	}
	handler.requestParams = &podReasonReq
	return nil
}

func (handler *rawDataHandler) ValidRequest() error {
	req := handler.requestParams
	if req.auditid == "" && req.plfid == "" {
		return fmt.Errorf("params error")
	}

	return nil
}

func (handler *rawDataHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("rawDataHandler.Process")
	debugApiCalledCounter("rawDataHandler", handler.request)

	if handler.requestParams.plfid != "" {
		podPhase := queryPodLifePhaseByID(handler.requestParams.plfid)
		return http.StatusOK, podPhase, nil
	} else if handler.requestParams.auditid != "" {
		auditLog := queryAuditLogByID(handler.requestParams.auditid)
		return http.StatusOK, auditLog, nil
	}
	return http.StatusOK, nil, nil
}

func rawDataFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &rawDataHandler{
		server:  s,
		request: r,
		writer:  w,
	}
}
