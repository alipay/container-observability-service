package slo

import (
	"container/list"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/shareutils"

	"github.com/alipay/container-observability-service/pkg/reason"

	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/reason/analyzers"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/spans"
	lua "github.com/yuin/gopher-lua"
	"k8s.io/apimachinery/pkg/types"

	"github.com/alipay/container-observability-service/pkg/xsearch"

	"k8s.io/klog/v2"

	v1 "k8s.io/api/core/v1"

	"github.com/alipay/container-observability-service/pkg/queue"
	"github.com/alipay/container-observability-service/pkg/utils"
)

// pod升级SLO相关
type PodUpgradeMileStone struct {
	Cluster              string
	Namespace            string
	PodName              string
	PodUID               string
	Type                 string
	TriggerAuditLog      string
	UpgradeResult        string
	UpgradeContainerName string
	UpdateStatus         string
	NodeIP               string
	CreatedTime          time.Time //开始升级时间
	UpgradeEndTime       time.Time
	UpgradeTimeoutTime   time.Time
	DebugUrl             string
	//内部变量
	key               string
	subKey            string
	mutex             sync.Mutex
	upgradeContainers []string
	trickTime         *time.Time
}

const (
	UPGRADE_TIMEOUT             = "timeout"
	UPGRADE_SUCCESS             = "success"
	UPGRADE_BEFOREFINISH        = "beforeFinish"
	UPGRADE_FAILED              = "failed"
	UPGRADE_RUN_CONTAINER_ERROR = "RunContainerError"

	timeoutDuration = 9 * time.Minute
)

var (
	podUpgradeMileStoneMap *utils.SafeMap // podKey -> *PodUpgradeMileStone
	upgradeQueue           *queue.BoundedQueue
	podUpgradeAuditLogMap  *utils.SafeMap
	auditTimeChan          chan *time.Time
)

func init() {
	podUpgradeMileStoneMap = utils.NewSafeMap()
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sloOngoingSize.WithLabelValues("podUpgrade").Set(float64(podUpgradeMileStoneMap.Size()))
			}
		}
	}()

	auditTimeChan = make(chan *time.Time, 10000)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		for {
			// 连续取最后一个
			t := <-auditTimeChan
			currentAuditLen := len(auditTimeChan)
			for i := 0; i < currentAuditLen; i++ {
				t = <-auditTimeChan
			}
			checkUpgradeTimeout(t)
		}
	}()

	podUpgradeAuditLogMap = utils.NewSafeMap()
	upgradeQueue = queue.NewBoundedQueue("slo-watcher-upgrade", 100000, nil)
	upgradeQueue.StartLengthReporting(10 * time.Second)
	upgradeQueue.StartConsumers(1, func(item interface{}) {
		defer utils.IgnorePanic("upgradeQueueConsumer")

		event, ok := item.(*shares.AuditEvent)
		if !ok || event == nil {
			return
		}

		processUpgradeTrigger(event)
		collectPodUpgradeAuditLog(event)
		processUpgradeStatus(event)
		syncAuditTimeForUpgrade(event)
	})

	spans.RegisterLuaHelperFunc("finish_upgrade", isFinishUpgradeHelper)
}

// syncAuditTime 同步审计日志时间
func syncAuditTimeForUpgrade(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("syncAuditTimeForUpgrade")
	t := auditEvent.StageTimestamp.Time
	select {
	case auditTimeChan <- &t:
	default:
	}
}

func checkUpgradeTimeout(auditTime *time.Time) {
	var toDeleteMs []*PodUpgradeMileStone
	podUpgradeMileStoneMap.IterateWithFunc(func(i interface{}) {
		upgradeMsMap, ok := i.(map[string]*PodUpgradeMileStone)
		if !ok {
			return
		}
		for _, upgradeMS := range upgradeMsMap {
			if auditTime.After(upgradeMS.UpgradeTimeoutTime) {
				if upgradeMS.UpgradeResult == "" {
					upgradeMS.UpgradeResult = UPGRADE_TIMEOUT
				}
				upgradeMS.trickTime = auditTime
				toDeleteMs = append(toDeleteMs, upgradeMS)
			}
		}
	})
	finishUpgradeMileStoneWithResult(toDeleteMs)
}

