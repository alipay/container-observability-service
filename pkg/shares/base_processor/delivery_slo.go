package base_processor

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/metas"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/klog/v2"
)

type SLOTimeComputer struct {
}

func (s *SLOTimeComputer) CanProcess(event *shares.AuditEvent) bool {
	if event.ObjectRef.Resource != "pods" || event.ObjectRef.Subresource != "" {
		return false
	}

	if event.ResponseRuntimeObj == nil {
		return false
	}

	return true
}

func (s *SLOTimeComputer) Process(event *shares.AuditEvent) error {
	defer utils.IgnorePanic("processPodCreation")

	resPod, ok := event.ResponseRuntimeObj.(*v1.Pod)
	if !ok {
		klog.Warningf("can not convert pod from response runtime obj")
		return nil
	}

	sloSpec := metas.FetchSloSpec(resPod)

	if event.Verb == "create" {
		if sloSpec["PodCreate"] == nil {
			sloTime, _ := metas.GetPodSLOByDeliveryPath(resPod)
			sloSpec["PodCreate"] = &metas.SloSpecItem{
				SloTime: sloTime.String(),
			}
		}

		if sloSpec["PodUpgrade"] == nil {
			upgradeTimeout := 9 * time.Minute
			sloSpec["PodUpgrade"] = &metas.SloSpecItem{
				SloTime: upgradeTimeout.String(),
			}
		}

		if sloSpec["PodDelete"] == nil {
			deleteTimeout := 10 * time.Minute
			sloSpec["PodDelete"] = &metas.SloSpecItem{
				SloTime: deleteTimeout.String(),
			}
		}
	}

	accessor, err := meta.Accessor(event.ResponseRuntimeObj)
	if err != nil {
		klog.Errorf("fetch slo spec error, err: %s", err)
		return err
	}

	anno := accessor.GetAnnotations()
	if anno == nil {
		anno = map[string]string{}
	}
	accessor.SetAnnotations(anno)

	return nil
}
