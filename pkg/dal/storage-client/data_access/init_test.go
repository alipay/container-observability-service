package data_access

import (
	"testing"

	"github.com/alipay/container-observability-service/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestNewDBClient(t *testing.T) {
	opts := common.NewDefaultOptions()
	_, err := NewDBClient(opts)
	assert.NotNil(t, err)

	opts = &common.DBOptions{
		Driver:       "mysql",
		MysqlOptions: common.NewMysqlOptions(),
	}
	_, err = NewDBClient(opts)
	assert.Nil(t, err)

	opts = &common.DBOptions{}
	_, err = NewDBClient(opts)
	assert.NotNil(t, err)
}
