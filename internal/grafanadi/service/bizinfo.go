package service

import (
	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertBizInfo2Frame(podInfos []*storagemodel.PodInfo) []model.BizInfoTable {
	if len(podInfos) < 1 {
		return nil
	}
	bit := model.BizInfoTable{
		ClusterName: podInfos[0].ClusterName,
		Namespace:   podInfos[0].Namespace,
	}

	return []model.BizInfoTable{bit}
}

func ConvertPodYaml2Frame(podYamls []*storagemodel.PodYaml) []model.PodYamlTable {
	if len(podYamls) < 1 {
		return nil
	}
	bit := model.PodYamlTable{
		AuditID:        podYamls[0].AuditID,
		ClusterName:    podYamls[0].ClusterName,
		HostIP:         podYamls[0].HostIP,
		IsBeginDelete:  podYamls[0].IsBeginDelete,
		IsDeleted:      podYamls[0].IsDeleted,
		Pod:            podYamls[0].Pod,
		PodIP:          podYamls[0].PodIP,
		PodName:        podYamls[0].PodName,
		PodUid:         podYamls[0].PodUid,
		StageTimestamp: podYamls[0].StageTimestamp,
	}

	return []model.PodYamlTable{bit}
}

func ConvertNodeYaml2Frame(nodeYamls []*storagemodel.NodeYaml) []model.NodeYamlTable {
	if len(nodeYamls) < 1 {
		return nil
	}
	bit := model.NodeYamlTable{
		AuditID:        nodeYamls[0].AuditID,
		ClusterName:    nodeYamls[0].ClusterName,
		NodeName:       nodeYamls[0].NodeName,
		NodeIp:         nodeYamls[0].NodeIp,
		UID:            nodeYamls[0].UID,
		Node:           nodeYamls[0].Node,
		StageTimeStamp: nodeYamls[0].StageTimeStamp,
	}

	return []model.NodeYamlTable{bit}
}
func ConvertPodphase2Frame(lifePhase []*storagemodel.LifePhase) []model.PhaseTable {
	if len(lifePhase) < 1 {
		return nil
	}
	bit := model.PhaseTable{
		EndTime:       lifePhase[0].EndTime,
		ClusterName:   lifePhase[0].ClusterName,
		DataSourceId:  lifePhase[0].DataSourceId,
		HasErr:        lifePhase[0].HasErr,
		Namespace:     lifePhase[0].Namespace,
		OperationName: lifePhase[0].OperationName,
		PodUID:        lifePhase[0].PodUID,
		PodName:       lifePhase[0].PodName,
		StartTime:     lifePhase[0].StartTime,
		ExtraInfo:     lifePhase[0].ExtraInfo,
		TraceStage:    lifePhase[0].TraceStage,
	}

	return []model.PhaseTable{bit}
}
