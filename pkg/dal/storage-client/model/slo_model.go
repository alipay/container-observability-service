package model

import "time"

type SloTraceData struct {
	DocID                         string             `gorm:"column:doc_id" json:"omitempty"`
	Cluster                       string             `gorm:"column:cluster"`
	InitImage                     string             `gorm:"column:init_image"`
	Namespace                     string             `gorm:"column:namespace"`
	PodName                       string             `gorm:"column:pod_name"`
	PodUID                        string             `gorm:"column:pod_uid"`
	Type                          string             `gorm:"column:type"`
	NodeIP                        string             `gorm:"column:node_ip"`
	NodeName                      string             `gorm:"column:node_name"`
	DebugUrl                      string             `gorm:"column:debug_url"`
	AppName                       string             `gorm:"column:app_name"`
	OwnerRefStr                   string             `gorm:"column:owner_ref_str"`
	BizName                       string             `gorm:"column:biz_name"`
	BizId                         string             `gorm:"column:biz_id"`
	DeliveryWorkload              string             `gorm:"column:delivery_workload"`
	WorkOrderId                   string             `gorm:"column:work_order_id"`
	SchedulingStrategy            string             `gorm:"column:scheduling_strategy"`
	StartUpResultFromCreate       string             `gorm:"column:start_up_result_from_create"`
	StartUpResultFromSchedule     string             `gorm:"column:start_up_result_from_schedule"`
	Situation                     string             `gorm:"column:situation"`
	ResourceType                  string             `gorm:"column:resource_type"`
	IsJob                         bool               `gorm:"column:is_job"`
	IsRafs                        bool               `gorm:"column:is_rafs"`
	Cores                         int64              `gorm:"column:cores"`
	WrittenToZsearch              bool               `gorm:"column:written_to_zsearch"`
	Finished                      bool               `gorm:"column:finished"`
	StopInterEvents               bool               `gorm:"column:stop_inter_events"`
	Created                       time.Time          `gorm:"column:created"`
	CreatedTime                   time.Time          `gorm:"column:created_time"`
	FinishTime                    time.Time          `gorm:"column:finish_time"`
	ActualFinishTimeAfterSchedule time.Time          `gorm:"column:actual_finish_time_after_schedule"`
	ActualFinishTimeAfterCreate   time.Time          `gorm:"column:actual_finish_time_after_create"`
	Scheduled                     time.Time          `gorm:"column:scheduled"`
	PodInitializedTime            time.Time          `gorm:"column:pod_initialized_time"`
	ContainersReady               time.Time          `gorm:"column:container_ready"`
	RunningAt                     time.Time          `gorm:"column:running_at"`
	SucceedAt                     time.Time          `gorm:"column:succeed_at"`
	FailedAt                      time.Time          `gorm:"column:failed_at"`
	ReadyAt                       time.Time          `gorm:"column:ready_at"`
	DeletedTime                   time.Time          `gorm:"column:deleted_time"`
	InitStartTime                 time.Time          `gorm:"column:init_start_time"`
	ImageNameToPullTime           map[string]float64 `gorm:"-"`
	PossibleReason                *string            `gorm:"-"`
	PossibleReasonStr             string             `gorm:"column:possible_reason"`
	SLOViolationReason            string             `gorm:"column:slo_violation_reason"`
	PodSLO                        int64              `gorm:"column:pod_slo"`
	DeliverySLO                   int64              `gorm:"column:delivery_slo"`
	DeliverySLOAdjusted           bool               `gorm:"column:delivery_slo_adjusted"`
	DeliveryDuration              time.Duration      `gorm:"column:delivery_duration"`
	DeliveryStatus                string             `gorm:"column:delivery_status"`
	DeliveryStatusOrig            string             `gorm:"column:delivery_status_orig"`
	SloHint                       string             `gorm:"column:slo_hint"`
	TrigerAuditLog                string             `gorm:"column:triger_audit_log"`
	DeleteResult                  string             `gorm:"column:delete_result"`
	KubeletKillingHost            string             `gorm:"column:kubelet_killing_host"`
	DeleteEndTime                 time.Time          `gorm:"column:delete_end_time"`
	RMZappInfoTime                time.Time          `gorm:"column:rm_zappinfo_time"`
	RMFQDNTime                    time.Time          `gorm:"column:rm_fqdn_time"`
	RMCMDBTime                    time.Time          `gorm:"column:rm_cmdb_time"`
	RMCNIAllocatedTime            time.Time          `gorm:"column:rm_cni_allocated_time"`
	RMAliCloudCNITime             time.Time          `gorm:"column:rm_ali_cloud_cni_time"`
	KubeletKillingTime            time.Time          `gorm:"column:kubelet_killing_time"`
	RemainingFinalizer            []string           `gorm:"column:remaining_finalizers;serializer:json"`
	DeleteTimeoutTime             time.Time          `gorm:"column:delete_timeout_time"`
	UpgradeResult                 string             `gorm:"column:upgrade_result"`
	UpgradeContainerName          string             `gorm:"column:upgrade_container_name"`
	UpdateStatus                  string             `gorm:"column:update_status"`
	UpgradeEndTime                time.Time          `gorm:"column:upgrade_end_time"`
	UpgradeTimeoutTime            time.Time          `gorm:"column:upgrade_timeout_time"`
	PVCName                       string             `gorm:"column:pvc_name"`
	PVCUID                        string             `gorm:"column:pvc_uid"`
	CreateResult                  string             `gorm:"column:create_result"`
	TimeoutTime                   time.Time          `gorm:"column:timeout_time"`
	PullTimeoutImageName          string             `gorm:"-"`
	SLOResult                     []string           `json:"sloResult,omitempty" gorm:"-"`
	SLOType                       string             `json:"sloType,omitempty" gorm:"-"`
}
type Slodata struct {
	DocID                         string             `gorm:"column:doc_id" json:"omitempty"`
	Cluster                       string             `gorm:"column:cluster" json:"cluster,omitempty"`
	InitImage                     string             `gorm:"column:init_image" json:"initimage,omitempty"`
	Namespace                     string             `gorm:"column:namespace" json:"namespace,omitempty"`
	PodName                       string             `gorm:"column:pod_name" json:"podName,omitempty"`
	PodUID                        string             `gorm:"column:pod_uid" json:"podUid,omitempty"`
	Type                          string             `gorm:"column:type" json:"type,omitempty"`
	NodeIP                        string             `gorm:"column:node_ip" json:"nodeIp,omitempty"`
	NodeName                      string             `gorm:"column:node_name" json:"nodeName,omitempty"`
	DebugUrl                      string             `gorm:"column:debug_url" json:"debugUrl,omitempty"`
	AppName                       string             `gorm:"column:app_name" json:"appname,omitempty"`
	OwnerRefStr                   string             `gorm:"column:owner_ref_str" json:"ownerRef,omitempty"`
	BizName                       string             `gorm:"column:biz_name" json:"bizName,omitempty"`
	BizId                         string             `gorm:"column:biz_id" json:"bizId,omitempty"`
	DeliveryWorkload              string             `gorm:"column:delivery_workload" json:"deliveryWorkload,omitempty"`
	WorkOrderId                   string             `gorm:"column:work_order_id" json:"workorderid,omitempty"`
	SchedulingStrategy            string             `gorm:"column:scheduling_strategy" json:"scheduleStrategy,omitempty"`
	StartUpResultFromCreate       string             `gorm:"column:start_up_result_from_create" json:"startupresultfromcreate,omitempty"`
	StartUpResultFromSchedule     string             `gorm:"column:start_up_result_from_schedule" json:"startupresultfromschedule,omitempty"`
	Situation                     string             `gorm:"column:situation" json:"situation,omitempty"`
	ResourceType                  string             `gorm:"column:resource_type" json:"resourceType,omitempty"`
	IsJob                         bool               `gorm:"column:is_job" json:"isjob,omitempty"`
	IsRafs                        bool               `gorm:"column:is_rafs" json:"israfs,omitempty"`
	Cores                         int64              `gorm:"column:cores" json:"cores,omitempty"`
	WrittenToZsearch              bool               `gorm:"column:written_to_zsearch" json:"writtentozsearch,omitempty"`
	Finished                      bool               `gorm:"column:finished" json:"finished,omitempty"`
	StopInterEvents               bool               `gorm:"column:stop_inter_events" json:"stopinterevents,omitempty"`
	Created                       time.Time          `gorm:"column:created" json:"CreatedTime,omitempty"`
	CreatedTime                   time.Time          `gorm:"column:created_time" json:"createdtime,omitempty"`
	FinishTime                    time.Time          `gorm:"column:finish_time" json:"finishTime,omitempty"`
	ActualFinishTimeAfterSchedule time.Time          `gorm:"column:actual_finish_time_after_schedule" json:"actualfinishtimeafterschedule,omitempty"`
	ActualFinishTimeAfterCreate   time.Time          `gorm:"column:actual_finish_time_after_create" json:"actualfinishtimeaftercreate,omitempty"`
	Scheduled                     time.Time          `gorm:"column:scheduled" json:"scheduled,omitempty"`
	PodInitializedTime            time.Time          `gorm:"column:pod_initialized_time" json:"podinitializedtime,omitempty"`
	ContainersReady               time.Time          `gorm:"column:container_ready" json:"ContainersReadyTime,omitempty"`
	RunningAt                     time.Time          `gorm:"column:running_at" json:"runningAt,omitempty"`
	SucceedAt                     time.Time          `gorm:"column:succeed_at" json:"succeedAt,omitempty"`
	FailedAt                      time.Time          `gorm:"column:failed_at" json:"failedAt,omitempty"`
	ReadyAt                       time.Time          `gorm:"column:ready_at" json:"readyAt,omitempty"`
	DeletedTime                   time.Time          `gorm:"column:deleted_time" json:"deletedAt,omitempty"`
	InitStartTime                 time.Time          `gorm:"column:init_start_time" json:"initStartTime,omitempty"`
	ImageNameToPullTime           map[string]float64 `gorm:"-"`
	PossibleReason                string             `gorm:"-"`
	PossibleReasonStr             string             `gorm:"column:possible_reason" json:"possiblereason,omitempty"`
	SLOViolationReason            string             `gorm:"column:slo_violation_reason" json:"sloviolationreason,omitempty"`
	PodSLO                        int64              `gorm:"column:pod_slo" json:"podslo,omitempty"`
	DeliverySLO                   int64              `gorm:"column:delivery_slo" json:"deliveryslo,omitempty"`
	DeliverySLOAdjusted           bool               `gorm:"column:delivery_slo_adjusted" json:"deliverysloadjusted,omitempty"`
	DeliveryDuration              time.Duration      `gorm:"column:delivery_duration" json:"deliveryduration,omitempty"`
	DeliveryStatus                string             `gorm:"column:delivery_status" json:"deliverystatus,omitempty"`
	DeliveryStatusOrig            string             `gorm:"column:delivery_status_orig" json:"deliverystatusorig,omitempty"`
	SloHint                       string             `gorm:"column:slo_hint" json:"slohint,omitempty"`
	PullTimeoutImageName          string             `gorm:"-"`
	DeleteEndTime                 time.Time
	DeleteTimeoutTime             time.Time
	UpgradeEndTime                time.Time
	UpgradeTimeoutTime            time.Time
	UpgradeResult                 string
	DeleteResult                  string
}

func (s *SloTraceData) TableName() string {
	return "slo_trace_data_daily"
}

func (s *SloTraceData) EsTableName() string {
	return "slo_trace_data_daily"
}

func (s *SloTraceData) TypeName() string {
	return "data"
}
func (s *Slodata) TableName() string {
	return "slo_data"
}

func (s *Slodata) EsTableName() string {
	return "slo_data"
}

func (s *Slodata) TypeName() string {
	return "data"
}
