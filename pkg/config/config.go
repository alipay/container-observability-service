package config

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// UserOnlineConfigMap:      用于指定一个特定的 namespace 的在线任务的超时时间，通常由 SRE 来配置
// UserAppConfigMap:         用于指定一个特定的 app 的超时时间，通常由 SRE 来配置
//
//	例如，"UserSLAConfigMap": {"linkeci":"10m"}则所有SLA<10m的将调整为10min
//
// IgnoredNamespaceForAudit: 用于指定 lunettes 可以忽略的 ns，通常由 测试开发 来配置
// PostStartHookTimeout:     用于指定 PostStartHookTimeout 超时时间
// ShouldIgnoreSinglePod:    用于指定 "资源交付SLO" 场景下，是否把一个单独的 Pod 给忽略掉
type LunettesConfig struct {
	UserOnlineConfigMap         map[string]string `json:"UserOnlineConfigMap,omitempty"`
	UserAppConfigMap            map[string]string `json:"UserAppConfigMap,omitempty"`
	UserSLAConfigMap            map[string]string `json:"UserSLAConfigMap,omitempty"`
	IgnoredNamespaceForAudit    []string          `json:"IgnoredNamespaceForAudit,omitempty"`
	PostStartHookTimeout        string            `json:"PostStartHookTimeout,omitempty"`
	ShouldIgnoreSinglePod       bool              `json:"ShouldIgnoreSinglePod,string,omitempty"`
	ShouldRetainOldMetrics      bool              `json:"ShouldRetainOldMetrics,string,omitempty"`
	IgnoreDeleteReasonNamespace []string          `json:"IgnoreDeleteReasonNamespace,omitempty"`
}

const (
	RetainOldMetrics      = false
	lunettesNs            = "lunettes"
	lunettesConfigMapName = "lunettes-config"
	kubeConfigPath        = "/etc/kubernetes/kubeconfig/admin.kubeconfig"
)

func GlobalLunettesConfig() *LunettesConfig {
	val := configValue.Load()
	if ptr, ok := val.(*LunettesConfig); ok && ptr != nil {
		return ptr
	}
	return &LunettesConfig{}
}

var (
	configValue          atomic.Value
	globalLunettesConfig LunettesConfig
)

func init() {
	configValue.Store(&globalLunettesConfig)
	stop := make(<-chan struct{})
	refreshLunettesConfigFromConfigMap(stop)
}

// 更新 lunettes configmap
func refreshLunettesConfigFromConfigMap(stop <-chan struct{}) {
	if EnableStressMode {
		return
	}

	var cfg *restclient.Config
	cfg, err := restclient.InClusterConfig()
	if err != nil {
		klog.Errorf("failed to build config, err is %v", err)
		return
	}

	cfg.UserAgent = "lunettes"
	cs, err := clientset.NewForConfig(cfg)
	if err != nil {
		klog.Errorf("failed to create clientset: %v", err)
		return
	}

	refreshConfigMap := func() {
		lunettesConfigMap, err := cs.CoreV1().ConfigMaps(lunettesNs).Get(context.TODO(), lunettesConfigMapName, metav1.GetOptions{})
		if err != nil {
			klog.Errorf("failed to get lunettes configmap: %v", err)
			return
		}

		lunettesStr := lunettesConfigMap.Data["lunettes-config"]
		if lunettesStr == "" {
			klog.Errorf("lunettes configmap data is empty")
			return
		}
		var tmpConfig LunettesConfig
		err = json.Unmarshal([]byte(lunettesStr), &tmpConfig)
		if err != nil {
			klog.Errorf("failed to unmarshal lunettes configmap: %v", err)
			return
		}

		klog.Infof("configmap is %v", tmpConfig)
		configValue.Store(&tmpConfig)
	}

	refreshConfigMap()
	go wait.JitterUntil(refreshConfigMap, 15*time.Second, 0.0, true, stop)
	return
}
