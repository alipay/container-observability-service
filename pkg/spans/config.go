package spans

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	lua "github.com/yuin/gopher-lua"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog"
	luajson "layeh.com/gopher-json"
)

const (
	StartFinish SpanMode = "start-finish"
	DirectInfo  SpanMode = "direct-info"

	CustomOwner SpanOwner = "custom"
	K8sOwner    SpanOwner = "k8s"
	OthersOwner SpanOwner = "others"
)

type SpanMode string

type SpanOwner string

type LuaMatcher struct {
	Scripts string `json:"Scripts,omitempty"`
}

func (l *LuaMatcher) match(event *shares.AuditEvent, spanName *string) bool {
	L := lua.NewState()
	defer L.Close()
	luajson.Preload(L)

	name := ""
	if spanName != nil {
		name = *spanName
	}

	reason := ""
	message := ""
	if e, ok := event.ResponseRuntimeObj.(*v1.Event); ok {
		reason = e.Reason
		message = e.Message
	}

	L.SetGlobal("requestObj", lua.LString("{}"))
	if event.RequestObject != nil {
		L.SetGlobal("requestObjStr", lua.LString(event.RequestObject.Raw))
	}

	L.SetGlobal("responseObjStr", lua.LString("{}"))
	if event.RequestObject != nil {
		L.SetGlobal("responseObjStr", lua.LString(event.ResponseObject.Raw))
		L.SetGlobal("responseObj", &lua.LUserData{Value: event.ResponseRuntimeObj})
	}
	L.SetGlobal("reason", lua.LString(reason))
	L.SetGlobal("message", lua.LString(message))

	L.SetGlobal("spanName", lua.LString(name))
	L.SetGlobal("verb", lua.LString(event.Verb))
	L.SetGlobal("userAgent", lua.LString(event.UserAgent))

	eventAnnoStr, err := json.Marshal(event.Annotations)
	if err == nil {
		L.SetGlobal("auditAnnotation", lua.LString(eventAnnoStr))
	}

	//inject helper func
	InjectHelperFuncToLua(L)

	//fmt.Printf("objectRef.Name: %s, verb: %s, l.Scripts：%s, requestObject.null: %t\n", event.ObjectRef.Name, event.Verb, l.Scripts, len(event.RequestObject.Raw) == 0)

	if err := L.DoString(l.Scripts); err != nil {
		klog.Errorf("do lua error: %s\n", err)
		return false
	}

	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	return lua.LVAsBool(ret)
}

type HyperEvent struct {
	Type        shares.AuditType `json:"Type,omitempty"`
	NameRex     string           `json:"NameRex,omitempty"`
	DurationRex string           `json:"DurationRex,omitempty"`
	Operation   string           `json:"Operation,omitempty"`
	Reason      string           `json:"Reason,omitempty"`
	LuaMatcher  *LuaMatcher      `json:"LuaMatcher,omitempty"`
}

func (h *HyperEvent) Match(event *shares.AuditEvent, spanName *string) (bool, time.Duration) {
	var names map[string]bool
	duration := time.Millisecond * 0

	if event.Type != h.Type {
		return false, duration
	}

	if h.Type == shares.AuditTypeEvent {
		names = h.GetName(event)
		duration = h.GetDuration(event)

		nameMatched := true
		if h.LuaMatcher != nil {
			nameMatched = h.LuaMatcher.match(event, spanName)
			return nameMatched, duration
		}

		if spanName != nil && len(h.NameRex) > 0 {
			_, nameMatched = names[*spanName]
		}

		if event.Reason == h.Reason {
			return nameMatched, duration
		}

	} else if h.Type == shares.AuditTypeOperation {
		nameMatched := true
		if h.LuaMatcher != nil {
			nameMatched = h.LuaMatcher.match(event, spanName)
			return nameMatched, duration
		}

		if _, ok := event.Operation[h.Operation]; ok {
			names = h.GetName(event)
			if spanName != nil && len(h.NameRex) > 0 {
				_, nameMatched = names[*spanName]
			}
			return nameMatched, duration
		}
	}
	return false, duration
}

func (h *HyperEvent) GetName(event *shares.AuditEvent) map[string]bool {
	result := make(map[string]bool)
	if h.NameRex == "" {
		return result
	}

	nameRegexp := regexp.MustCompile(h.NameRex)
	if h.Type == shares.AuditTypeEvent {
		e, ok := event.ResponseRuntimeObj.(*v1.Event)
		if !ok || e == nil {
			return result
		}

		rs := nameRegexp.FindStringSubmatch(e.Message)
		if rs != nil && len(rs) == 2 {
			result[rs[1]] = true
			return result
		}
	}

	if h.Type == shares.AuditTypeOperation {
		for _, info := range event.Operation[h.Operation] {
			rs := nameRegexp.FindStringSubmatch(info)
			if rs != nil && len(rs) == 2 {
				result[rs[1]] = true
			}
		}
		return result
	}

	return result
}

