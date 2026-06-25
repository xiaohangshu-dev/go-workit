package observability

import (
	"context"
	"testing"
)

func TestHealthRegistryAggregatesReadinessChecks(t *testing.T) {
	options := NewOptions()
	options.Health.AddCheck("database", func(ctx context.Context) CheckResult {
		return Healthy("database is reachable")
	}, "ready")
	options.Health.AddCheck("cache", func(ctx context.Context) CheckResult {
		return Unhealthy("cache is down")
	}, "ready")

	registry := NewHealthRegistry(options)
	report := registry.Readiness(context.Background())

	if report.Status != HealthStatusUnhealthy {
		t.Fatalf("expected unhealthy readiness, got %s", report.Status)
	}
	if report.Checks["database"].Status != HealthStatusHealthy {
		t.Fatalf("expected database check to be healthy")
	}
	if report.Checks["cache"].Status != HealthStatusUnhealthy {
		t.Fatalf("expected cache check to be unhealthy")
	}
}

func TestHealthRegistryLivenessDoesNotRunDependencyChecks(t *testing.T) {
	options := NewOptions()
	options.Health.AddCheck("database", func(ctx context.Context) CheckResult {
		return Unhealthy("database is down")
	}, "ready")

	registry := NewHealthRegistry(options)
	report := registry.Liveness(context.Background())

	if report.Status != HealthStatusHealthy {
		t.Fatalf("expected liveness to be healthy, got %s", report.Status)
	}
	if len(report.Checks) != 0 {
		t.Fatalf("expected liveness to skip dependency checks")
	}
}
