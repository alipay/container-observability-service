package model

type PodInfo struct {
	ClusterName   string `json:"ClusterName,omitempty"`
	Namespace     string `json:"Namespace,omitempty"`
	PodName       string `json:"PodName,omitempty"`
	PodUID        string `json:"PodUID,omitempty"`
	PodIP         string `json:"PodIP,omitempty"`
	NodeName      string `json:"NodeName,omitempty"`
	AppName       string `json:"AppName,omitempty"`
	NodeIP        string `json:"NodeIP,omitempty"`
	Zone          string `json:"Zone,omitempty"`
	PodPhase      string `json:"PodPhase,omitempty"`
	LastTimeStamp string `json:"LastActiveAt,omitempty"`
	CreateTime    string `json:"CreatedAt,omitempty"`
	State         string `json:"State,omitempty"`
	Site          string `json:"Site,omitempty"`

	BizName string `json:"Bizname,omitempty"`
}
type DebugPodRestResult struct {
	PodInfos    PodInfo
	DebugPodRes DebugPodResult
}
type DebugPodResult struct {
	DebugStage  string
	ResultCode  string
	Description string      `json:"Description,omitempty"`
	Summary     interface{} `json:"Summary,omitempty"`
	Component   interface{} `json:"Component,omitempty"`
	Detail      string      `json:"Detail,omitempty"`
	Action      string
	Contact     string
	Info        string
}
type ResetResult struct {
	Code    int
	Status  string
	Message string
	Data    interface{}
}
type SLOResult struct {
	Create  []Create  `json:"create"`
	Delete  []Delete  `json:"delete"`
	Upgrade []Upgrade `json:"upgrade"`
}
type Create struct {
	Result            string `json:"Result"`
	IsCustomFailed    bool   `json:"IsCustomFailed"`
	ResultDescription string `json:"ResultDescription"`
	NeedSreAlert      bool   `json:"NeedSreAlert"`
	NeedSloAlert      bool   `json:"NeedSloAlert"`
}
type Delete struct {
	Result            string `json:"Result"`
	IsCustomFailed    bool   `json:"IsCustomFailed"`
	ResultDescription string `json:"ResultDescription"`
	NeedSreAlert      bool   `json:"NeedSreAlert"`
	NeedSloAlert      bool   `json:"NeedSloAlert"`
}
type Upgrade struct {
	Result            string `json:"Result"`
	IsCustomFailed    bool   `json:"IsCustomFailed"`
	ResultDescription string `json:"ResultDescription"`
	NeedSreAlert      bool   `json:"NeedSreAlert"`
	NeedSloAlert      bool   `json:"NeedSloAlert"`
}