func (h *HyperEvent) GetDuration(event *shares.AuditEvent) time.Duration {
	duration := 0 * time.Millisecond
	if h.Type != shares.AuditTypeEvent {
		return duration
	}

	e, ok := event.ResponseRuntimeObj.(*v1.Event)
	if !ok || e == nil {
		return duration
	}

	if h.DurationRex != "" {
		durationRex := regexp.MustCompile(h.DurationRex)
		rs := durationRex.FindStringSubmatch(e.Message)
		if rs != nil && len(rs) == 2 {
			dur, err := time.ParseDuration(rs[1])
			if err == nil {
				duration = dur
			}
		}
	}

	return duration
}

type SpanConfig struct {
	Name        string        `json:"Name,omitempty"`    // span name
	Type        string        `json:"Type,omitempty"`    // span type
	NameRef     *string       `json:"NameRef,omitempty"` // json path
	Mode        SpanMode      `json:"Mode,omitempty"`
	DirectEvent []*HyperEvent `json:"DirectEvent,omitempty"`
	StartEvent  []*HyperEvent `json:"StartEvent,omitempty"`
	EndEvent    []*HyperEvent `json:"EndEvent,omitempty"`
	Omitempty   bool          `json:"Omitempty,omitempty"`
	SpanOwner   SpanOwner     `json:"SpanOwner,omitempty"`
	NeedClose   bool          `json:"NeedClose,omitempty"` // if need to close the span which does not have end time
}

func (s *SpanConfig) Initial(object runtime.Object) map[string]*Span {
	result := make(map[string]*Span, 0)

	if s.NameRef == nil {
		span := Span{
			Name:      s.Name,
			Type:      s.Type,
			Omitempty: s.Omitempty,
			config:    s,
		}
		result[span.Name] = &span
		return result
	}

	fieldRef := NewFieldRef(*s.NameRef, ".")
	valueMap := fieldRef.GetFieldValue(object)
	if valueMap == nil {
		return result
	}

	for key, val := range valueMap {
		span := Span{
			Name:      val.String(),
			Type:      s.Type,
			Omitempty: s.Omitempty,
			config:    s,
		}
		result[fmt.Sprintf("%s.%s", key, s.Type)] = &span
	}
	return result
}

func (s *SpanConfig) Update(event *shares.AuditEvent, span *Span) {
	if s.Mode == DirectInfo {
		if event.Type != shares.AuditTypeEvent {
			return
		}

		for _, directEvent := range s.DirectEvent {
			matched := false
			dur := 0 * time.Millisecond
			if s.NameRef != nil {
				matched, dur = directEvent.Match(event, &span.Name)
			} else {
				matched, dur = directEvent.Match(event, nil)
			}

			if matched {
				curEnd := event.RequestReceivedTimestamp.Time
				curBegin := curEnd.Add(-dur)

				if span.End.IsZero() {
					span.End = curEnd
				}

				if span.Begin.IsZero() {
					span.Begin = curBegin
				}

				if span.End.Before(curEnd) {
					span.End = curEnd
				}

				if span.Begin.After(curBegin) {
					span.Begin = curBegin
				}

				span.SetElapsed(span.End.Sub(span.Begin))

			}
		}
	} else if s.Mode == StartFinish {
		matched := false
		sName := &span.Name
		if s.NameRef == nil {
			sName = nil
		}

		for _, startEvent := range s.StartEvent {
			matched, _ = startEvent.Match(event, sName)
			if matched {
				//set the latest time for start event
				if span.Begin.IsZero() || span.Begin.Before(event.RequestReceivedTimestamp.Time) {
					span.Begin = event.RequestReceivedTimestamp.Time
				}
				if !span.End.IsZero() {
					span.SetElapsed(span.End.Sub(span.Begin))
				}
			}
		}

		for _, endEvent := range s.EndEvent {

			matched, _ = endEvent.Match(event, sName)
			if matched {
				//set the earliest time for start event
				if span.End.IsZero() || span.End.After(event.RequestReceivedTimestamp.Time) {
					span.End = event.RequestReceivedTimestamp.Time
				}

				if !span.Begin.IsZero() {
					span.SetElapsed(span.End.Sub(span.Begin))
				}
			}
		}
	}
}

type LifeFlag struct {
	StartName   *string       `json:"StartName,omitempty"`
	FinishName  *string       `json:"FinishName,omitempty"`
	StartEvent  []*HyperEvent `json:"StartEvent,omitempty"`
	FinishEvent []*HyperEvent `json:"FinishEvent,omitempty"`
}

func (l *LifeFlag) isStartToTrack(ev *shares.AuditEvent) bool {
	if l.StartEvent == nil || ev == nil {
		return false
	}

	for _, e := range l.StartEvent {
		matched, _ := e.Match(ev, l.StartName)
		if matched {
			return true
		}
	}
	return false
}

