package slo

import (
	"container/list"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/shareutils"

	"github.com/alipay/container-observability-service/pkg/config"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/spans"
	lua "github.com/yuin/gopher-lua"

	"github.com/alipay/container-observability-service/pkg/queue"

	"github.com/alipay/container-observability-service/pkg/metas"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"

	v1 "k8s.io/api/core/v1"
	//k8s_audit "k8s.io/apiserver/pkg/apis/audit"

	"k8s.io/klog/v2"
)

const (
	sloInitialized     = "initialized"
	sloScheduled       = "scheduled"
	sloContainersReady = "containersReady"
	sloRunningAt       = "runningAt"
	sloSucceedAt       = "succeedAt"
	sloFailedAt        = "failedAt"
	sloReady           = "ready"
	// SLOFinish is combined of running/succeed/failed/ready
	sloFinish = "finish"
)

const (
	PVCScheduleDelay         = "PVCScheduleDelay"         // PVC调度导致的调度延迟
	PreemptionDelay          = "PreemptionDelay"          // 抢占调度
	ScheduleDelay            = "ScheduleDelay"            // 调度延迟
	ResourceNotEnough        = "ResourceNotEnough"        // 因为资源不足导致的调度耗时过长
	PullImageTooMuchTime     = "PullImageTooMuchTime"     // 拉取镜像超时或者耗时过长
	KMPullImageTooMuchTime   = "KMPullImageTooMuchTime"   // kubemaker 拉取镜像超时或者耗时过长
	DMPullImageTooMuchTime   = "DMPullImageTooMuchTime"   // 采用device mapper 导致镜像拉取超时或者耗时过长
	PostStartHookTooMuchTime = "PostStartHookTooMuchTime" // PostStartHook 耗时过长
	InvalidMountConfig       = "InvalidMountConfig"       // 挂载配置非法，用户侧错误
	CREATE_RESULT_SUCCESS    = "success"                  // 创建成功
	beforeFinish             = "beforeFinish"
	resultAPIFailed          = "api_failed"      // API错误
	timeoutUnknown           = "timeout_unknown" // 创建失败，但原因不明
	NoEvent                  = "NoEvent"         // 没有event，分析不出原因
	FailedPostStartHook      = "FailedPostStartHook"
	KubeletDelay             = "KubeletDelay" //kubelet节点处理慢

	SERVICE_POD_TIMEOUT_TIME = 480 * time.Second // 在线服务类Pod超时时间
	JOB_POD_TIMEOUT_TIME     = 90 * time.Second  // Job类Pod超时时间
)

var (
	createQueue       *queue.BoundedQueue
	podEventsMap      *utils.SafeMap // podKey -> array of events
	podAuditLogMap    *utils.SafeMap
	notifyQueue       chan string
	podMilestoneMap   *utils.SafeMap
	timeBroadcastChan chan *time.Time
)

func init() {
	createQueue = queue.NewBoundedQueue("slo-watcher-create", 200000, nil)
	createQueue.StartLengthReporting(10 * time.Second)
	createQueue.IsLockOSThread = true
	createQueue.IsDropEventOnFull = false
	createQueue.StartConsumers(1, func(item interface{}) {
		metrics.EventConsumedCount.Inc()
		defer utils.IgnorePanic("createQueueConsumer")

		time1 := time.Now()
		event, ok := item.(*shares.AuditEvent)
		if !ok || event == nil {
			return
		}
		event.Wait()

		// api code metrics
		time2 := time.Now()
		genPodCreateAPIResultMetrics(event)
		// 处理创建动作
		time3 := time.Now()
		processPodCreateLog(event)
		// collect pod events
		time4 := time.Now()
		collectPodEvents(event)
		collectPodAuditLog(event)
		// 同步审计日志时间
		time5 := time.Now()
		syncAuditTime(event)
		time6 := time.Now()

		metrics.MethodDurationMilliSeconds.WithLabelValues("CreateConsumerTotal").Set(utils.TimeDiffInMilliSeconds(time1, time6))
		metrics.MethodDurationMilliSeconds.WithLabelValues("genPodCreateAPIResultMetrics").Set(utils.TimeDiffInMilliSeconds(time2, time3))
		metrics.MethodDurationMilliSeconds.WithLabelValues("processPodCreateLog").Set(utils.TimeDiffInMilliSeconds(time3, time4))
		metrics.MethodDurationMilliSeconds.WithLabelValues("collectPodEvents").Set(utils.TimeDiffInMilliSeconds(time4, time5))
		metrics.MethodDurationMilliSeconds.WithLabelValues("syncAuditTime").Set(utils.TimeDiffInMilliSeconds(time5, time6))
	})

	timeBroadcastChan = make(chan *time.Time, 10000)
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		for {
			// 连续取最后一个
			t := <-timeBroadcastChan
			currentAuditLen := len(timeBroadcastChan)
			for i := 0; i < currentAuditLen; i++ {
				t = <-timeBroadcastChan
			}

			podMilestoneMap.IterateWithFunc(func(obj interface{}) {
				podMilestone := obj.(*PodStartupMilestones)
				if !podMilestone.Finished {
					if len(podMilestone.auditTimeQueue) < 10 {
						podMilestone.auditTimeQueue <- t
					}
				}
			})
		}
	}()

	notifyQueue = make(chan string, 10000)
	go func() {
		for {
			select {
			case key := <-notifyQueue:
				podMilestoneMap.Delete(key)
				podEventsMap.Delete(key)
				podAuditLogMap.Delete(key)
			}
		}
	}()

	podEventsMap = utils.NewSafeMap()
	podAuditLogMap = utils.NewSafeMap()
	podMilestoneMap = utils.NewSafeMap()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				sloOngoingSize.WithLabelValues("podCreate").Set(float64(podMilestoneMap.Size()))
				sloOngoingSize.WithLabelValues("podEvent").Set(float64(podEventsMap.Size()))
				sloOngoingSize.WithLabelValues("podAuditLog").Set(float64(podAuditLogMap.Size()))
			}
		}
	}()

	spans.RegisterLuaHelperFunc("get_pod_slo_time", getSLOTimeHelper)
	metas.RegisterPublisher(metas.PodCreateSLO)
}

