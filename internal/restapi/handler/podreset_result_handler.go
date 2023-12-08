package handler

import (
	"errors"
	// "fmt"
	"time"
	"net/http"
	"regexp"
	"sync"

	"github.com/alipay/container-observability-service/internal/restapi/model"
	"github.com/alipay/container-observability-service/internal/restapi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	clientmodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/utils"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/metrics"
)

type PodResetResultHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodResetResultParams
	storage       data_access.StorageInterface
}

type PodResetResultParams struct {
	PodUIDName string
	PodUID     string
}

// type PodResetResultParams struct {
// 	podUID   string
// 	podName  string
// 	podIP    string
// 	hostname string
// }

func (handler *PodResetResultHandler) RequestParams() interface{} {

	return handler.requestParams
}

func (handler *PodResetResultHandler) ParseRequest() error {
	params := PodResetResultParams{}
	// if handler.request.Method == http.MethodGet {
	// 	params.podUID = handler.request.URL.Query().Get("uid")
	// 	params.podName = handler.request.URL.Query().Get("name")
	// 	params.podIP = handler.request.URL.Query().Get("podip")
	// 	params.hostname = handler.request.URL.Query().Get("hostname")
	// }
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodResetResultHandler) ValidRequest() error {
	// reqParam := handler.requestParams
	// if reqParam.podUID == "" && reqParam.podName == "" && reqParam.podIP == "" && reqParam.hostname == "" {
	// 	return fmt.Errorf("uid or name or podip or hostname needed")
	// }
	return nil
}

func (handler *PodResetResultHandler) QueryPodResetResult(key, value string) (int, interface{}, error) {
	podInfos := make([]*clientmodel.PodInfo, 0)
	var podRest = model.DebugPodRestResult{}
	var responseResult clientmodel.Response
	slotracedata := make([]*clientmodel.SloTraceData, 0)
	lifephases := make([]*clientmodel.LifePhase, 0)
	podYamls := make([]*clientmodel.PodYaml, 0)
	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryNodeYaml").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)
	podUid := value

	match, _ := regexp.MatchString("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$", podUid)
	if !match {
		responseResult = clientmodel.Response{
			Code:    http.StatusBadRequest,
			Status:  "failed",
			Message: "query nil,parameters maybe error",
		}
		return http.StatusBadRequest, responseResult, nil
	}

	var err error
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		err = handler.storage.QueryPodInfoWithPodUid(&podInfos, podUid)
	}()
	go func() {
		defer wg.Done()
		err = handler.storage.QuerySloTraceDataWithPodUID(&slotracedata, podUid)
	}()
	go func() {
		defer wg.Done()
		err = handler.storage.QueryLifePhaseWithPodUid(&lifephases, podUid)
	}()
	if err != nil {
		responseResult.Message = "query failed" + err.Error()

		return http.StatusBadRequest, responseResult, errors.New(err.Error())
	}
	wg.Wait()

	var summary = make(map[interface{}]int)
	var component interface{}
	var phases []*clientmodel.LifePhase
	if len(lifephases) > 0 {
		for _, ph := range lifephases {
			if ph.HasErr {
				phases = append(phases, ph)
			}
		}
	}

	for _, ph := range phases {
		info, ok := ph.ExtraInfo.(map[string]interface{})
		if ok {
			if info["Message"] != nil {
				Message := info["Message"]
				if len(Message.(string)) > 4 {
					if _, ok := summary[Message]; !ok {
						summary[Message] = 1
					} else {
						summary[Message] += 1
					}
				}
				if info["UserAgent"] != nil {
					component = info["UserAgent"]
				}
			}
			if obj, ok := info["eventObject"]; ok {
				v, ok := obj.(map[string]interface{})
				if ok {
					msg := v["message"]
					if len(msg.(string)) > 4 {
						if _, ok := summary[msg]; !ok {
							summary[msg] = 1
						} else {
							summary[msg] += 1
						}
					}
				}
				if info["UserAgent"] != nil {
					component = info["UserAgent"]
				}
			}

		}
	}
	var str interface{}
	max := 0
	for key, v := range summary {
		if v > max {
			max = v
			str = key
		}
	}

	if len(podInfos) > 0 {
		info := podInfos[0]
		podRest.PodInfos.Site = "mainsite"
		podRest.PodInfos.ClusterName = info.ClusterName
		podRest.PodInfos.BizName = info.BizSource
		podRest.PodInfos.Namespace = info.Namespace
		podRest.PodInfos.PodName = info.PodName
		podRest.PodInfos.PodUID = info.PodUID
		podRest.PodInfos.PodIP = info.PodIP
		podRest.PodInfos.NodeName = info.NodeName
	}
	podResult := model.DebugPodResult{
		Action:  "Contact Admin",
		Contact: "https://github.com/alipay/container-observability-service",
		Info:    "CopyRight&ServiceContact",
	}

	if len(slotracedata) > 0 {

		if slotracedata[0].Type == "create" {
			podResult.ResultCode = slotracedata[0].SLOViolationReason
			podResult.DebugStage = slotracedata[0].Type
		}

		if slotracedata[0].Type == "delete" {
			podResult.ResultCode = slotracedata[0].DeleteResult
			podResult.DebugStage = slotracedata[0].Type

		}
		if slotracedata[0].Type == "pod_upgrade" {
			podResult.ResultCode = slotracedata[0].UpgradeResult
			podResult.DebugStage = slotracedata[0].Type
		}
		if podResult.ResultCode != "success" {
			podResult.Summary = str
			podResult.Component = component
		}

	}
	if len(podInfos) == 0 && len(slotracedata) == 0 {
		responseResult.Code = http.StatusBadRequest
		responseResult.Status = http.StatusText(responseResult.Code)
		responseResult.Message = "query parameters maybe error"
	} else {
		responseResult.Code = http.StatusOK
		responseResult.Message = "query success"
		responseResult.Status = http.StatusText(responseResult.Code)
		podRest.DebugPodRes = podResult
		responseResult.Data = podRest
	}
	dataFrame := service.ConvertResetResult2Frame(responseResult)
	// dataFrame := service.ConvertPodInfo2Frame(podInfos)

	return http.StatusOK, dataFrame, nil
}

func (handler *PodResetResultHandler) Process() (int, interface{}, error) {
	var result interface{}
	var err error
	var httpStatus int

	// httpStatus, result, err = handler.QueryPodResetResult()
	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryPodResetResult(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}

	return httpStatus, result, err
}

func PodResetResultFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodResetResultHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
