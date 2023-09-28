package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitConfig(t *testing.T) {
	cfgFile := "./testdata/config.yaml"
	options, err := InitConfig(cfgFile)
	assert.Nil(t, err)
	fmt.Printf("options is %+v\n, %+v\n, %+v\n, ", options, options.MysqlOptions, options.ESOptions)

	badCfgFile := "./testdata/config-bad.yaml"
	_, err = InitConfig(badCfgFile)
	assert.NotNil(t, err)

	noCfgFile := "./testdata/config-not-exist.yaml"
	_, err = InitConfig(noCfgFile)
	assert.NotNil(t, err)

}
