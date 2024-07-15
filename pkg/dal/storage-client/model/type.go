package model

import "time"

type SloOptions struct {
	Result         string
	Cluster        string
	BizName        string
	Count          string
	Type           string    // create 或者 delete, 默认是 create
	DeliveryStatus string    // FAIL/KILL/ALL/SUCCESS
	SloTime        string    // 20s/30m0s/10m0s
	Env            string    // prod, test
	From           time.Time // range query
	To             time.Time // range query
}
type NodeParams struct {
	NodeUid  string
	NodeIp   string
	NodeName string
}
type PodParams struct {
	Name     string
	Uid      string
	Hostname string
	Podip    string
	From     time.Time
	To       time.Time
}