func finishUpgradeMileStoneWithResult(milestones []*PodUpgradeMileStone) {
	defer utils.IgnorePanic("finishUpgradeMileStoneWithResult")

	for _, milestone := range milestones {
		milestone.mutex.Lock()
		mileStoneMap, ok := podUpgradeMileStoneMap.Get(milestone.key)
		if !ok || mileStoneMap == nil {
			milestone.mutex.Unlock()
			continue
		}
		_, ok = mileStoneMap.(map[string]*PodUpgradeMileStone)[milestone.subKey]
		if !ok {
			milestone.mutex.Unlock()
			continue
		}

		milestone.UpgradeEndTime = *milestone.trickTime
		if milestone.UpgradeResult != UPGRADE_SUCCESS && milestone.UpgradeResult != UPGRADE_BEFOREFINISH {
			orgResult := milestone.UpgradeResult
			reason := milestone.analyzeFailedReason()
			if !strings.Contains(strings.ToLower(reason), "sandbox") {
				milestone.UpgradeResult = reason
			}
			klog.V(8).Infof("analysis upgrade for pod %s, result: %s, orgResult: %s, start:%s, end:%s\n", milestone.PodName, milestone.UpgradeResult, orgResult,
				milestone.CreatedTime, *milestone.trickTime)
		}
		sloData, err := json.Marshal(milestone)
		if err == nil {
			e := xsearch.SaveSloTraceData(milestone.Cluster, milestone.Namespace, milestone.PodName, milestone.PodUID, "upgrade", sloData)
			if e != nil {
				klog.Info(e)
			}
		}

		klog.Infof("upgrade finish for %s result %s\n", milestone.PodName, milestone.UpgradeResult)
		metrics.PodUpgradeResultCounter.WithLabelValues(milestone.Cluster, milestone.Namespace, milestone.NodeIP, milestone.UpgradeResult).Inc()

		delete(mileStoneMap.(map[string]*PodUpgradeMileStone), milestone.subKey)
		if len(mileStoneMap.(map[string]*PodUpgradeMileStone)) == 0 {
			podUpgradeMileStoneMap.Delete(milestone.key)
			podUpgradeAuditLogMap.Delete(milestone.key)
		}

		milestone.mutex.Unlock()
	}

}

func processUpgradeStatus(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processUpgradeStatus")

	if (auditEvent.Verb != "patch" && auditEvent.Verb != "update" && auditEvent.Verb != "delete") || auditEvent.ObjectRef.Resource != "pods" {
		return
	}

	// delete verb
	if auditEvent.Verb == "delete" {
		processDelete(auditEvent)
		return
	}

	// update/patch verb
	processUpdateUpgrade(auditEvent)
}

