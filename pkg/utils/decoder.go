package utils

import (
	"encoding/json"
)

type JsoniterDecoder struct{}

// Decode decodes with json.Unmarshal from the Go standard library.
func (u *JsoniterDecoder) Decode(data []byte, v interface{}) error {
	//var json = jsoniter.ConfigCompatibleWithStandardLibrary
	return json.Unmarshal(data, v)
}
