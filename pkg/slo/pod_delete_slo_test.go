package slo

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/alipay/container-observability-service/pkg/metas"
	"github.com/alipay/container-observability-service/pkg/metrics"
	"github.com/alipay/container-observability-service/pkg/shares"
	"github.com/alipay/container-observability-service/pkg/xsearch"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	k8saudit "k8s.io/apiserver/pkg/apis/audit"
)

func generatePodDeleteMileStone(podname string, finalizers []string, tp string) *xsearch.PodDeleteMileStone {
	return &xsearch.PodDeleteMileStone{
		Cluster:             "eu95",
		Namespace:           "test",
		PodName:             podname,
		PodUID:              fmt.Sprintf("poduid-%d", rand.Uint32()),
		Type:                tp,
		CreatedTime:         time.Time{},
		RemainingFinalizers: finalizers,
		DeleteTimeoutTime:   time.Time{},
		Key:                 "",
		Mutex:               sync.Mutex{},
	}
}

func Test_checkTimeout(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "delete_success",
		},
		{
			name: "stale_success",
		},
		{
			name: "termi_success",
		},
		{
			name: "finalizer_test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

type DelSLOResult struct {
	lock         sync.Mutex
	DeleteResult map[string]int64 `json:"deleteResult,omitempty"` // key is reason, value is count
	StaleResult  map[string]int64 `json:"staleResult,omitempty"`
	TermiResult  map[string]int64 `json:"termiResult,omitempty"`
}

func (r *DelSLOResult) EqualTo(exp *DelSLOResult) bool {
	r.lock.Lock()
	exp.lock.Lock()
	defer r.lock.Unlock()
	defer exp.lock.Unlock()
	for key, val := range r.DeleteResult {
		if v, ok := exp.DeleteResult[key]; !ok || v != val {
			return false
		}
	}
	for key, val := range r.StaleResult {
		if v, ok := exp.StaleResult[key]; !ok || v != val {
			return false
		}
	}
	for key, val := range r.TermiResult {
		if v, ok := exp.TermiResult[key]; !ok || v != val {
			return false
		}
	}

	for key, val := range exp.DeleteResult {
		if v, ok := r.DeleteResult[key]; !ok || v != val {
			return false
		}
	}
	for key, val := range exp.StaleResult {
		if v, ok := r.StaleResult[key]; !ok || v != val {
			return false
		}
	}
	for key, val := range exp.TermiResult {
		if v, ok := r.TermiResult[key]; !ok || v != val {
			return false
		}
	}
	return true
}

func Test_doDeleteSLO(t *testing.T) {}

func Test_finishMileStoneWithResult(t *testing.T) {
	type args struct {
		podKey      string
		result      string
		currentTime time.Time
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_processDeleteOp(t *testing.T) {
	type args struct {
		auditEvent *k8saudit.Event
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_processEvents(t *testing.T) {
	type args struct {
		auditEvent *k8saudit.Event
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_processPatchOp(t *testing.T) {
	type args struct {
		auditEvent *k8saudit.Event
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_PodDeleteLatencyQuantiles(t *testing.T) {

	metrics.PodDeleteLatencyQuantiles.Reset()
	var err error
	var q50 = 0.5
	var q90 = 0.9
	var q99 = 0.99
	var j50 = 1.0
	var j90 = 2.0
	var j99 = 3.0

	var a50 = 5.0
	var a90 = 6.0
	var a99 = 7.0

	type latencyRecord struct {
		val   float64 // latency value
		count int     // count of records
	}
	testcases := []struct {
		name          string
		latencies     map[string][]latencyRecord // key is label, value is latencyRecords
		expectedQuans map[string][]*dto.Quantile // key is label, value is expected quantile values
	}{
		{
			name: "only one label",
			latencies: map[string][]latencyRecord{
				"job": []latencyRecord{
					{1.0, 550},
					{2.0, 360},
					{3.0, 81},
					{4.0, 9},
				},
				"app": []latencyRecord{
					{5.0, 550},
					{6.0, 360},
					{7.0, 81},
					{8.0, 9},
				},
			},
			expectedQuans: map[string][]*dto.Quantile{
				"job": []*dto.Quantile{
					&dto.Quantile{
						Quantile: &q50,
						Value:    &j50,
					},
					&dto.Quantile{
						Quantile: &q90,
						Value:    &j90,
					},
					&dto.Quantile{
						Quantile: &q99,
						Value:    &j99,
					},
				},
				"app": []*dto.Quantile{
					&dto.Quantile{
						Quantile: &q50,
						Value:    &a50,
					},
					&dto.Quantile{
						Quantile: &q90,
						Value:    &a90,
					},
					&dto.Quantile{
						Quantile: &q99,
						Value:    &a99,
					},
				},
			},
		},
	}

	for _, tc := range testcases {

		for key, val := range tc.latencies {
			for _, rec := range val {
				for i := 0; i < rec.count; i++ {
					metrics.PodDeleteLatencyQuantiles.WithLabelValues(key).Observe(rec.val)
				}
			}
		}

		metricCh := make(chan prometheus.Metric, len(tc.latencies))
		go func() {
			metrics.PodDeleteLatencyQuantiles.Collect(metricCh)
			close(metricCh)
		}()

		for metric := range metricCh {
			actualMetric := &dto.Metric{}
			err = metric.Write(actualMetric)
			if err != nil {
				t.Errorf("error while write to client_metrics %+v", err)
			}
			const labelname = "pod_type"
			var labelValue = ""
			for _, item := range actualMetric.GetLabel() {
				if *item.Name == labelname {
					labelValue = *item.Value
				}
			}
			quans := tc.expectedQuans[labelValue]
			for _, q := range quans {
				for _, aq := range actualMetric.Summary.Quantile {
					if *q.Quantile == *aq.Quantile {
						if *q.Value != *aq.Value {
							t.Errorf("actual value is %f while expected is %f", *aq.Value, *q.Value)
						}
					}
				}
			}

		}
	}

}

func GenerateData(event *k8saudit.Event) *shares.AuditEvent {
	shareEvent := &shares.AuditEvent{Event: event}

	if obj := metas.GetObjectFromRuntimeUnknown(event.ResponseObject, nil); obj != nil {
		shareEvent.ResponseRuntimeObj = obj
	}

	if obj := metas.GetObjectFromRuntimeUnknown(event.RequestObject, nil); obj != nil {
		shareEvent.RequestRuntimeObj = obj
	}

	return shareEvent
}
