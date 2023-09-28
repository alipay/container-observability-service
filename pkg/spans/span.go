package spans

import (
	"sync"
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog/v2"
)

type Span struct {
	Name       string
	Type       string
	Begin      time.Time
	End        time.Time
	elapsed    time.Duration
	Elapsed    int64
	Cluster    string
	ActionType string
	TimeStamp  time.Time
	Omitempty  bool

	config *SpanConfig
}

func (s *Span) GetConfig() *SpanConfig {
	return s.config
}

func (s *Span) SetConfig(sc *SpanConfig) {
	s.config = sc
}

func (s *Span) SetElapsed(d time.Duration) {
	if d < 0 {
		d = 0
	}
	s.elapsed = d
	s.Elapsed = s.elapsed.Milliseconds()
}

func (s *Span) Emptry() bool {
	return s.Begin.IsZero() && s.End.IsZero()
}

func (s *Span) Reset() {
	s.Type = ""
	s.Name = ""
	s.Begin = time.Time{}
	s.End = time.Time{}
	s.elapsed = 0
	s.Elapsed = 0
	s.Cluster = ""
	s.ActionType = ""
	s.TimeStamp = time.Time{}
	s.Omitempty = false
	s.config = nil
}

type ExtraProperty struct {
	Name       string `json:"Name,omitempty"`
	Value      string `json:"Value,omitempty"`      //json path to Value
	NeedMetric bool   `json:"NeedMetric,omitempty"` //is need metric
}

type SpanMeta struct {
	ObjectRef *audit.ObjectReference
	Spans     []*Span
	spanKeys  utils.ConcurrentMap
	//span config
	config *ResourceSpanConfig
	//properties
	Cluster           string
	CreationTimestamp time.Time
	ExtraProperties   utils.ConcurrentMap

	mutex   *sync.Mutex
	written bool
}

func NewSpanMeta(config *ResourceSpanConfig, cluster string, createTime time.Time, event *shares.AuditEvent) *SpanMeta {
	spanMeta := &SpanMeta{
		config:            config,
		ObjectRef:         event.ObjectRef,
		Cluster:           cluster,
		CreationTimestamp: createTime,
		Spans:             make([]*Span, 0),
		spanKeys:          utils.New(),
		ExtraProperties:   utils.New(),
		mutex:             &sync.Mutex{},
		written:           false,
	}
	spanMeta.tryUpdateSpan(event)
	spanMeta.ObjectRef.UID, _ = event.GetObjectUID()

	return spanMeta
}

func (s *SpanMeta) TrackSpan(event *shares.AuditEvent) {
	defer HandleCrash()
	oldLen := len(s.Spans)
	s.tryUpdateSpan(event)
	for idx, span := range s.Spans {
		uid, _ := event.GetObjectUID()
		if span == nil {
			klog.Errorf("event.ObjectRef.uid: %s, oldLen(%d)/curLen(%d), idx: %d, track span error, span nil\n", uid, oldLen, len(s.Spans), idx)
		} else if span.config == nil {
			klog.Errorf("event.ObjectRef.uid: %s, oldLen(%d)/curLen(%d), idx: %d, track span error, span.config nil\n", uid, oldLen, len(s.Spans), idx)
		}
		span.config.Update(event, span)
	}
}

func (s *SpanMeta) tryUpdateSpan(event *shares.AuditEvent) {
	propertiesMap, spanMap := s.config.Initial(event.ResponseRuntimeObj, string(s.ObjectRef.UID), s.ExtraProperties)

	for key, pro := range propertiesMap {
		s.ExtraProperties.Set(key, &ExtraProperty{
			NeedMetric: s.config.ExtraProperties[key].NeedMetric,
		})
		p, _ := s.ExtraProperties.Get(key)
		p.(*ExtraProperty).Name = key
		p.(*ExtraProperty).Value = pro
	}

	for key, span := range spanMap {
		if _, ok := s.spanKeys.Get(key); !ok {
			span.TimeStamp = event.StageTimestamp.Time
			span.Cluster = s.Cluster
			span.ActionType = s.config.ActionType
			s.spanKeys.Set(key, true)
			s.Spans = append(s.Spans, span)
		}
	}
}

func (s *SpanMeta) SpanConfig() *ResourceSpanConfig {
	return s.config
}

func (s *SpanMeta) finishOpenSpanNow(now time.Time) {
	if s.Spans == nil || now.IsZero() {
		return
	}
	for _, span := range s.Spans {
		if span == nil || span.config == nil || !span.config.NeedClose {
			continue
		}
		if !span.Begin.IsZero() && span.End.IsZero() {
			span.End = now
			span.SetElapsed(now.Sub(span.Begin))
		}
	}
}
