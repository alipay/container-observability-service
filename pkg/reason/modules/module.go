package modules

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"
)

type DeliveryModule interface {
	Do(AuditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool)
	Name() string
	CanDo() bool
	SetParents([]DeliveryModule)
	SetChildren([]DeliveryModule)
	Stop()
}

type ProcessState string

const (
	ProcessingState ProcessState = "Processing"
	FinishedState   ProcessState = "Finished"
)

type DAGDeliveryModule struct {
	name          string
	state         ProcessState
	stateNotifier chan string //用于parent通知child状态变化
	stop          chan string //用于终止分析信号
	parents       []*DAGDeliveryModule
	children      []*DAGDeliveryModule
	analysisFunc  func(AuditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool)
}

func NewDAGDeliveryModule(name string,
	analysisFunc func(AuditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool)) *DAGDeliveryModule {
	return &DAGDeliveryModule{
		name:          name,
		state:         ProcessingState,
		stateNotifier: make(chan string, 10),
		stop:          make(chan string, 10),
		parents:       []*DAGDeliveryModule{},
		children:      []*DAGDeliveryModule{},
		analysisFunc:  analysisFunc,
	}
}

func (a *DAGDeliveryModule) Name() string {
	return a.name
}

func (a *DAGDeliveryModule) Do(auditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool) {
	if a.analysisFunc == nil {
		return "", false
	}

	return a.analysisFunc(auditEvents, beginTime, endTime)
}

// 判断依赖的父亲节点是否已经结束
func (a *DAGDeliveryModule) CanDo() bool {
	for {
		if a.IsParentsReady() {
			return true
		}
		select {
		case <-a.stateNotifier:
		case <-a.stop:
			return false
		}
	}
}

func (a *DAGDeliveryModule) Children() []*DAGDeliveryModule {
	return a.children
}

func (a *DAGDeliveryModule) SetParents(parents []DeliveryModule) {
	for _, parent := range parents {
		p, ok := parent.(*DAGDeliveryModule)
		if ok && p != nil {
			a.parents = append(a.parents, p)
		}
	}
}

func (a *DAGDeliveryModule) SetChildren(children []DeliveryModule) {
	for _, child := range children {
		c, ok := child.(*DAGDeliveryModule)
		if ok && c != nil {
			a.children = append(a.children, c)
		}
	}
}

func (a *DAGDeliveryModule) IsParentsReady() bool {
	for _, p := range a.parents {
		if p.state != FinishedState {
			return false
		}
	}
	return true
}

func (a *DAGDeliveryModule) FinishProcess() {
	a.finish()
	a.notify()
}

func (a *DAGDeliveryModule) SetAnalysisFunc(f func(AuditEvents []*shares.AuditEvent, beginTime *time.Time, endTime *time.Time) (result string, hasError bool)) {
	a.analysisFunc = f
}

func (a *DAGDeliveryModule) Stop() {
	a.stop <- "stop"
}

func (a *DAGDeliveryModule) finish() {
	a.state = FinishedState
}

func (a *DAGDeliveryModule) notify() {
	for _, c := range a.children {
		c.stateNotifier <- "ok"
	}
}
