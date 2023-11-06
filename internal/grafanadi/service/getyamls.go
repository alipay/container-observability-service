package service

import (
	"encoding/json"
	"strings"

	"github.com/alipay/container-observability-service/internal/grafanadi/model"
)

func TransformYaml2Html(yamlStruct interface{}) []byte {
	tmp, _ := json.Marshal(yamlStruct)
	respBody := []byte(strings.Replace(model.TPL, "%s", string(tmp), -1))
	return respBody
}
