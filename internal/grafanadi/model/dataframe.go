package model

type DataFrame struct {
	Schema SchemaType `json:"schema,omitempty"`
	Data   DataType   `json:"data,omitempty"`
}

type SchemaType struct {
	Fields []FieldType `json:"fields,omitempty"`
	Name   string      `json:"name,omitempty"`
}

type FieldType struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type,omitempty"`
}

type DataType struct {
	Values []interface{} `json:"values,omitempty"`
}
