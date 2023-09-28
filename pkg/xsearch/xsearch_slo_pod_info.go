package xsearch

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/olivere/elastic"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	podInfoIndexName = "slo_pod_info"
	podInfoTypeName  = "_doc"

	sloPodInfoMapping = `
	{
		"mappings" : {
		  "_doc" : {
			"properties" : {
			  "podName" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },          
			  "bizId" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "clusterName" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "namespace" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },          
			  "deliveryStatus" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "startTime" : {
				"type" : "date"
			  },
			  "bizSource" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },  
			  "stageTimestamp" : {
				"type" : "date"
			  },          
			  "deliveryProgress" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "currentTime" : {
				"type" : "date"
			  },
			  "auditID" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "podIP" : {
				"type" : "keyword",
				"ignore_above" : 256
			  },
			  "podUID" : {
				"type" : "keyword",
				"ignore_above" : 256
			  }
			}
		  }
	  }
	}
	`
)

var (
	podInfoCache     utils.LRU
	podInfoCacheSize int        = 20000
	cacheLock        sync.Mutex = sync.Mutex{}

	podInfoCacheForDocID utils.LRU
	cacheForDocIDLock    = sync.Mutex{}

	podInfoCriticalReginLocks = utils.LRUCache("pod_info_doc_id_locker", 20000)
	criticalForDocIDLock      = sync.Mutex{}
)

type CondLocker struct {
	locker *sync.Mutex
	wg     *sync.WaitGroup
	canDo  bool
}

func init() {
	flag.IntVar(&podInfoCacheSize, "pod_info_cache_size", 200000, "cache size for pod info")
}

func getPodInfoCacheForUID() utils.LRU {
	if podInfoCache == nil {
		cacheLock.Lock()
		if podInfoCache == nil {
			klog.V(7).Infof("pod_info_cache_size: %d", podInfoCacheSize)
			podInfoCache = utils.LRUCache("pod_info_uid", podInfoCacheSize)
			InitSloPodInfo(cluster, podInfoCache)
		}
		cacheLock.Unlock()
	}

	return podInfoCache
}

func getPodInfoCacheForDocID() utils.LRU {
	if podInfoCacheForDocID == nil {
		cacheForDocIDLock.Lock()
		if podInfoCacheForDocID == nil {
			klog.V(7).Infof("pod_info_cache_size: %d", podInfoCacheSize)
			podInfoCacheForDocID = utils.LRUCache("pod_info_docid", podInfoCacheSize)
		}
		cacheForDocIDLock.Unlock()
	}

	return podInfoCacheForDocID
}

func getPodInfoCriticalLockForUid(docID string) *CondLocker {
	if condLocker := podInfoCriticalReginLocks.Get(docID); condLocker == nil {
		criticalForDocIDLock.Lock()
		if condLocker = podInfoCriticalReginLocks.Get(docID); condLocker == nil {
			podInfoCriticalReginLocks.Put(docID, &CondLocker{
				locker: &sync.Mutex{},
				wg:     &sync.WaitGroup{},
				canDo:  true,
			})
		}
		criticalForDocIDLock.Unlock()
	}

	return podInfoCriticalReginLocks.Get(docID).(*CondLocker)
}

type SloPodInfo struct {
	PodUid         string `json:"podUID"`
	PodName        string `json:"podName"`
	DeliveryStatus string `json:"deliveryStatus"`
}

// SavePodInfoToZSearch 在 slo_pod_info index 增加 pod
var podInfoBuffer *utils.BufferUtils = nil

func SavePodInfoToZSearch(cluster string, pod *corev1.Pod, deliveryStatus string, currentTime time.Time, auditID string, deliveryProgress string, isUpdate bool) error {
	defer utils.IgnorePanic("SavePodInfoToZSearch")

	if podInfoBuffer == nil {
		podInfoBuffer = utils.NewBufferUtils(podInfoIndexName, 1000, 10*time.Second, false, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do bulk, data size: %d", len(datas))
			err := utils.ReTry(func() error {
				bulkService := esClient.Bulk().Index(podInfoIndexName).Refresh("true")
				for id, data := range datas {
					doc := elastic.NewBulkUpdateRequest().Type(podInfoTypeName).Id(id).Doc(data).UseEasyJSON(true).Upsert(data)
					bulkService = bulkService.Add(doc)
				}
				bulkRs, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}

				for _, itemMap := range bulkRs.Items {
					if itemMap == nil {
						continue
					}
					for _, item := range itemMap {
						if item != nil {
							getPodInfoCacheForDocID().Put(item.Id, item.Index)
						}
					}
				}
				return nil
			}, 1*time.Second, 5)

			if err != nil {
				return err
			}
			return nil
		},
		)

		podInfoBuffer.DoClearData()
		//add graceful clear
		XSearchClear.AddCleanWork(func() {
			podInfoBuffer.Stop()
		})
	}

	begin := time.Now()
	defer func() {
		cost := utils.TimeSinceInMilliSeconds(begin)
		metrics.DebugMethodDurationMilliSeconds.WithLabelValues("SavePodInfoToZSearch").Observe(cost)
	}()

	dic := make(map[string]interface{})
	if pod != nil {
		// 下面四个为提供给前端搜索的字段
		dic["podName"] = pod.Name
		dic["bizId"] = "N/A"
		dic["clusterName"] = cluster
		dic["namespace"] = pod.Namespace
		dic["deliveryStatus"] = deliveryStatus // 交付状态：失败，进行中，未开始，已完成
		dic["startTime"] = pod.ObjectMeta.CreationTimestamp.Time

		// 下面四个为提供给前端展示的字段
		dic["currentTime"] = currentTime           // 前端字段名称为"耗时"
		dic["deliveryProgress"] = deliveryProgress // 前端字段名称为"交付进度"，即交付大类

		// 暂时
		dic["podUID"] = pod.UID
		dic["podIP"] = pod.Status.PodIP
		dic["stageTimestamp"] = currentTime
		if !isUpdate {
			dic["auditID"] = auditID
		}

	}

	docID := fmt.Sprintf("%s_%s", cluster, dic["podUID"])

	dic["index"] = podInfoIndexName
	//insert to es with retry
	err := podInfoBuffer.SaveData(docID, dic)

	if err != nil {
		klog.Errorf("SavePodInfoToZSearch Err: %s", err)
	}

	//更新缓存
	result := []*SloPodInfo{
		{
			PodUid:         string(pod.UID),
			PodName:        pod.Name,
			DeliveryStatus: "未开始",
		},
	}
	getPodInfoCacheForUID().Put(string(pod.UID), result)

	return nil
}

