package utils

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestGetPersistentVolumeClaimNamesFromPod(t *testing.T) {
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetPersistentVolumeClaimNamesFromPod(tt.args.pod); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetPersistentVolumeClaimNamesFromPod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInAnnotations(t *testing.T) {
	type args struct {
		annotations map[string]string
		k           string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
				k: "key1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InAnnotations(tt.args.annotations, tt.args.k); got != tt.want {
				t.Errorf("InAnnotations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInFinalizers(t *testing.T) {
	type args struct {
		finalizers []string
		k          string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				finalizers: []string{"f1", "f2"},
				k:          "f1",
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				finalizers: []string{"f1", "f2"},
				k:          "f3",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := InFinalizers(tt.args.finalizers, tt.args.k); got != tt.want {
				t.Errorf("InFinalizers() = %v, want %v", got, tt.want)
			}
		})
	}
}