// PodEvent is Pod and Event struct
type PodEvent struct {
	Pod   *v1.Pod
	Event *shares.AuditEvent
}

// PodStartupMilestones keeps all milestone timestamps from Pod creation.
type PodStartupMilestones struct {
	Cluster                       string
	InitImage                     string
	Namespace                     string
	PodName                       string
	PodUID                        string
	Type                          string
	NodeIP                        string
	NodeName                      string // in some cases, when a pod is scheduled to a specified node, but kubelet doesn't post any pod status to apiserver, nodeIP is empty, but we do want know which node is not working properly.
	DebugUrl                      string
	OwnerRefStr                   string
	SchedulingStrategy            string
	StartUpResultFromCreate       string
	StartUpResultFromSchedule     string
	IsJob                         bool
	Cores                         int64
	WrittenToZsearch              bool
	Finished                      bool
	StopInterEvents               bool      //停止接受新的event
	Created                       time.Time //创建开始时间
	CreatedTime                   time.Time //创建开始时间
	FinishTime                    time.Time
	ActualFinishTimeAfterSchedule time.Time
	ActualFinishTimeAfterCreate   time.Time
	Scheduled                     time.Time //调度时间
	PodInitializedTime            time.Time
	ContainersReady               time.Time
	RunningAt                     time.Time
	SucceedAt                     time.Time
	FailedAt                      time.Time
	ReadyAt                       time.Time
	DeletedTime                   time.Time
	InitStartTime                 time.Time
	ImageNameToPullTime           map[string]float64 //每个镜像拉取的时间
	PossibleReason                *string            //创建失败或超时时针对未知原因附加说明（timeout_unknown和kubeletDelay）
	// SLO 越界推理
	SLOViolationReason  string
	PodSLO              int64
	DeliverySLO         int64
	DeliverySLOAdjusted bool
	DeliveryDuration    time.Duration
	DeliveryStatus      string
	DeliveryStatusOrig  string
	SloHint             string // why current slo class
	//内部变量
	mutex            *sync.RWMutex
	inputQueue       chan *PodEvent
	closeCh          chan struct{}
	notifyQueue      chan string
	key              string
	auditTimeQueue   chan *time.Time
	latestPod        *v1.Pod
	shouldFinishTime *time.Time
	trickTime        *time.Time
}

// finish 结束Pod跟踪，线程安全的
func (data *PodStartupMilestones) finish() {
	defer utils.IgnorePanic("PodStartupMilestones.finish ")

	if data.Finished {
		return
	}
	data.Finished = true

	//记录镜像拉拉取时间
	v, ok := podEventsMap.Get(data.key)
	if ok && v != nil {
		eventsNormalOrder := make([]*v1.Event, 0)
		for item := v.(*list.List).Front(); nil != item; item = item.Next() {
			eventsNormalOrder = append(eventsNormalOrder, item.Value.(*v1.Event))
		}
		data.ImageNameToPullTime = calculateImagePullTime(eventsNormalOrder, data.shouldFinishTime)
	}

	data.FinishTime = time.Now()
	data.DebugUrl = ""
	if data.PodUID != "" {
		data.DebugUrl = "http://host:port/api/v1/debugpod?uid=" + data.PodUID
	} else if data.PodName != "" {
		data.DebugUrl = "http://host:port/api/v1/debugpod?name=" + data.PodName
	}

	//save milestone to zsearch
	data.saveMileStone()

	close(data.closeCh)
}

func (data *PodStartupMilestones) saveMileStone() {
	sloData, err := json.Marshal(data)
	if err == nil {
		_ = xsearch.SaveSloData(data.Cluster, data.Namespace, data.PodName, data.PodUID, "create", sloData)
		_ = xsearch.SaveSloTraceData(data.Cluster, data.Namespace, data.PodName, data.PodUID, "create", sloData)
	}
	//及时释放
	sloData = nil
}

