package kube

import (
	"context"

	"github.com/alipay/container-observability-service/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNodeSummary(name string) string {
	if config.EnableStressMode {
		return ""
	}

	node, err := KubeClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return ""
	}

	nodeStatus := ""
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady {
			nodeStatus = nodeStatus + string(v1.NodeReady) + ":" + string(cond.Status)
			nodeStatus = nodeStatus + ":" + cond.Message + " "
		} else if cond.Type == v1.NodeNetworkUnavailable {
			nodeStatus = nodeStatus + string(v1.NodeNetworkUnavailable) + ":" + string(cond.Status)
			nodeStatus = nodeStatus + ":" + cond.Message + " "
		}
	}
	return nodeStatus
}

func GetNodeInfo(name string) map[string]interface{} {
	result := make(map[string]interface{})
	if config.EnableStressMode {
		return result
	}

	node, err := KubeClient.CoreV1().Nodes().Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return result
	}

	result["Spec"] = node.Spec
	result["status.conditions"] = node.Status.Conditions
	result["status.nodeInfo"] = node.Status.NodeInfo

	return result
}
