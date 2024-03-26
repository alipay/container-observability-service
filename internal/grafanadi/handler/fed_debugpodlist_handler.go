package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"k8s.io/klog/v2"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/data_access"
	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
)

type FedDebugPodListHandler struct {
	request       *http.Request
	writer        http.ResponseWriter
	requestParams *FedDebugPodListParams
	storage       data_access.StorageInterface
}

type FedDebugPodListParams struct {
	PodUIDName string
	PodUID     string
}

func (handler *FedDebugPodListHandler) RequestParams() interface{} {
	return handler.requestParams
}

func (handler *FedDebugPodListHandler) ParseRequest() error {
	params := FedDebugPodListParams{}
	if handler.request.Method == http.MethodGet {
		key := handler.request.URL.Query().Get("searchkey")
		value := handler.request.URL.Query().Get("searchvalue")
		params.PodUIDName = key
		params.PodUID = value
	}

	handler.requestParams = &params
	return nil
}

func (handler *FedDebugPodListHandler) ValidRequest() error {

	return nil
}

func (handler *FedDebugPodListHandler) QueryFedDebugPodListWithPodUid(key, value string) (int, interface{}, error) {
	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.QueryMethodDurationMilliSeconds.WithLabelValues("QueryFedDebugPodListWithPodUid").Observe(cost)
	}()
	ch := make(chan *model.DebugPodListTable, 1000)
	var wg sync.WaitGroup
	concurrency := 5                        // 并发数量
	sem := make(chan struct{}, concurrency) // 信号量用于控制并发数量

	for _, v := range common.GSiteOptions.SiteInfos {
		siteName := v.SiteName
		dashboardUrl := v.DashboardUrl
		diUrl := v.DiUrl

		// 仅主 Lunettes 支持联邦查询功能
		if v.SiteName == "main" {
			// 解析 URL
			parsedURL, err := url.Parse(diUrl)
			if err != nil {
				fmt.Println("无法解析 URL:", err)
				return http.StatusOK, nil, nil
			}
			klog.Infof("parsedURL is %+v", parsedURL.Host)
			// 替换 host
			hostParts := strings.Split(parsedURL.Host, ":")
			hostParts[0] = "localhost"
			parsedURL.Host = strings.Join(hostParts, ":")
			// 生成新的 URL
			diUrl = parsedURL.String()
			klog.Infof("newURL is %s", diUrl)
		}

		// 请求各个站点的 lunettesdi server
		reqUrl := fmt.Sprintf("%s/debugpodlist?searchkey=%s&searchvalue=%s", diUrl, key, value)
		klog.Infof("reqUrl is %s", reqUrl)

		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			podlist := make([]*model.DebugPodListTable, 0)

			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			if err != nil {
				klog.Infof("failed to create http request for %s error: %s", reqUrl, err)
				return
			}
			req.SetBasicAuth("admin", "admin")

			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				// 错误先跳过
				klog.Infof("failed to issue http request for %s error: %s", reqUrl, err)
				return
			}
			defer resp.Body.Close()

			// 读取响应内容
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				klog.Infof("ReadAll req %s body error: %s", diUrl, err)
				return
			}

			err = json.Unmarshal(body, &podlist)
			if err != nil {
				klog.Infof("Unmarshal req %s body error: %s", diUrl, err)
				return
			}
			for _, v := range podlist {
				klog.Infof("v.podname is %+v", v.Podname)
				copyV := &model.DebugPodListTable{
					Podname:    v.Podname,
					PodIP:      v.PodIP,
					Cluster:    v.Cluster,
					PodUID:     v.PodUID,
					NodeIP:     v.NodeIP,
					CreateTime: v.CreateTime,
					Namespace:  v.Namespace,
					State:      v.State,
					PodPhase:   v.PodPhase,
					Site:       siteName,
					DebugUrl:   fmt.Sprintf("%s/d/lunettes-debugpod/lunettes-debugpod?orgId=1&var-podinfo=uid&var-podinfovalue=%s", dashboardUrl, v.PodUID),
					Dignosis:   "诊断",
				}
				ch <- copyV
			}
		}()

	}
	wg.Wait()
	close(ch)

	fedPodlist := make([]*model.DebugPodListTable, 0)
	for res := range ch {
		fedPodlist = append(fedPodlist, res)
	}

	return http.StatusOK, fedPodlist, nil

}

func (handler *FedDebugPodListHandler) Process() (int, interface{}, error) {
	defer utils.IgnorePanic("FedDebugPodListHandler.Process ")

	var result interface{}
	var err error
	var httpStatus int

	httpStatus, result, err = handler.QueryFedDebugPodListWithPodUid(handler.requestParams.PodUIDName, handler.requestParams.PodUID)

	return httpStatus, result, err
}

func FedDebugPodListFactory(w http.ResponseWriter, r *http.Request, storage data_access.StorageInterface) Handler {
	return &FedDebugPodListHandler{
		request: r,
		writer:  w,
		storage: storage,
	}
}
