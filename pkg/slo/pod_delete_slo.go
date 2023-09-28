package slo

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/alipay/container-observability-service/pkg/xsearch"

	"github.com/alipay/container-observability-service/pkg/metas"

	v1 "k8s.io/api/core/v1"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/utils"

	"k8s.io/klog/v2"
)

var saveSLOData = saveSLODataToZSearch

const (
	TIMEOUT        = "timeout"
	SUCCESS        = "success"
	CNIALLOC       = "cnialloc"
	FINISH         = "finish"
	KUBEKILLINGPOD = "killingpod"
	// maximum latency between pod.deletionTimestamp and auditEvent.StageTimestamp
	// Periodically sending request to delete a pod that is stuck in terminating state will not count to podDeleteResult.
	auditMaxLatency = time.Minute * 30

	DeleteMileStoneType              = "delete"
	StaleDeletionMileStoneType       = "stale" // stale means pod deletion duration in 10min to 24h
	TerminatingDeletionMileStoneType = "terminating"

	PodDeleteTimeoutPeriod      = time.Minute * 10
	PodStaleTimeoutPeriod       = time.Hour * 1  // an hour
	PodTerminatingTimeoutPeriud = time.Hour * 24 // a day
)

func podUniqueKey(ms *xsearch.PodDeleteMileStone) string {
	return fmt.Sprintf("%s/%s/%s", ms.Namespace, ms.PodName, ms.PodUID)
}

type PodCache struct {
	processedCache *sync.Map // key is podUniqueKey, value is timestamp it's first added into this map
}

// true if pod has processed, false if pod is the first time seen
func (c *PodCache) SeenPod(ms *xsearch.PodDeleteMileStone) bool {
	if ms == nil {
		return true
	}
	key := podUniqueKey(ms)
	_, ok := c.processedCache.Load(key)
	return ok
}

func (c *PodCache) RecordPod(ms *xsearch.PodDeleteMileStone, tm time.Time) {
	if ms == nil {
		return
	}
	key := podUniqueKey(ms)
	if _, ok := c.processedCache.Load(key); ok {
		return
	}
	c.processedCache.Store(key, &tm)
}

// remove entries elder than 30 min (auditMaxLatency) periodically
func (c *PodCache) compact() {
	wait.Forever(func() {
		now := currentDeleteAuditTime
		c.processedCache.Range(func(key, value interface{}) bool {
			if stageTime, ok := value.(*time.Time); ok {
				if now.Sub(*stageTime) > auditMaxLatency {
					c.processedCache.Delete(key)
				}
			}
			return true
		})
	}, time.Minute)
}

var (
	podCache = &PodCache{processedCache: &sync.Map{}}

	PodDeleteMileStoneMap  *utils.SafeMap // podKey -> podDeleteMileStone
	deleteQueue            *queue.BoundedQueue
	currentDeleteAuditTime time.Time
)

func init() {
	PodDeleteMileStoneMap = utils.NewSafeMap()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if PodDeleteMileStoneMap == nil {
					PodDeleteMileStoneMap = utils.NewSafeMap()
				}
				sloOngoingSize.WithLabelValues("podDelete").Set(float64(PodDeleteMileStoneMap.Size()))
				checkTimeout()
			}
		}
	}()

	//add graceful clear
	xsearch.XSearchClear.AddCleanWork(func() {
		xsearch.SaveDeleteSloMilestoneMapToZsearch(PodDeleteMileStoneMap)
	})

	go podCache.compact()

	deleteQueue = queue.NewBoundedQueue("slo-watcher-delete", 100000, nil)
	deleteQueue.StartLengthReporting(10 * time.Second)
	deleteQueue.IsLockOSThread = true
	deleteQueue.IsDropEventOnFull = false
	deleteQueue.StartConsumers(1, func(item interface{}) {
		event, ok := item.(*shares.AuditEvent)
		if !ok || event == nil {
			return
		}
		doDeleteSLO(event)
	})
}

