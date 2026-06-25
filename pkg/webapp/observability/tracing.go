package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Telemetry struct {
	options  *Options
	provider *sdktrace.TracerProvider
}

func NewTelemetry(ctx context.Context, options *Options) (*Telemetry, error) {
	if options == nil {
		options = NewOptions()
	}
	options.normalize()

	telemetry := &Telemetry{
		options: options,
	}

	exporter, err := newTraceExporter(ctx, options.Tracing)
	if err != nil {
		return nil, err
	}

	resource, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", options.ServiceName),
			attribute.String("deployment.environment", options.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("build telemetry resource: %w", err)
	}

	providerOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sampler(options.Tracing.SampleRatio)),
	}
	if exporter != nil {
		providerOptions = append(providerOptions, sdktrace.WithBatcher(exporter))
	}

	telemetry.provider = sdktrace.NewTracerProvider(providerOptions...)
	otel.SetTracerProvider(telemetry.provider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return telemetry, nil
}

func (t *Telemetry) Options() *Options {
	return t.options
}

func (t *Telemetry) Enabled() bool {
	return t != nil && t.provider != nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil || t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

func newTraceExporter(ctx context.Context, options TracingOptions) (sdktrace.SpanExporter, error) {
	switch options.Exporter {
	case TraceExporterNone, "":
		return nil, nil
	case TraceExporterStdout:
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case TraceExporterOTLPGRPC:
		otlpOptions := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(options.Endpoint),
		}
		if options.Insecure {
			otlpOptions = append(otlpOptions, otlptracegrpc.WithInsecure())
		}
		if len(options.Headers) > 0 {
			otlpOptions = append(otlpOptions, otlptracegrpc.WithHeaders(options.Headers))
		}
		return otlptracegrpc.New(ctx, otlpOptions...)
	default:
		return nil, fmt.Errorf("unsupported trace exporter: %s", options.Exporter)
	}
}

func sampler(ratio float64) sdktrace.Sampler {
	switch {
	case ratio >= 1:
		return sdktrace.AlwaysSample()
	case ratio <= 0:
		return sdktrace.NeverSample()
	default:
		return sdktrace.TraceIDRatioBased(ratio)
	}
}
