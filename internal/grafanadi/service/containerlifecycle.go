package service

import (
	"log"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

var (
	PodReadySpan   string = "pod_ready_span"
	PodUpgradeSpan string = "pod_upgrade_span"
	PodDeleteSpan  string = "pod_delete_span"

	PodCreate  string = "PodCreate"
	PodUpgrade string = "PodUpgrade"
	PodDelete  string = "PodDelete"
	PodRunning string = "PodRunning"
	None       string = "None"
)

func ConvertLifePhase2State(lifephases []*storagemodel.LifePhase, slolist []*storagemodel.SloTraceData) []model.StateData {

	stateDataSlice := make([]model.StateData, 0)
	log.Printf("lifephase len is %d\n", len(lifephases))
	for _, slo := range slolist {
		if !slo.CreatedTime.IsZero() && slo.Type == "create" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.CreatedTime.Format(time.RFC3339),
				Status: PodCreate,
			})

		}
		if !slo.ContainersReady.IsZero() && slo.Type == "create" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.ContainersReady.Format(time.RFC3339),
				Status: PodRunning,
			})
		}
		if !slo.RunningAt.IsZero() && slo.Type == "create" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.RunningAt.Format(time.RFC3339),
				Status: PodRunning,
			})
		}
		if !slo.ReadyAt.IsZero() && slo.Type == "create" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.ReadyAt.Format(time.RFC3339),
				Status: PodRunning,
			})
		}

		// 处理升级阶段
		if slo.Type == "pod_upgrade" {

			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.CreatedTime.Format(time.RFC3339),
				Status: PodUpgrade,
			})
			// 升级结束作为 running
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.UpgradeEndTime.Format(time.RFC3339),
				Status: PodRunning,
			})
		}

		if !slo.CreatedTime.IsZero() && slo.Type == "delete" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   slo.CreatedTime.Format(time.RFC3339),
				Status: PodRunning,
			})
		}
	}

	for _, v := range lifephases {
		if v.OperationName == "apidelete" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   v.StartTime.Format(time.RFC3339),
				Status: PodDelete,
			})
		}
		if v.OperationName == "apideleted" {
			stateDataSlice = append(stateDataSlice, model.StateData{
				Time:   v.StartTime.Format(time.RFC3339),
				Status: None,
			})
		}
	}

	return stateDataSlice
}
