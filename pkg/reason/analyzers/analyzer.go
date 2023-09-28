package analyzers

import (
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/spans"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
)

type DeliveryType string

const (
	PodCreate  DeliveryType = "PodCreate"
	PodDelete  DeliveryType = "PodDelete"
	PodUpgrade DeliveryType = "PodUpgrade"
)

type DAGAnalyzer struct {
	deliveryType DeliveryType
	result       *share.ReasonResult
	AuditEvents  []*shares.AuditEvent
	BeginTime    *time.Time
	EndTime      *time.Time

	maxConsumingFunc func([]*spans.Span, []*shares.AuditEvent, *time.Time) string
	deliveryModule   modules.DeliveryModule
	finishNotifier   chan string
	modules          map[string]modules.DeliveryModule
	Spans            []*spans.Span
}

func NewDAGAnalyzer(deliveryType DeliveryType, deliveryModule modules.DeliveryModule, maxSpanFunc func(spans []*spans.Span, events []*shares.AuditEvent, curTime *time.Time) string) *DAGAnalyzer {
	return &DAGAnalyzer{
		result:           &share.ReasonResult{},
		deliveryType:     deliveryType,
		deliveryModule:   deliveryModule,
		maxConsumingFunc: maxSpanFunc,
		finishNotifier:   make(chan string, 100),
		modules:          make(map[string]modules.DeliveryModule, 0),
	}
}

func (a *DAGAnalyzer) Analysis(cluster string, podName string, podUid string) *share.ReasonResult {
	defer utils.IgnorePanic("new_analysis_reason ")
	a.result.PodName = podName
	a.result.PodUid = podUid

	//a.PrintHyperEvent()
	//a.PrintMaxSpan()

	a.StartAnalysis()

	if !a.result.HasError {
		a.maxConsumingSpan()
	}
	return a.result
}

func (a *DAGAnalyzer) PrintHyperEvent() {
	for _, event := range a.AuditEvents {
		if event.Type == shares.AuditTypeEvent {
			fmt.Printf("Hyper event: %s, msg: %s\n", event.Reason, event.ResponseRuntimeObj.(*v1.Event).Message)
		} else {
			fmt.Printf("Hyper operation: %s\n", event.Operation)

		}
	}
}

func (a *DAGAnalyzer) PrintMaxSpan() {
	if a.Spans == nil {
		return
	}
	maxSpan := a.Spans[0]
	for _, span := range a.Spans {
		if maxSpan.Elapsed < span.Elapsed {
			maxSpan = span
		}
		fmt.Printf("pod name: %s, DAG analyzer span: %s, %s, begin: %s, end: %s\n", a.result.PodName, span.Name, span.Type,
			span.Begin.Format(time.RFC3339), span.End.Format(time.RFC3339))
	}
	//fmt.Printf("pod name: %s, DAGAnalyzer max span: %s, %s, %s\n", a.result.PodName, maxSpan.Name, maxSpan.Type, time.Duration(maxSpan.Elapsed*1000*1000))
}

func (a *DAGAnalyzer) StartAnalysis() {
	a.startAnalysisAll(a.deliveryModule)
	finishCount := 0
	timer := time.NewTimer(20 * time.Second)
	for {
		if len(a.modules) == finishCount {
			return
		}

		if a.isFinish() {
			a.stopAll()
			return
		}

		select {
		case <-a.finishNotifier:
			finishCount++
		case <-timer.C:
			klog.Warningf("wait for finish timeout for %s, module: %d, finish: %d\n", a.result.PodName, len(a.modules), finishCount)
			return
		}

	}
}

func (a *DAGAnalyzer) startAnalysisAll(deliveryModule modules.DeliveryModule) {
	dagModule, ok := deliveryModule.(*modules.DAGDeliveryModule)
	if !ok {
		return
	}

	if _, ok := a.modules[dagModule.Name()]; ok {
		return
	}

	go a.doModuleAnalysis(dagModule)
	a.modules[dagModule.Name()] = dagModule
	for _, c := range dagModule.Children() {
		a.startAnalysisAll(c)
	}
}

func (a *DAGAnalyzer) doModuleAnalysis(deliveryModule modules.DeliveryModule) {
	dagModule, ok := deliveryModule.(*modules.DAGDeliveryModule)
	if !ok {
		return
	}

	can := dagModule.CanDo()
	if !can {
		return
	}

	rs, hasErr := dagModule.Do(a.AuditEvents, a.BeginTime, a.EndTime)
	if rs != "" && !a.isFinish() {
		a.result.Result = rs
		a.result.HasError = hasErr
	}

	a.finishNotifier <- "ok"
	dagModule.FinishProcess()
}

func (a *DAGAnalyzer) stopAll() {
	for _, module := range a.modules {
		module.Stop()
	}
}

func (a *DAGAnalyzer) GetResult() *share.ReasonResult {
	return a.result
}

func (a *DAGAnalyzer) isFinish() bool {
	if a.result.Result != "" && a.result.HasError {
		return true
	}
	return false
}

func (a *DAGAnalyzer) maxConsumingSpan() {
	if a.result.HasError || a.maxConsumingFunc == nil {
		//fmt.Printf("pod name: %s retrun for hasError or maxFunc is nil\n", a.result.PodName)
		return
	}
	rs := a.maxConsumingFunc(a.Spans, a.AuditEvents, a.EndTime)
	if rs != "" {
		a.result.Result = rs
	}
}