// 一个 PodStartupMilestones 开始处理 podEvent 或者 timeEvent
func (data *PodStartupMilestones) start() {
	klog.V(6).Infof("pod [%s] to consume PodEvent: %d", data.PodName, utils.GoID())
	for {
		select {
		case pe := <-data.inputQueue:
			if pe != nil {
				data.processEvent(pe)
			}
		case t := <-data.auditTimeQueue:
			data.processTime(*t)
		case <-data.closeCh:
			auditTimeLen := len(data.auditTimeQueue)
			for i := 0; i < auditTimeLen; i++ {
				<-data.auditTimeQueue //chain可能满了，外部product可能已经阻塞写入了
			}
			klog.V(8).Infof("finish for [%s], GoId: %d", data.PodName, utils.GoID())
			data.notifyQueue <- data.key
			//send slo data to publisher
			metas.GetPubLister(metas.PodCreateSLO).Publish(data)
			return
		}
	}
}

func (data *PodStartupMilestones) getPodDeliverySLO() (podSLO, deliverySLO int64) {

	podSLO = int64(SERVICE_POD_TIMEOUT_TIME)
	if data.IsJob {
		podSLO = int64(JOB_POD_TIMEOUT_TIME)
	}

	_, cPodSLO := getSloTimeoutFromConfigMap(data)
	if cPodSLO > 0 {
		podSLO = cPodSLO
	}

	deliverySLO = podSLO

	return
}

// 审计日志时钟
func (data *PodStartupMilestones) processTime(t time.Time) {
	data.mutex.Lock()
	defer data.mutex.Unlock()
	defer utils.IgnorePanic("processTime")

	if data.IsComplete() || data.Finished {
		return
	}

	// TODO 超时退出(30分钟还没有创建完成，不再跟踪)
	// 35 minutes to avoid 30min SLO class
	if !data.Created.IsZero() && t.After(data.Created.Add(35*time.Minute)) {
		klog.Infof("processTime for pod[%s] return, because 超时退出(35分钟还没有创建完成，不再跟踪", data.PodName)

		if data.StartUpResultFromCreate == "" {
			// 分析失败原因
			data.trickTime = &t
			failedReason := data.analyzeFailedReason()
			data.StartUpResultFromCreate = failedReason

			metrics.PodStartupResult.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr,
				failedReason, data.SchedulingStrategy, fmt.Sprintf("%d", data.Cores), fmt.Sprintf("%t", data.IsJob), data.NodeIP, "SUCCESS", strconv.FormatInt(data.PodSLO, 10)).Inc()
		}

		data.finish()

		return
	}

	//有必要更新zsearch
	needUpdateMileStone := false
	defer func() {
		if needUpdateMileStone {
			data.saveMileStone()
		}
	}()

	// ## begin test for slo time
	if spans, callBack := shareutils.GetSpansByUIDAndType(data.latestPod.UID, nil, "PodCreate"); spans != nil {
		defer callBack()
		deliveryCost := DeliveryTimeCalcNew(spans, t, data.latestPod.UID)

		dsloDuration := time.Duration(data.DeliverySLO)

		// for pod with BESTEFFORT SLO class, 10min gives the current reason
		//SLO时间为0，不保障
		isBestEffortPod := false
		if dsloDuration == 0 {
			dsloDuration = 10 * time.Minute
			isBestEffortPod = true
		}

		data.DeliveryDuration = deliveryCost

		// track DeliverySLO
		if deliveryCost > dsloDuration && data.SLOViolationReason == "" {
			// 分析失败原因
			data.trickTime = &t
			failedReason := data.analyzeFailedReason()
			data.SLOViolationReason = failedReason

			sloReason := metas.IsTypicalPodNew(data.latestPod)
			deliveryStatus := "FAIL"
			if isBestEffortPod && strings.Contains(failedReason, "TooMuchTime") {
				deliveryStatus = "SUCCESS"
			}
			data.DeliveryStatus = deliveryStatus

			metrics.PodStartupSLOResult.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr, failedReason,
				time.Duration(data.DeliverySLO).String(), fmt.Sprintf("%d", data.Cores), fmt.Sprintf("%t", data.IsJob), "", "", "", deliveryStatus, sloReason, strconv.FormatBool(data.DeliverySLOAdjusted)).Inc()

			needUpdateMileStone = true
		}
	}
	// ## end test for slo time without kubespeed

	// 创建开始的超时，分析错误原因
	if !data.Created.IsZero() {
		// klog.Infof("create timeout for pod [%s]", data.PodName)
		if data.StartUpResultFromCreate != "" {
			return
		}

		finishTimeForSlo := data.GetFinishTime()

		if t.After(finishTimeForSlo) {
			klog.V(6).Infof("Pod[%s] hash timout(auditTime:%s, finishTime:%s, CreatedTime:%s), to analysis reason...创建开始的超时", data.PodName, t.String(), finishTimeForSlo.String(), data.Created.String())
			data.ActualFinishTimeAfterCreate = t
			data.shouldFinishTime = &finishTimeForSlo
			data.trickTime = &t
			// 分析失败原因
			failedReason := data.analyzeFailedReason()

			metrics.PodStartupResult.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr, failedReason,
				data.SchedulingStrategy, fmt.Sprintf("%d", data.Cores), fmt.Sprintf("%t", data.IsJob), data.NodeIP, "FAIL", strconv.FormatInt(data.PodSLO, 10)).Inc()
			data.StartUpResultFromCreate = failedReason
			data.DeliveryStatusOrig = "FAIL"

			needUpdateMileStone = true
		}

		if data.WrittenToZsearch {
			return
		}
		if t.After(finishTimeForSlo) {
			// 更新 slo_pod_info 状态到 失败
			err := xsearch.SavePodInfoToZSearch(data.Cluster, data.latestPod, "失败", t, "", "", true)
			if err != nil {
				klog.Errorf("error updating podinfo %s, %v", data.PodUID, err)
			}

		}

	} else if !data.Scheduled.IsZero() { // 调度开始的超时，分析错误原因
		// klog.Infof("schedule timeout for pod [%s]", data.PodName)
		if data.StartUpResultFromSchedule != "" {
			return
		}

		finishTimeForSlo := data.Scheduled.Add(SERVICE_POD_TIMEOUT_TIME)
		if data.IsJob {
			finishTimeForSlo = data.Scheduled.Add(JOB_POD_TIMEOUT_TIME)
		}

		//finish time增加init耗时
		if !data.InitStartTime.IsZero() && !data.PodInitializedTime.IsZero() && data.PodInitializedTime.After(data.InitStartTime) {
			finishTimeForSlo = finishTimeForSlo.Add(data.PodInitializedTime.Sub(data.InitStartTime))
		}

		timeoutFromConfigMap, _ := getSloTimeoutFromConfigMap(data)
		if !timeoutFromConfigMap.IsZero() {
			finishTimeForSlo = timeoutFromConfigMap
		}

		if t.After(finishTimeForSlo) {
			klog.V(6).Infof("Pod[%s] hash timout(auditTime:%s, finishTime:%s, CreatedTime:%s), to analysis reason...调度开始的超时", data.PodName, t.String(), finishTimeForSlo.String(), data.Created.String())
			//判断已经超时，记录一下，用于collectEvents时判断是否有新的在"超时时间"之前的事件更新时将needReReason设为true
			data.shouldFinishTime = &finishTimeForSlo
			data.ActualFinishTimeAfterSchedule = t
			data.trickTime = &t
			failedReason := data.analyzeFailedReason() // 分析超时错误

			metrics.PodStartupResultExcludingScheduling.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr,
				failedReason, fmt.Sprintf("%t", data.IsJob), "FAIL").Inc()
			data.StartUpResultFromSchedule = failedReason
			needUpdateMileStone = true
		}
	}

}

