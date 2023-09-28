package utils

import (
	corev1 "k8s.io/api/core/v1"
)

// 获取pod中所有pvc的资源名列表
func GetPersistentVolumeClaimNamesFromPod(pod *corev1.Pod) []string {
	result := make([]string, 0)

	for _, value := range pod.Spec.Volumes {
		if value.PersistentVolumeClaim != nil {
			result = append(result, value.PersistentVolumeClaim.ClaimName)
		}
	}

	return result
}

func InFinalizers(finalizers []string, k string) bool {
	if finalizers == nil {
		return false
	}

	for i := range finalizers {
		if finalizers[i] == k {
			return true
		}
	}

	return false
}

func InAnnotations(annotations map[string]string, k string) bool {
	v, found := annotations[k]
	return found && v != ""
}
