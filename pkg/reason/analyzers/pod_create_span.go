package analyzers

import (
	"strings"
	"time"

	"github.com/alipay/container-observability-service/pkg/reason/share"
	"github.com/alipay/container-observability-service/pkg/reason/utils"
	"github.com/alipay/container-observability-service/pkg/shares"
	spanpkg "github.com/alipay/container-observability-service/pkg/spans"
	"k8s.io/klog"
)

type SpanType string

var (
	DEFAULT_SCHEDULE_SPAN SpanType = "default_schedule_span"
	//使用Schedule代替
	SCHEDULE_SPAN SpanType = "schedule_span"

	IP_ALLOCATE_SPAN      SpanType = "ip_allocate_span"
	KUBELET_DELAY_SPAN    SpanType = "kubelet_delay_span"
	VOLUE_MOUNT_SPAN      SpanType = "volume_mount_span"
	VOLUE_ATTACH_SPAN     SpanType = "volume_attach_span"
	SANDBOX_CREATE_SPAN   SpanType = "sandbox_create_span"
	IMAGE_PULL_SPAN       SpanType = "image_pull_span"
	CONTAINER_CREATE_SPAN SpanType = "container_create_span"
	CONTAINER_START_SPAN  SpanType = "container_start_span"
	POSTSTART_HOOK_SPAN   SpanType = "container_poststart_span"
	POD_INITIAL_SPAN      SpanType = "pod_init_span"
	CONTIANER_READY_SPAN  SpanType = "container_readiness_span"
	POD_READY_SPAN        SpanType = "pod_readiness_span"

	resultMap = map[SpanType]string{
		DEFAULT_SCHEDULE_SPAN: "DefaultScheduleTooMuchTime",
		KUBELET_DELAY_SPAN:    "KubeletDelayTooMuchTime",
		IP_ALLOCATE_SPAN:      "IpAllocateTooMuchTime",
		VOLUE_MOUNT_SPAN:      "VolumeMountTooMuchTime",
		VOLUE_ATTACH_SPAN:     "VolumeAttachTooMuchTime",
		SANDBOX_CREATE_SPAN:   "SandboxCreateTooMuchTime",
		IMAGE_PULL_SPAN:       "ImagePullTooMuchTime",
		CONTAINER_CREATE_SPAN: "ContainerCreateTooMuchTime",
		CONTAINER_START_SPAN:  "ContainerStartTooMuchTime",
		POSTSTART_HOOK_SPAN:   "PostStartHookTooMuchTime",
		POD_INITIAL_SPAN:      "PodInitialTooMuchTime",
		CONTIANER_READY_SPAN:  "ContainerReadyTooMuchTime",
		POD_READY_SPAN:        "PodReadyTooMuchTime",
	}
	moduleMap = map[SpanType]string{
		DEFAULT_SCHEDULE_SPAN: share.SCHEDULER,
		IP_ALLOCATE_SPAN:      share.NETWORK,
		VOLUE_MOUNT_SPAN:      share.VOLUME,
		VOLUE_ATTACH_SPAN:     share.VOLUME,
		SANDBOX_CREATE_SPAN:   share.RUNTIME,
		IMAGE_PULL_SPAN:       share.IMAGE,
		CONTAINER_CREATE_SPAN: share.CONTAINER_READINESS,
		CONTAINER_START_SPAN:  share.CONTAINER_START,
		POSTSTART_HOOK_SPAN:   share.CONTAINER_POST_START,
		POD_INITIAL_SPAN:      share.POD_INITIALIZE,
		CONTIANER_READY_SPAN:  share.CONTAINER_READINESS,
		POD_READY_SPAN:        share.POD_READINESS,
	}
)

func (s SpanType) getResultName() string {
	return resultMap[s]
}
func (s SpanType) getModuleName() string {
	return moduleMap[s]
}

func analysisMaxSpan(spans []*spanpkg.Span, events []*shares.AuditEvent, curTime *time.Time) string {
	if curTime == nil {
		return ""
	}

	pod := utils.GetPodYamlFromHyperEvents(events, curTime)

	var maxSpan *spanpkg.Span
	sandboxBegin := time.Time{}

	for _, span := range spans {
		if (strings.Contains(span.Type, "volume") || strings.Contains(span.Type, "ip_allocate")) && sandboxBegin.Before(span.End) {
			sandboxBegin = span.End
		}
	}

	for idx, _ := range spans {
		span := spans[idx]
		// custom span
		if span.GetConfig().SpanOwner != spanpkg.K8sOwner && span.GetConfig().SpanOwner != spanpkg.CustomOwner {
			continue
		}

		if strings.Contains(span.Type, "sandbox") && span.Begin.Before(sandboxBegin) {
			span.Begin = sandboxBegin
		}

		if !span.Begin.IsZero() && span.End.IsZero() {
			span.End = *curTime
		}

		//对于调整后的span进行耗时重置
		if !span.Begin.IsZero() && !span.End.IsZero() {
			span.SetElapsed(span.End.Sub(span.Begin))
		}

		//对于hostnetwork的pod不需要ip分配
		if pod != nil && pod.Spec.HostNetwork && strings.Contains(span.Type, "ip_allocate") {
			span.SetElapsed(0)
		}

		if maxSpan == nil || maxSpan.Elapsed < span.Elapsed {
			maxSpan = span
		}
	}

	if maxSpan != nil {
		if pod != nil && resultMap[SpanType(maxSpan.Type)] == "" {
			klog.V(8).Infof("pod name: %s, maxSpan: %s/%s, begin:%s, end:%s\n", pod.Name, maxSpan.Name, maxSpan.Type, maxSpan.Begin, maxSpan.End)
		}
		return resultMap[SpanType(maxSpan.Type)]
	}
	return ""
}
