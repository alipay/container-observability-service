package trace

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/metas"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
)

type Span struct {
	Name        string
	Type        string
	Begin       time.Time
	End         time.Time
	elapsed     time.Duration
	Elapsed     int64
	Cluster     string
	ActionType  string
	TimeStamp   time.Time
	Omitempty   bool
	Children    []*Span
	parent      *Span
	config      *SpanConfig
	spanMeta    *SpanMeta
	errorEvents []*v1.Event

	tracer       trace.Tracer
	traceContext context.Context
	traceSpan    trace.Span
}

func (s *Span) SetElapsed(d time.Duration) {
	if d < 0 {
		d = 0
	}
	s.elapsed = d
	s.Elapsed = s.elapsed.Milliseconds()
}

func (s *Span) Empty() bool {
	return s.Begin.IsZero() && s.End.IsZero()
}

func (s *Span) Update(event *shares.AuditEvent) {
	s.config.Update(event, s)
	//更新trace信息
	s.updateTraceInfo(time.Time{})

	for _, child := range s.Children {
		child.Update(event)
	}
}

func (s *Span) Finish(endTime time.Time) {

	s.updateTraceInfo(endTime)
	for _, child := range s.Children {
		child.Finish(endTime)
	}
}

func (s *Span) InjectAttributesToSpan() {
	if s.traceSpan == nil || s.spanMeta == nil {
		return
	}

	//固定元数据
	s.traceSpan.SetAttributes(attribute.String("deliver.action", string(s.spanMeta.config.ActionType)))
	s.traceSpan.SetAttributes(attribute.String("k8s.resource", string(s.spanMeta.ObjectRef.Resource)))
	s.traceSpan.SetAttributes(attribute.String("k8s.uid", string(s.spanMeta.ObjectRef.UID)))
	s.traceSpan.SetAttributes(attribute.String("k8s.name", string(s.spanMeta.ObjectRef.Name)))
	s.traceSpan.SetAttributes(attribute.String("k8s.namespace", string(s.spanMeta.ObjectRef.Namespace)))
	s.traceSpan.SetAttributes(attribute.String("k8s.cluster", string(s.spanMeta.Cluster)))

	//用户定义的attribute
	for _, value := range s.spanMeta.ExtraProperties.Items() {
		pro := value.(*ExtraProperty)
		s.traceSpan.SetAttributes(attribute.String(pro.Name, pro.Value))
	}
}

func (s *Span) isFinished() bool {
	if s.config.Mode == StartFinish && !s.Begin.IsZero() && !s.End.IsZero() {
		return true
	}
	return false
}

func (s *Span) updateTraceInfo(endTime time.Time) {
	if s.traceSpan != nil && !s.traceSpan.IsRecording() {
		return
	}

	//尝试开始parent的span
	if !s.Begin.IsZero() && s.parent != nil {
		s.parent.tryStartSpan(s.Begin)
	}

	// 1. start-finish一旦捕获完成，直接结束span
	// 2. 整个trace结束，endTime不为zero
	if s.isFinished() || !endTime.IsZero() {
		isFinish := s.isFinished()

		if s.End.IsZero() {
			s.End = s.Begin
			//如果需要关闭span，则直接使用
			if s.config.NeedClose {
				s.End = endTime
			}
		}

		if s.Begin.IsZero() {
			s.Begin = s.End
		}

		if s.End.IsZero() {
			s.Begin = endTime
			s.End = endTime
		}

		s.tryStartSpan(s.Begin)
		if s.traceSpan != nil {
			if len(s.errorEvents) > 0 {
				for _, e := range s.errorEvents {
					s.traceSpan.RecordError(fmt.Errorf(e.Message), trace.WithTimestamp(e.LastTimestamp.Time))
				}
				if !isFinish {
					s.traceSpan.SetStatus(codes.Error, "")
				}
			}
			//更新属性信息
			s.InjectAttributesToSpan()

			s.traceSpan.End(trace.WithTimestamp(s.End))
			klog.V(8).Infof("pod %s , end span, type: %s name: %s , trace id: %s, span id: %s, begin: %s, end: %s\n", s.spanMeta.ObjectRef.UID, s.Type, s.Name,
				s.traceSpan.SpanContext().TraceID().String(), s.traceSpan.SpanContext().SpanID().String(), s.Begin.Format(time.RFC3339Nano), s.End.Format(time.RFC3339Nano))
		}
	}
}