func (data *PodStartupMilestones) processEvent(pe *PodEvent) {
	data.mutex.Lock()
	defer data.mutex.Unlock()
	defer utils.IgnorePanic("processEvent")

	//latency
	pod := pe.Pod
	event := pe.Event

	data.NodeIP = pod.Status.HostIP
	data.NodeName = pod.Spec.NodeName
	data.latestPod = pod

	for _, cond := range pod.Status.Conditions {
		if cond.Status != v1.ConditionTrue {
			continue
		}

		if cond.Type == v1.PodInitialized && data.PodInitializedTime.IsZero() {
			data.PodInitializedTime = cond.LastTransitionTime.Time
			data.updateLatencyMetrics(sloInitialized, data.PodInitializedTime)
		} else if cond.Type == v1.PodScheduled && data.Scheduled.IsZero() {
			data.Scheduled = cond.LastTransitionTime.Time
			data.updateLatencyMetrics(sloScheduled, data.Scheduled)
		} else if cond.Type == v1.ContainersReady && data.ContainersReady.IsZero() {
			data.ContainersReady = cond.LastTransitionTime.Time
			data.updateLatencyMetrics(sloContainersReady, data.ContainersReady)
		} else if cond.Type == v1.PodReady && data.ReadyAt.IsZero() {
			data.ReadyAt = cond.LastTransitionTime.Time
			data.updateLatencyMetrics(sloReady, data.ReadyAt)
			if !data.IsJob {
				data.updateLatencyMetrics(sloFinish, data.ReadyAt)
			}
		}
	}

	phase := pod.Status.Phase
	if phase == v1.PodRunning && data.RunningAt.IsZero() {
		data.RunningAt = event.StageTimestamp.Time
		data.updateLatencyMetrics(sloRunningAt, data.RunningAt)
		if data.IsJob {
			data.updateLatencyMetrics(sloFinish, data.RunningAt)
		}

		err := xsearch.SavePodInfoToZSearch(data.Cluster, data.latestPod, "已完成", data.RunningAt, "", "运行阶段", true)
		if err != nil {
			klog.Errorf("error updating podinfo %s, %v", data.PodUID, err)
		}

	} else if phase == v1.PodSucceeded && data.SucceedAt.IsZero() {
		data.SucceedAt = event.StageTimestamp.Time
		data.updateLatencyMetrics(sloSucceedAt, data.SucceedAt)
		if data.IsJob {
			data.updateLatencyMetrics(sloFinish, data.SucceedAt)
		}
	} else if phase == v1.PodFailed && data.FailedAt.IsZero() {
		data.FailedAt = event.StageTimestamp.Time
		data.updateLatencyMetrics(sloFailedAt, data.FailedAt)
		if data.IsJob {
			data.updateLatencyMetrics(sloFinish, data.FailedAt)
		}
	}

	//create result:
	if data.IsComplete() {
		coresStr := fmt.Sprintf("%d", data.Cores)
		isJobStr := fmt.Sprintf("%t", data.IsJob)
		if data.StartUpResultFromCreate == "" {
			data.StartUpResultFromCreate = CREATE_RESULT_SUCCESS
			data.DeliveryStatusOrig = "SUCCESS"
			metrics.PodStartupResult.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr,
				CREATE_RESULT_SUCCESS, data.SchedulingStrategy, coresStr, isJobStr, data.NodeIP, "SUCCESS", strconv.FormatInt(data.PodSLO, 10)).Inc()
		}

		d := data.DeliveryDuration.Seconds()

		sloTime := time.Duration(data.DeliverySLO)
		metrics.PodStartupSLOLatency.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr, "ready", "",
			sloTime.String(), fmt.Sprintf("%d", data.Cores)).Observe(d)

		if data.SLOViolationReason == "" {
			data.SLOViolationReason = CREATE_RESULT_SUCCESS
			data.DeliveryStatus = "SUCCESS"
			// update
			sloTime, adjusted := metas.GetPodSLOByDeliveryPath(pod)
			priority := metas.GetPriority(pod)
			sloReason := metas.IsTypicalPodNew(data.latestPod)
			metrics.PodStartupSLOResult.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr,
				CREATE_RESULT_SUCCESS, sloTime.String(), coresStr, isJobStr, priority, "SUCCESS", sloReason, strconv.FormatBool(adjusted)).Inc()
		}

		if data.StartUpResultFromSchedule == "" {
			data.StartUpResultFromSchedule = CREATE_RESULT_SUCCESS
			metrics.PodStartupResultExcludingScheduling.WithLabelValues(data.Cluster, data.Namespace,
				data.OwnerRefStr, CREATE_RESULT_SUCCESS, isJobStr, "SUCCESS").Inc()
		}
		data.finish() // finish 结束Pod跟踪，线程安全的
		return
	}

	traceProcessingTime := time.Now().Sub(pe.Event.StageTimestamp.Time).Seconds()
	metrics.TraceProcessingLatency.WithLabelValues("slo_create").Observe(traceProcessingTime)
}

