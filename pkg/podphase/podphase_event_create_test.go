package podphase

import (
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apiserver/pkg/apis/audit"
)

func Test_isErrorEvent(t *testing.T) {
	type args struct {
		event *v1.Event
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
			if got := isErrorEvent(tt.args.event); got != tt.want {
				t.Errorf("isErrorEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processPodEventCreation(t *testing.T) {
	type args struct {
		auditEvent *audit.Event
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

func Test_extractVictims(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				str: "victims are: [xxx,  yyyy]",
			},
			want: []string{
				"xxx", "yyyy",
			},
		},
		{
			name: "test2",
			args: args{
				str: "victims are:[xxxx, yyyy ]",
			},
			want: []string{
				"xxxx", "yyyy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractVictims(tt.args.str); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractVictims() = %v, want %v", got, tt.want)
			}
		})
	}
}
