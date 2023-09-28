package slo

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
)

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}

func Test_calculateImagePullTime(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name string
		args args
		want map[string]float64
	}{
		{
			name: "test1",
			args: args{[]*v1.Event{
				{
					Reason:         "Pulled",
					Message:        `Container image "reg.docker.com/swiftimage/basementtask-assets:21fe4e20200320223124636" already present on machine`,
					FirstTimestamp: metav1.NewTime(parseTime("2020-03-31T11:12:09.126788+08:00")),
				},
			}},
			want: map[string]float64{
				"reg.docker.com/swiftimage/basementtask-assets:21fe4e20200320223124636": 0,
			},
		},
		{
			name: "test1",
			args: args{[]*v1.Event{
				{
					Message:        `Container image "reg.docker.com/aci/jenkins-slave-jnlp:2019-09-24" already present on machine`,
					Reason:         "Pulled",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-24T13:37:54.888178+08:00")),
				},
				{
					Message:        `Pulling image "reg.docker.com/alps/alps-ci:alios7u2_python3_tf1.13_alps-20200721165447"`,
					Reason:         "Pulling",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-24T13:37:58.138593+08:00")),
				},
				{
					Message:        `Successfully pulled image "reg.docker.com/alps/alps-ci:alios7u2_python3_tf1.13_alps-20200721165447"`,
					Reason:         "Pulled",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-24T13:39:38.94661+08:00")),
				},
			}},
			want: map[string]float64{
				"reg.docker.com/aci/jenkins-slave-jnlp:2019-09-24":                        0,
				"reg.docker.com/alps/alps-ci:alios7u2_python3_tf1.13_alps-20200721165447": 100.808017,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateImagePullTime(tt.args.events, nil); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateImagePullTime() = %v, want %v", got, tt.want)
			}
			times := calculateImagePullTime(tt.args.events, nil)
			isJob := false
			for _, value := range times {
				if isJob && value > 30 || !isJob && value > 60 {
					fmt.Println("aaaaaaaaa")
				} else {
					fmt.Println("bbbbbbbbb")
				}
			}
		})
	}
}

