package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	trace_sdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func InitializeWithCustomTracer(tp trace.TracerProvider, propagator propagation.TextMapPropagator) {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagator)
}

func Initialize(ctx context.Context, opts ...Option) (trace.TracerProvider, error) {
	conf := &config{
		endpointType: EndpointType_GRPC,
		insecure:     true,
	}

	for _, opt := range opts {
		opt(conf)
	}

	// create the exporter
	var err error
	var exporter *otlptrace.Exporter
	switch conf.endpointType {
	case EndpointType_GRPC:
		exporter, err = otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(conf.endpoint),
			otlptracegrpc.WithInsecure(),
		)
	case EndpointType_HTTP:
		exporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(conf.endpoint),
			otlptracehttp.WithInsecure(),
		)
	default:
		return nil, fmt.Errorf("invalid endpoint type: %s", conf.endpointType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create otlp exporter: %w", err)
	}

	// create resource
	metaAttributes := []attribute.KeyValue{
		semconv.ServiceNameKey.String(conf.serviceName),
		semconv.ServiceVersionKey.String(conf.serviceVersion),
	}
	for k, v := range conf.metadata {
		metaAttributes = append(metaAttributes, attribute.String(k, v))
	}

	res, err := resource.New(ctx, resource.WithAttributes(metaAttributes...))
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// create sampler
	sampler := trace_sdk.AlwaysSample()

	// create tracer provider
	tp := trace_sdk.NewTracerProvider(
		trace_sdk.WithBatcher(exporter),
		trace_sdk.WithResource(res),
		trace_sdk.WithSampler(sampler),
	)

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)

	InitializeWithCustomTracer(tp, propagator)
	return tp, nil
}
