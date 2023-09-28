package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type sloHandler struct {
	server        *Server
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *sloReq
}

type sloReq struct {
	Result         string
	Cluster        string
	Count          string
	Type           string    // create 或者 delete, 默认是 create
	DeliveryStatus string    // FAIL/KILL/ALL/SUCCESS
	SloTime        string    // 20s/30m0s/10m0s
	Env            string    // prod, test
	From           time.Time // range query
	To             time.Time // range query
}

func (handler *sloHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *sloHandler) ParseRequest() error {
	r := handler.request
	slorequest := sloReq{}
	if r.Method == http.MethodGet {
		setSP(r.URL.Query(), "result", &slorequest.Result)
		setSP(r.URL.Query(), "cluster", &slorequest.Cluster)
		setSP(r.URL.Query(), "count", &slorequest.Count)
		setSP(r.URL.Query(), "deliverystatus", &slorequest.DeliveryStatus)
		setSP(r.URL.Query(), "slotime", &slorequest.SloTime)
		setSP(r.URL.Query(), "env", &slorequest.Env)
		setSP(r.URL.Query(), "type", &slorequest.Type)
		setTPLayout(r.URL.Query(), "from", &slorequest.From)
		setTPLayout(r.URL.Query(), "to", &slorequest.To)

	}
	handler.requestParams = &slorequest
	return nil
}

func (handler *sloHandler) ValidRequest() error {
	req := handler.requestParams
	if req.Result == "" {
		return fmt.Errorf("result needed")
	}

	return nil
}

func (handler *sloHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("sloHandler.Process")

	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("QueryDebugSlo", begin)
	}()

	debugApiCalledCounter("debugsloHandler", handler.request)

	var slodatas []*slodata
	if handler.requestParams.Type == "delete" {
		slodatas = handler.server.queryDeleteSloByResult(handler.requestParams)
	} else if strings.Contains(handler.requestParams.Type, "upgrade") {
		slodatas = handler.server.queryUpgradeSloByResult(handler.requestParams)
	} else {
		slodatas = handler.server.querySloByResult(handler.requestParams)
	}

	result := make(map[string]interface{})

	namespaceCount := make(map[string]int)
	nodeCount := make(map[string]int)
	imageCount := make(map[string]int)
	podTypeCount := make(map[string]int)
	clusterCount := make(map[string]int)
	errors := make([]map[string]string, 0)
	for _, slo := range slodatas {
		dic := transSlodata(slo, handler.requestParams.Env)
		if slo.DebugUrl != "" {
			podUID, ok := dic["PodUID"]
			if ok {
				dic["DebugPodUrl"] = fmt.Sprintf("http://%s/api/v1/debugpod?uid=%s", handler.request.Host, podUID)
			}
		}

		errors = append(errors, dic)
		if v, ok := namespaceCount[slo.Namespace]; ok {
			namespaceCount[slo.Namespace] = v + 1
		} else {
			namespaceCount[slo.Namespace] = 1
		}
		nodeKey := fmt.Sprintf("%s/%s", slo.NodeName, slo.NodeIP)
		if v, ok := nodeCount[nodeKey]; ok {
			nodeCount[nodeKey] = v + 1
		} else {
			nodeCount[nodeKey] = 1
		}
		if v, ok := podTypeCount[fmt.Sprintf("IsJob:%t", slo.IsJob)]; ok {
			podTypeCount[fmt.Sprintf("IsJob:%t", slo.IsJob)] = v + 1
		} else {
			podTypeCount[fmt.Sprintf("IsJob:%t", slo.IsJob)] = 1
		}
		if v, ok := clusterCount[slo.Cluster]; ok {
			clusterCount[slo.Cluster] = v + 1
		} else {
			clusterCount[slo.Cluster] = 1
		}
		if v, ok := imageCount[slo.PullTimeoutImageName]; ok {
			imageCount[slo.PullTimeoutImageName] = v + 1
		} else {
			imageCount[slo.PullTimeoutImageName] = 1
		}
	}

	result["Namespace distribution of pods"] = namespaceCount
	result["Node distribution of pods"] = nodeCount
	result["Type distribution of pods"] = podTypeCount
	if handler.requestParams.Cluster == "" {
		result["Cluster distribution of pods"] = clusterCount
	}
	result["PodList"] = errors

	return http.StatusOK, result, nil
}

func sloFactory(s *Server, w http.ResponseWriter, r *http.Request) handler {
	return &sloHandler{
		server:  s,
		request: r,
		writer:  w,
	}
}