func (data *PodStartupMilestones) updateLatencyMetrics(milestone string, end time.Time) {
	d := utils.TimeDiffInSeconds(data.Created, end)
	if d >= 0 {
		metrics.PodStartupLatency.WithLabelValues(data.Cluster, data.Namespace, data.OwnerRefStr,
			milestone, fmt.Sprintf("%t", data.IsJob), data.SchedulingStrategy, fmt.Sprintf("%d", data.Cores)).Observe(d)
	}

	if milestone != sloScheduled {
		d := utils.TimeDiffInSeconds(data.Scheduled, end)
		if d >= 0 {
			metrics.PodStartupLatencyExcludingShceduling.WithLabelValues(
				data.Cluster, data.Namespace, data.OwnerRefStr, milestone,
				fmt.Sprintf("%t", data.IsJob)).Observe(d)
		}
	}
}

// IsComplete returns true is data is complete (ready to be included in the metric)
// and if it haven't been included in the metric yet.
func (data *PodStartupMilestones) IsComplete() bool {
	return !data.ReadyAt.IsZero()
}

func (data *PodStartupMilestones) GetFinishTime() time.Time {
	finishTimeForSlo := data.Created.Add(SERVICE_POD_TIMEOUT_TIME)
	if data.IsJob {
		finishTimeForSlo = data.Created.Add(JOB_POD_TIMEOUT_TIME)
	}

	//finish time增加init耗时
	if !data.InitStartTime.IsZero() && !data.PodInitializedTime.IsZero() && data.PodInitializedTime.After(data.InitStartTime) {
		finishTimeForSlo = finishTimeForSlo.Add(data.PodInitializedTime.Sub(data.InitStartTime))
	}

	timeoutFromConfigMap, _ := getSloTimeoutFromConfigMap(data)
	if !timeoutFromConfigMap.IsZero() {
		finishTimeForSlo = timeoutFromConfigMap
	}
	return finishTimeForSlo
}

