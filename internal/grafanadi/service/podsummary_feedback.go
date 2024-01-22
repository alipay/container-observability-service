package service

import (
	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertSummaryFeedback2Frame(summaryFeedback storagemodel.PodSummaryFeedback) []model.PodSummaryFeedback {
	bit := model.PodSummaryFeedback{
		ClusterName: summaryFeedback.ClusterName,
		Namespace: summaryFeedback.Namespace,
		PodName: summaryFeedback.PodName,
		PodUID: summaryFeedback.PodUID,
		PodIP: summaryFeedback.PodIP,
		NodeName: summaryFeedback.NodeName,
		Feedback: summaryFeedback.Feedback,
		Score:   summaryFeedback.Score,
		Comment: summaryFeedback.Comment,
		Summary: summaryFeedback.Summary,
		CreateTime: summaryFeedback.CreateTime,
	}

	return []model.PodSummaryFeedback{bit}
}