func checkTimeout() {
	toDeleteKeys := make([]string, 0)
	PodDeleteMileStoneMap.IterateWithFunc(func(i interface{}) {
		deleteMS, ok := i.(*xsearch.PodDeleteMileStone)
		if !ok {
			return
		}
		if currentDeleteAuditTime.After(deleteMS.DeleteTimeoutTime) {
			toDeleteKeys = append(toDeleteKeys, deleteMS.Key)
		}
	})
	for _, key := range toDeleteKeys {
		finishMileStoneWithResult(key, TIMEOUT, currentDeleteAuditTime)
	}
}

func doDeleteSLO(auditEvent *shares.AuditEvent) {
	currentDeleteAuditTime = auditEvent.StageTimestamp.Time
	//处理patch操作
	processPatchOp(auditEvent)
	//delete
	processDeleteOp(auditEvent)
	//event
	processEvents(auditEvent)
}

func processEvents(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processEvents")
	if auditEvent.ResponseStatus.Code >= 300 || auditEvent.ObjectRef.Resource != "events" {
		return
	}
	if auditEvent.Verb != "create" || auditEvent.ObjectRef.Subresource != "" {
		return
	}
	/*event := &v1.Event{}
	if err := json.Unmarshal(auditEvent.RequestObject.Raw, event); err != nil {
		klog.Infof("Unmarshal failed: %s", auditEvent.AuditID)
		return
	}*/
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	ok := false
	var event *v1.Event
	if event, ok = auditEvent.ResponseRuntimeObj.(*v1.Event); !ok {
		return
	}

	if event.InvolvedObject.Kind != "Pod" {
		return
	}
	clusterName := auditEvent.Annotations["cluster"]
	namespace := event.InvolvedObject.Namespace
	name := event.InvolvedObject.Name
	podKey := genPodKey(clusterName, namespace, name)
	v, ok := PodDeleteMileStoneMap.Get(podKey)
	if !ok {
		return
	}
	milestone := v.(*xsearch.PodDeleteMileStone)
	milestone.Mutex.Lock()
	defer milestone.Mutex.Unlock()
	if event.Reason == "Killing" {
		if milestone.KubeletKillingTime.IsZero() {
			milestone.KubeletKillingTime = auditEvent.StageTimestamp.Time
			milestone.KubeletKillingHost = event.Source.Host
		}
		d := utils.TimeDiffInSeconds(milestone.CreatedTime, milestone.KubeletKillingTime)
		if d >= 0 {
			metrics.PodDeleteLatency.WithLabelValues(milestone.Cluster, milestone.Namespace, KUBEKILLINGPOD).Observe(d)
		}
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_delete_events").Observe(traceProcessingTime)
}

func processDeleteOp(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processDeleteOp")
	if auditEvent.ObjectRef.Resource != "pods" || auditEvent.Verb != "delete" || auditEvent.ObjectRef.Subresource != "" {
		return
	}
	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	namespace := auditEvent.ObjectRef.Namespace

	//delete api res code
	metrics.PodDeleteApiCode.WithLabelValues(clusterName, namespace, fmt.Sprintf("%d", auditEvent.ResponseStatus.Code)).Inc()

	/*resPod := &v1.Pod{}
	if err := json.Unmarshal(auditEvent.ResponseObject.Raw, resPod); err != nil {
		klog.Error(err)
		return
	}*/
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	ok := false
	var resPod *v1.Pod
	if resPod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok {
		return
	}

	podKey := genPodKey(clusterName, resPod.Namespace, resPod.Name)
	shouldTrace := addToTraceMap(podKey, resPod, auditEvent)
	if !shouldTrace {
		return
	}

	if len(resPod.ObjectMeta.GetFinalizers()) == 0 && (resPod.DeletionGracePeriodSeconds == nil || *resPod.DeletionGracePeriodSeconds == 0) {
		finishMileStoneWithResult(podKey, SUCCESS, currentDeleteAuditTime)
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_delete_deleteOp").Observe(traceProcessingTime)
}

func processPatchOp(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPatchOp")
	if auditEvent == nil || auditEvent.ObjectRef.Resource != "pods" || auditEvent.Verb != "patch" ||
		auditEvent.ObjectRef.Subresource != "" {
		return
	}
	clusterName := auditEvent.Annotations["cluster"]
	/*resPod := &v1.Pod{}
	if err := json.Unmarshal(auditEvent.ResponseObject.Raw, resPod); err != nil {
		klog.Error(err)
		return
	}*/
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	ok := false
	var resPod *v1.Pod
	if resPod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok {
		return
	}

	podKey := genPodKey(clusterName, resPod.Namespace, resPod.Name)

	/**
	 * skip add to TraceMap, since we cannot make sure audit events are processed in same sequence as the requests processed.
	 * We cannot use *requestReceivedTimestamp* , as different apiserver instance may have different local time.
	 *  stageTimestamp is not a good choice, since it relies on local time as well.
	shouldTrace := addToTraceMap(podKey, resPod, auditEvent)
	if !shouldTrace {
		return
	}
	*/
	v, isExist := PodDeleteMileStoneMap.Get(podKey)
	if !isExist {
		return
	}
	mileStone := v.(*xsearch.PodDeleteMileStone)

	// 检测finalizer的移除
	var jsonData interface{}
	var err error
	err = json.Unmarshal(auditEvent.RequestObject.Raw, &jsonData)
	if err != nil {
		klog.Warningf("unmarshal %s with error %v", auditEvent.RequestObject.Raw, err)
		return
	}

	if err != nil {
		klog.V(5).Infof("not found $deleteFromPrimitiveList for %s with error %v", podKey, err)
	}

	mileStone.Mutex.Lock()

	mileStone.RemainingFinalizers = resPod.ObjectMeta.Finalizers
	mileStone.Mutex.Unlock()

	if len(mileStone.RemainingFinalizers) == 0 && (resPod.DeletionGracePeriodSeconds == nil || *resPod.DeletionGracePeriodSeconds == 0) {
		finishMileStoneWithResult(podKey, SUCCESS, currentDeleteAuditTime)
	}

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_delete_patchOp").Observe(traceProcessingTime)
}

func finishMileStoneWithResult(podKey string, result string, currentTime time.Time) {
	defer utils.IgnorePanic("finishMileStoneWithResult")
	v, ok := PodDeleteMileStoneMap.Get(podKey)
	if !ok {
		return
	}

	milestone := v.(*xsearch.PodDeleteMileStone)
	milestone.Mutex.Lock()
	defer milestone.Mutex.Unlock()

	if result == SUCCESS {
		// add to podCache
		podCache.RecordPod(milestone, milestone.CreatedTime)
	}

	//失败时解析具体的finalizer
	if result == TIMEOUT && milestone.RemainingFinalizers != nil && len(milestone.RemainingFinalizers) > 0 {
		result = milestone.RemainingFinalizers[0]
	}

	if milestone.DeleteResult == "" {
		milestone.DeleteResult = result
		milestone.DeleteEndTime = currentTime
	}

	saveSLOData(milestone)
	if milestone.Type == DeleteMileStoneType {
		metrics.PodDeleteResult.WithLabelValues(milestone.Cluster, milestone.Namespace, milestone.NodeIP, result).Inc()
		if result != SUCCESS {
			milestone.Type = StaleDeletionMileStoneType
			milestone.DeleteResult = ""
			milestone.DeleteTimeoutTime = milestone.DeleteTimeoutTime.Add(PodStaleTimeoutPeriod - PodDeleteTimeoutPeriod)
		} else {
			// also count success records for 24h / 7d metrics
			metrics.PodDeleteResultInDay.WithLabelValues(milestone.Cluster, milestone.Namespace, result).Inc()
			metrics.PodDeleteResultInWeek.WithLabelValues(milestone.Cluster, milestone.Namespace, result).Inc()
			if utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime) >= 0 {
				metrics.PodDeleteLatency.WithLabelValues(milestone.Cluster, milestone.Namespace, FINISH).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
				metrics.PodDeleteLatencyQuantiles.WithLabelValues(getPodType(milestone)).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
			}
			PodDeleteMileStoneMap.Delete(podKey)
		}
	} else if milestone.Type == StaleDeletionMileStoneType {
		metrics.PodDeleteResultInDay.WithLabelValues(milestone.Cluster, milestone.Namespace, result).Inc()
		if result != SUCCESS {
			milestone.Type = TerminatingDeletionMileStoneType
			milestone.DeleteResult = ""
			milestone.DeleteTimeoutTime = milestone.DeleteTimeoutTime.Add(PodTerminatingTimeoutPeriud - PodStaleTimeoutPeriod)
		} else {
			// also count this as success for 7d metrics
			metrics.PodDeleteResultInWeek.WithLabelValues(milestone.Cluster, milestone.Namespace, result).Inc()
			if utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime) >= 0 {
				metrics.PodDeleteLatency.WithLabelValues(milestone.Cluster, milestone.Namespace, FINISH).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
				metrics.PodDeleteLatencyQuantiles.WithLabelValues(getPodType(milestone)).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
			}
			PodDeleteMileStoneMap.Delete(podKey)
		}
	} else if milestone.Type == TerminatingDeletionMileStoneType {
		metrics.PodDeleteResultInWeek.WithLabelValues(milestone.Cluster, milestone.Namespace, result).Inc()
		if utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime) >= 0 {
			// no matter whether finally this pod is deleted, just mark it as 1 week
			metrics.PodDeleteLatency.WithLabelValues(milestone.Cluster, milestone.Namespace, FINISH).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
			metrics.PodDeleteLatencyQuantiles.WithLabelValues(getPodType(milestone)).Observe(utils.TimeDiffInSeconds(milestone.CreatedTime, currentTime))
		}
		PodDeleteMileStoneMap.Delete(podKey)
	} else {
		klog.Warningf("unknown type %q for %s", milestone.Type, podKey)
		PodDeleteMileStoneMap.Delete(podKey)
	}
}

