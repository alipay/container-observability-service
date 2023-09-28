package model

import "time"

type LifePhase struct {
	// stage 代表 Pod 的 trace 链路中的一个大类
	// OperationName 和 Reason 可以用来决定 stage
	DocID         string      `json:"omitempty" gorm:"column:doc_id"`
	PlfID         string      `json:"plf_id"`
	TraceStage    string      `json:"traceStage"`
	DataSourceId  string      `json:"dataSourceId" gorm:"column:data_source_id"`
	ClusterName   string      `json:"clusterName" gorm:"column:cluster_name"`
	Namespace     string      `json:"namespace" gorm:"column:namespace"`
	PodName       string      `json:"podName" gorm:"column:pod_name"`
	PodUID        string      `json:"podUID" gorm:"column:pod_uid"`
	OperationName string      `json:"operationName" gorm:"column:operation_name"`
	HasErr        bool        `json:"hasErr" gorm:"column:has_err"`
	SqlHasErr     string      `gorm:"column:has_err"`
	StartTime     time.Time   `json:"startTime" gorm:"column:start_time"`
	EndTime       time.Time   `gorm:"column:end_time"`
	ExtraInfo     interface{} `json:"extraInfo,omitempty" gorm:"-"`
	Reason        interface{} `json:"reason,omitempty" gorm:"-"`
	Message       interface{} `json:"message,omitempty" gorm:"-"`
	Info          string      `gorm:"column:extra_info"`
}

func (s *LifePhase) TableName() string {
	return "pod_phase"
}

// 当前 es 和 ob 版本的表名还不一致, 过度阶段增加一个EsTableName()返回 Es的表名
// 后续表名统一后, 删除 EsTableName() 仅使用 TableName()
func (s *LifePhase) EsTableName() string {
	return "pod_life_phase"
}

func (s *LifePhase) TypeName() string {
	return "_doc"
}

type NodeLifePhase struct {
	DocID         string `gorm:"column:doc_id"`
	PlfID         string
	TraceStage    string      `json:"traceStage"`
	DataSourceId  string      `json:"dataSourceId" gorm:"column:data_source_id"`
	ClusterName   string      `json:"clusterName" gorm:"column:cluster_name"`
	NodeName      string      `json:"nodeName" gorm:"column:node_name"`
	UID           string      `json:"uid" gorm:"column:node_uid"`
	OperationName string      `json:"operationName" gorm:"column:operation_name"`
	HasErr        bool        `json:"hasErr" gorm:"column:has_err"`
	StartTime     time.Time   `json:"startTime" gorm:"column:start_time"`
	EndTime       time.Time   `gorm:"column:end_time"`
	ExtraInfo     interface{} `json:"extraInfo,omitempty" gorm:"-"`
	Reason        interface{} `json:"reason,omitempty" gorm:"-"`
	Message       interface{} `json:"message,omitempty" gorm:"-"`
	SqlExtraInfo  string      `gorm:"column:extra_info"`
}

func (s *NodeLifePhase) TableName() string {
	return "node_phase"
}
func (s *NodeLifePhase) EsTableName() string {
	return "node_life_phase"
}

func (s *NodeLifePhase) TypeName() string {
	return "_doc"
}
