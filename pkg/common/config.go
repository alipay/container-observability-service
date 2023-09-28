package common

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

func InitConfig(cfgFile string) (*DBOptions, error) {
	options := NewDefaultOptions()
	configData, err := os.ReadFile(cfgFile)
	if err != nil {
		klog.Infof("read cfgFile %s err:%s", cfgFile, err.Error())
		return nil, err
	}
	// unmarshal cfgFile bytes to options
	if err = yaml.Unmarshal(configData, options); err != nil {
		log.Printf("unmarshal cfgFile %s err:%s", cfgFile, err.Error())
		return nil, err
	}
	return options, nil
}
