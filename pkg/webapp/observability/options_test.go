package observability

import "testing"

func TestOptionsProvidesDefaults(t *testing.T) {
	options := NewOptions()

	options.normalize()

	if options.ServiceName != "go-workit" {
		t.Fatalf("expected default service name, got %q", options.ServiceName)
	}
	if options.Environment != "development" {
		t.Fatalf("expected default environment, got %q", options.Environment)
	}
	if options.Metrics.Path != "/metrics" {
		t.Fatalf("expected default metrics path, got %q", options.Metrics.Path)
	}
	if options.Health.HealthPath != "/health" {
		t.Fatalf("expected default health path, got %q", options.Health.HealthPath)
	}
}

func TestOptionsKeepsGroupedConfiguration(t *testing.T) {
	options := NewOptions()
	options.Metrics.Namespace = "service"
	options.Metrics.Path = "/service-metrics"
	options.Health.LivenessPath = "/service-live"
	options.Health.ReadinessPath = "/service-ready"

	options.normalize()

	if options.Metrics.Namespace != "service" || options.Metrics.Path != "/service-metrics" {
		t.Fatalf("expected grouped metrics configuration to be kept")
	}
	if options.Health.LivenessPath != "/service-live" || options.Health.ReadinessPath != "/service-ready" {
		t.Fatalf("expected grouped health configuration to be kept")
	}
}
