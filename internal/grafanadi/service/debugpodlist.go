package service

import (
	"sort"
	"time"

	"github.com/alipay/container-observability-service/pkg/dal/storage-client/model"
)

func ConvertPodYamls2Table(podyamls []*model.PodYaml) []*model.DebugPodListTable {
	podyamlsSet := DeduplicateAndGetTopK(podyamls, 5)

	tables := make([]*model.DebugPodListTable, 0)
	for _, v := range podyamlsSet {
		var state string = "在线"
		if v.IsBeginDelete == "true" {
			state = "发起删除"
		}
		if v.IsDeleted == "true" {
			state = "已删除"
		}

		t := &model.DebugPodListTable{
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

func DeduplicateAndGetTopK(slice []*model.PodYaml, topK int) []*model.PodYaml {
	// 创建一个 map 用于记录已经出现过的元素
	set := make(map[string]*model.PodYaml)

	for _, item := range slice {
		if v, ok := set[string(item.Pod.UID)]; ok {
			// 如果该元素存在, 检查当前元素的时间是否更新, 保存最新的记录
			if v.StageTimestamp.Before(item.StageTimestamp) {
				set[string(item.Pod.UID)] = item
			}
		} else {
			// 如果该元素在 set 中不存在，则将其添加到 set 中，并将其添加到 deduplicated 切片中
			set[string(item.Pod.UID)] = item
		}
	}

	// 创建一个新的切片用于存放去重后的元素
	deduplicated := make([]*model.PodYaml, 0)
	for _, podyaml := range set {
		deduplicated = append(deduplicated, podyaml)
	}

	sort.Slice(deduplicated, func(i, j int) bool {
		// 按照时间降序
		return deduplicated[i].StageTimestamp.After(deduplicated[j].StageTimestamp)
	})

	// 返回前 topK 个元素
	if len(deduplicated) > topK {
		return deduplicated[:topK]
	}
	return deduplicated
}
