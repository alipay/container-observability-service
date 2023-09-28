package slo

import (
	"testing"

	k8saudit "k8s.io/apiserver/pkg/apis/audit"
)

func Test_checkUpgradeTimeout(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func Test_finishUpgradeMileStoneWithResult(t *testing.T) {
	type args struct {
		podKey string
		result string
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

func Test_processUpgradeStatus(t *testing.T) {
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

func Test_processUpgradeTrigger(t *testing.T) {
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