// 处理 pod 创建的审计日志
func processPodCreateLog(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("processPodCreateLog")

	if auditEvent.ObjectRef.Resource != "pods" {
		return
	}

	//用户问题，直接忽略
	if auditEvent.ResponseStatus.Code >= 400 && auditEvent.ResponseStatus.Code < 500 && auditEvent.ResponseStatus.Code != 403 {
		return
	}

	/*pod, err := metas.GeneratePodFromEvent(auditEvent)
	if err != nil || pod == nil {
		return
	}*/
	var pod *v1.Pod
	if pod = auditEvent.TryGetPodFromEvent(); pod == nil {
		return
	}

	//一些特殊的用户错误
	if auditEvent.ResponseStatus.Code >= 500 {
		/*stats := &metav1.Status{}
		err := json.Unmarshal(auditEvent.ResponseObject.Raw, stats)*/
		var stats *metav1.Status
		ok := false
		if auditEvent.RequestRuntimeObj == nil || auditEvent.ResponseRuntimeObj == nil {
			return
		}
		if stats, ok = auditEvent.RequestRuntimeObj.(*metav1.Status); !ok {
			return
		}

		if stats != nil {
			//admission error
			adWebhookStr := "admission webhook"
			denyStr := "denied the request"
			if strings.Contains(stats.Message, adWebhookStr) && strings.Contains(stats.Message, denyStr) {
				return
			}
			//missing resource
			missingResStr := "the server could not find the requested resource"
			if strings.Contains(stats.Message, adWebhookStr) && strings.Contains(stats.Message, missingResStr) {
				return
			}
			//stateful set server timeout
			if len(pod.ObjectMeta.OwnerReferences) > 0 && pod.ObjectMeta.OwnerReferences[0].Kind == "StatefulSet" {
				msg := "The POST operation against Pod could not be completed at this time, please try again"
				if strings.Contains(stats.Message, msg) && stats.Reason == metav1.StatusReasonServerTimeout {
					return
				}
			}
		}
	}

	clusterName := auditEvent.Annotations["cluster"]
	ownerRefStr := getOwnerRefStr(pod)

	podKey := genPodKey(clusterName, pod.Namespace, pod.Name)
	auditResponseCode := auditEvent.ResponseStatus.Code
	milestone, exist := podMilestoneMap.Get(podKey)

	if auditEvent.Verb == "create" && (auditResponseCode >= 500 || auditResponseCode == 403) && auditEvent.ObjectRef.Subresource == "" {
		// create failed.
		if auditResponseCode == 403 && !strings.Contains(string(auditEvent.ResponseObject.Raw), "quota") {
			return
		}
		Status := metav1.Status{}
		if err := json.Unmarshal(auditEvent.ResponseObject.Raw, &Status); err != nil {
			return
		}
		if !exist || milestone == nil {
			result := analysisApiFailed(Status.Message)

			ss, cores := getSchedulingStrategyAndCores(pod)

			// podslo/deliveryslo
			apiFailedMilestone := &PodStartupMilestones{
				Cluster:                 clusterName,
				PodName:                 pod.Name,
				StartUpResultFromCreate: result,
				Namespace:               pod.Namespace,
				Created:                 auditEvent.StageTimestamp.Time,
				CreatedTime:             auditEvent.StageTimestamp.Time,
				Type:                    "create",
				OwnerRefStr:             ownerRefStr,
				SchedulingStrategy:      ss,
				Cores:                   cores,
				IsJob:                   metas.IsJobPod(pod),
				DebugUrl:                string(auditEvent.ResponseObject.Raw),
				SLOViolationReason:      result,
				DeliveryStatus:          "FAIL",
			}

			sloTime, adjusted := metas.GetPodSLOByDeliveryPath(pod)
			podslo, _ := apiFailedMilestone.getPodDeliverySLO()
			apiFailedMilestone.PodSLO = podslo
			apiFailedMilestone.DeliverySLO = int64(sloTime)

			metrics.PodStartupResult.WithLabelValues(clusterName, pod.Namespace, ownerRefStr, result, ss,
				fmt.Sprintf("%d", cores), fmt.Sprintf("%t", metas.IsJobPod(pod)), pod.Status.HostIP, "FAIL", strconv.FormatInt(podslo, 10)).Inc()
			priority := metas.GetPriority(pod)

			sloReason := metas.IsTypicalPodNew(pod)
			metrics.PodStartupSLOResult.WithLabelValues(clusterName, pod.Namespace, ownerRefStr, result, sloTime.String(),
				fmt.Sprintf("%d", cores), fmt.Sprintf("%t", metas.IsJobPod(pod)), priority, "FAIL", sloReason, strconv.FormatBool(adjusted)).Inc()

			metrics.PodCreateTotal.WithLabelValues(clusterName, pod.Namespace, fmt.Sprintf("%d", cores), fmt.Sprintf("%t", metas.IsJobPod(pod))).Inc()

			apiFailedMilestone.SloHint = sloReason

			apiFailedMilestone.PodUID = "NOPODID" + time.Now().Format(time.RFC3339)
			apiFailedMilestone.saveMileStone()
		}
	} else if auditEvent.Verb == "create" && auditResponseCode == 201 && auditEvent.ObjectRef.Subresource == "" {
		if !exist || milestone == nil {
			imageName := getInitContainerImage(pod)
			ss, cores := getSchedulingStrategyAndCores(pod)
			newMilestone := &PodStartupMilestones{
				mutex:              &sync.RWMutex{},
				Cluster:            clusterName,
				Namespace:          pod.Namespace,
				PodName:            pod.Name,
				InitImage:          imageName,
				PodUID:             string(pod.UID),
				OwnerRefStr:        ownerRefStr,
				Type:               "create",
				key:                podKey,
				Created:            pod.CreationTimestamp.Time,
				CreatedTime:        pod.CreationTimestamp.Time,
				IsJob:              metas.IsJobPod(pod),
				SchedulingStrategy: ss,
				Cores:              cores,
				notifyQueue:        notifyQueue,
				closeCh:            make(chan struct{}),
				inputQueue:         make(chan *PodEvent, 10000),
				auditTimeQueue:     make(chan *time.Time, 200),
				latestPod:          pod,
			}

			sloTime, adjusted := metas.GetPodSLOByDeliveryPath(pod)
			podslo, _ := newMilestone.getPodDeliverySLO()
			newMilestone.PodSLO = podslo
			newMilestone.DeliverySLO = int64(sloTime)
			newMilestone.DeliverySLOAdjusted = adjusted

			sloReason := metas.IsTypicalPodNew(pod)
			newMilestone.SloHint = sloReason

			podMilestoneMap.Set(podKey, newMilestone)
			metrics.PodCreateTotal.WithLabelValues(newMilestone.Cluster, newMilestone.Namespace, fmt.Sprintf("%d", newMilestone.Cores), fmt.Sprintf("%t", newMilestone.IsJob)).Inc()
			go newMilestone.start()
		}
	} else if auditEvent.Verb == "delete" && auditResponseCode < 300 && auditEvent.ObjectRef.Subresource == "" {
		if exist && milestone != nil {
			podMs := milestone.(*PodStartupMilestones)
			podMs.StopInterEvents = true
			go func() {
				if !podMs.IsComplete() && !podMs.Finished && podMs.StartUpResultFromCreate == "" {
					ss, cores := getSchedulingStrategyAndCores(pod)
					coresStr := fmt.Sprintf("%d", cores)
					isJobStr := fmt.Sprintf("%t", podMs.IsJob)
					podMs.DeletedTime = auditEvent.StageTimestamp.Time
					//等待 PodEvent被消耗完
					klog.V(8).Infof("wait for [%s] to consume PodEvent: %d", podMs.PodName, utils.GoID())
					for len(podMs.inputQueue) > 0 {
						if podMs.Finished {
							klog.V(8).Infof("before finish for create [%s]", podMs.PodName)
							return
						}
						time.Sleep(500 * time.Millisecond)
					}

					//分析失败原因
					reason := beforeFinish
					if auditEvent.StageTimestamp.Time.After(podMs.CreatedTime.Add(3*time.Second)) &&
						!utils.SliceContainsString(config.GlobalLunettesConfig().IgnoreDeleteReasonNamespace, podMs.Namespace) {
						shouldFinishTime := podMs.GetFinishTime()
						podMs.shouldFinishTime = &shouldFinishTime
						podMs.trickTime = &auditEvent.StageTimestamp.Time
						reason = podMs.analyzeFailedReason()
						if reason == timeoutUnknown || reason == NoEvent {
							reason = beforeFinish
						}
					}

					if podMs.StartUpResultFromCreate == "" {
						podMs.StartUpResultFromCreate = reason
						//输出prometheus指标
						metrics.PodStartupResult.WithLabelValues(podMs.Cluster, podMs.Namespace, ownerRefStr,
							reason, ss, coresStr, isJobStr, podMs.NodeIP, "KILL", strconv.FormatInt(podMs.PodSLO, 10)).Inc()
						podMs.DeliveryStatusOrig = "KILL"
					}

					// Might already violate slo then trigger the delete operation
					if podMs.SLOViolationReason == "" {
						podMs.SLOViolationReason = reason
						podMs.DeliveryStatus = "KILL"
						// update PodStartupSLOResult
						//slo_time := xsearch.GetScheduleTimeLimit(pod)
						sloTime, adjusted := metas.GetPodSLOByDeliveryPath(pod)
						priority := metas.GetPriority(pod)

						sloReason := metas.IsTypicalPodNew(podMs.latestPod)
						podMs.DeliverySLO = int64(sloTime)
						metrics.PodStartupSLOResult.WithLabelValues(podMs.Cluster, podMs.Namespace, ownerRefStr,
							reason, sloTime.String(), coresStr, isJobStr, priority, "KILL", sloReason, strconv.FormatBool(adjusted)).Inc()
					}

					// for best-effort pod, if user kill in 10min, delivery-status will be KILL, else DeliveryStatus will be FAIL/SUCCESS
				}

				podMs.finish()
			}()
		}
	} else if exist && milestone != nil && auditResponseCode < 300 && !milestone.(*PodStartupMilestones).Finished && !milestone.(*PodStartupMilestones).StopInterEvents {
		// in creating, add to queue
		// 其它审计事件，放入 inputQueue，用于 processEvent 的处理
		milestone.(*PodStartupMilestones).inputQueue <- &PodEvent{
			Pod:   pod,
			Event: auditEvent,
		}
	}
}