func (s *Span) tryStartSpan(begin time.Time) {
	if s.traceContext == nil {
		var ctx context.Context
		if s.parent != nil {
			if s.parent.tracer == nil {
				s.parent.tryStartSpan(begin)
			}
			ctx = trace.ContextWithSpanContext(context.Background(), trace.SpanContextFromContext(s.parent.traceContext))
			s.tracer = s.parent.tracer

			klog.V(8).Infof("pod %s ctx is nil for span: %s, tracer is nil: %t, parent is nil: %t, parent tracecontext is nil %t\n", s.spanMeta.ObjectRef.UID, s.Type, s.tracer == nil, s.parent == nil, s.parent.traceContext == nil)
		}

		if len(s.config.Component) > 0 {
			s.tracer = getProvider(s.config.Component).Tracer(s.Name)
		}

		if spanID, ok := s.spanMeta.componentsTraces[s.config.Component]; ok {
			sid, _ := trace.SpanIDFromHex(spanID)
			ctx = context.WithValue(ctx, current_span_key, sid)
		}

		if s.tracer != nil {
			if s.Begin.IsZero() {
				s.Begin = begin
			}
			s.traceContext, s.traceSpan = s.tracer.Start(ctx, fmt.Sprintf("%s:%s", s.Type, s.Name),
				trace.WithTimestamp(s.Begin))
		}
	}
}

type ExtraProperty struct {
	Name       string `json:"Name,omitempty"`
	Value      string `json:"Value,omitempty"`      //json path to Value
	NeedMetric bool   `json:"NeedMetric,omitempty"` //is need metric
}

type SpanMeta struct {
	ObjectRef *audit.ObjectReference
	topSpan   *Span
	Spans     []*Span
	spanKeys  utils.ConcurrentMap
	//span config
	config *ResourceSpanConfig
	//properties
	Cluster           string
	CreationTimestamp time.Time
	ExtraProperties   utils.ConcurrentMap
	//slo time
	sloTime time.Duration

	mutex   *sync.Mutex
	written bool

	trace.SpanContextConfig
	ParentSpanID trace.SpanID

	Begin            time.Time
	End              time.Time
	traceInfo        *TraceInfo
	componentsTraces map[string]string
}

func NewSpanMeta(config *ResourceSpanConfig, cluster string, event *shares.AuditEvent) *SpanMeta {
	spanMeta := &SpanMeta{
		config:            config,
		ObjectRef:         event.ObjectRef,
		Cluster:           cluster,
		CreationTimestamp: event.RequestReceivedTimestamp.Time,
		Spans:             make([]*Span, 0),
		spanKeys:          utils.New(),
		ExtraProperties:   utils.New(),
		mutex:             &sync.Mutex{},
		written:           false,
		Begin:             event.RequestReceivedTimestamp.Time,
	}
	spanMeta.ObjectRef.UID, _ = event.GetObjectUID()
	sloSpec := metas.FetchSloSpec(event.ResponseRuntimeObj)
	if sloSpec[config.ActionType] != nil {
		sloTime, err := time.ParseDuration(sloSpec[config.ActionType].SloTime)
		if err == nil {
			spanMeta.sloTime = sloTime
		}
	}

	if spanMeta.sloTime == 0 {
		spanMeta.sloTime = 10 * time.Minute
	}

	spanMeta.fetchTraceContext(event)
	spanMeta.tryInitialNewSpan(event)
	return spanMeta
}

func (s *SpanMeta) TackSpan(event *shares.AuditEvent) {
	//fmt.Printf("SpanMeta.ActionType: [%s] TackSpan for %s\n", s.config.ActionType, s.ObjectRef.UID)
	s.tryInitialNewSpan(event)
	for _, span := range s.Spans {
		span.Update(event)
	}
}

