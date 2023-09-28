package spans

import (
	"fmt"
	"reflect"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

const (
	matchKeyDelimiter = ","
)

type Field struct {
	Name      string
	MatchKeys []string // 用户数组元素匹配; match key对应的value必须为string

}

// parse field
func (f *Field) parseField(path string) {
	if len(path) == 0 {
		return
	}

	if strings.Count(path, "[") == 1 && strings.Count(path, "]") == 1 &&
		strings.HasPrefix(path, "[") && !strings.HasPrefix(path, "]") && !strings.HasSuffix(path, "]") {
		f.Name = path[strings.Index(path, "]")+1:]
		f.MatchKeys = strings.Split(path[1:strings.Index(path, "]")], matchKeyDelimiter)
	} else {
		f.Name = path
	}
}

func (f *Field) GetFieldKey(value reflect.Value) string {
	if f.MatchKeys == nil || len(f.MatchKeys) == 0 {
		return f.Name
	}

	if !value.IsValid() || (value.Type().Kind() != reflect.Struct && value.Type().Kind() != reflect.String) {
		return ""
	}

	key := ""
	for _, mk := range f.MatchKeys {
		if mk == "$" {
			return fmt.Sprintf("%s-%s", key, value.String())
		}
		for i := 0; i < value.NumField(); i++ {
			if EqualsField(value.Type().Field(i), mk) {
				key = fmt.Sprintf("%s-%s", key, value.Field(i).String())
			}
		}
	}
	key = strings.Trim(key, "-")

	return fmt.Sprintf("[%s]%s", key, f.Name)

}

type FieldRef struct {
	FieldSelector string   `json:"fieldSelector,omitempty"`
	FieldPaths    []*Field `json:"-"`
	PathDelimiter string   `json:"pathDelimiter,omitempty"`
}

func NewFieldRef(fieldSelector string, pathDelimiter string) *FieldRef {
	f := &FieldRef{
		FieldSelector: fieldSelector,
		PathDelimiter: pathDelimiter,
	}
	f.parseFieldPath()
	return f
}

// 只支持int string 和 指针类型的字段
func (f *FieldRef) parseFieldPath() {
	if len(f.FieldSelector) < 0 {
		return
	}
	pathStrings := strings.Split(f.FieldSelector, f.PathDelimiter)
	f.FieldPaths = make([]*Field, 0, len(pathStrings))
	for i, _ := range pathStrings {
		field := &Field{}
		field.parseField(pathStrings[i])
		f.FieldPaths = append(f.FieldPaths, field)
	}
}

func (f *FieldRef) getFieldValue(currentValue reflect.Value, currentKeyPrefix string, currentIdx int) map[string]reflect.Value {
	if currentIdx >= len(f.FieldPaths) {
		return map[string]reflect.Value{currentKeyPrefix: currentValue}
	}
	//ptr
	if currentValue.Type().Kind() == reflect.Ptr && !currentValue.IsNil() {
		return f.getFieldValue(currentValue.Elem(), currentKeyPrefix, currentIdx)
	}

	//string
	if currentValue.Type().Kind() == reflect.String || currentValue.Type().Kind() == reflect.Int || currentValue.Type().Kind() == reflect.Bool {
		return map[string]reflect.Value{currentKeyPrefix: currentValue}
	}

	//struct
	if currentValue.Type().Kind() == reflect.Struct {
		for i := 0; i < currentValue.NumField(); i++ {
			if EqualsField(currentValue.Type().Field(i), f.FieldPaths[currentIdx].Name) {
				newKeyPrefix := currentKeyPrefix
				if key := f.FieldPaths[currentIdx].GetFieldKey(reflect.Value{}); len(key) > 0 {
					newKeyPrefix = fmt.Sprintf("%s_%s", currentKeyPrefix, key)
				}
				return f.getFieldValue(currentValue.Field(i), newKeyPrefix, currentIdx+1)
			}
		}
	}

	//slice	and array
	if currentValue.Type().Kind() == reflect.Slice || currentValue.Type().Kind() == reflect.Array {
		result := make(map[string]reflect.Value)
		if currentValue.IsNil() {
			return result
		}
		for i := 0; i < currentValue.Len(); i++ {
			newKeyPrefix := fmt.Sprintf("%s_%s", currentKeyPrefix, f.FieldPaths[currentIdx-1].GetFieldKey(currentValue.Index(i)))
			tmpRS := f.getFieldValue(currentValue.Index(i), newKeyPrefix, currentIdx)
			for k, v := range tmpRS {
				result[k] = v
			}
		}

		return result
	}

	//map
	if currentValue.Type().Kind() == reflect.Map {
		result := make(map[string]reflect.Value, 0)

		for i := 0; i < currentValue.Len(); i++ {
			mapKey := currentValue.MapKeys()[i]
			mapValue := currentValue.MapIndex(mapKey)

			if mapKey.String() != f.FieldPaths[currentIdx].Name {
				continue
			}

			newKeyPrefix := ""
			if mapKey.Type().Kind() == reflect.Ptr {
				newKeyPrefix = fmt.Sprintf("%s_%s", currentKeyPrefix, mapKey.Elem().String())
			} else {
				newKeyPrefix = fmt.Sprintf("%s_%s", currentKeyPrefix, mapKey.String())
			}

			tmpRS := f.getFieldValue(mapValue, newKeyPrefix, currentIdx+1)
			for k, v := range tmpRS {
				result[k] = v
			}
		}
		return result
	}

	return map[string]reflect.Value{}
}

func (f *FieldRef) GetFieldValue(obj interface{}) map[string]reflect.Value {
	if obj == nil {
		return nil
	}
	return f.getFieldValue(reflect.ValueOf(obj), "", 0)
}

// return 0 when RestartPolicy is Always, otherwise, return 1
func podIsJob(p *corev1.Pod) int {
	if p == nil {
		return 0
	}
	if p.Spec.RestartPolicy == corev1.RestartPolicyAlways {
		return 0
	}
	return 1
}

// return 0 when pod is not in a pod group, otherwise, return 1
func podInPodGroup(pod *corev1.Pod) int {
	if pod == nil {
		return 0
	}
	if len(pod.OwnerReferences) != 0 {
		for i := range pod.OwnerReferences {
			if pod.OwnerReferences[i].Kind == "PodGroup" {
				return 1
			}
		}
	}
	return 0
}

// 获取字段的tag，支持json和protobuf两种tag设置
func ParseJsonAndProtoTag(tag reflect.StructTag) []string {
	rs := make([]string, 0, 2)
	if tag.Get("json") != "" {
		rs = append(rs, strings.Split(tag.Get("json"), ",")[0])
	}
	if tag.Get("proto") != "" {
		rs = append(rs, strings.Split(tag.Get("json"), ",")[0])
	}
	return rs
}

// 对比字段的tag与给定的field是否想等
func EqualsField(rField reflect.StructField, fieldName string) bool {
	names := ParseJsonAndProtoTag(rField.Tag)
	for _, name := range names {
		if strings.EqualFold(name, fieldName) {
			return true
		}
	}
	return false
}
