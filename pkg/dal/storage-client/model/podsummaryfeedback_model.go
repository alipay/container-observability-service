package model

import (
	"time"
)

type PodSummaryFeedback struct {
	ClusterName      string    `json:"ClusterName,omitempty" gorm:"column:cluster_name"`
	Namespace        string    `json:"Namespace,omitempty" gorm:"column:namespace"`
	PodName          string    `json:"PodName,omitempty" gorm:"column:pod_name"`
	PodUID           string    `json:"PodUID,omitempty" gorm:"column:pod_uid"`
	PodIP            string    `json:"PodIP,omitempty" gorm:"column:pod_ip"`
	NodeName         string    `json:"NodeName,omitempty" gorm:"-"`
	Feedback         string    `json:"Feedback"`
	Score            int	   `json:"Score"`
	Comment          string	   `json:"Comment"`
	Summary			 string    `json:"Summary"`
	CreateTime    	 time.Time `json:"CreateTime"`
}

func (s *PodSummaryFeedback) TableName() string {
	return "pod_summary_feedback"
}
func (s *PodSummaryFeedback) EsTableName() string {
	return "pod_summary_feedback"
}

func (s *PodSummaryFeedback) TypeName() string {
	return "_doc"
}