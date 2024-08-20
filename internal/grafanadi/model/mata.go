package model

type RecordTimeTable struct {
	Cluster      string `json:"cluster,omitempty"`
	TimeDuration int64  `json:"timeDuration,omitempty"`
	LastRecord   string `json:"lastRecord,omitempty"`
}