func Test_getScheduleStatus(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 time.Time
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getScheduleStatus(tt.args.pod)
			if got != tt.want {
				t.Errorf("getScheduleStatus() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("getScheduleStatus() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_isKubeletDelay(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := isKubeletDelay(tt.args.events); got != tt.want {
				t.Errorf("isKubeletDelay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isPreemption(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{events: []*v1.Event{
				{
					Reason: "PreemptionSuccess",
				},
				{
					Reason: "ToDoPreemption",
				},
			}},
			want: true,
		},
		{
			name: "test2",
			args: args{[]*v1.Event{
				{
					Reason: "xxxx",
				},
				{
					Reason: "yyyyy",
				},
			}},
			want: false,
		},
		{
			name: "test3",
			args: args{[]*v1.Event{
				{
					Reason: "PreemptionSuccess",
				},
				{
					Reason: "xxxxx",
				},
			}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPreemption(tt.args.events); got != tt.want {
				t.Errorf("isPreemption() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodStartupMilestones_analyzeFailedReason(t *testing.T) {
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
		want   string
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
			if got := data.analyzeFailedReason(); got != tt.want {
				t.Errorf("analyzeFailedReason() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_analyzeFailReason(t *testing.T) {
	type args struct {
		podyaml           *v1.Pod
		eventsNormalOrder []*v1.Event
		cTime             time.Time
		isJob             bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "没有event事件，调度失败",
			args: args{
				podyaml: &v1.Pod{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       v1.PodSpec{},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodScheduled,
								Status: v1.ConditionFalse,
								Reason: v1.PodReasonUnschedulable,
							},
						},
					},
				},
				eventsNormalOrder: make([]*v1.Event, 0),
				cTime:             time.Time{},
				isJob:             false,
			},
			want: "FailedScheduling",
		},
		{
			name: "没有event事件，没有调度标识别",
			args: args{
				podyaml: &v1.Pod{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       v1.PodSpec{},
					Status:     v1.PodStatus{},
				},
				eventsNormalOrder: make([]*v1.Event, 0),
				cTime:             time.Time{},
				isJob:             false,
			},
			want: "FailedScheduling",
		},
		{
			name: "test3",
			args: args{
				podyaml: &v1.Pod{
					TypeMeta:   metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       v1.PodSpec{},
					Status: v1.PodStatus{
						Conditions: []v1.PodCondition{
							{
								Type:               v1.PodScheduled,
								Status:             v1.ConditionTrue,
								LastTransitionTime: metav1.NewTime(parseTime("2020-03-31T10:42:32.491419+08:00")),
							},
						},
					},
				},
				eventsNormalOrder: []*v1.Event{
					{
						Message:        "Successfully assigned sts-yjbbp7ugdvbtc2kq-9httd to 215229390-c",
						Reason:         "Scheduled",
						FirstTimestamp: metav1.NewTime(parseTime("2020-03-31T10:42:32.495588+08:00")),
					},
				},
				cTime: parseTime("2020-03-31T10:42:31.923762+08:00"),
				isJob: false,
			},
			want: "FailedScheduling",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldFinishTime := tt.args.cTime.Add(10 * time.Minute)
			possibleReason := ""
			if got := analyzeFailureReasonDetails(tt.args.podyaml, tt.args.eventsNormalOrder, tt.args.cTime, &shouldFinishTime, tt.args.isJob, &possibleReason); got != tt.want {
				t.Errorf("analyzeFailReason() = %v, want %v", got, tt.want)
			}
		})
	}
}

// calculatePostStartHookTime
func TestCalculatePostStartHookTime(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "test1",
			args: args{[]*v1.Event{
				{
					Reason:         "Pulled",
					Message:        `Container image "reg.docker.com/swiftimage/basementtask-assets:21fe4e20200320223124636" already present on machine`,
					FirstTimestamp: metav1.NewTime(parseTime("2020-03-31T11:12:09.126788+08:00")),
				},
				{
					Message:        "Started container odp-sidecar-container",
					Reason:         "Started",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:16:13.280156+08:00")),
				},
				{
					Message:        "Container odp-sidecar-container execute poststart hook success",
					Reason:         "SucceedPostStartHook",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:16:13.576626+08:00")),
				},
			}},
			want: 0,
		},
		{
			name: "test2",
			args: args{[]*v1.Event{
				{
					Reason:         "Pulled",
					Message:        `Container image "reg.docker.com/swiftimage/basementtask-assets:21fe4e20200320223124636" already present on machine`,
					FirstTimestamp: metav1.NewTime(parseTime("2020-03-31T11:12:09.126788+08:00")),
				},
				{
					Message:        "Started container odp-sidecar-container",
					Reason:         "Started",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:16:13.280156+08:00")),
				},
				{
					Message:        "Container odp-sidecar-container execute poststart hook success",
					Reason:         "SucceedPostStartHook",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:16:13.576626+08:00")),
				},
				{
					Message:        "Started container mobilecashier",
					Reason:         "Started",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:18:00.85551+08:00")),
				},
				{
					Message:        "Container mobilecashier execute poststart hook success",
					Reason:         "SucceedPostStartHook",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:23:57.023865+08:00")),
				},
			}},
			want: 356,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculatePostStartHookTime(tt.args.events); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculatePostStartHookTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

// isRuntimeOperationFailure
func TestIsRuntimeOperationFailure(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{[]*v1.Event{
				{
					Message:        "Container odp-sidecar-container execute poststart hook success",
					Reason:         "SucceedPostStartHook",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-02T20:16:13.576626+08:00")),
				},
				{
					Message: `Error: failed to update resolv file of container: 7dbd132a34f678646087844bb391025a6ef96bcde047374e01d6bcc467db063e, ` +
						`error: failed to exec &{/usr/bin/docker [docker cp /home/t4/docker/containers/28549c22fd0d552398cf083c1ed5cc3a766cd199ccbba334e8b7949c2ecad10e/resolv.conf 7dbd132a34f678646087844bb391025a6ef96bcde047374e01d6bcc467db063e:/etc/] [] 0xc4260718b0 [] 0xc422021560 signal: killed 0xc427ccff20 true [0xc422cfc760 0xc422cfc778 0xc422cfc790] [0xc422cfc760 0xc422cfc778 0xc422cfc790] [0xc422cfc770 0xc422cfc788] [0xa86d80 0xa86d80] 0xc4216561e0 0xc4245552c0}, out: "", err: signal: killed"`,
					Reason:         "Failed",
					FirstTimestamp: metav1.NewTime(parseTime("2020-07-09T17:16:24.581848+08:00")),
				},
			}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRuntimeOperationFailure(tt.args.events); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("isRuntimeOperationFailure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_reverseEvent(t *testing.T) {
	type args struct {
		events []*v1.Event
	}
	tests := []struct {
		name  string
		args  args
		want  []*v1.Event
		want1 bool
	}{
		{
			name:  "test1",
			args:  args{events: []*v1.Event{}},
			want:  []*v1.Event{},
			want1: false,
		},
		{
			name: "test2",
			args: args{events: []*v1.Event{
				{Reason: "1", Type: v1.EventTypeNormal},
				{Reason: "2", Type: v1.EventTypeNormal},
			}},
			want: []*v1.Event{
				{Reason: "2", Type: v1.EventTypeNormal},
				{Reason: "1", Type: v1.EventTypeNormal},
			},
			want1: false,
		},
		{
			name: "test3",
			args: args{events: []*v1.Event{
				{Reason: "1", Type: v1.EventTypeNormal},
				{Reason: "2", Type: v1.EventTypeNormal},
				{Reason: "3", Type: v1.EventTypeNormal},
			}},
			want: []*v1.Event{
				{Reason: "3", Type: v1.EventTypeNormal},
				{Reason: "2", Type: v1.EventTypeNormal},
				{Reason: "1", Type: v1.EventTypeNormal},
			},
			want1: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := reverseEvent(tt.args.events)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverseEvent() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("reverseEvent() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_sortConditiosByTime(t *testing.T) {
	type args struct {
		conditions []v1.PodCondition
	}
	t1 := metav1.NewTime(time.Now())
	t2 := metav1.NewTime(t1.Add(20 * time.Second))
	t3 := metav1.NewTime(t1.Add(10 * time.Second))

	tests := []struct {
		name  string
		args  args
		want  []v1.PodCondition
		want1 bool
	}{
		{
			name: "test1",
			args: args{conditions: []v1.PodCondition{
				{Type: v1.PodReady, LastTransitionTime: t2},
				{Type: v1.PodInitialized, LastTransitionTime: t3},
				{Type: v1.PodScheduled, LastTransitionTime: t1},
			}},
			want: []v1.PodCondition{
				{Type: v1.PodReady, LastTransitionTime: t2},
				{Type: v1.PodInitialized, LastTransitionTime: t3},
				{Type: v1.PodScheduled, LastTransitionTime: t1},
			},
		},
		{
			name: "test2",
			args: args{conditions: []v1.PodCondition{
				{Type: v1.PodScheduled, LastTransitionTime: t1},
			}},
			want: []v1.PodCondition{
				{Type: v1.PodScheduled, LastTransitionTime: t1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortConditionsBytime(tt.args.conditions)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("reverseEvent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
