package tkpReqProvider

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/klog"
)

var tkpRequestConfigMap map[string]string

func InitTkpReqConfig(cfgFile string) error {
	jsonFile, err := os.ReadFile(cfgFile)
	if err != nil {
		klog.Errorf("Failed to read tkpReqConfig file %s: %v", cfgFile, err)
		return err
	}

	var dataMap map[string]interface{}
	if err := json.Unmarshal(jsonFile, &dataMap); err != nil {
		klog.Errorf("Failed to parse tkpReqConfig JSON content from %s: %v", cfgFile, err)
		return err
	}

	stringMap := make(map[string]string)
	for key, value := range dataMap {
		if strValue, ok := value.(string); ok {
			stringMap[key] = strValue
		} else {
			stringMap[key] = fmt.Sprintf("%v", value)
		}
	}

	tkpRequestConfigMap = stringMap
	return nil
}

func GetTkpReqUrl(key string) string {
	if tkpRequestConfigMap == nil {
		klog.Warningf("tkpReqConfig Configuration map is uninitialized. Returning empty string for key '%s'.", key)
		return ""
	}
	return tkpRequestConfigMap[key]
}