// todo 收集 pod event 事件
func collectPodEvents(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("collectPodEvents")
	if auditEvent.ResponseStatus.Code >= 300 {
		return
	}
	if auditEvent.Verb != "create" || auditEvent.ObjectRef.Resource != "events" || auditEvent.ObjectRef.Subresource != "" {
		return
	}

	/*event := &v1.Event{}
	if err := json.Unmarshal(auditEvent.RequestObject.Raw, event); err != nil {
		klog.Infof("Unmarshal failed: %s", auditEvent.AuditID)
		return
	}*/
	var event *v1.Event
	ok := false
	if auditEvent.RequestRuntimeObj == nil {
		return
	}
	if event, ok = auditEvent.RequestRuntimeObj.(*v1.Event); !ok || event == nil {
		return
	}

	clusterName := auditEvent.Annotations["cluster"]
	if event.InvolvedObject.Kind == "Pod" {
		podKey := genPodKey(clusterName, event.InvolvedObject.Namespace, event.InvolvedObject.Name)
		v, ok := podMilestoneMap.Get(podKey)
		if !ok || v == nil {
			return
		}

		podMilestone := v.(*PodStartupMilestones)
		//标记Init Container开始拉取镜像时间
		if podMilestone.InitImage != "" && podMilestone.InitStartTime.IsZero() {
			if event.Reason == "Pulled" && strings.Contains(event.Message, podMilestone.InitImage) {
				podMilestone.InitStartTime = auditEvent.StageTimestamp.Time
			} else if event.Reason == "Pulling" && strings.Contains(event.Message, podMilestone.InitImage) {
				podMilestone.InitStartTime = auditEvent.StageTimestamp.Time
			}
		}

		if podMilestone.Finished {
			return
		}

		vv, isExist := podEventsMap.Get(podKey)
		if !isExist {
			newList := list.New()
			vv = newList
			podEventsMap.Set(podKey, newList)
		}
		vv.(*list.List).PushBack(event)
	}
}

