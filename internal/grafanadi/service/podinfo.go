package service

import (
	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertPodInfo2Frame(podInfo storagemodel.PodInfo, podYamls storagemodel.PodYaml) []model.PodInfoTable {
	bit := model.PodInfoTable{
		PodName:  podInfo.PodName,
		PodUID:   podInfo.PodUID,
		PodIP:    podInfo.PodIP,
		PodYaml:  "PodYaml",
		NodeName: podYamls.Pod.Spec.NodeName,
		NodeIP:   podYamls.Pod.Status.HostIP,
		NodeYaml: "NodeYaml",
	}

	return []model.PodInfoTable{bit}
}
