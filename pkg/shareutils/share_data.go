package shareutils

import (
	"fmt"
	"sync"

	"github.com/alipay/container-observability-service/pkg/shares"

	"github.com/alipay/container-observability-service/pkg/spans"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

var slicePool = sync.Pool{
	New: func() interface{} {
		s := make([]*spans.Span, 0, 60)
		return s
	},
}

var spanPool = sync.Pool{
	New: func() interface{} {
		return &spans.Span{}
	},
}

func GetSpansByUIDAndType(objectUID types.UID, event *shares.AuditEvent, actionType string) ([]*spans.Span, func()) {
	uid := objectUID
	var err error
	if len(uid) == 0 && event != nil {
		uid, err = event.GetObjectUID()
		if err != nil {
			return nil, nil
		}
	}

	tmpList, ok := spans.DeliverySpanProcessor.SpanMetas.Load(string(uid))
	if !ok || tmpList == nil {
		return nil, nil
	}
	spanMetaList, ok := tmpList.([]*spans.SpanMeta)
	if !ok || spanMetaList == nil {
		return nil, nil
	}

	for idx, _ := range spanMetaList {
		spanMeta := spanMetaList[idx]
		if spanMeta == nil {
			continue
		}
		if spanMeta.SpanConfig().ActionType == actionType {
			klog.V(8).Info(fmt.Sprintf("GetSpansByUIDAndType, slo uid: %s, spans.len:%d\n", uid, len(spanMeta.Spans)))
			//result = make([]*spans.Span, 0, len(spanMeta.Spans))
			result, _ := slicePool.Get().([]*spans.Span)
			for i, _ := range spanMeta.Spans {
				tmpSpan := spanPool.Get().(*spans.Span)
				*tmpSpan = *spanMeta.Spans[i]
				result = append(result, tmpSpan)
			}
			return result, func() {
				for idx, _ := range result {
					if result[idx] != nil {
						result[idx].Reset()
						spanPool.Put(result[idx])
					}
				}
				result = result[:0]
				slicePool.Put(result)
			}
		}
	}

	return nil, nil
}
