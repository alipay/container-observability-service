package model

import "time"

// 对外数据模型
type Span struct {
	DocId             string    `gorm:"column:doc_id"`
	Name              string    `json:"name,omitempty" gorm:"column:or_name"`
	OrResource        string    `gorm:"column:or_resource"`
	OrNamespace       string    `gorm:"column:or_namespace"`
	OrUID             string    `gorm:"column:or_uid"`
	OrAPIGroup        string    `gorm:"column:or_apigroup"`
	OrAPIVersion      string    `gorm:"column:or_apiversion"`
	OrResourceVersion string    `gorm:"column:or_resource_version"`
	OrSubresource     string    `gorm:"column:or_subresource"`
	SpanName          string    `gorm:"column:span_name" `
	Type              string    `json:"type,omitempty" gorm:"column:span_type"`
	Begin             time.Time `json:"begin,omitempty" gorm:"column:span_begin"`
	End               time.Time `json:"end,omitempty" gorm:"column:span_end"`
	Elapsed           int64     `json:"elapsed,omitempty" gorm:"column:span_elapsed"`
	Cluster           string    `json:"cluster,omitempty" gorm:"column:span_cluster"`
	ActionType        string    `json:"actionType,omitempty" gorm:"column:span_action_type"`
	TimeStamp         time.Time `json:"timeStamp,omitempty" gorm:"column:span_timestamp"`
	Omitempty         bool      `json:"omitempty,omitempty" gorm:"column:span_omitempty"`
	SpanConfig        string    `gorm:"column:span_config"`
	Sqlroperties      string    `gorm:"column:properties"`
}

func (s *Span) TableName() string {
	return "span"
}

// 当前 es 和 ob 版本的表名还不一致, 过度阶段增加一个EsTableName()返回 Es的表名
// 后续表名统一后, 删除 EsTableName() 仅使用 TableName()
func (s *Span) EsTableName() string {
	return "spans_consuming"
}

func (s *Span) TypeName() string {
	return "_doc"
}
