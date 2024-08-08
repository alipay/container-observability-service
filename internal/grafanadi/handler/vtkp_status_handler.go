package handler

import (
	"encoding/json"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"io"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"time"
)

type VTkpStatusHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *VTkpStatusParams
	storage       data_access.StorageInterface
}

type VTkpStatusParams struct {
	VtkpNamespace string `json:"vtkpNamespace"`
	VtkpName      string `json:"vtkpName"`
}

func (handler *VTkpStatusHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *VTkpStatusHandler) ParseRequest() error {
	params := VTkpStatusParams{}
	if handler.request.Method == http.MethodGet {
		params.VtkpNamespace = handler.request.URL.Query().Get("vtkp_namespace")
		params.VtkpName = handler.request.URL.Query().Get("vtkp_name")
	}
	handler.requestParams = &params
	klog.Infof("vtkpStatus request params: %+v", params)
	return nil
}

func (handler *VTkpStatusHandler) ValidRequest() error {
	return nil
}

type VTkpStatusResp struct {
	Message string         `json:"message"`
	Code    int            `json:"code"`
	Data    VTkpStatusData `json:"data"`
}

type VTkpStatusData struct {
	Type            string   `json:"type"`
	Status          string   `json:"status"`
	CurPhase        string   `json:"cur_phase"`
	TotalPhases     []string `json:"total_phases"`
	CompletedPhases []string `json:"completed_phases"`
}

func (handler *VTkpStatusHandler) VTkpStatus(params *VTkpStatusParams) (int, interface{}, error) {
	var tkpResp VTkpStatusResp
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues(" VTkpStatus ").Observe(cost)
	}()
	reqUrl := buildReqUrl(tkpSvcName, tkpNamespace, "/apis/v2/turnkeypods/progress")
	tkpStatusReq, err := url.Parse(reqUrl)
	if err != nil {
		klog.Errorf("url parse error: %s\n", err)
		return http.StatusOK, nil, err
	}
	queryParams := url.Values{}
	queryParams.Set("vtkp_namespace", params.VtkpNamespace)
	queryParams.Set("vtkp_name", params.VtkpName)
	tkpStatusReq.RawQuery = queryParams.Encode()
	klog.Infof("VTkpStatusHandler req : %s\n", tkpStatusReq.String())

	req, err := http.NewRequest(http.MethodGet, tkpStatusReq.String(), nil)
	if err != nil {
		klog.Errorf("new request: %s error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		klog.Errorf("get url: %s error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Errorf("ReadAll req %s body error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}

	err = json.Unmarshal(body, &tkpResp)
	if err != nil {
		klog.Errorf("Unmarshal req %s body error: %s\n", reqUrl, err)
		return http.StatusOK, nil, err
	}
	klog.Infof("vtkpStatusResp: %+v\n", tkpResp)
	//结构转换
	return http.StatusOK, convertTkpStatusForGData(&tkpResp), nil
}

type GVTkpStatus struct {
	PhaseList []Phase `json:"phase_list"`
}

type Phase struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Trigger bool   `json:"trigger"`
	Reason  string `json:"reason"`
}

func convertTkpStatusForGData(v *VTkpStatusResp) interface{} {
	var PhaseList []Phase
	// HTTP code check
	if v.Code < 200 || v.Code >= 300 {
		klog.Errorf("http code error: %d\n", v.Code)
		return PhaseList
	}
	var set []string
	mapPhase := make(map[string]Phase, 4)
	for _, phaseName := range v.Data.TotalPhases {
		set = append(set, phaseName)
		mapPhase[phaseName] = Phase{
			Name:    phaseName,
			Status:  "",
			Trigger: false,
			Reason:  "",
		}
	}

	for _, compPhase := range v.Data.CompletedPhases {
		mapPhase[compPhase] = Phase{
			Name:    compPhase,
			Status:  "success",
			Trigger: true,
			Reason:  "",
		}
	}
	mapPhase[v.Data.CurPhase] = Phase{
		Name:    v.Data.CurPhase,
		Status:  v.Data.Status,
		Trigger: true,
		Reason:  "",
	}

	//根据 set 返回
	for _, phaseName := range set {
		value, exists := mapPhase[phaseName]
		if exists {
			PhaseList = append(PhaseList, value)
		}
	}
	return PhaseList
}

func FindStringIndex(slice []string, str string) int {
	for i, v := range slice {
		if v == str {
			return i
		}
	}
	return -1
}

func (handler *VTkpStatusHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic(" VTkpStatusHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.VTkpStatus(handler.requestParams)

	return httpStatus, result, err
}

func VTkpStatusFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &VTkpStatusHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
