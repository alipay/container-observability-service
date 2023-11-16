package service

import (
	"fmt"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConverPodCreate2Frame(sloTraces []*storagemodel.SloTraceData) []model.DeliveryPodCreateOrDeleteTable {
	bit := make([]model.DeliveryPodCreateOrDeleteTable, 0)
	if len(sloTraces) == 0 {
		return nil
	}
	for _, slo := range sloTraces {
		bit = []model.DeliveryPodCreateOrDeleteTable{
			{Key: "DeliveryEnv", Value: convertNil("")},
			{Key: "PodType", Value: convertNil(time.Duration(slo.PodSLO).String())},
			{Key: "SloHint", Value: convertNil(slo.SloHint)},
			{Key: "CreationResult", Value: convertNil(slo.SLOViolationReason)},
			{Key: "CreatedAt", Value: convertNil(slo.CreatedTime.Format(time.RFC3339Nano))},
			{Key: "ReadyAt", Value: convertNil(slo.ReadyAt.Format(time.RFC3339Nano))},
		}
	}

	return bit
}
func convertNil(value string) string {
	if len(value) == 0 {
		return "Nil"
	}

	return value
}
func ConvertPodDelete2Frame(sloTraces []*storagemodel.SloTraceData) []model.DeliveryPodCreateOrDeleteTable {

	for _, slo := range sloTraces {
		if slo.Type == "delete" {
			bit := []model.DeliveryPodCreateOrDeleteTable{
				{Key: "DeletedAt", Value: convertNil(slo.CreatedTime.Format(time.RFC3339Nano))},
				{Key: "DeletionSLO", Value: "10min"},
				{Key: "DeletionSloHint", Value: "Nil"},
				{Key: "DeletionResult", Value: convertNil(slo.DeleteResult)},
				{Key: "DeleteEndAt", Value: convertNil(slo.DeleteEndTime.Format(time.RFC3339Nano))},
			}
			return bit
		}
	}

	return []model.DeliveryPodCreateOrDeleteTable{}
}
func ConvertPodUpgrade2Frame(sloLists []*storagemodel.SloTraceData) []model.DeliveryPodUpgradeTable {
	var bit []model.DeliveryPodUpgradeTable
	if len(sloLists) == 0 {
		return nil
	}
	for i, slo := range sloLists {
		id := fmt.Sprintf("Upgrade-%d", i+1)
		if slo.Type == "pod_upgrade" {

			upRec := []model.DeliveryPodUpgradeTable{
				{Index: id, Key: "UpgradeSLO", Value: "9min"},
				{Index: id, Key: "UpgradeSloHint", Value: "Nil"},
				{Index: id, Key: "UpgradeResult", Value: convertNil(slo.UpgradeResult)},
				{Index: id, Key: "UpgradedAt", Value: convertNil(slo.CreatedTime.Format(time.RFC3339Nano))},
				{Index: id, Key: "UpgradeFinishAt", Value: convertNil(slo.UpgradeEndTime.Format(time.RFC3339Nano))},
			}
			bit = append(bit, upRec...)
		}
	}

	return bit
}
