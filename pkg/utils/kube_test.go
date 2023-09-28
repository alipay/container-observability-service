package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prashantv/gostub"
)

func TestInitKube(t *testing.T) {
	stubs := gostub.Stub(&EnableKubeClient, false)

	err := InitKube("/etc/kubernetes/kubeconfig/admin.kubeconfig")
	assert.Nil(t, err)
	stubs.Reset()

	err = InitKube("/etc/kubernetes/kubeconfig/admin.kubeconfig")
	assert.Error(t, err)

	_, err = GetClientFromFile("http://localhost:6443", "", 1024, 1024)
	assert.NoError(t, err)
}