func (s *SpanMeta) tryInitialNewSpan(event *shares.AuditEvent) {
	propertiesMap, rootSpans, addedChildren := s.config.Initial(event.ResponseRuntimeObj, s.spanKeys, s.ExtraProperties)

	// expend properties
	for key, pro := range propertiesMap {
		if _, ok := s.ExtraProperties.Get(key); !ok {
			s.ExtraProperties.Set(key, &ExtraProperty{
				NeedMetric: s.config.ExtraProperties[key].NeedMetric,
			})
		}
		p, _ := s.ExtraProperties.Get(key)
		p.(*ExtraProperty).Name = key
		p.(*ExtraProperty).Value = pro
	}

	//current level span
	for key, span := range rootSpans {
		if _, ok := s.spanKeys.Get(key); !ok {
			span.TimeStamp = event.StageTimestamp.Time
			span.Cluster = s.Cluster
			span.ActionType = s.config.ActionType
			span.parent = s.topSpan
			span.spanMeta = s

			s.spanKeys.Set(key, span)
			s.Spans = append(s.Spans, span)

			//s.injectTraceContext(span)
		}
	}

	//child level span
	for _, span := range addedChildren {
		span.TimeStamp = event.StageTimestamp.Time
		span.Cluster = s.Cluster
		span.ActionType = s.config.ActionType
		span.spanMeta = s
		//s.injectTraceContext(span)
	}
}

func (s *SpanMeta) fetchTraceContext(event *shares.AuditEvent) {
	if event == nil {
		return
	}
	metaObj, err := meta.Accessor(event.ResponseRuntimeObj)
	if err != nil {
		klog.Warningf("return from get meta, for %s\n", event.ObjectRef.Name)
		return
	}

	traceContextAnnotation, ok := metaObj.GetAnnotations()[TraceContextAnnotation]
	ctx := context.Background()
	if ok {
		traceInfos := make([]*TraceInfo, 0)
		err = json.Unmarshal([]byte(traceContextAnnotation), &traceInfos)
		if err != nil {
			klog.Infof("return from unmarshal, for %s\n", event.ObjectRef.Name)
			return
		}

		componentsTraces := make(map[string]string)
		var tid trace.TraceID
		var sid trace.SpanID

		for _, traceInfo := range traceInfos {
			if traceInfo.DeliveryType == s.config.ActionType && traceInfo.Status != ClosedTraceStatus {
				tid, _ = trace.TraceIDFromHex(traceInfo.TraceID)
				s.ParentSpanID, _ = trace.SpanIDFromHex(traceInfo.ParentSpanID)
				sid, _ = trace.SpanIDFromHex(traceInfo.RootSpanID)
				for _, service := range traceInfo.Services {
					componentsTraces[service.Component] = service.SpanID
				}
				s.componentsTraces = componentsTraces
				s.traceInfo = traceInfo
			}
		}

		parentSpanContext := trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: tid,
			SpanID:  s.ParentSpanID,
			Remote:  true,
		})

		ctx = trace.ContextWithRemoteSpanContext(context.Background(), parentSpanContext)
		ctx = context.WithValue(ctx, current_span_key, sid)
	} else {
		klog.Warningf("annotation %s is nil, for %s\n", event.ObjectRef.Name, TraceContextAnnotation)
	}

	s.topSpan = &Span{
		spanMeta: s,
		tracer:   getProvider("pod_delivery").Tracer("root"),
	}
	//承接上有trace id
	tc, ts := s.topSpan.tracer.Start(ctx, fmt.Sprintf("%s_%s", s.config.ActionType, s.ObjectRef.UID),
		trace.WithTimestamp(s.Begin))
	s.topSpan.traceContext = tc
	s.topSpan.traceSpan = ts

	s.TraceID = ts.SpanContext().TraceID()
	s.SpanID = ts.SpanContext().SpanID()
}

func (s *SpanMeta) Finish(ev *shares.AuditEvent) {
	if ev != nil {
		s.End = ev.StageTimestamp.Time
	} else {
		s.End = time.Now()
	}

	for _, span := range s.Spans {
		span.Finish(s.End)
	}

	if s.topSpan != nil {
		s.topSpan.InjectAttributesToSpan()
		s.topSpan.traceSpan.End(trace.WithTimestamp(s.End))
		klog.V(8).Infof("pod %s is finished, trace id: %s\n", s.ObjectRef.Name, s.topSpan.traceSpan.SpanContext().TraceID())
	}
}
