package base_processor

import (
	"strings"

	"github.com/alipay/container-observability-service/pkg/metas"
	"github.com/alipay/container-observability-service/pkg/shares"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type ObjectDecoder struct {
}

func (m *ObjectDecoder) CanProcess(event *shares.AuditEvent) bool {
	return true
}

func (m *ObjectDecoder) Process(event *shares.AuditEvent) error {
	reqObjGVK := &schema.GroupVersionKind{Group: event.ObjectRef.APIGroup, Version: event.ObjectRef.APIVersion, Kind: event.ObjectRef.Resource}

	if len(event.ObjectRef.Subresource) > 0 && !strings.EqualFold(strings.ToLower(event.Verb), "patch") {
		reqObjGVK.Kind = event.ObjectRef.Subresource
	}

	if obj := metas.GetObjectFromRuntimeUnknown(event.ResponseObject, nil); obj != nil {
		event.ResponseRuntimeObj = obj
	}

	if obj := metas.GetObjectFromRuntimeUnknown(event.RequestObject, reqObjGVK); obj != nil {
		event.RequestRuntimeObj = obj
	}

	event.RequestMetaJson = metas.GetJsonFromRuntimeUnknown(event.RequestObject)

	klog.V(8).Infof("auditId: %s, requestObjNil: %t \n", event.AuditID, event.RequestRuntimeObj == nil)
	return nil
}
