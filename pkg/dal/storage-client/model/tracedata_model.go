package model

type JSONData struct {
	Data   []Data      `json:"data"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
	Errors interface{} `json:"errors"`
}
type References struct {
	RefType string `json:"refType"`
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}
type Tags struct {
	Key   string      `json:"key"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}
type Spans struct {
	TraceID       string        `json:"traceID"`
	SpanID        string        `json:"spanID"`
	OperationName string        `json:"operationName"`
	References    []References  `json:"references"`
	StartTime     int64         `json:"startTime"`
	Duration      int64         `json:"duration"`
	Tags          []Tags        `json:"tags"`
	Logs          []interface{} `json:"logs"`
	ProcessID     string        `json:"processID"`
	Warnings      interface{}   `json:"warnings"`
}
type P1 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P2 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P3 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P4 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P5 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P6 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type P7 struct {
	ServiceName string        `json:"serviceName"`
	Tags        []interface{} `json:"tags"`
}
type Processes struct {
	P1 P1 `json:"p1"`
	P2 P2 `json:"p2"`
	P3 P3 `json:"p3"`
	P4 P4 `json:"p4"`
	P5 P5 `json:"p5"`
	P6 P6 `json:"p6"`
	P7 P7 `json:"p7"`
}
type Data struct {
	TraceID   string      `json:"traceID"`
	Spans     []Spans     `json:"spans"`
	Processes Processes   `json:"processes"`
	Warnings  interface{} `json:"warnings"`
}
