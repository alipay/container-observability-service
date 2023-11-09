package xsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/olivere/elastic/v7"
	"k8s.io/klog/v2"
)

const (
	PodDeleteMileStoneIndex = "pod_delete_milestone_cache"
	PodDeleteMileStoneType  = "_doc"
)

func init() {

}

// pod删除SLO相关
type PodDeleteMileStone struct {
	Cluster             string
	Namespace           string
	PodName             string
	PodUID              string
	Type                string
	TrigerAuditLog      string
	NodeIP              string
	DeleteResult        string //删除结果：success、failed reason
	KubeletKillingHost  string
	DebugUrl            string
	HostingStatus       string //删除托管状态
	CreatedTime         time.Time
	DeleteEndTime       time.Time
	KubeletKillingTime  time.Time
	LifeDuration        time.Duration // DeletionTimeStamp - CreationTimeStamp
	RemainingFinalizers []string
	DeleteTimeoutTime   time.Time
	IsJob               bool
	Key                 string
	Mutex               sync.Mutex
}

func SaveDeleteSloMilestoneMapToZsearch(podDeleteMileStoneMap *utils.SafeMap) {
	klog.Infof("Saving podDeleteMileStoneMap to zsearch, map size is %d", podDeleteMileStoneMap.Size())
	podDeleteMileStoneMap.IterateWithFunc(func(i interface{}) {
		deleteMS, ok := i.(*PodDeleteMileStone)
		if !ok {
			return
		}
		err := saveDeleteSloMilestoneToZsearch(deleteMS)
		if err != nil {
			klog.Errorf("failed to save podDeleteMileStoneMap %s, err is %v", deleteMS.Key, err)
		}
	})
	klog.Infof("Finished saving podDeleteMileStoneMap to zsearch, map size is %d", podDeleteMileStoneMap.Size())
}

func saveDeleteSloMilestoneToZsearch(podDeleteMs *PodDeleteMileStone) error {
	podDeleteMsStr, err := json.Marshal(podDeleteMs)
	if err != nil {
		klog.Errorf("saveDeleteSloMilestoneToZsearch Error marshalling deleteSloMilestone: %s", err)
	}

	defer utils.IgnorePanic("saveDeleteSloMilestoneToZsearch")

	begin := time.Now()
	defer func() {
		metrics.DebugMethodDurationMilliSeconds.
			WithLabelValues("saveDeleteSloMilestoneToZsearch").Observe(utils.TimeSinceInMilliSeconds(begin))
	}()

	docID := fmt.Sprintf("%s", podDeleteMs.Key)

	indexName := GetIndexNameForPodDeleteMileStone(podDeleteMs.Cluster)

	err = utils.ReTry(func() error {
		_, err := esClient.Index().Index(indexName).Type(PodDeleteMileStoneType).
			Id(docID).
			BodyString(string(podDeleteMsStr)).
			Do(context.Background())
		if err != nil {
			klog.Errorf("saveDeleteSloMilestoneToZsearch Error saving deleteSloMilestone: %s", err)
			return err
		}

		return nil
	}, 1*time.Second, 20)

	if err != nil {
		klog.Errorf("saveDeleteSloMilestoneToZsearch Error saving deleteSloMilestone after retry: %s", err)
	}

	return nil
}

// 获取所有 PodDeleteMilestone
func GetAllPodDeleteMilestoneByScroll(cluster string) (*utils.SafeMap, error) {
	begin := time.Now()
	defer func() {
		metrics.ObserveQueryMethodDuration("GetAllPodDeleteMilestone", begin)
	}()

	klog.Infof("GetAllPodDeleteMilestoneByScroll from zsearch, %s", cluster)

	result := utils.NewSafeMap()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("Cluster: \"%s\"", cluster))
	klog.Infof("trimmed clusterName is %s", cluster)

	query := elastic.NewBoolQuery().Must(stringQuery)

	indexName := GetIndexNameForPodDeleteMileStone(cluster)
	scroller := esClient.Scroll().
		Index(indexName).
		Type(PodDeleteMileStoneType).
		Query(query).
		Sort("CreatedTime", true).
		Size(100)

	for {
		searchResult, err := scroller.Do(context.TODO())
		if err == io.EOF {
			// No remaining documents matching the search
			break
		}
		if err != nil {
			klog.Infof("xsearch indexName: %s", indexName)
			klog.Errorf("GetAllPodDeleteMilestoneByScroll failed, err is %v", err)
			return result, err
		}
		if searchResult == nil {
			klog.Errorf("GetAllPodDeleteMilestoneByScroll search result is nil, %v", err)
			return result, err
		}

		for _, hit := range searchResult.Hits.Hits {
			podDeleteMileStone := &PodDeleteMileStone{}
			if er := json.Unmarshal(hit.Source, podDeleteMileStone); er == nil {
				result.Set(podDeleteMileStone.Key, podDeleteMileStone)
			}
		}
	}

	klog.Infof("Successfully fetched %d podDeleteMileStone from zsearch for recovery.", result.Size())
	return result, nil
}

func GetIndexNameForPodDeleteMileStone(cluster string) string {
	return PodDeleteMileStoneIndex + "_" + strings.Replace(cluster, "-", "_", -1)
}

// 删除所有 podDeleteMileStone
func DeleteAllPodDeleteMilestone(cluster string) *utils.SafeMap {

	result := utils.NewSafeMap()

	stringQuery := elastic.NewQueryStringQuery(fmt.Sprintf("Cluster: \"%s\"", cluster))
	klog.Infof("clusterName is %s", cluster)

	query := elastic.NewBoolQuery().Must(stringQuery)

	indexName := GetIndexNameForPodDeleteMileStone(cluster)

	searchResult, err := esClient.DeleteByQuery(indexName).Type(PodDeleteMileStoneType).
		Query(query).
		Do(context.Background())
	if err != nil {
		klog.Errorf("DeleteAllPodDeleteMilestone failed to get from zsearch %v", err)
		return utils.NewSafeMap()
	}

	klog.Infof("Successfully deleted %d podDeleteMileStone from zsearch for recovery.", searchResult.Deleted)
	return result
}
