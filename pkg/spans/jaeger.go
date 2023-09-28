package spans

import (
	"context"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	//tracesdk "go.opentelemetry.io/otel/sdk/export/trace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

var providers = make(map[string]*sdktrace.TracerProvider, 0)
var spanProcessor sdktrace.SpanProcessor

//var spanExporter tracesdk.SpanExporter

/*
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
	fmt.Printf("trace error: %s", err.Error())
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
		fmt.Println(fmt.Errorf("failed to create resource: %w", err))
	}

	providers[service] = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(spanProcessor),
	)

	return providers[service]
}
