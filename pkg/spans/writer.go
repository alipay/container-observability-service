package spans

import (
	"context"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/alipay/container-observability-service/pkg/featuregates"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/utils"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	"github.com/olivere/elastic/v7"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/apis/audit"
	"k8s.io/klog/v2"
)

type Writer interface {
	Write(p *SpanMeta) error
}

var _ Writer = &XSearchWriter{}
var spanBuffer *utils.BufferUtils = nil

func NewXSearchWriter() (*XSearchWriter, error) {
	client := xsearch.GetXSearchClient()
	if client == nil {
		return nil, fmt.Errorf("xsearch client is not initialized")
	}
	return &XSearchWriter{
		esClient: client,
	}, nil
}

type XSearchWriter struct {
	esClient       *elastic.Client
	processedCount int
}

func concatDocIdForPod(cluster string, p *corev1.Pod) (docId string) {
	if p == nil {
		return
	}
	return fmt.Sprintf("%s.%s.%s.%s", cluster, p.Namespace, p.Name, string(p.GetUID()))
}

func concatDocIdForMeta(p *SpanMeta) (docId string) {
	if p == nil {
		return
	}
	return fmt.Sprintf("%s.%s.%s.%s", p.Cluster, p.ObjectRef.Namespace, p.ObjectRef.Name, p.ObjectRef.UID)
}

func concatDocIdForSpan(pm *SpanMeta, span *Span) (docId string) {
	if pm == nil {
		return
	}
	return fmt.Sprintf("%s.%s.%s.%s", concatDocIdForMeta(pm), span.Type, span.Name, utils.RandString(5))
}

func (x *XSearchWriter) Write(p *SpanMeta) error {
	if nil == p {
		return nil
	}
	properties := p.ExtraProperties.Items()
	attrs := convertAttributes(properties, string(p.ObjectRef.UID))

	// TODO make error aggregated
	// fetch begin/end time from spans
	var beginTime = time.Now().Add(365 * 24 * time.Hour)
	var endTime time.Time
	for idx, _ := range p.Spans {
		if !p.Spans[idx].Begin.IsZero() && p.Spans[idx].Begin.Before(beginTime) {
			beginTime = p.Spans[idx].Begin
		}
		if !p.Spans[idx].End.IsZero() && p.Spans[idx].End.After(endTime) {
			endTime = p.Spans[idx].End
		}
	}

	var ctx context.Context
	var rootSpan trace.Span
	if featuregates.IsEnabled(JaegerFeature) {
		ctx, rootSpan = x.startRootSpan(p.ObjectRef, p.config.ActionType, attrs, beginTime)
	}
	var err error
	for idx, _ := range p.Spans {
		body := struct {
			OwnerRef *audit.ObjectReference
			*Span
			Properties map[string]interface{}
		}{
			OwnerRef:   p.ObjectRef,
			Span:       p.Spans[idx],
			Properties: properties,
		}
		if p.Spans[idx].Emptry() && p.Spans[idx].Omitempty {
			continue
		}

		err = x.writePart(body, concatDocIdForSpan(p, p.Spans[idx]))
		x.writeMetric(p.Cluster, p.ObjectRef.Namespace, p.ObjectRef.Resource, p.Spans[idx].ActionType, p.Spans[idx].Type, properties, float64(p.Spans[idx].Elapsed))

		if featuregates.IsEnabled(JaegerFeature) {
			//spanSnapShot := x.buildSpanSnapshot(p.Cluster, p.ObjectRef, p.Spans[idx], attrs, p.config.ActionType, p.CreationTimestamp)
			//spanSnapshots = append(spanSnapshots, spanSnapShot)
			x.buildSpan(ctx, p.Spans[idx], attrs, endTime)
		}
	}
	if featuregates.IsEnabled(JaegerFeature) {
		//x.emitSpanSnapshot(spanSnapshots, p.ObjectRef, p.config.ActionType, attrs, p.CreationTimestamp)
		rootSpan.End(trace.WithTimestamp(endTime))
	}
	return err
}

