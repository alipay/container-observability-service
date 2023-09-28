package trace

const (
	TraceFeature           = "TraceFeature"
	TraceContextAnnotation = "meta.lunettes.com/trace-context"
)

type TraceService struct {
	Component string `json:"component"`
	SpanID    string `json:"span_id"`
}

type TraceInfo struct {
	TraceID      string          `json:"trace_id"`
	ParentSpanID string          `json:"parent_id"`
	RootSpanID   string          `json:"root_span_id"`
	DeliveryType string          `json:"delivery_type"`
	Status       TraceStatus     `json:"status"`
	Services     []*TraceService `json:"services"`
	StartAt      string          `json:"start_at"`
	FinishAt     string          `json:"finish_at"`
}

type TraceStatus string

const (
	OpenTraceStatus   TraceStatus = "open"
	ClosedTraceStatus TraceStatus = "closed"

	CreateDelivery string = "create"
	StartDelivery  string = "start"
	StopDelivery   string = "stop"
)
