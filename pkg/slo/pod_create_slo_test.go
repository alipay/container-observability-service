package slo

import (
	"fmt"
	"sync"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	k8s_audit "k8s.io/apiserver/pkg/apis/audit"
)

func TestPodStartupMilestones_IsComplete(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			if got := data.IsComplete(); got != tt.want {
				t.Errorf("IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodStartupMilestones_finish(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			fmt.Print(data)
		})
	}
}

func TestPodStartupMilestones_processEvent(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	type args struct {
		pe *PodEvent
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			fmt.Print(data)
		})
	}
}

func TestPodStartupMilestones_processTime(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	type args struct {
		t time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			fmt.Print(data)
		})
	}
}

func TestPodStartupMilestones_start(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			fmt.Print(data)
		})
	}
}

func TestPodStartupMilestones_updateLatencyMetrics(t *testing.T) {
	type fields struct {
		Cluster                       string
		InitImage                     string
		Namespace                     string
		PodName                       string
		PodUID                        string
		Type                          string
		NodeIP                        string
		DebugUrl                      string
		OwnerRefStr                   string
		SchedulingStrategy            string
		StartUpResultFromCreate       string
		StartUpResultFromSchedule     string
		IsJob                         bool
		Cores                         int64
		Finished                      bool
		Created                       time.Time
		CreatedTime                   time.Time
		FinishTime                    time.Time
		ActualFinishTimeAfterSchedule time.Time
		ActualFinishTimeAfterCreate   time.Time
		Scheduled                     time.Time
		PodInitializedTime            time.Time
		ContainersReady               time.Time
		RunningAt                     time.Time
		SucceedAt                     time.Time
		FailedAt                      time.Time
		ReadyAt                       time.Time
		DeletedTime                   time.Time
		InitStartTime                 time.Time
		ImageNameToPullTime           map[string]float64
		mutex                         *sync.RWMutex
		inputQueue                    chan *PodEvent
		closeCh                       chan struct{}
		notifyQueue                   chan string
		key                           string
		auditTimeQueue                chan *time.Time
		latestPod                     *v1.Pod
	}
	type args struct {
		milestone string
		end       time.Time
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PodStartupMilestones{
				Cluster:                       tt.fields.Cluster,
				InitImage:                     tt.fields.InitImage,
				Namespace:                     tt.fields.Namespace,
				PodName:                       tt.fields.PodName,
				PodUID:                        tt.fields.PodUID,
				Type:                          tt.fields.Type,
				NodeIP:                        tt.fields.NodeIP,
				DebugUrl:                      tt.fields.DebugUrl,
				OwnerRefStr:                   tt.fields.OwnerRefStr,
				SchedulingStrategy:            tt.fields.SchedulingStrategy,
				StartUpResultFromCreate:       tt.fields.StartUpResultFromCreate,
				StartUpResultFromSchedule:     tt.fields.StartUpResultFromSchedule,
				IsJob:                         tt.fields.IsJob,
				Cores:                         tt.fields.Cores,
				Finished:                      tt.fields.Finished,
				Created:                       tt.fields.Created,
				CreatedTime:                   tt.fields.CreatedTime,
				FinishTime:                    tt.fields.FinishTime,
				ActualFinishTimeAfterSchedule: tt.fields.ActualFinishTimeAfterSchedule,
				ActualFinishTimeAfterCreate:   tt.fields.ActualFinishTimeAfterCreate,
				Scheduled:                     tt.fields.Scheduled,
				PodInitializedTime:            tt.fields.PodInitializedTime,
				ContainersReady:               tt.fields.ContainersReady,
				RunningAt:                     tt.fields.RunningAt,
				SucceedAt:                     tt.fields.SucceedAt,
				FailedAt:                      tt.fields.FailedAt,
				ReadyAt:                       tt.fields.ReadyAt,
				DeletedTime:                   tt.fields.DeletedTime,
				InitStartTime:                 tt.fields.InitStartTime,
				ImageNameToPullTime:           tt.fields.ImageNameToPullTime,
				mutex:                         tt.fields.mutex,
				inputQueue:                    tt.fields.inputQueue,
				closeCh:                       tt.fields.closeCh,
				notifyQueue:                   tt.fields.notifyQueue,
				key:                           tt.fields.key,
				auditTimeQueue:                tt.fields.auditTimeQueue,
				latestPod:                     tt.fields.latestPod,
			}
			fmt.Print(data)
		})
	}
}

func Test_collectPodEvents(t *testing.T) {
	type args struct {
		auditEvent *k8s_audit.Event
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

func Test_genPodCreateAPIResultMetrics(t *testing.T) {
	type args struct {
		auditEvent *k8s_audit.Event
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

func Test_processPodCreateLog(t *testing.T) {
	type args struct {
		auditEvent *k8s_audit.Event
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

func Test_syncAuditTime(t *testing.T) {
	type args struct {
		auditEvent *k8s_audit.Event
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