func (l *LifeFlag) isFinishToTrack(ev *shares.AuditEvent) bool {
	if l.FinishEvent == nil || ev == nil {
		return false
	}

	for _, e := range l.FinishEvent {
		matched, _ := e.Match(ev, l.FinishName)
		if matched {
			return true
		}
	}
	return false
}

type LuaFetcher struct {
	Scripts string `json:"Scripts,omitempty"`
}

func (l *LuaFetcher) fetchValue(object runtime.Object) (string, bool) {
	L := lua.NewState()
	defer L.Close()
	luajson.Preload(L)

	objByte, err := json.Marshal(object)
	if err != nil {
		klog.Errorf("json marshal error when fetch value: %s\n", err)
		return "", false
	}

	L.SetGlobal("objectJsonStr", lua.LString(objByte))
	//inject helper func
	InjectHelperFuncToLua(L)

	if err := L.DoString(l.Scripts); err != nil {
		klog.Errorf("do lua error when fetch value for property, err: %s\n", err)
		return "", false
	}

	ret := L.Get(-1) // returned value
	L.Pop(1)         // remove received value
	return lua.LVAsString(ret), true
}

type ExtraPropertyConfig struct {
	Name         string      `json:"Name,omitempty"`
	ValueRex     string      `json:"ValueRex,omitempty"`     //json path to Value
	ValueFetcher *LuaFetcher `json:"ValueFetcher,omitempty"` //json path to Value
	NeedMetric   bool        `json:"NeedMetric,omitempty"`   //is need metric
}

type ResourceSpanConfig struct {
	ObjectRef       *audit.ObjectReference          `json:"ObjectRef,omitempty"`
	ActionType      string                          `json:"ActionType,omitempty"` //span计算类型
	LifeFlag        *LifeFlag                       `json:"LifeFlag,omitempty"`   //标记Span的开始和结束
	Spans           []*SpanConfig                   `json:"Spans,omitempty"`
	ExtraProperties map[string]*ExtraPropertyConfig `json:"ExtraProperties,omitempty"` //需要提取的属性 map[name] = [json.path.to.value]
}

func (r *ResourceSpanConfig) IsStartToTrack(ev *shares.AuditEvent) bool {
	if r.LifeFlag == nil || ev == nil {
		return false
	}
	rs := r.LifeFlag.isStartToTrack(ev)

	return rs
}

func (r *ResourceSpanConfig) IsFinishToTrack(ev *shares.AuditEvent) bool {
	if r.LifeFlag == nil || ev == nil {
		return false
	}
	return r.LifeFlag.isFinishToTrack(ev)
}

func (r *ResourceSpanConfig) Initial(object runtime.Object, uid string, properties utils.ConcurrentMap) (map[string]string, map[string]*Span) {
	//get extra-properties from object
	propertiesResult := make(map[string]string, 0)
	if r.ExtraProperties != nil {
		for pName, pConfig := range r.ExtraProperties {
			if _, ok := properties.Get(pName); ok {
				continue
			}
			//fetch property from field rex
			if len(pConfig.ValueRex) > 0 {
				fieldRef := NewFieldRef(pConfig.ValueRex, "#")
				valueMap := fieldRef.GetFieldValue(object)
				if valueMap != nil {
					for _, val := range valueMap {
						propertiesResult[pName] = val.String()
					}
				}
			}

			// fetch property from lua
			if pConfig.ValueFetcher != nil {
				val, ok := pConfig.ValueFetcher.fetchValue(object)
				if ok {
					propertiesResult[pName] = val
				}
			}
		}
	}

	//get span
	spanResult := make(map[string]*Span, 0)
	for _, span := range r.Spans {
		rs := span.Initial(object)
		if rs != nil {
			for k, v := range rs {
				spanResult[k] = v
			}
		}
	}
	return propertiesResult, spanResult
}

func (r *ResourceSpanConfig) GetExtraPropertyNames() []string {
	result := make([]string, 0)
	if r.ExtraProperties == nil {
		return result
	}

	for k, v := range r.ExtraProperties {
		if v.NeedMetric {
			result = append(result, k)
		}
	}
	return result
}

type ResourceSpanConfigList []*ResourceSpanConfig

func NewResourceSpanConfigList() ResourceSpanConfigList {
	list := make([]*ResourceSpanConfig, 0)
	return list
}

func (r *ResourceSpanConfigList) GetConfigByRef(ref *audit.ObjectReference) []*ResourceSpanConfig {
	result := make([]*ResourceSpanConfig, 0)
	if ref == nil {
		return result
	}
	for _, config := range *r {
		if config.ObjectRef.Resource == ref.Resource && config.ObjectRef.APIVersion == ref.APIVersion {
			result = append(result, config)
		}
	}
	return result
}

func (r *ResourceSpanConfigList) GetExtraPropertyNames() []string {
	result := make([]string, 0)
	for idx, _ := range *r {
		result = append(result, (*r)[idx].GetExtraPropertyNames()...)
	}
	return result
}
