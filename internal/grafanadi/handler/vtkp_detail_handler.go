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

type VTkpDetailHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *VTkpDetailParams
	storage       data_access.StorageInterface
}

type VTkpDetailParams struct {
	Cluster       string
	VtkpNamespace string
	VtkpName      string
}

func (handler *VTkpDetailHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *VTkpDetailHandler) ParseRequest() error {
	params := VTkpDetailParams{}
	if handler.request.Method == http.MethodGet {
		klog.Infof("vtkpDetail request params: %+v", handler.request.URL.Query())
		params.Cluster = handler.request.URL.Query().Get("cluster")
		params.VtkpNamespace = handler.request.URL.Query().Get("vtkp_namespace")
		params.VtkpName = handler.request.URL.Query().Get("vtkp_name")
	}
	handler.requestParams = &params
	klog.Infof("vtkpDetail request params: %+v", params)
	return nil
}

func (handler *VTkpDetailHandler) ValidRequest() error {
	return nil
}

type VTkpDetailResp struct {
	Message string   `json:"message"`
	Code    int      `json:"code"`
	Data    VTkpData `json:"data"`
}

type VTkpData struct {
	VtkpNamespace string       `json:"vtkp_namespace"`
	VtkpId        string       `json:"vtkp_uid"`
	VtkpName      string       `json:"vtkp_name"`
	PeerPod       string       `json:"peer_pod"`
	PodInfos      []TkpPodInfo `json:"pod_infos"`
}

type TkpPodInfo struct {
	Namespace string      `json:"namespace"`
	Name      string      `json:"name"`
	Age       string      `json:"age"`
	Ip        string      `json:"ip"`
	Node      string      `json:"node"`
	Ready     string      `json:"ready"`
	Status    string      `json:"status"`
	Restarts  int         `json:"restarts"`
	OwnerRefs []OwnerRefs `json:"ownerRefs"`
}

type OwnerRefs struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Uid        string `json:"uid"`
}

func (handler *VTkpDetailHandler) VTkpDetail(params *VTkpDetailParams) (int, interface{}, error) {
	var tkpResp VTkpDetailResp
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues(" VTkpDetail ").Observe(cost)
	}()
	reqUrl := buildReqUrl(params.Cluster, "/apis/v2/turnkeypods/profiles")
	tkpDetailReq, err := url.Parse(reqUrl)
	if err != nil {
		klog.Errorf("url parse error: %s\n", err)
		return http.StatusOK, nil, err
	}
	queryParams := url.Values{}
	queryParams.Set("vtkp_namespace", params.VtkpNamespace)
	queryParams.Set("vtkp_name", params.VtkpName)
	tkpDetailReq.RawQuery = queryParams.Encode()
	klog.Infof("vTkpDetailHandler req: %s\n", tkpDetailReq.String())

	req, err := http.NewRequest(http.MethodGet, tkpDetailReq.String(), nil)
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
	klog.Infof("vtkpDetailResp: %+v\n", tkpResp)
	//结构转换
	return http.StatusOK, convertTkpForGData(&tkpResp), nil
}

type PodType string

type GVTkpDetail struct {
	PodList      []GPodInfos `json:"podList"`
	WorkloadList []OwnerRefs `json:"workloadList"`
}

type GPodInfos struct {
	RefId   []string   `json:"refIds"`
	PodInfo TkpPodInfo `json:"detail"`
}

func convertTkpForGData(v *VTkpDetailResp) GVTkpDetail {
	var gVTkpDetail GVTkpDetail
	// HTTP code check
	if v.Code < 200 || v.Code >= 300 {
		klog.Errorf("http code error: %d\n", v.Code)
		return gVTkpDetail
	}
	workloadMap := make(map[string]OwnerRefs)
	for _, podInfo := range v.Data.PodInfos {
		var refIds []string
		for _, ownerRefDetail := range podInfo.OwnerRefs {
			if _, exists := workloadMap[ownerRefDetail.Uid]; !exists {
				workloadMap[ownerRefDetail.Uid] = ownerRefDetail
			}
			refIds = append(refIds, ownerRefDetail.Uid)
		}
		gVTkpDetail.PodList = append(gVTkpDetail.PodList, GPodInfos{
			RefId:   refIds,
			PodInfo: podInfo,
		})
	}

	for _, ownerRefDetail := range workloadMap {
		gVTkpDetail.WorkloadList = append(gVTkpDetail.WorkloadList, ownerRefDetail)
	}
	return gVTkpDetail
}

func DeduplicateByUid(refs []OwnerRefs) []OwnerRefs {
	uidMap := make(map[string]bool)
	var uniqueRefs []OwnerRefs

	for _, ref := range refs {
		if _, exists := uidMap[ref.Uid]; !exists {
			uidMap[ref.Uid] = true
			uniqueRefs = append(uniqueRefs, ref)
		}
	}
	return uniqueRefs
}

func (handler *VTkpDetailHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic(" VTkpDetailHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.VTkpDetail(handler.requestParams)

	return httpStatus, result, err
}

func VTkpDetailFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &VTkpDetailHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
