package data_access

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

type AuditInterface interface {
	QueryAuditWithAuditId(data interface{}, auditid string) error
	QueryEventPodsWithPodUid(data interface{}, uid string) error
	QueryEventNodeWithPodUid(data interface{}, uid string) error
	QueryEventWithTimeRange(data interface{}, from, to time.Time) error
}

type PodInterface interface {
	QueryPodYamlsWithPodUID(data interface{}, uid string) error
	QueryPodYamlsWithPodName(data interface{}, name string) error
	QueryPodYamlsWithHostName(data interface{}, hostname string) error
	QueryPodYamlsWithPodIp(data interface{}, ip string) error
	QueryPodYamlsWithNodeIP(data interface{}, ip string) error
	QueryPodListWithNodeip(data interface{}, nodeIp string, isDeleted bool) error
	QueryPodUIDListByHostname(data interface{}, hostname string) error
	QueryPodUIDListByPodIP(data interface{}, ip string) error
	QueryPodUIDListByPodName(data interface{}, podname string) error
	QueryPodYamlWithParams(data interface{}, opts *model.PodParams) error
}

type NodeInterface interface {
	QueryNodeYamlsWithNodeUid(data interface{}, uid string) error
	QueryNodeYamlsWithNodeName(data interface{}, name string) error
	QueryNodeYamlsWithNodeIP(data interface{}, ip string) error
	QueryNodeYamlWithParams(data interface{}, opts *model.NodeParams) error
	QueryNodeUIDListWithNodeIp(data interface{}, ip string) error
}

type NodePhaseInterface interface {
	QueryNodephaseWithNodeName(data interface{}, name string) error
	QueryNodephaseWithNodeUID(data interface{}, uid string) error
}

type SloTraceDataInterface interface {
	QuerySloTraceDataWithPodUID(data interface{}, uid string) error
	QueryDeleteSloWithResult(data interface{}, opts *model.SloOptions) error
	QueryUpgradeSloWithResult(data interface{}, opts *model.SloOptions) error
	QueryCreateSloWithResult(data interface{}, opts *model.SloOptions) error
}

type PodLifePhaseInterface interface {
	QueryLifePhaseWithPodUid(data interface{}, uid string) error
}

type SpanInterface interface {
	QuerySpanWithPodUid(data interface{}, uid string) error
}

type PodInfoInterface interface {
	QueryPodInfoWithPodUid(data interface{}, uid string) error
}

type PodSummaryFeedbackInterface interface {
	StorePodSummaryFeedbackWithPodUid(data interface{}, podSummaryFeedback model.PodSummaryFeedback) error
}

type StorageInterface interface {
	AuditInterface
	PodInterface
	NodeInterface
	NodePhaseInterface
	SloTraceDataInterface
	PodLifePhaseInterface
	SpanInterface
	PodInfoInterface
	PodSummaryFeedbackInterface
}