func collectPodAuditLog(auditEvent *shares.AuditEvent) {
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

	v, ok := podMilestoneMap.Get(podKey)
	if !ok || v == nil {
		return
	}

	logList, isExist := podAuditLogMap.Get(podKey)
	if !isExist {
		newList := list.New()
		logList = newList
		podAuditLogMap.Set(podKey, newList)
	}
	logList.(*list.List).PushBack(auditEvent)
}

// PodCreateAPIResult 记录pod create的api返回结果
func genPodCreateAPIResultMetrics(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("genPodCreateAPIResultMetrics")

	clusterName := auditEvent.Annotations["cluster"]
	/*pod, err := metas.GeneratePodFromEvent(auditEvent)
	if err != nil || pod == nil {
		return
	}*/

	var pod *v1.Pod
	if pod = auditEvent.TryGetPodFromEvent(); pod == nil {
		return
	}

	if auditEvent.Verb == "create" && auditEvent.ObjectRef.Resource == "pods" && auditEvent.ObjectRef.Subresource == "" {

		resCode := fmt.Sprintf("%d", auditEvent.ResponseStatus.Code)
		metrics.PodCreateAPIResult.WithLabelValues(clusterName, auditEvent.ObjectRef.Namespace, resCode).Inc()
	}
}

// syncAuditTime 同步审计日志时间
func syncAuditTime(auditEvent *shares.AuditEvent) {
	defer utils.IgnorePanic("syncAuditTime")
	t := auditEvent.StageTimestamp.Time
	select {
	case timeBroadcastChan <- &t:
	default:
	}
}

// 从 configmap 里面获取特定的 超时时间
// 1）查看 namespace 的超时时间，注：同时修改在线任务和离线任务
// 默认是在线 8 min，离线 90 s
func getSloTimeoutFromConfigMap(data *PodStartupMilestones) (time.Time, int64) {

	var finishTimeForSlo time.Time
	var slo time.Duration

	// 1）先看 namespace 的超时时间
	// 从 configmap 获取用户自己配置的 configmap，注：同时修改在线任务和离线任务
	customizedTimeoutStr, ok := config.GlobalLunettesConfig().UserOnlineConfigMap[data.Namespace]
	if ok {
		customizedTimeout, err := time.ParseDuration(customizedTimeoutStr)
		if err != nil {
			klog.Errorf("Error parsing Namespace timeout from configmap, %v", err)
		} else {
			finishTimeForSlo = data.Created.Add(customizedTimeout)
			slo = customizedTimeout
			// klog.Infof("customized timeout in configmap is %s for %s", customizedTimeout, data.Namespace)
		}
	}

	return finishTimeForSlo, int64(slo)
}

func getSLOTimeHelper(L *lua.LState) int {
	podStr := L.ToString(1) /* get argument */

	pod := &v1.Pod{}
	err := json.Unmarshal([]byte(podStr), pod)
	if err != nil {
		klog.Errorf("unmarshal pod error, err:%s", err)
		L.Push(lua.LString(""))
		return 1
	}

	sloTime, _ := metas.GetPodSLOByDeliveryPath(pod)
	L.Push(lua.LString(sloTime.String()))
	return 1
}
