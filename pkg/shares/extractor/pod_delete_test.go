package extractor

import (
	"testing"

	"k8s.io/apiserver/pkg/apis/audit"
)

func Test_processPodDeletion(t *testing.T) {
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