func processUpdateUpgrade(auditEvent *shares.AuditEvent) {
	var reqPod *v1.Pod
	ok := false
	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	if reqPod, ok = auditEvent.RequestRuntimeObj.(*v1.Pod); !ok || reqPod == nil {
		return
	}

	var resPod *v1.Pod
	ok = false
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	if resPod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok {
		return
	}

	//正处于删除过程中
	if resPod.DeletionTimestamp != nil {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	podKey := genPodKey(clusterName, resPod.Namespace, resPod.Name)
	v, ok := podUpgradeMileStoneMap.Get(podKey)
	if !ok {
		return
	}

	podMsMap, ok := v.(map[string]*PodUpgradeMileStone)
	if !ok {
		return
	}

	var toFinish []*PodUpgradeMileStone
	for _, podMs := range podMsMap {
		isFinished, result := isFinishUpgrade(podMs, resPod)
		podMs.UpgradeResult = result
		if isFinished {
			podMs.trickTime = &auditEvent.StageTimestamp.Time
			toFinish = append(toFinish, podMs)
		}
	}
	finishUpgradeMileStoneWithResult(toFinish)

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_upgrade_status").Observe(traceProcessingTime)
}
func processDelete(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processDeleteUpgrade")
	if auditEvent.ObjectRef.Resource != "pods" || auditEvent.Verb != "delete" || auditEvent.ObjectRef.Subresource != "" {
		return
	}
	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	if auditEvent.ResponseRuntimeObj == nil {
		return
	}
	ok := false
	var resPod *v1.Pod
	if resPod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok {
		return
	}

	podKey := genPodKey(clusterName, resPod.Namespace, resPod.Name)
	v, ok := podUpgradeMileStoneMap.Get(podKey)
	if !ok {
		return
	}

	podMsMap, ok := v.(map[string]*PodUpgradeMileStone)
	if !ok {
		return
	}

	var toFinish []*PodUpgradeMileStone
	for _, podMs := range podMsMap {
		if time.Now().Before(podMs.CreatedTime.Add(10 * time.Second)) {
			podMs.UpgradeResult = UPGRADE_BEFOREFINISH
		}
		podMs.trickTime = &auditEvent.StageTimestamp.Time
		toFinish = append(toFinish, podMs)
	}
	finishUpgradeMileStoneWithResult(toFinish)

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_upgrade_status").Observe(traceProcessingTime)
}

func processUpgradeTrigger(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processUpgradeTrigger")

	if auditEvent == nil || (auditEvent.Verb != "patch" && auditEvent.Verb != "update") || auditEvent.ObjectRef.Resource != "pods" {
		return
	}

	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	ok := false
	var reqPod *v1.Pod
	if reqPod, ok = auditEvent.RequestRuntimeObj.(*v1.Pod); !ok || reqPod == nil {
		return
	}

	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	ok = false
	var resPod *v1.Pod
	if resPod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); !ok || resPod == nil {
		return
	}

	var pod *v1.Pod
	if pod = auditEvent.TryGetPodFromEvent(); pod == nil {
		return
	}

	isUpgrade, containers, timeStamp := isNewUpgradeDelivery(resPod)
	if !isUpgrade {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	podKey := genPodKey(clusterName, resPod.Namespace, resPod.Name)
	podMsMap, ok := podUpgradeMileStoneMap.Get(podKey)
	if ok {
		if _, has := podMsMap.(map[string]*PodUpgradeMileStone)[timeStamp]; has {
			return
		}
	} else {
		podMsMap = map[string]*PodUpgradeMileStone{}
		podUpgradeMileStoneMap.Set(podKey, podMsMap)
	}
	klog.V(6).Infof("upgrade start for %s timestamp: %s, audit: %s\n", resPod.Name, timeStamp, auditEvent.AuditID)

	podUpgradeMs := &PodUpgradeMileStone{
		Cluster:              clusterName,
		Namespace:            auditEvent.ObjectRef.Namespace,
		PodName:              auditEvent.ObjectRef.Name,
		PodUID:               string(resPod.UID),
		Type:                 POD_UPGRADE,
		TriggerAuditLog:      string(auditEvent.AuditID),
		UpgradeResult:        "",
		UpgradeContainerName: strings.Join(containers, ","),
		CreatedTime:          auditEvent.StageTimestamp.Time,
		UpgradeTimeoutTime:   auditEvent.StageTimestamp.Time.Add(timeoutDuration),
		DebugUrl:             fmt.Sprintf("http://host:port/api/v1/debugpod?uid=%s", string(resPod.UID)),
		key:                  podKey,
		subKey:               timeStamp,
		mutex:                sync.Mutex{},
		NodeIP:               resPod.Status.HostIP,
		upgradeContainers:    containers,
	}

	podMsMap.(map[string]*PodUpgradeMileStone)[timeStamp] = podUpgradeMs
	podUpgradeMileStoneMap.Set(podKey, podMsMap)

	traceProcessingTime := time.Now().Sub(auditEvent.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_upgrade_trigger").Observe(traceProcessingTime)
}

func isNewUpgradeDelivery(pod *v1.Pod) (bool, []string, string) {
	var upgradeContainers []string

	if pod == nil {
		return false, upgradeContainers, ""
	}

	return false, upgradeContainers, ""
}

// return: 1.是否结束  2.是否成功
func isFinishUpgrade(podMs *PodUpgradeMileStone, pod *v1.Pod) (bool, string) {
	if pod == nil {
		return false, UPGRADE_FAILED
	}

	// updateStatus, err := getUpdateStatusFromAnnotation(pod)
	// if err != nil {
	// 	return false, UPGRADE_FAILED
	// }

	podContainers := make(map[string]*v1.Container)
	for idx, _ := range pod.Spec.Containers {
		podContainers[pod.Spec.Containers[idx].Name] = &pod.Spec.Containers[idx]
	}

	hasReadinessProbe := false
	isSuccess := true
	for _, co := range podMs.upgradeContainers {
		// if status, ok := updateStatus.Statuses[api.ContainerInfo{Name: co}]; !ok {
		// 	return false, UPGRADE_RUN_CONTAINER_ERROR
		// } else {
		// 	if !status.FinishTimestamp.After(podMs.CreatedTime) {
		// 		return false, UPGRADE_RUN_CONTAINER_ERROR
		// 	}

		// 	if status.Success != true {
		// 		isSuccess = false
		// 	}
		// }
		if pc, ok := podContainers[co]; ok {
			if pc.ReadinessProbe != nil {
				hasReadinessProbe = true
			}
		}
	}

	if !isSuccess {
		return false, UPGRADE_RUN_CONTAINER_ERROR
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
			return false, "PodNotReady"
		}

		// need new ready probe
		if condition.Type == v1.PodReady && hasReadinessProbe && condition.LastTransitionTime.Time.Before(podMs.CreatedTime) {
			return false, "PodNotReady"
		}
	}

	return true, UPGRADE_SUCCESS
}

func collectPodUpgradeAuditLog(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("collectPodAuditLog")

	if auditEvent.ObjectRef.Resource != "events" && auditEvent.ObjectRef.Resource != "pods" {
		return
	}

	podKey := ""

	if auditEvent.ObjectRef.Resource == "events" {
		var event *v1.Event
		ok := false
		if auditEvent.ResponseRuntimeObj == nil {
			return
		}
		if event, ok = auditEvent.ResponseRuntimeObj.(*v1.Event); !ok || event == nil {
			return
		}

		clusterName := auditEvent.Annotations["cluster"]
		if event.InvolvedObject.Kind == "Pod" {
			podKey = genPodKey(clusterName, event.InvolvedObject.Namespace, event.InvolvedObject.Name)
		}
	}

	if auditEvent.ObjectRef.Resource == "pods" {
		if auditEvent.Verb == "create" && auditEvent.ObjectRef.Subresource == "" {
			if auditEvent.ResponseRuntimeObj != nil {
				var responsePod *v1.Pod
				ok := false
				if responsePod, ok = auditEvent.ResponseRuntimeObj.(*v1.Pod); ok && responsePod != nil {
					clusterName := auditEvent.Annotations["cluster"]
					podKey = genPodKey(clusterName, responsePod.Namespace, responsePod.Name)
				}
			}
		} else {
			clusterName := auditEvent.Annotations["cluster"]
			podKey = genPodKey(clusterName, auditEvent.ObjectRef.Namespace, auditEvent.ObjectRef.Name)
		}
	}

	v, ok := podUpgradeMileStoneMap.Get(podKey)
	if !ok || v == nil {
		return
	}

	logList, isExist := podUpgradeAuditLogMap.Get(podKey)
	if !isExist {
		newList := list.New()
		logList = newList
		podUpgradeAuditLogMap.Set(podKey, newList)
	}
	logList.(*list.List).PushBack(auditEvent)
}

func (data *PodUpgradeMileStone) analyzeFailedReason() string {
	//get audit event
	v, ok := podUpgradeAuditLogMap.Get(data.key)
	auditEvents := make([]*shares.AuditEvent, 0)
	if ok && v != nil {
		for item := v.(*list.List).Front(); nil != item; item = item.Next() {
			auditEvents = append(auditEvents, item.Value.(*shares.AuditEvent))
		}
	}

	if featuregates.IsEnabled(reason.NewReasonFeature) {
		analyzerDAG := analyzers.ShareAnalyzerFactory.GetAnalyzerByType(analyzers.PodUpgrade)
		if len(auditEvents) > 0 {
			//fmt.Printf("pod name: %s, uid: %s,lastEvent name: %s, lastEvent time: %s \n", data.PodName, data.PodUID, auditEvents[len(auditEvents)-1].AuditID, auditEvents[len(auditEvents)-1].RequestReceivedTimestamp.String())
			analyzerDAG.AuditEvents = auditEvents
		}

		// todo: PodUID?
		if upgradeSpans, callBack := shareutils.GetSpansByUIDAndType(types.UID(data.PodUID), nil, "PodUpgrade"); upgradeSpans != nil {
			analyzerDAG.Spans = upgradeSpans
			defer callBack()
		}

		// DAG
		analyzerDAG.BeginTime = &data.CreatedTime
		analyzerDAG.EndTime = data.trickTime
		analyzerDAG.Analysis(data.Cluster, data.PodName, data.PodUID)

		return analyzerDAG.GetResult().Result
	}
	return ""
}

func isFinishUpgradeHelper(L *lua.LState) int {
	/*podStr := L.ToString(1)

	pod := &v1.Pod{}
	err := json.Unmarshal([]byte(podStr), pod)
	if err != nil {
		klog.Errorf("unmarshal pod error, err:%s", err)
		L.Push(lua.LBool(false))
		return 1
	}*/
	userData := L.ToUserData(1)

	pod := &v1.Pod{}
	ok := false
	if pod, ok = userData.Value.(*v1.Pod); !ok || pod == nil {
		L.Push(lua.LBool(false))
		return 1
	}

	podContainers := make(map[string]*v1.Container)
	for idx, _ := range pod.Spec.Containers {
		podContainers[pod.Spec.Containers[idx].Name] = &pod.Spec.Containers[idx]
	}

	isSuccess := true

	if !isSuccess {
		L.Push(lua.LBool(false))
		return 1
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == v1.PodReady && condition.Status != v1.ConditionTrue {
			L.Push(lua.LBool(false))
			return 1
		}
	}

	L.Push(lua.LBool(true))
	return 1
}
