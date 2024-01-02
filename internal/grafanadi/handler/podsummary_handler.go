package handler

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"

	"github.com/alipay/container-observability-service/pkg/slo"
	"github.com/alipay/container-observability-service/pkg/utils"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/internal/grafanadi/service"
)

type PodSummaryHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodSummaryParams
	storage       data_access.StorageInterface
}

type PodSummaryParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *PodSummaryHandler) RequestParams() interface{} {

	return handler.requestParams
}

func (handler *PodSummaryHandler) ParseRequest() error {
	params := PodSummaryParams{}
	if handler.request.Method == http.MethodGet {
		params.PodUIDName = handler.request.URL.Query().Get("searchkey")
		params.PodUID = handler.request.URL.Query().Get("searchvalue")
	}

	handler.requestParams = &params
	return nil
}

func (handler *PodSummaryHandler) ValidRequest() error {
	return nil
}

func (handler *PodSummaryHandler) QueryPodSummary(key, value string) (int, interface{}, error) {
	podInfos := make([]*storagemodel.PodInfo, 0)
	var podSummary = model.PodSummary{}
	slotracedata := make([]*storagemodel.SloTraceData, 0)
	lifephases := make([]*storagemodel.LifePhase, 0)
	podYamls := make([]*storagemodel.PodYaml, 0)
	if value == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryPodSummary").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, key, &value)
	podUid := value

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
		return http.StatusBadRequest, nil, errors.New(err.Error())
	}
	wg.Wait()

	var summary = make(map[interface{}]int)
	var component interface{}
	var phases []*storagemodel.LifePhase
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

	var summary_description string
	summary_description = "The current stage is the pod \"creation\" phase. \n\nIf you do not configure a custom slo timeout time, by default the service pod timeout time is \"" + slo.SERVICE_POD_TIMEOUT_TIME.String() + "\", and the job pod timeout time is \"" + slo.JOB_POD_TIMEOUT_TIME.String() + "\"."
	
	if len(slotracedata) > 0 {

		if slotracedata[0].Type == "create" {
			podSummary.ResultCode = slotracedata[0].SLOViolationReason
			podSummary.DebugStage = slotracedata[0].Type
		}

		if slotracedata[0].Type == "delete" {
			podSummary.ResultCode = slotracedata[0].DeleteResult
			podSummary.DebugStage = slotracedata[0].Type

		}
		if slotracedata[0].Type == "pod_upgrade" {
			podSummary.ResultCode = slotracedata[0].UpgradeResult
			podSummary.DebugStage = slotracedata[0].Type
		}

		if podSummary.ResultCode != "success" {
			podSummary.Component = component
			summary_description = "At \"" + podSummary.DebugStage + "\" stage, error \"" + podSummary.ResultCode + "\" occurred, useragent is \"" + podSummary.Component.(string) + "\"。\n\nThe following is a summary of error: \n\"" + str.(string)+"\""
		} else {
			summary_description = "Pod deliver successfully!"
		}
	}

	if len(podInfos) == 0 && len(slotracedata) == 0 {
		podSummary.Summary = "query parameters maybe error"
	} else {
		podSummary.Summary = summary_description
	}

	dataFrame := service.ConvertPodSummary2Frame(podSummary)

	return http.StatusOK, dataFrame, nil
}

func (handler *PodSummaryHandler) Process() (int, interface{}, error) {
	var result interface{}
	var err error
	var httpStatus int

	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.QueryPodSummary(handler.requestParams.PodUIDName, handler.requestParams.PodUID)
	}

	return httpStatus, result, err
}

func PodSummaryFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodSummaryHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}