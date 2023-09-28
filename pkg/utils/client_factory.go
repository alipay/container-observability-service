package utils

import (
	"fmt"

	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog"
)

func GetClientFromIncluster(qps float32, burst int) (clientset.Interface, error) {
	cfg, err := restclient.InClusterConfig()
	if err != nil {
		klog.Errorf("failed to build config, err is %v", err)
		return nil, fmt.Errorf("failed to build config, err is %v", err)
	}

	cfg.UserAgent = "lunettes"
	kubeClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to create clientset: %v", err)
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}
	return kubeClient, nil
}
