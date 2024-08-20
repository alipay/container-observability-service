package model

type LunettesMeta struct {
	ClusterName    string      `json:"clusterName"`
	LastReadTime   interface{} `json:"LastReadTime"`
	LivePodLatency interface{} `json:"LivePodLatency"`
}

func (*LunettesMeta) TableName() string {
	return "lunettes_meta"
}
func (s *LunettesMeta) EsTableName() string {
	return "lunettes_meta"
}

func (s *LunettesMeta) TypeName() string {
	return "_doc"
}
