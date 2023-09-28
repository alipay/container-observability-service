package nodeyaml

import (
	v12 "k8s.io/api/authentication/v1"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/apis/audit"
)

func Test_processAuditEvent(t *testing.T) {
	type args struct {
		v interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				v: &audit.Event{
					TypeMeta:                 v1.TypeMeta{},
					Level:                    "",
					AuditID:                  "",
					Stage:                    "",
					RequestURI:               "",
					Verb:                     "",
					User:                     v12.UserInfo{},
					ImpersonatedUser:         nil,
					SourceIPs:                nil,
					UserAgent:                "",
					ObjectRef:                nil,
					ResponseStatus:           nil,
					RequestObject:            nil,
					ResponseObject:           nil,
					RequestReceivedTimestamp: v1.MicroTime{},
					StageTimestamp:           v1.MicroTime{},
					Annotations:              nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