// 查询一个 PodInfo 所属于的 Index
// 由于 slo_pod_info 是天级别的索引，需要先找到这个 docID 对应的索引名称，然后再写入
func GetSloPodInfoIndexByDocID(docID string) string {

	defer utils.IgnorePanic("GetSloPodInfoIndexByDocID")

	begin := time.Now()
	defer func() {
		metrics.DebugMethodDurationMilliSeconds.
			WithLabelValues("GetSloPodInfoIndexByDocID").Observe(utils.TimeSinceInMilliSeconds(begin))
	}()

	if index := getPodInfoCacheForDocID().Get(docID); index != nil {
		return index.(string)
	}

	//对于同一个docID，有一把临界区锁，下边从db获取podInfo的代码只能执行一次，避免大并发查询给db带来压力
	criticalLocker := getPodInfoCriticalLockForUid(docID)
	criticalLocker.locker.Lock()
	//只有第一个go routine才去db上查询
	if !criticalLocker.canDo {
		criticalLocker.locker.Unlock()
		criticalLocker.wg.Wait()
	} else {
		//进入查询工作时应立即释放锁，并且设置后续不可继续工作，以便在查询时让后续的携程进入等待
		criticalLocker.canDo = false
		criticalLocker.wg.Add(1)
		criticalLocker.locker.Unlock()
		//无论结果如何，结束后设置后续go routine可以工作
		defer func() {
			criticalLocker.canDo = true
			criticalLocker.wg.Done()
		}()
	}

	//对于被唤醒的goroutine，再次拿一下缓存中的数据，对于不是第一个的将直接命中
	if index := getPodInfoCacheForDocID().Get(docID); index != nil {
		return index.(string)
	}

	index := ""
	var idx *string = &index
	err := utils.ReTry(func() error {
		stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("_id: \"%s\"", docID))
		query := elastic.NewBoolQuery().Must(stringQuery)

		searchResult, err := esClient.Search().Index(podInfoIndexName).Type(podInfoTypeName).
			Query(query).Size(1).
			Sort("startTime", false).
			Do(context.Background())
		if err != nil {
			//如果出错，则只放一个进来继续重试
			err = fmt.Errorf("[GetSloPodInfoIndexByPodUID] Failed to get podInfo %s, err is %v", docID, err)
			fmt.Println(err)
			return err
		}

		for _, hit := range searchResult.Hits.Hits {
			getPodInfoCacheForDocID().Put(docID, hit.Index)
			*idx = hit.Index
			return nil
		}

		//没查到pod info应该是新创建的，则存入空值(读取的时候要做判断)，后续goroutine不需
		getPodInfoCacheForDocID().Put(docID, "")
		return nil
	}, 10*time.Millisecond, 10)

	if err != nil {
		klog.Errorf("[GetSloPodInfoIndexByPodUID] doc id %s not found in index slo_pod_info", docID)
	}

	return index
}

// 查询 SLO Pod
func InitSloPodInfo(cluster string, podInfoCacheUid utils.LRU) error {
	defer utils.IgnorePanic("QuerySloPodInfo")

	now := time.Now()
	rangeQuery := elastic.NewRangeQuery("stageTimestamp").
		From(now.Add(-1 * time.Hour)).To(now).
		IncludeLower(true).
		IncludeUpper(false).
		TimeZone("UTC")
	query := elastic.NewBoolQuery().Must(rangeQuery)

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("clusterName: \"%s\"", cluster))
	query = query.Must(stringQuery)
	s, _ := query.Source()
	klog.Infof("query form init: %s", utils.Dumps(s))

	pageSize := 5000
	scroll := esClient.Scroll(podInfoIndexName).Type(podInfoTypeName).Query(query).Sort("stageTimestamp", true).Size(pageSize).Scroll("5m")
	total := 0
	for {
		results, err := scroll.Do(context.TODO())
		if results != nil && results.Error != nil {
			klog.Errorf("results failed error: %s, FailedShards len: %d", results.Error.Reason, len(results.Error.FailedShards))
			return err
		}

		if err == io.EOF {
			break
		}

		if err != nil && err != io.EOF {
			klog.Errorf("init pod info to cache, err: %s", err.Error())
			return err
		}

		if results != nil {
			for _, hit := range results.Hits.Hits {
				podInfo := &SloPodInfo{}
				if er := json.Unmarshal(*hit.Source, podInfo); er == nil {
					//add pod info list
					rs := []*SloPodInfo{podInfo}
					podInfoCacheUid.Put(podInfo.PodUid, rs)
					getPodInfoCacheForDocID().Put(hit.Id, hit.Index)
					klog.Infof("cache put pod uid :%s , doc id: %s", podInfo.PodUid, hit.Id)
					total += 1
				}
			}
		}
	}

	klog.Infof("query form init, cache size: %d", total)
	return nil
}
