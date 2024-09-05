package model

type QueryPodListTable struct {
	Podname      string        `json:"podname,omitempty"`
	PodIP        string        `json:"podip,omitempty"`
	Cluster      string        `json:"cluster,omitempty"`
	PodUID       string        `json:"poduid,omitempty"`
	NodeIP       string        `json:"nodeip,omitempty"`
	NodeName     string        `json:"nodename,omitempty"`
	CreateTime   string        `json:"createTime,omitempty"`
	Namespace    string        `json:"namespace,omitempty"`
	State        string        `json:"state,omitempty"`
	PodPhase     string        `json:"podphase,omitempty"`
	WorkloadInfo WorkloadTable `json:"workloadInfo,omitempty"`
}

type WorkloadTable struct {
	ClusterName string `json:"ClusterName,omitempty"`
	Namespace   string `json:"Namespace,omitempty"`
	Name        string `json:"Name,omitempty"`
	UID         string `json:"Uid,omitempty"`
	Kind        string `json:"Kind,omitempty"`
}
