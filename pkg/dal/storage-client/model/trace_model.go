package model

import "time"

// trace list meta model

type Action struct {
	Name      string `json:"name"`
	FrontName string `json:"frontName"`
	Count     int    `json:"count"`
}
type DeliveryInfo struct {
	Total       int       `json:"total"`
	ErrorCount  int       `json:"errorCount"`
	NormalCount int       `json:"normalCount"`
	Actions     []*Action `json:"actions"`
}
type Resource struct {
	Cluster   string `json:"cluster"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}
type TraceListMetaData struct {
	DeliveryInfo DeliveryInfo `json:"deliveryInfo"`
	Resource     Resource     `json:"resource"`
}

// trace detail meta model

type TraceInfo struct {
	SpanTotal      int     `json:"spanTotal"`
	ErrorSpanCount int     `json:"errorSpanCount"`
	ServiceTotal   int     `json:"serviceTotal"`
	TotalDuration  string  `json:"totalDuration"`
	MaxConsumeSpan string  `json:"maxConsumeSpan"`
	Action         *Action `json:"action"`
}

type TraceDetailMetaData struct {
	TraceInfo TraceInfo `json:"traceInfo"`
	Resource  Resource  `json:"resource"`
}
type TraceStats struct {
	Cluster        string    `json:"Cluster"`
	Namespace      string    `json:"Namespace"`
	Name           string    `json:"Name"`
	UID            string    `json:"UID"`
	TraceType      string    `json:"TraceType"`
	StartTimestamp time.Time `json:"StartTimestamp"`
	Error          bool      `json:"Error"`
	FailModule     string    `json:"FailModule"`
	FailReason     string    `json:"FailReason"`
	FailMessage    string    `json:"FailMessage"`
	Owner          string    `json:"接口人"`
	HostIP         string
}
