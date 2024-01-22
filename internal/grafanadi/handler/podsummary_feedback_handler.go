package handler

import (
	"time"
	"net/http"
	"encoding/json"
	"fmt"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	"github.com/alipay/container-observability-service/internal/grafanadi/service"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/utils"
	interutils "github.com/alipay/container-observability-service/internal/grafanadi/utils"
	"github.com/alipay/container-observability-service/pkg/metrics"
)

type PodSummaryFeedbackHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *PodSummaryFeedbackParams
	storage       data_access.StorageInterface
}

type PodSummaryFeedbackParams struct {
	PodUIDName string
	PodUID     string
	PodSummaryFeedback	 model.PodSummaryFeedback
}

func (handler *PodSummaryFeedbackHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *PodSummaryFeedbackHandler) ParseRequest() error {
	params := PodSummaryFeedbackParams{}
	if handler.request.Method == http.MethodPost {
		params.PodUIDName = handler.request.URL.Query().Get("searchkey")
		params.PodUID = handler.request.URL.Query().Get("searchvalue")

		var summaryFeedback model.PodSummaryFeedback
		err := json.NewDecoder(handler.request.Body).Decode(&summaryFeedback)
		if err != nil {
			return err
		}
		params.PodSummaryFeedback = summaryFeedback
	}
	handler.requestParams = &params
	return nil
}

func (handler *PodSummaryFeedbackHandler) ValidRequest() error {
	return nil
}

func (handler *PodSummaryFeedbackHandler) StorePodSummaryFeedback(podUIDName string, podUID string, summaryFeedback model.PodSummaryFeedback) (int, interface{}, error) {
	podInfos := make([]*storagemodel.PodInfo, 0)
	podYamls := make([]*storagemodel.PodYaml, 0)
	var podSummaryFeedback = storagemodel.PodSummaryFeedback{}
	if podUID == "" {
		return http.StatusOK, nil, nil
	}

	begin := time.Now()
	podSummaryFeedback.CreateTime = begin
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("StorePodSummaryFeedback").Observe(cost)
	}()
	util := interutils.Util{
		Storage: handler.storage,
	}
	util.GetUid(podYamls, podUIDName, &podUID)

	errInfo := handler.storage.QueryPodInfoWithPodUid(&podInfos, podUID)
	if errInfo != nil {
		return http.StatusOK, nil, fmt.Errorf("QueryPodInfoWithPodUid error, error is %s", errInfo)
	}
	
	if len(podInfos) > 0 {
		info := podInfos[0]
		podSummaryFeedback.ClusterName = info.ClusterName
		podSummaryFeedback.Namespace = info.Namespace
		podSummaryFeedback.PodName = info.PodName
		podSummaryFeedback.PodUID = info.PodUID
		podSummaryFeedback.PodIP = info.PodIP
		podSummaryFeedback.NodeName = info.NodeName
	}

	podSummaryFeedback.Feedback = summaryFeedback.Feedback
	podSummaryFeedback.Score = summaryFeedback.Score
	podSummaryFeedback.Comment = summaryFeedback.Comment
	podSummaryFeedback.Summary = summaryFeedback.Summary

	errInfo = handler.storage.StorePodSummaryFeedbackWithPodUid(podSummaryFeedback, podSummaryFeedback)
	if errInfo != nil {
		return http.StatusOK, nil, fmt.Errorf("StorePodSummaryFeedbackWithPodUid error, error is %s", errInfo)
	}

	dataFrame := service.ConvertSummaryFeedback2Frame(podSummaryFeedback)
	return http.StatusOK, dataFrame, nil
}

func (handler *PodSummaryFeedbackHandler) Process() (int, interface{}, error) {
	var result interface{}
	var err error
	var httpStatus int

	if handler.requestParams.PodUID != "" {
		httpStatus, result, err = handler.StorePodSummaryFeedback(handler.requestParams.PodUIDName, handler.requestParams.PodUID, handler.requestParams.PodSummaryFeedback)
	}

	return httpStatus, result, err
}

func PodSummaryFeedbackFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &PodSummaryFeedbackHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
