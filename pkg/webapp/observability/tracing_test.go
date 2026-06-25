package observability

import (
	"context"
	"testing"
)

func TestTelemetryInitializesWithDefaultExporter(t *testing.T) {
	telemetry, err := NewTelemetry(context.Background(), NewOptions())
	if err != nil {
		t.Fatalf("expected telemetry to initialize: %v", err)
	}
	if !telemetry.Enabled() {
		t.Fatalf("expected telemetry to be enabled when it is constructed")
	}
	if err := telemetry.Shutdown(context.Background()); err != nil {
		t.Fatalf("expected telemetry shutdown to succeed: %v", err)
	}
}

func TestTelemetryRejectsUnsupportedExporter(t *testing.T) {
	options := NewOptions()
	options.Tracing.Exporter = TraceExporter("bad_exporter")

	if _, err := NewTelemetry(context.Background(), options); err == nil {
		t.Fatalf("expected unsupported exporter to return an error")
	}
}
