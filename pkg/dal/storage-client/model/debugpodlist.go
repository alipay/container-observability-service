package model

type DebugPodListTable struct {
	Podname    string `json:"podname,omitempty"`
	PodIP      string `json:"podip,omitempty"`
	Cluster    string `json:"cluster,omitempty"`
	PodUID     string `json:"poduid,omitempty"`
	NodeIP     string `json:"nodeip,omitempty"`
	CreateTime string `json:"createTime,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	State      string `json:"state,omitempty"`
	PodPhase   string `json:"podphase,omitempty"`
	Site       string `json:"site,omitempty"`
	DebugUrl   string `json:"debugurl,omitempty"`
	Dignosis   string `json:"诊断,omitempty"`
}
