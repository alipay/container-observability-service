package extractor

import (
	"testing"

	"k8s.io/apiserver/pkg/apis/audit"
)

func Test_processPatchNoSubresource(t *testing.T) {
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

func Test_processPatchSubresourceStatus(t *testing.T) {
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

func Test_processPodPatch(t *testing.T) {
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
