package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"io"
	"k8s.io/klog"
	"log"
	"net/http"
	"time"
)

const (
	tkpNamespace = "lunettes"
	tkpSvcName   = "alipay-tkp-manager"
)

type TkpHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *TkpParams
	storage       data_access.StorageInterface
}

type TkpParams struct {
	Namespace string `json:"pod_namespace"`
	PodName   string `json:"pod_name"`
}

func (handler *TkpHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *TkpHandler) ParseRequest() error {
	params := TkpParams{}
	if handler.request.Method == http.MethodPost {
		err := json.NewDecoder(handler.request.Body).Decode(&params)
		if err != nil {
			klog.Errorf("parse request body error: %s", err)
			return err
		}
	}
	handler.requestParams = &params
	klog.Infof("tkp request params: %+v", params)
	return nil
}

func (handler *TkpHandler) ValidRequest() error {
	return nil
}

type TkpResp struct {
	Message string  `json:"message"`
	Code    int     `json:"code"`
	Data    TkpBody `json:"data"`
}

type TkpBody struct {
	PodNamespace string `json:"pod_namespace"`
	PodName      string `json:"pod_name"`
	//被托管的pod 生成的vtk 对应的名称
	Vtkp string `json:"vtkp"`
}

func buildReqUrl(svc, tkpSNS, uri string) string {
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:9999%s", svc, tkpSNS, uri)
}

func (handler *TkpHandler) Tkp(params *TkpParams) (int, interface{}, error) {
	var tkpResp TkpResp
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues(" Tkp").Observe(cost)
	}()
	reqUrl := buildReqUrl(tkpSvcName, tkpNamespace, "/apis/v2/turnkeypods/pods")
	klog.Infof("tkp request url: %s", reqUrl)
	jsonData, err := json.Marshal(params)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return http.StatusOK, nil, err
	}
	//post tkp
	req, err := http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("new request: %s error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("get url: %s error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ReadAll req %s body error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}

	err = json.Unmarshal(body, &tkpResp)
	if err != nil {
		log.Printf("Unmarshal req %s body error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	return http.StatusOK, tkpResp, nil

}

func (handler *TkpHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic(" TkpHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.Tkp(handler.requestParams)

	return httpStatus, result, err
}

func TkpFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &TkpHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
