package kube

import (
	"github.com/alipay/container-observability-service/pkg/config"
	"github.com/alipay/container-observability-service/pkg/utils"
	clientset "k8s.io/client-go/kubernetes"
)

var (
	KubeClient clientset.Interface
)

func InitKube(kubeConfig string) {
	if config.EnableStressMode {
		return
	}

	var err error
	KubeClient, err = utils.GetClientFromIncluster(1024, 1024)

	if err != nil {
		panic(err)
	}
}
