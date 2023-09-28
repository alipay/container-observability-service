package service

import (
	"strconv"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
)

func ConvertClusterDistribute2Frame(result map[string]int) []model.ClusterTable {
	bit := make([]model.ClusterTable, 0)
	for res, count := range result {
		table := model.ClusterTable{
			Cluster: res,
			Number:  strconv.Itoa(count),
		}
		bit = append(bit, table)
	}

	return bit
}
func ConvertNameSpaceDistribute2Frame(result map[string]int) []model.NamespaceTable {
	bit := make([]model.NamespaceTable, 0)
	for res, count := range result {
		table := model.NamespaceTable{
			Namespace: res,
			Number:    strconv.Itoa(count),
		}
		bit = append(bit, table)
	}

	return bit
}
func ConvertNodeDistribute2Frame(result map[string]int) []model.NodeTable {
	bit := make([]model.NodeTable, 0)
	for res, count := range result {
		table := model.NodeTable{
			Node:   res,
			Number: strconv.Itoa(count),
		}
		bit = append(bit, table)
	}

	return bit
}
func ConvertSloDistribute2Frame(result map[string]int) []model.SloTable {
	bit := make([]model.SloTable, 0)
	for res, count := range result {
		table := model.SloTable{
			Slo:    res,
			Number: strconv.Itoa(count),
		}
		bit = append(bit, table)
	}

	return bit
}
func ConvertPodList2Frame(result []map[string]string) []model.PodListTable {
	bit := make([]model.PodListTable, 0)
	for _, res := range result {
		table := model.PodListTable{
			DeliveryTime:       res["CreatedTime"],
			Namespace:          res["Namespace"],
			Cluster:            res["Cluster"],
			PodUID:             res["PodUID"],
			PodName:            res["PodName"],
			NodeIP:             res["NodeIP"],
			SLO:                res["SLO"],
			Result:             res["DeliveryStatus"],
			DebugUrl:           res["PodDebugUrl"],
			SLOViolationReason: res["SLOViolationReason"],
		}
		bit = append(bit, table)
	}

	return bit
}
