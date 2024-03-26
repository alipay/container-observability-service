package common

import (
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

// SiteInfo stores the site information for a Kubernetes cluster.
// This is used for Lunettes' federatin API for querying pod lists across multiple clusters.
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

// InitFedConfig retrieves the content from cfgFile (in cmd line) to initiate SiteOptions
func InitFedConfig(cfgFile string) (*SiteOptions, error) {
	options := NewSiteOptions()
	configData, err := os.ReadFile(cfgFile)
	if err != nil {
		klog.Infof("read cfgFile %s err:%s", cfgFile, err.Error())
		return nil, err
	}
	// unmarshal cfgFile to SiteOptions
	if err = yaml.Unmarshal(configData, options); err != nil {
		klog.Infof("unmarshal cfgFile %s err:%s", cfgFile, err.Error())
		return nil, err
	}
	klog.Infof("fed options is: %+v\n", options)
	return options, nil
}
