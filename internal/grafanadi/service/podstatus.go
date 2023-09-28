package service

import (
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertPodStatus2Frame(podInfo storagemodel.PodInfo, podYaml storagemodel.PodYaml) []model.PodStatusTable {
	upTime := time.Duration(0)
	if !podYaml.Pod.CreationTimestamp.IsZero() {
		upTime = time.Now().Sub(podYaml.Pod.CreationTimestamp.Time)
	}

	if !podYaml.StageTimestamp.IsZero() {
		upTime = podYaml.StageTimestamp.Sub(podYaml.Pod.CreationTimestamp.Time)
	}
	var state string
	if podYaml.IsBeginDelete == "true" {
		state = "发起删除"
	}
	if podYaml.IsDeleted == "true" {
		state = "已删除"
	}

	bit := model.PodStatusTable{
		PodPhase:      string(podYaml.Pod.Status.Phase),
		State:         state,
		CreateTime:    podYaml.Pod.CreationTimestamp.Format(time.RFC3339Nano),
		UpTime:        upTime.String(),
		LastTimeStamp: podYaml.StageTimestamp.Format(time.RFC3339Nano),
	}

	return []model.PodStatusTable{bit}
}
