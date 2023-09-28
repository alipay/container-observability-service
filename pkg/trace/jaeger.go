package trace

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/klog/v2"
	"math/rand"
	"sync"
	"time"
)

const (
	current_trace_key = "current_trace_key"
	current_span_key  = "current_span_key"
)

var providers = make(map[string]*sdktrace.TracerProvider, 0)
var spanProcessor sdktrace.SpanProcessor
var signalCtx context.Context

/*var spanExporter tracesdk.SpanExporter

func setupOTLP(ctx context.Context, addr string, headers string, secured bool) (tracesdk.SpanExporter, error) {
	klog.Info("Setting up OTLP Exporter", "addr", addr)

	var exp *otlp.Exporter
	var err error

	headersMap := make(map[string]string)
	if headers != "" {
		ha := strings.Split(headers, ",")
		for _, h := range ha {
			parts := strings.Split(h, "=")
			if len(parts) != 2 {
				klog.Error(errors.New("Error parsing OTLP header"), "header parts length is not 2", "header", h)
				continue
			}
			headersMap[parts[0]] = parts[1]
		}
	}

	if secured {
		exp, err = otlp.NewExporter(
			ctx,
			otlpgrpc.NewDriver(
				otlpgrpc.WithEndpoint(addr),
				otlpgrpc.WithHeaders(headersMap),
				otlpgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, "")),
			),
		)
	} else {
		exp, err = otlp.NewExporter(
			ctx,
			otlpgrpc.NewDriver(
				otlpgrpc.WithEndpoint(addr),
				otlpgrpc.WithHeaders(headersMap),
				otlpgrpc.WithInsecure(),
			),
		)
	}
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.TraceContext{})
	return exp, err
}

func GetSpanExporter() tracesdk.SpanExporter {
	return spanExporter
}*/

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initOtlpProcessor(addr string) error {
	ctx := context.Background()

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}
	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	spanProcessor = sdktrace.NewBatchSpanProcessor(traceExporter)
	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return nil
}

type TraceErrorHandler struct{}

func (*TraceErrorHandler) Handle(err error) {
	klog.Errorf("trace error: %s", err.Error())
}

func getProvider(service string) *sdktrace.TracerProvider {
	if pv, ok := providers[service]; ok {
		return pv
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceNameKey.String(service),
		),
	)
	if err != nil {
		klog.Errorf("failed to create resource: %w", err)
	}

	providers[service] = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProcessor),
		sdktrace.WithIDGenerator(newContextIDGenerator()),
	)

	return providers[service]
}

type contextIDGenerator struct {
	sync.Mutex
	randSource *rand.Rand
}

func newContextIDGenerator() sdktrace.IDGenerator {
	gen := &contextIDGenerator{}
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	gen.randSource = rand.New(rand.NewSource(rngSeed))
	return gen
}

// NewSpanID returns a non-zero span ID from a randomly-chosen sequence.
func (gen *contextIDGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.Lock()
	defer gen.Unlock()
	if v := ctx.Value(current_span_key); v != nil {
		sid, ok := v.(trace.SpanID)
		if ok {
			return sid
		}
	}

	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])
	return sid
}

// NewIDs returns a non-zero trace ID and a non-zero span ID from a
// randomly-chosen sequence.
func (gen *contextIDGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	gen.Lock()
	defer gen.Unlock()
	tid := trace.TraceID{}
	_, _ = gen.randSource.Read(tid[:])
	if v := ctx.Value(current_trace_key); v != nil {
		customID, ok := v.(trace.TraceID)
		if ok {
			tid = customID
		}
	}

	sid := trace.SpanID{}
	_, _ = gen.randSource.Read(sid[:])
	if v := ctx.Value(current_span_key); v != nil {
		customID, ok := v.(trace.SpanID)
		if ok {
			sid = customID
		}
	}

	return tid, sid
}
