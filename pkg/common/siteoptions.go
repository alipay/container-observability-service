package common

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

var GSiteOptions SiteOptions = SiteOptions{
	[]SiteInfo{
		{
			DashboardUrl: "http://xxx:3003",
			DiUrl:        "http:/xxx:1888",
			SiteName:     "mainsite",
		},
	},
}

type SiteInfo struct {
	DashboardUrl string `yaml:"dashboard-url"`
	DiUrl        string `yaml:"di-url"`
	SiteName     string `yaml:"site-name"`
}

type SiteOptions struct {
	SiteInfos []SiteInfo `yaml:"site-infos"`
}

func NewSiteOptions() *SiteOptions {
	return &SiteOptions{
		SiteInfos: make([]SiteInfo, 0),
	}
}

func InitFedConfig(cfgFile string) (*SiteOptions, error) {
	options := NewSiteOptions()
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
	log.Printf("fed options %+v\n", options)
	return options, nil
}
