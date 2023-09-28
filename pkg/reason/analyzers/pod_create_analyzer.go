package analyzers

import (
	"github.com/alipay/container-observability-service/pkg/reason/modules"
	_ "github.com/alipay/container-observability-service/pkg/reason/modules/pods"
	"github.com/alipay/container-observability-service/pkg/reason/share"
)

func init() {
	ShareAnalyzerFactory.Register(PodCreate, GeneratePodCreateAnalyzer)
}

func GeneratePodCreateAnalyzer() *DAGAnalyzer {
	return NewDAGAnalyzer(PodCreate, GeneratePodCreateDAG(), analysisMaxSpan)
}

// 用于构建pod创建链路DAG
func GeneratePodCreateDAG() modules.DeliveryModule {
	scheduleModule := modules.ShareModuleFactory.GetModuleByName(share.SCHEDULER)
	networkModule := modules.ShareModuleFactory.GetModuleByName(share.NETWORK)
	volumeModule := modules.ShareModuleFactory.GetModuleByName(share.VOLUME)
	admissionModule := modules.ShareModuleFactory.GetModuleByName(share.ADMISSION)
	sandboxModule := modules.ShareModuleFactory.GetModuleByName(share.SANDBOX)
	kubeletDelayModule := modules.ShareModuleFactory.GetModuleByName(share.KUBELET_DELAY)
	imageModule := modules.ShareModuleFactory.GetModuleByName(share.IMAGE)
	containerCreateModule := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_CREATE)
	containerStartReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_START)
	containerPostHookReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_POST_START)
	containerReadinessReason := modules.ShareModuleFactory.GetModuleByName(share.CONTAINER_READINESS)
	podReadinessReason := modules.ShareModuleFactory.GetModuleByName(share.POD_READINESS)

	// build DAG
	scheduleModule.SetChildren([]modules.DeliveryModule{networkModule, volumeModule, admissionModule})

	networkModule.SetParents([]modules.DeliveryModule{scheduleModule})
	networkModule.SetChildren([]modules.DeliveryModule{sandboxModule})

	volumeModule.SetParents([]modules.DeliveryModule{scheduleModule})
	volumeModule.SetChildren([]modules.DeliveryModule{sandboxModule})

	admissionModule.SetParents([]modules.DeliveryModule{scheduleModule})
	admissionModule.SetChildren([]modules.DeliveryModule{sandboxModule})

	sandboxModule.SetParents([]modules.DeliveryModule{networkModule, volumeModule, admissionModule})
	sandboxModule.SetChildren([]modules.DeliveryModule{kubeletDelayModule})

	kubeletDelayModule.SetParents([]modules.DeliveryModule{sandboxModule})
	kubeletDelayModule.SetChildren([]modules.DeliveryModule{imageModule})

	imageModule.SetParents([]modules.DeliveryModule{kubeletDelayModule})
	imageModule.SetChildren([]modules.DeliveryModule{containerCreateModule})

	containerCreateModule.SetParents([]modules.DeliveryModule{imageModule})
	containerCreateModule.SetChildren([]modules.DeliveryModule{containerStartReason})

	containerStartReason.SetParents([]modules.DeliveryModule{containerCreateModule})
	containerStartReason.SetChildren([]modules.DeliveryModule{containerPostHookReason})

	containerPostHookReason.SetParents([]modules.DeliveryModule{containerStartReason})
	containerPostHookReason.SetChildren([]modules.DeliveryModule{containerReadinessReason})

	containerReadinessReason.SetParents([]modules.DeliveryModule{containerPostHookReason})
	containerReadinessReason.SetChildren([]modules.DeliveryModule{podReadinessReason})

	podReadinessReason.SetParents([]modules.DeliveryModule{containerReadinessReason})
	return scheduleModule
}
