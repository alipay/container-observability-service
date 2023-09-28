package featuregates

import (
	"strings"

	"k8s.io/klog/v2"
)

var (
	fg = FeatureGates{}
)

type FeatureGates struct {
	gates    string
	features map[string]bool
}

func Parse(gates string) {
	fg.gates = gates
	fg.features = make(map[string]bool)
	ss := strings.Split(gates, ",")
	for _, s := range ss {
		fg.features[s] = true
	}
	klog.Infof("enabled feature: %s", gates)
}

func IsEnabled(f string) bool {
	return fg.features[f]
}
