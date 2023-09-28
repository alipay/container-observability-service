package model

type DataModelInterface interface {
	// 返回表名
	TableName() string
	// type: _doc, ...
	TypeName() string
}
