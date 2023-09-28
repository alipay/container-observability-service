package model

import (
	"time"

	corev1 "k8s.io/api/core/v1"
)

type NodeYaml struct {
	AuditID        string       `gorm:"column:audit_id"`
	SqlNode        string       `gorm:"column:node"`
	NodeName       string       `gorm:"column:node_name"`
	NodeIp         string       `gorm:"column:node_ip"`
	UID            string       `gorm:"column:uid"`
	ClusterName    string       `gorm:"column:cluster_name"`
	Node           *corev1.Node `gorm:"-"`
	StageTimeStamp time.Time    `gorm:"column:stage_timestamp"`
}
type NodeInfo struct {
	IP                string
	NodeSn            string
	Cluster           string
	Zone              string
	AppEnv            string
	NodeName          string
	CreationTimestamp time.Time
	NodeStatus        string
	ScheduleStatus    string
	RemedyStatus      string
	Resource          string
	TotalCount        int
	ReadyAt           time.Time
	OtherInfo         interface{}
}

func (s *NodeYaml) TableName() string {
	return "node_yaml"
}

// 当前 es 和 ob 版本的表名还不一致, 过度阶段增加一个EsTableName()返回 Es的表名
// 后续表名统一后, 删除 EsTableName() 仅使用 TableName()
func (s *NodeYaml) EsTableName() string {
	return "node_yaml"
}

func (s *NodeYaml) TypeName() string {
	return "_doc"
}
