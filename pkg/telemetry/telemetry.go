package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracerProvider *sdktrace.TracerProvider
)

// InitTracer initializes the global tracer
func InitTracer(serviceName string) (func(), error) {
	ctx := context.Background()

	// Create OTLP exporter
	conn, err := grpc.DialContext(ctx, "localhost:4317",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global trace provider
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Return cleanup function
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tracerProvider.Shutdown(ctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}, nil
}

// StartSpan starts a new span with the given name and returns the context and span
func StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name)
}

// StartSpanWithTracer starts a new span with the given name and tracer name
func StartSpanWithTracer(ctx context.Context, tracerName, name string) (context.Context, trace.Span) {
	return otel.Tracer(tracerName).Start(ctx, name)
}

// GetTraceID returns the trace ID from the context
func GetTraceID(ctx context.Context) string {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

// GetSpanID returns the span ID from the context
func GetSpanID(ctx context.Context) string {
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		return spanCtx.SpanID().String()
	}
	return ""
}

// GetTracer returns a tracer for the given name
func GetTracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