func (x *XSearchWriter) writePart(obj interface{}, docID string) error {
	if spanBuffer == nil {
		spanBuffer = utils.NewBufferUtils(SpanIndex, 1000, 10*time.Second, false, func(datas map[string]interface{}) error {
			if datas == nil {
				return nil
			}

			klog.V(6).Infof("do spans bulk, data size: %d", len(datas))
			err := utils.ReTry(func() error {
				bulkService := x.esClient.Bulk()
				for id, data := range datas {
					doc := elastic.NewBulkIndexRequest().Index(SpanIndex).Type(SpanDocType).Id(id).Doc(data).UseEasyJSON(true)
					bulkService = bulkService.Add(doc)
					data = nil
				}
				_, err := bulkService.Do(context.Background())
				if err != nil {
					return err
				}
				return nil
			}, 1*time.Second, 5)

			if err != nil {
				return err
			}
			return nil
		},
		)

		spanBuffer.DoClearData()
		//add graceful clear
		xsearch.XSearchClear.AddCleanWork(func() {
			spanBuffer.Stop()
		})
	}

	//insert to es with retry
	err := spanBuffer.SaveData(docID, obj)

	if err != nil {
		klog.Errorf("write part for [%+v] error: %v", obj, err)
		return err
	}

	/*err := utils.ReTry(func() (err error) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		_, err = x.esClient.Index().Index(SpanIndex).Type(SpanDocType).
			Id(docId).BodyJson(obj).Do(ctx)

		if err != nil {
			return err
		}
		return nil
	}, 1*time.Second, 5)

	if err != nil {
		klog.Errorf("write part for [%+v] error: %v", obj, err)
		return err
	}*/

	return nil
}

// write metrics
func (x *XSearchWriter) writeMetric(cluster, namespace, resource, actionType string, spanType string, properties map[string]interface{}, value float64) error {
	labels := map[string]string{
		"cluster": cluster,
		//"namespace":   namespace,
		"resource":    resource,
		"type":        spanType,
		"action_type": actionType,
	}

	for k, v := range properties {
		if v.(*ExtraProperty).NeedMetric {
			labels[k] = v.(*ExtraProperty).Value
		}
	}

	//对齐指标
	// delete label
	toDelete := make([]string, 0)
	for k, _ := range labels {
		if _, ok := metrics.SpanConsumingLabels[k]; !ok {
			toDelete = append(toDelete, k)
		}
	}
	for idx, _ := range toDelete {
		delete(labels, toDelete[idx])
	}

	// add label
	for k, _ := range metrics.SpanConsumingLabels {
		if _, ok := labels[k]; !ok {
			labels[k] = ""
		}
	}

	//write value
	metrics.SpansConsumingStatistic.With(labels).Observe(value)
	return nil
}

/*func (x *XSearchWriter) buildSpanSnapshot(cluster string, objectRef *audit.ObjectReference, span *Span, attrs []attribute.KeyValue, actionType string, createTime time.Time) *tracesdk.SpanSnapshot {
	res := resource.NewWithAttributes(semconv.ServiceNameKey.String(span.Type), semconv.ServiceNamespaceKey.String(objectRef.Namespace))
	start := span.Begin
	end := span.End

	if start.IsZero() {
		start = end
	}
	if end.IsZero() {
		end = start
	}

	if end.Before(start) {
		end = start
	}

	spanSnapShot := &tracesdk.SpanSnapshot{
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: UIDToTraceID(fmt.Sprintf("%s_%s_%s", actionType, string(objectRef.UID), createTime.String())),
			SpanID:  SpanToSpanID(span.Name),
		}),
		SpanKind:        trace.SpanKindInternal,
		Name:            span.Name,
		StartTime:       start,
		EndTime:         end,
		Attributes:      attrs,
		StatusCode:      codes.Ok,
		HasRemoteParent: true,
		Resource:        res,
	}
	return spanSnapShot
}*/

func convertAttributes(properties map[string]interface{}, uid string) []attribute.KeyValue {
	attrs := []attribute.KeyValue{}

	for _, v := range properties {
		pro := v.(*ExtraProperty)
		attrs = append(attrs, attribute.String(pro.Name, pro.Value))
	}
	attrs = append(attrs, attribute.String("uid", uid))
	return attrs

}
func UIDToTraceID(uid string) trace.TraceID {
	f := fnv.New64()
	_, _ = f.Write([]byte(uid))
	var h trace.TraceID
	_ = f.Sum(h[:0])
	return h
}

