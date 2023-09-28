package utils

import (
	"flag"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	EnableKubeClient = false
)

func init() {
	flag.BoolVar(&EnableKubeClient, "enable-kube-client", true, "if enable kube client")
}

var (
	KubeClient kubernetes.Interface
)

func InitKube(kubeConfigPath string) error {
	if !EnableKubeClient {
		return nil
	}
	var err error
	KubeClient, err = GetClientFromFile("", kubeConfigPath, 1024, 1024)
	return err
}

// GetClientFromFile create a k8s client from config file
func GetClientFromFile(masterUrl, kubeConfigFile string, qps float32, burst int) (kubernetes.Interface, error) {
	clientConfig, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigFile)
	if err != nil {
		return nil, err
	}

	clientConfig.QPS = qps
	clientConfig.Burst = burst
	clientConfig.UserAgent = "lunettes"
	kubeClient, err := kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return nil, err
	}

	return kubeClient, nil
}
