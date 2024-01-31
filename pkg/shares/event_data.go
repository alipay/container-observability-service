package shares

import (
	"encoding/json"
	"sort"
	"sync"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8s_audit "k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog/v2"
)

type AuditType string

const (
	AuditTypeEvent     AuditType = "event"
	AuditTypeOperation AuditType = "operation"
)

type AuditEvent struct {
	sync.WaitGroup
	*k8s_audit.Event
	ResponseRuntimeObj runtime.Object
	RequestRuntimeObj  runtime.Object
	RequestMetaJson    map[string]interface{}

	processor Processor
	Type      AuditType
	Operation map[string][]string //Operation操作
	Reason    string              //event类型的reason

	//process DAG
	processDAG *AuditProcessDAG
}

func NewAuditEvent(event *k8s_audit.Event) *AuditEvent {
	return &AuditEvent{
		Event:      event,
		processor:  metaProcessor,
		Operation:  make(map[string][]string),
		processDAG: InitAuditProcessADG(),
	}
}

func (a *AuditEvent) GetResponseOrRequestObj() runtime.Object {
	if a.ResponseRuntimeObj != nil {
		return a.ResponseRuntimeObj
	}
	if a.RequestRuntimeObj != nil {
		return a.RequestRuntimeObj
	}
	return nil
}

func (a *AuditEvent) TryGetPodFromEvent() *v1.Pod {
	if a.ResponseRuntimeObj != nil {
		pod, ok := a.ResponseRuntimeObj.(*v1.Pod)
		if pod != nil && ok {
			return pod
		}
	}

	if a.RequestRuntimeObj != nil {
		pod, ok := a.RequestRuntimeObj.(*v1.Pod)
		if pod != nil && ok {
			return pod
		}
	}

	return nil
}

func (a *AuditEvent) Process() {
	a.Add(1)
	go func() {
		a.processor.Process(a)
		//fmt.Printf("Operation:%v Reason:%s, uid: %s \n", a.Operation, a.Reason, a.GetObjectUID())
		a.Done()
	}()
}

// 判断依赖的父亲节点是否已经结束
func (a *AuditEvent) CanProcess(DAGNodeName string) {
	a.Wait()
	dagNode, ok := a.processDAG.allSpans[DAGNodeName]
	if !ok {
		return
	}
	for {
		if dagNode.IsParentsReady() {
			return
		}

		select {
		case <-dagNode.stateNotifier:
		}
	}
}

func (a *AuditEvent) FinishProcess(DAGNodeName string) {
	dagNode, ok := a.processDAG.allSpans[DAGNodeName]
	if !ok {
		return
	}

	dagNode.finish()
	dagNode.notify()
}

func (a *AuditEvent) GetObjectUID() (types.UID, error) {
	var resultUID types.UID
	var err error
	if a.ObjectRef.Resource == "events" {
		if event, ok := a.ResponseRuntimeObj.(*v1.Event); ok {
			resultUID = event.InvolvedObject.UID
		}
	} else {
		if a.ResponseRuntimeObj != nil {
			metaAccessor := meta.NewAccessor()
			resultUID, err = metaAccessor.UID(a.ResponseRuntimeObj)
			if err != nil {
				klog.Warningf("failed get uid from response runtime object eventID: %s ; %s \n", a.AuditID, err.Error())
			}
		}

		if err != nil && a.RequestRuntimeObj != nil {
			metaAccessor := meta.NewAccessor()
			resultUID, err = metaAccessor.UID(a.RequestRuntimeObj)
			if err != nil {
				klog.Warningf("failed get uid from request runtime object eventID: %s ; %s \n", a.AuditID, err.Error())
			}
		}
		if resultUID != "" {
			return resultUID, nil
		}

		if a.RequestRuntimeObj != nil {
			metaAccessor := meta.NewAccessor()
			resultUID, err = metaAccessor.UID(a.RequestRuntimeObj)
			if err != nil {
				klog.Warningf("failed get uid from runtime object eventID: %s ; %s \n", a.AuditID, err.Error())
			}
		}
	}
	return resultUID, err
}

type HitEvent struct {
	Event    *k8s_audit.Event
	HasError bool
	sync.WaitGroup
}

func NewHitEvent() *HitEvent {
	return &HitEvent{
		Event:    &k8s_audit.Event{},
		HasError: false,
	}
}

func (h *HitEvent) UnmarshalToEvent(hit *json.RawMessage) {
	h.Add(1)
	go func() {
		//var json = jsoniter.ConfigCompatibleWithStandardLibrary
		err := json.Unmarshal(*hit, &h.Event)
		if err != nil {
			klog.Errorf("failed unmarshal %s, data %s", err.Error(), string(*hit))
			h.HasError = true
		}
		h.Done()
	}()
}

type HitEventSlice struct {
	sync.Mutex
	hits []*HitEvent
}

func NewHitEventSlice(capacity int) *HitEventSlice {
	return &HitEventSlice{
		hits: make([]*HitEvent, 0, capacity),
	}

}

func (h *HitEventSlice) Wait() {
	for _, hit := range h.hits {
		hit.Wait()
	}
}

func (h *HitEventSlice) Append(hit *HitEvent) {
	h.Lock()
	h.hits = append(h.hits, hit)
	h.Unlock()
}

func (h *HitEventSlice) Hits() []*HitEvent {
	return h.hits
}

func (h *HitEventSlice) SortByTimeStamp(ascending bool) {
	sort.Slice(h.hits, func(i, j int) bool {
		if ascending {
			return h.hits[i].Event.StageTimestamp.Before(&h.hits[j].Event.StageTimestamp)
		} else {
			return !(h.hits[i].Event.StageTimestamp.Before(&h.hits[j].Event.StageTimestamp))
		}
	})
}

type ProcessState string

var (
	ProcessingState ProcessState = "Processing"
	FinishedState   ProcessState = "Finished"

	SLOProcessNode  string = "SLONode"
	SpanProcessNode string = "SpanNode"
)

type AuditProcessDAG struct {
	name          string
	state         ProcessState
	stateNotifier chan string //用于parent通知child状态变化
	parents       []*AuditProcessDAG
	children      []*AuditProcessDAG

	allSpans map[string]*AuditProcessDAG
}

func (d *AuditProcessDAG) finish() {
	d.state = FinishedState
}

func (d *AuditProcessDAG) notify() {
	for _, child := range d.children {
		child.stateNotifier <- "ok"
	}
}

func (d *AuditProcessDAG) IsParentsReady() bool {
	for _, parent := range d.parents {
		if parent.state != FinishedState {
			return false
		}
	}
	return true
}

func InitAuditProcessADG() *AuditProcessDAG {
	allSpans := make(map[string]*AuditProcessDAG, 0)

	spanNode := &AuditProcessDAG{
		name:          SpanProcessNode,
		state:         ProcessingState,
		stateNotifier: make(chan string, 60),
		parents:       []*AuditProcessDAG{},
		children:      []*AuditProcessDAG{},

		allSpans: allSpans,
	}

	sloNode := &AuditProcessDAG{
		name:          SLOProcessNode,
		state:         ProcessingState,
		stateNotifier: make(chan string, 60),
		parents:       []*AuditProcessDAG{},
		children:      []*AuditProcessDAG{},

		allSpans: allSpans,
	}
	spanNode.children = append(spanNode.children, sloNode)
	sloNode.parents = append(sloNode.parents, spanNode)

	allSpans[spanNode.name] = spanNode
	allSpans[sloNode.name] = sloNode

	return spanNode
}
