package service

import (
	"sort"
	"time"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
	storagemodel "github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertPodyamls2Table(podyamls []*storagemodel.PodYaml) []*model.QueryPodListTable {
	podyamlsSet := DeduplicateAndTopK(podyamls, 5)

	tables := make([]*model.QueryPodListTable, 0)
	for _, v := range podyamlsSet {
		var state string = "running"
		if v.IsBeginDelete == "true" {
			state = "terminating"
		}
		if v.IsDeleted == "true" {
			state = "terminated"
		}

		t := &model.QueryPodListTable{
			Podname:    v.PodName,
			PodIP:      v.PodIP,
			Cluster:    v.ClusterName,
			PodUID:     v.PodUid,
			NodeIP:     v.Pod.Status.HostIP,
			CreateTime: v.Pod.CreationTimestamp.Format(time.RFC3339),
			Namespace:  v.Namespace,
			State:      state,
			PodPhase:   string(v.Pod.Status.Phase),
		}
		tables = append(tables, t)

	}

	return tables

}

func DeduplicateAndTopK(slice []*storagemodel.PodYaml, topK int) []*storagemodel.PodYaml {
	set := make(map[string]*storagemodel.PodYaml)

	for _, item := range slice {
		if v, ok := set[string(item.Pod.UID)]; ok {
			if v.StageTimestamp.Before(item.StageTimestamp) {
				set[string(item.Pod.UID)] = item
			}
		} else {
			set[string(item.Pod.UID)] = item
		}
	}

	deduplicated := make([]*storagemodel.PodYaml, 0)
	for _, podyaml := range set {
		deduplicated = append(deduplicated, podyaml)
	}

	sort.Slice(deduplicated, func(i, j int) bool {
		return deduplicated[i].StageTimestamp.After(deduplicated[j].StageTimestamp)
	})

	if len(deduplicated) > topK {
		return deduplicated[:topK]
	}
	return deduplicated
}
