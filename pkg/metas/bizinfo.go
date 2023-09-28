package metas

import (
	"time"

	"github.com/alipay/container-observability-service/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

var NoPriority string = ""

func GetPriority(pod *v1.Pod) string {
	return NoPriority
}

// < 3 container, <= 8C16G, no user readiness-gate, no FlexVolume
func IsTypicalPodNew(pod *v1.Pod) string {
	if pod == nil {
		return "UNKNOWN"
	}

	cpu, mem := utils.CalculateCpuAndMem(pod)
	mem = mem / 1024.0
	if cpu > 32 || mem > 64 {
		return "NotTypicalResource"
	}

	if len(pod.Spec.Containers) > 10 {
		return "MoreThan10Containers"
	}

	if cpu > 8 || mem > 16 {
		return "NONTYPICAL2"
	}

	containerNum := len(pod.Spec.Containers)
	if containerNum > 3 || len(pod.Spec.Volumes) > 15 {
		return "NONTYPICAL1"
	}

	return "TYPICAL"
}

func GetPodSLOByDeliveryPath(pod *v1.Pod) (time.Duration, bool) {
	result := IsTypicalPodNew(pod)
	sloSpec := 0 * time.Second
	if result == "TYPICALPATH" {
		sloSpec = 90 * time.Second
	}

	if result == "NONTYPICAL1" {
		sloSpec = 600 * time.Second
	}

	if result == "NONTYPICAL2" {
		sloSpec = 1800 * time.Second
	}

	return sloSpec, false
}
