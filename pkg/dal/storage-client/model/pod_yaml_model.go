package model

import (
	"time"

	v1 "k8s.io/api/core/v1"
)

// 对外数据模型
type PodYaml struct {
	AuditID           string    `json:"auditID,omitempty" gorm:"column:audit_id"`
	ClusterName       string    `json:"clusterName,omitempty" gorm:"column:cluster_name"`
	Namespace         string    `json:"namespace,omitempty" gorm:"column:namespace"`
	SLO               string    `json:"slo,omitempty" gorm:"-"`
	Status            string    `json:"Status,omitempty" gorm:"-"`
	HostIP            string    `json:"hostIP,omitempty" gorm:"column:host_ip"`
	PodIP             string    `json:"podIP,omitempty" gorm:"column:pod_ip"`
	Hostname          string    `json:"hostname,omitempty" gorm:"column:hostname"`
	IsDeleted         string    `json:"isDeleted,omitempty" gorm:"column:is_deleted"`
	Pod               *v1.Pod   `json:"pod,omitempty" gorm:"-"`
	SqlPod            string    `json:"sqlpod,omitempty" gorm:"column:pod"`
	PodName           string    `json:"podName,omitempty" gorm:"column:pod_name"`
	PodUid            string    `json:"podUid,omitempty" gorm:"column:pod_uid"`
	CreationTimestamp time.Time `json:"creationTimestamp,omitempty" gorm:"column:stage_timestamp"`
	StageTimestamp    time.Time `json:"stageTimestamp,omitempty" gorm:"-"`
	SLOResult         string    `json:"sloResult,omitempty" gorm:"-"`
	SLOType           string    `json:"sloType,omitempty" gorm:"-"`
	IsBeginDelete     string    `json:"isBeginDelete,omitempty" gorm:"-"`
	Fqdn              string    `json:"fqdn,omitempty" gorm:"column:fqdn"`
	DebugUrl          string    `json:"DebugUrl,omitempty" gorm:"-"`
	SqlImages         string    `json:"sqlimages,omitempty" gorm:"column:images"`
}

func (s *PodYaml) TableName() string {
	return "pod_yaml"
}

// 当前 es 和 ob 版本的表名还不一致, 过度阶段增加一个EsTableName()返回 Es的表名
// 后续表名统一后, 删除 EsTableName() 仅使用 TableName()
func (s *PodYaml) EsTableName() string {
	return "pod_yaml"
}

func (s *PodYaml) TypeName() string {
	return "_doc"
}
