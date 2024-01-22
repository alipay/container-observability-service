package handler

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"

	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/slo"
	"github.com/alipay/container-observability-service/pkg/utils"
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

	var podInfoErr, sloTraceDataErr, lifePhaseErr error
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		podInfoErr = handler.storage.QueryPodInfoWithPodUid(&podInfos, podUid)
	}()
	go func() {
		defer wg.Done()
		sloTraceDataErr = handler.storage.QuerySloTraceDataWithPodUID(&slotracedata, podUid)
	}()
	go func() {
		defer wg.Done()
		lifePhaseErr = handler.storage.QueryLifePhaseWithPodUid(&lifephases, podUid)
	}()
	wg.Wait()
	if podInfoErr != nil {
		return http.StatusBadRequest, nil, errors.New(podInfoErr.Error())
	}
	if sloTraceDataErr != nil {
		return http.StatusBadRequest, nil, errors.New(sloTraceDataErr.Error())
	}
	if lifePhaseErr != nil {
		return http.StatusBadRequest, nil, errors.New(lifePhaseErr.Error())
	}

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
		info := ph.ExtraInfo.(map[string]interface{})
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
			v := obj.(map[string]interface{})
			msg := v["message"]
			if len(msg.(string)) > 4 {
				if _, ok := summary[msg]; !ok {
					summary[msg] = 1
				} else {
					summary[msg] += 1
				}
			}
			if info["UserAgent"] != nil {
				component = info["UserAgent"]
			}
		}
	}
	var summaryStr interface{}
	max := 0
	for key, v := range summary {
		if v > max {
			max = v
			summaryStr = key
		}
	}

	var summaryDescription string
	summaryDescription = "The current stage is the pod \"creation\" phase. \n\nIf you do not configure a custom slo timeout time yet, by default the service pod timeout time is \"" + slo.SERVICE_POD_TIMEOUT_TIME.String() + "\", and the job pod timeout time is \"" + slo.JOB_POD_TIMEOUT_TIME.String() + "\"."

	if len(slotracedata) > 0 {

		sloViolationFlag := false

		for i := len(slotracedata) - 1; i >= 0; i-- {
			if slotracedata[i].SLOViolationReason != "" && slotracedata[i].SLOViolationReason != "success" {
				sloViolationFlag = true
				podSummary.DebugStage = append(podSummary.DebugStage, slotracedata[i].Type)
				podSummary.ResultCode = append(podSummary.ResultCode, slotracedata[i].SLOViolationReason)
			}
		}

		if sloViolationFlag {
			podSummary.Component = component
			summaryDescription = ""
			for i := 0; i < len(podSummary.DebugStage); i++ {
				summaryDescription += "At \"" + podSummary.DebugStage[i] + "\" stage, error \"" + podSummary.ResultCode[i] + "\" occurred, useragent is \"" + podSummary.Component.(string) + "\"。\n\nThe following is a summary of error: \n\"" + summaryStr.(string) + "\"\n\n\n"
			}
		} else {
			summaryDescription = "Pod deliver successfully!"
		}
	}

	if len(podInfos) == 0 && len(slotracedata) == 0 {
		podSummary.Summary = "query parameters maybe error"
	} else {
		podSummary.Summary = summaryDescription
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
