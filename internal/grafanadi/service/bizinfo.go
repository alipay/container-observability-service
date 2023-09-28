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
