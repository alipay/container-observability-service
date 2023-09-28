package model

type Event struct {
	Agent          Agent       `json:"agent"`
	Log            Log         `json:"log"`
	Annotations    Annotations `json:"annotations"`
	ResponseObject interface{} `json:"responseObject"`
}
type Agent struct {
	Hostname    string `json:"hostname"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	EphemeralID string `json:"ephemeral_id"`
	Version     string `json:"version"`
}
type Log struct {
	File   File `json:"file"`
	Offset int  `json:"offset"`
}
type File struct {
	Path string `json:"path"`
}
type Annotations struct {
	AuthorizationK8SIoDecision string `json:"authorization.k8s.io/decision"`
	Cluster                    string `json:"cluster"`
	AuthorizationK8SIoReason   string `json:"authorization.k8s.io/reason"`
}

type PodCreateOrDeleteOrUpgrade struct {
	Result  string    `json:"result"`
	SLO     string    `json:"slo"`
	TraceID string    `json:"traceID"`
	Life    Lifecycle `json:"Lifecycle,omitempty"`
}

type Lifecycle struct {
	FinishedTime     string `json:"finishedTime,omitempty"`
	RunningAt        string `json:"runningAt,omitempty"`
	SucceedAt        string `json:"succeedAt,omitempty"`
	FailedAt         string `json:"failedAt,omitempty"`
	ReadyAt          string `json:"readyAt,omitempty"`
	LastActiveAt     string `json:"lastActiveAt,omitempty"`
	CreatedAt        string `json:"createdAt,omitempty"`
	ScheduledAt      string `json:"scheduledAt,omitempty"`
	DeleteEndAt      string `json:"deleteEndAt,omitempty"`
	DeletedAt        string `json:"deletedAt,omitempty"`
	UpgradeAt        string `json:"upgradeAt,omitempty"`
	UpgradeFinshedAt string `json:"upgradeFinshedAt,omitempty"`
}
