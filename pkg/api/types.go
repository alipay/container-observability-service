package api

import (
	"net/url"
	"time"

	"github.com/alipay/container-observability-service/pkg/utils"
	"k8s.io/klog/v2"
)

// PerfRequest is request query parameters
type PerfRequest struct {
	// 查询开始时间时间戳（trace开始时间）
	StartTimestamp time.Time `json:"startTimestamp,omitempty"`
	// 查询结束时间时间戳
	FinishTimestamp time.Time `json:"finishTimestamp,omitempty"`
	Cluster         string    `json:"cluster,omitempty"`
	Namespace       string    `json:"namespace,omitempty"`
}

// TraceRequest is request query parameters
type TraceRequest struct {
	// create/delete/start/stop/upgrade/update
	TraceType string `json:"traceType,omitempty"`
	// 查询开始时间时间戳（trace开始时间）
	StartTimestamp time.Time `json:"startTimestamp,omitempty"`
	// 查询结束时间时间戳
	FinishTimestamp time.Time `json:"finishTimestamp,omitempty"`
	Hostname        string    `json:"hostname,omitempty"`
	RequestID       string    `json:"requestID,omitempty"`
	// Pod's meta data
	UID       string `json:"uid,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Error     bool   `json:"error,omitempty"`
	HasError  string `json:"haserr,omitempty"`
}

// Response is response for query result
type TraceResponse []TraceItem

// TraceItem is one trace item for a pod creation/deletion
type TraceItem struct {
	UID       string `json:"uid,omitempty"`
	Cluster   string `json:"cluster,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	TraceType string `json:"traceType,omitempty"`
	// true 表示失败的trace
	Error bool `json:"error,omitempty"`
	// 失败原因
	FailedReason string `json:"failedReason,omitempty"`
	// 所属模块，network/storage/image/...
	Module          string    `json:"module,omitempty"`
	FailedMessage   string    `json:"failedMessage,omitempty"`
	StartTimestamp  time.Time `json:"startTimestamp,omitempty"`
	FinishTimestamp time.Time `json:"finishTimestamp,omitempty"`
	HostIP          string    `json:"hostIP,omitempty"`
	NodeName        string    `json:"nodeName,omitempty"`
	// 管理组件，直接创建的Pod为空
	OwnerType string `json:"ownerType,omitempty"`
	// jaeger里的trace id，可以直接拼接到jaeger ui的链接后面打开jaeger
	JaegerTraceID string `json:"jaegerTraceID,omitempty"`
	JaegerURL     string `json:"jaegerURL,omitempty"`
	// 是否为job类型的pod
	IsJob bool `json:"isJob,omitempty"`
	// 每个span的概要。
	Spans []SpanItem `json:"spans,omitempty"`
}

// SpanItem is one span in a trace
type SpanItem struct {
	// 服务名
	Service string `json:"service,omitempty"`
	// 处理内容
	Operation       string    `json:"operation,omitempty"`
	StartTimestamp  time.Time `json:"startTimestamp,omitempty"`
	FinishTimestamp time.Time `json:"finishTimestamp,omitempty"`
	// 是否出错
	Error bool `json:"error,omitempty"`
	// 该span在总的trace中耗时的百分比
	DurationPercent float64 `json:"durationPercent,omitempty"`
}

func setSP(values url.Values, name string, f *string) {
	s := values.Get(name)
	if s != "" {
		*f = s
	}
}

func setBP(values url.Values, name string, f *bool) {
	s := values.Get(name)
	if s == "t" || s == "true" || s == "1" {
		*f = true
	}
}

func setTP(values url.Values, name string, f *time.Time) {
	s := values.Get(name)
	if s != "" {
		t, err := utils.ParseTime(s)
		if err != nil {
			klog.Errorf("failed to parse time %s for %s: %s", s, name, err.Error())
		} else {
			*f = t
		}
	}
}

func setTPLayout(values url.Values, name string, f *time.Time) {
	s := values.Get(name)
	layOut := "2006-01-02T15:04:05"
	if s != "" {
		t, err := time.ParseInLocation(layOut, s, time.Local)
		if err != nil {
			klog.Errorf("failed to parse time %s for %s: %s", s, name, err.Error())
		} else {
			*f = t
		}
	}
}
