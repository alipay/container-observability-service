package model

import "time"

type PodInfo struct {
	DocID            string    `json:"docid,omitempty" gorm:"column:doc_id"`
	ClusterName      string    `json:"ClusterName,omitempty" gorm:"column:cluster_name"`
	Namespace        string    `json:"Namespace,omitempty" gorm:"column:namespace"`
	PodName          string    `json:"PodName,omitempty" gorm:"column:pod_name"`
	PodUID           string    `json:"PodUID,omitempty" gorm:"column:pod_uid"`
	PodIP            string    `json:"PodIP,omitempty" gorm:"column:pod_ip"`
	NodeName         string    `json:"NodeName,omitempty" gorm:"-"`
	AppName          string    `json:"AppName,omitempty" gorm:"-"`
	NodeIP           string    `json:"NodeIP,omitempty" gorm:"-"`
	Zone             string    `json:"Zone,omitempty" gorm:"column:zone"`
	PodPhase         string    `json:"PodPhase,omitempty" gorm:"-"`
	LastTimeStamp    string    `json:"LastActiveAt,omitempty" gorm:"-"`
	CreateTime       string    `json:"CreatedAt,omitempty" gorm:"-"`
	State            string    `json:"State,omitempty" gorm:"-"`
	Resource         string    `json:"Resource,omitempty" gorm:"-"`
	SLO              string    `json:"SLO,omitempty" gorm:"-"`
	Quota            string    `json:"Quota,omitempty" gorm:"-"`
	IpPool           string    `json:"IPPool,omitempty" gorm:"-"`
	DeliveryStatus   string    `gorm:"column:deliver_status" json:"deliverStatus,omitempty"`
	StartTime        time.Time `gorm:"column:start_time"`
	CurrentTime      time.Time `gorm:"column:slo_current_time"`
	DeliveryProgress string    `gorm:"column:delivery_progress" json:"deliveryprogress,omitempty"`
	StageTimestamp   time.Time `gorm:"column:stage_timestamp"`
	AuditID          string    `gorm:"column:audit_id" json:"auditid,omitempty"`
	// 为了 SLO 前端
	BizSource   string `json:"BizSource,omitempty" gorm:"column:biz_source"`
	BizId       string `json:"BizId,omitempty" gorm:"column:biz_id"`
	WorkOrderId string `json:"WorkOrderId,omitempty" gorm:"column:work_order_id"`
	BizSize     string `json:"BizSize,omitempty" gorm:"-"`
}

func (s *PodInfo) TableName() string {
	return "slo_pod_info"
}
func (s *PodInfo) EsTableName() string {
	return "slo_pod_info"
}

func (s *PodInfo) TypeName() string {
	return "_doc"
}
