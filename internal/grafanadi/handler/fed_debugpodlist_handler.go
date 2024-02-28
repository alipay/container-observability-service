package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

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
			log.Printf("parsedURL is %+v\n", parsedURL.Host)
			// 替换 host
			hostParts := strings.Split(parsedURL.Host, ":")
			hostParts[0] = "localhost"
			parsedURL.Host = strings.Join(hostParts, ":")
			// 生成新的 URL
			diUrl = parsedURL.String()
			log.Printf("newURL is %s\n", diUrl)
		}

		// 请求各个站点的 lunettesdi server
		reqUrl := fmt.Sprintf("%s/debugpodlist?searchkey=%s&searchvalue=%s", diUrl, key, value)
		log.Printf("reqUrl is %s\n", reqUrl)

		wg.Add(1)
		sem <- struct{}{} // 获取信号量
		go func() {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			podlist := make([]*model.DebugPodListTable, 0)
			resp, err := http.Get(reqUrl)
			if err != nil {
				// 错误先跳过
				log.Printf("get url: %s error: %s\n", diUrl, err)
				return
			}
			defer resp.Body.Close()

			// 读取响应内容
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Printf("ReadAll req %s body error: %s\n", diUrl, err)
				return
			}

			err = json.Unmarshal(body, &podlist)
			if err != nil {
				log.Printf("Unmarshal req %s body error: %s\n", diUrl, err)
				return
			}
			for _, v := range podlist {
				log.Printf("v.podname is %+v\n", v.Podname)
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
					DebugUrl:   fmt.Sprintf("%s/d/eavesdropping-debugpod/eavesdropping-debugpod?orgId=1&var-podinfo=uid&var-podinfovalue=%s", dashboardUrl, v.PodUID),
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
