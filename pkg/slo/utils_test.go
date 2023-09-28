package slo

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
)

func Test_genPodKey(t *testing.T) {
	type args struct {
		clusterName  string
		podNamespace string
		podName      string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				clusterName:  "cluster",
				podNamespace: "namespace",
				podName:      "podname",
			},
			want: "cluster/namespace/podname",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genPodKey(tt.args.clusterName, tt.args.podNamespace, tt.args.podName); got != tt.want {
				t.Errorf("genPodKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getInitContainerImage(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getInitContainerImage(tt.args.pod); got != tt.want {
				t.Errorf("getInitContainerImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getOwnerRefStr(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOwnerRefStr(tt.args.pod); got != tt.want {
				t.Errorf("getOwnerRefStr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getSchedulingStrategyAndCores(t *testing.T) {
	type args struct {
		pod *v1.Pod
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getSchedulingStrategyAndCores(tt.args.pod)
			if got != tt.want {
				t.Errorf("getSchedulingStrategyAndCores() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getSchedulingStrategyAndCores() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_isPodWithPVC(t *testing.T) {
	type args struct {
		pod *v1.Pod
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
			if got := isPodWithPVC(tt.args.pod); got != tt.want {
				t.Errorf("isPodWithPVC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_r2m(t *testing.T) {
	type args struct {
		r v1.ResourceList
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r2m(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("r2m() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("r2m() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEsReader(t *testing.T) {}