func saveSLODataToZSearch(milestone *xsearch.PodDeleteMileStone) {
	sloData, err := json.Marshal(milestone)
	if err == nil {
		e := xsearch.SaveSloTraceData(milestone.Cluster, milestone.Namespace, milestone.PodName, milestone.PodUID, milestone.Type, sloData)
		if e != nil {
			klog.Info(e)
		}
	}
}

// return true for pod that in podDeleteMileStoneMap, otherwise, return false
func addToTraceMap(podKey string, resPod *v1.Pod, auditEvent *shares.AuditEvent) bool {
	clusterName := auditEvent.Annotations["cluster"]
	_, ok := PodDeleteMileStoneMap.Get(podKey)
	if ok {
		return true
	}
	var pod *v1.Pod
	if pod = auditEvent.TryGetPodFromEvent(); pod == nil {
		return false
	}
	// multi delete requests will not update pod.DeletionTimestamp,
	// that is once DeletionTimestamp is set, it will not be updated.
	// TODO how to make sure each pod is processed only once?
	if resPod != nil && resPod.DeletionTimestamp != nil && !resPod.DeletionTimestamp.IsZero() && auditEvent.StageTimestamp.Sub(resPod.DeletionTimestamp.Time) < auditMaxLatency {
		if resPod.DeletionTimestamp != nil && !resPod.DeletionTimestamp.IsZero() {
			deleteMs := &xsearch.PodDeleteMileStone{
				Cluster:             clusterName,
				Namespace:           resPod.Namespace,
				PodName:             resPod.Name,
				PodUID:              string(resPod.UID),
				NodeIP:              resPod.Status.HostIP,
				TrigerAuditLog:      string(auditEvent.AuditID),
				Type:                DeleteMileStoneType,
				CreatedTime:         auditEvent.StageTimestamp.Time,
				DeleteTimeoutTime:   auditEvent.StageTimestamp.Time.Add(PodDeleteTimeoutPeriod),
				RemainingFinalizers: resPod.ObjectMeta.GetFinalizers(),
				DeleteResult:        "",
				DebugUrl:            fmt.Sprintf("http://host:port/api/v1/debugpod?uid=%s", string(resPod.UID)),
				IsJob:               metas.IsJobPod(resPod),
				Key:                 podKey,
				Mutex:               sync.Mutex{},
				LifeDuration:        resPod.DeletionTimestamp.Sub(resPod.CreationTimestamp.Time),
			}
			if podCache.SeenPod(deleteMs) {
				return false
			}
			PodDeleteMileStoneMap.Set(podKey, deleteMs)
			return true
		}
	}
	return false
}

func getPodType(milestone *xsearch.PodDeleteMileStone) string {
	var podType = "app"
	if milestone != nil && milestone.IsJob {
		podType = "job"
	}
	return podType
}
