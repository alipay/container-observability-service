package analyzers

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/modules"
	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	spanpkg "github.com/alipay/container-observability-service/pkg/spans"
	"k8s.io/klog"
)

func init() {
	ShareAnalyzerFactory.Register(PodUpgrade, GeneratePodUpgradeAnalyzer)
}

func GeneratePodUpgradeAnalyzer() *DAGAnalyzer {
	return NewDAGAnalyzer(PodUpgrade, GeneratePodUpgradeDAG(), analysisUpgradeMaxSpan)
}

// GeneratePodUpgradeDAG 用于构建pod升级链路DAG
func GeneratePodUpgradeDAG() modules.DeliveryModule {
	kubeletDelayModule := modules.ShareModuleFactory.GetModuleByName(share.KUBELET_DELAY)
	imageModule := modules.ShareModuleFactory.GetModuleByName(share.IMAGE)
	containerKillModule := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_KILL)
	containerCreateModule := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_CREATE)
	containerStartReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_START)
	containerPostHookReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_POST_START)
	containerReadinessReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_READINESS)
	podReadinessReason := modules.ShareModuleFactory.GetModuleByName(share.POD_READINESS)

	// build DAG
	kubeletDelayModule.SetChildren([]modules.DeliveryModule{imageModule})

	imageModule.SetParents([]modules.DeliveryModule{kubeletDelayModule})
	imageModule.SetChildren([]modules.DeliveryModule{containerKillModule})

	containerKillModule.SetParents([]modules.DeliveryModule{imageModule})
	containerKillModule.SetChildren([]modules.DeliveryModule{containerCreateModule})

	containerCreateModule.SetParents([]modules.DeliveryModule{containerKillModule})
	containerCreateModule.SetChildren([]modules.DeliveryModule{containerStartReason})

	containerStartReason.SetParents([]modules.DeliveryModule{containerCreateModule})
	containerStartReason.SetChildren([]modules.DeliveryModule{containerPostHookReason})

	containerPostHookReason.SetParents([]modules.DeliveryModule{containerStartReason})
	containerPostHookReason.SetChildren([]modules.DeliveryModule{containerReadinessReason})

	containerReadinessReason.SetParents([]modules.DeliveryModule{containerPostHookReason})
	containerReadinessReason.SetChildren([]modules.DeliveryModule{podReadinessReason})

	podReadinessReason.SetParents([]modules.DeliveryModule{containerReadinessReason})
	return kubeletDelayModule
}

func analysisUpgradeMaxSpan(spans []*spanpkg.Span, events []*shares.AuditEvent, curTime *time.Time) string {
	if curTime == nil {
		return ""
	}

	pod := utils.GetPodYamlFromHyperEvents(events, curTime)

	var maxSpan *spanpkg.Span

	for idx, _ := range spans {
		span := spans[idx]
		if span.GetConfig().SpanOwner != spanpkg.K8sOwner && span.GetConfig().SpanOwner != spanpkg.CustomOwner {
			continue
		}
		klog.Infof("pod:%+v, span:%+v", pod.UID, span)

		if !span.Begin.IsZero() && span.End.IsZero() {
			span.End = *curTime
		}

		//对于调整后的span进行耗时重置
		if !span.Begin.IsZero() && !span.End.IsZero() {
			span.SetElapsed(span.End.Sub(span.Begin))
		}

		if maxSpan == nil || maxSpan.Elapsed < span.Elapsed {
			maxSpan = span
		}
	}

	if maxSpan != nil {
		if pod != nil && resultMap[SpanType(maxSpan.Type)] == "" {
			klog.V(8).Infof("pod name: %s, maxUpgradeSpan: %s/%s, begin:%s, end:%s\n", pod.Name, maxSpan.Name, maxSpan.Type, maxSpan.Begin, maxSpan.End)
		}
		return resultMap[SpanType(maxSpan.Type)]
	}
	return ""
}
