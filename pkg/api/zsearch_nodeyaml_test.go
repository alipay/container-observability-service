package api

import (
	"reflect"
	"testing"
)

func Test_queryNodeYamlWithNodeName(t *testing.T) {
	type args struct {
		nodeName string
	}
	tests := []struct {
		name string
		args args
		want *nodeOpStruct
	}{
		{
			name: "test1",
			args: args{
				nodeName: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryNodeYamlWithNodeName(tt.args.nodeName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("queryNodeYamlWithNodeName() = %v, want %v", got, tt.want)
			}
		})
	}
}
