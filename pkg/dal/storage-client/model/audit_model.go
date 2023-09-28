package model

import (
	"time"

	k8saudit "k8s.io/apiserver/pkg/apis/audit"
)

type Audit struct {
	AuditId           string         `json:"auditid" gorm:"audit_id"` // 主键
	SqlStageTimestamp time.Time      `gorm:"stage_timestamp"`         // stage_timestamp
	Cluster           string         `gorm:"cluster"`                 // cluster
	Namespace         string         `gorm:"namespace"`               // namespace
	Resource          string         `gorm:"resource"`                // resource
	Content           string         `gorm:"content"`                 // content
	AuditLog          k8saudit.Event `gorm:"-"`
}

func (*Audit) TableName() string {
	return "audit"
}
func (s *Audit) EsTableName() string {
	return "audit_*"
}

func (s *Audit) TypeName() string {
	return "_doc"
}
