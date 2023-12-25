package service

import (
    "github.com/alipay/container-observability-service/internal/grafanadi/model"
)

func ConvertPodSummary2Frame(podSummary model.PodSummary) []model.PodSummary {
    bit := model.PodSummary{
		DebugStage: podSummary.DebugStage,
		ResultCode: podSummary.ResultCode,
		Component: podSummary.Component,
		Summary: podSummary.Summary,
    }

    return []model.PodSummary{bit}
}