func SpanToSpanID(name string) trace.SpanID {
	f := fnv.New64()
	_, _ = f.Write([]byte(name))
	var h trace.SpanID
	_ = f.Sum(h[:0])
	return h
}

/*
func (x *XSearchWriter) emitSpanSnapshot(spans []*tracesdk.SpanSnapshot, objectRef *audit.ObjectReference, actionType string, attrs []attribute.KeyValue, createTime time.Time) {
	res := resource.NewWithAttributes(semconv.ServiceNameKey.String("pod_delivery"), semconv.ServiceNamespaceKey.String(objectRef.Namespace))
	parentSpanSnapshot := &tracesdk.SpanSnapshot{
		SpanContext: trace.NewSpanContext(trace.SpanContextConfig{
			TraceID: UIDToTraceID(fmt.Sprintf("%s_%s_%s", actionType, string(objectRef.UID), createTime.String())),
			SpanID:  SpanToSpanID(objectRef.Name),
		}),
		SpanKind:        trace.SpanKindInternal,
		Name:            fmt.Sprintf("%s_%s", actionType, objectRef.UID),
		Attributes:      attrs,
		StatusCode:      codes.Ok,
		HasRemoteParent: true,
		Resource:        res,
	}
	for idx, _ := range spans {
		if parentSpanSnapshot.StartTime.IsZero() {
			parentSpanSnapshot.StartTime = spans[idx].StartTime
		}
		if parentSpanSnapshot.EndTime.IsZero() {
			parentSpanSnapshot.EndTime = spans[idx].EndTime
		}

		if !spans[idx].StartTime.IsZero() && parentSpanSnapshot.StartTime.After(spans[idx].StartTime) {
			parentSpanSnapshot.StartTime = spans[idx].StartTime
		}
		if !spans[idx].EndTime.IsZero() && parentSpanSnapshot.EndTime.Before(spans[idx].EndTime) {
			parentSpanSnapshot.EndTime = spans[idx].EndTime
		}

		spans[idx].ParentSpanID = parentSpanSnapshot.SpanContext.SpanID()
	}
	spans = append(spans, parentSpanSnapshot)
	for idx, _ := range spans {
		if spans[idx].StartTime.IsZero() || spans[idx].EndTime.IsZero() {
			spans[idx].StartTime = parentSpanSnapshot.EndTime
			spans[idx].EndTime = parentSpanSnapshot.EndTime
		}
	}

	klog.V(8).Infof("emitting span for pod %s, action: %s, traceId: %s, spans.len: %d", objectRef.Name, actionType, parentSpanSnapshot.SpanContext.TraceID(), len(spans))
	err := spanExporter.ExportSpans(context.TODO(), spans)

	if err != nil {
		klog.Errorf("failed to emit span for pod %s, msg: %s", objectRef.Name, err.Error())
	}
}*/

func (x *XSearchWriter) startRootSpan(objectRef *audit.ObjectReference, actionType string, attrs []attribute.KeyValue, createTime time.Time) (context.Context, trace.Span) {
	tracer := getProvider("pod_delivery").Tracer("root_tracer")

	ctx := context.Background()
	ctx, span := tracer.Start(ctx, fmt.Sprintf("%s_%s", actionType, objectRef.UID),
		trace.WithAttributes(attrs...),
		trace.WithNewRoot(),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithTimestamp(createTime))

	return ctx, span
}

func (x *XSearchWriter) buildSpan(ctx context.Context, span *Span, attrs []attribute.KeyValue, latestTime time.Time) {
	start := span.Begin
	end := span.End

	if start.IsZero() {
		start = end
	}
	if end.IsZero() {
		end = start
	}

	if end.Before(start) {
		end = start
	}

	if start.IsZero() || end.IsZero() {
		start = latestTime
		end = latestTime
	}

	tracer := getProvider(span.Type).Tracer("sub_tracer")
	ctx, traceSpan := tracer.Start(ctx, span.Name,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindInternal),
		trace.WithTimestamp(start))
	traceSpan.End(trace.WithTimestamp(end))
}
