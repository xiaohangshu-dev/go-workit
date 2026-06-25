package app

import (
	"testing"
)

func TestNewBuilderReturnsBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("expected NewBuilder() to return non-nil builder")
	}
}

func TestBuilderDefaultsAppName(t *testing.T) {
	builder := NewBuilder()
	name := builder.Config().GetString("app.name")
	if name != "go-workit" {
		t.Fatalf("expected default app name 'go-workit', got %q", name)
	}
}

func TestBuilderDefaultsLogLevel(t *testing.T) {
	builder := NewBuilder()
	level := builder.Config().GetString("log.level")
	if level != "info" {
		t.Fatalf("expected default log level 'info', got %q", level)
	}
}

func TestBuilderBuildReturnsApplication(t *testing.T) {
	builder := NewBuilder()
	app := builder.Build()

	if app == nil {
		t.Fatal("expected Build() to return non-nil application")
	}
	if app.Config() == nil {
		t.Fatal("expected application to have config")
	}
	if app.Logger() == nil {
		t.Fatal("expected application to have logger")
	}
}

func TestApplicationConfigReturnsConfig(t *testing.T) {
	builder := NewBuilder()
	app := builder.Build()

	cfg := app.Config()
	if cfg == nil {
		t.Fatal("expected Config() to return non-nil viper instance")
	}
}

func TestApplicationLoggerReturnsLogger(t *testing.T) {
	builder := NewBuilder()
	app := builder.Build()

	logger := app.Logger()
	if logger == nil {
		t.Fatal("expected Logger() to return non-nil zap logger")
	}
}

func TestApplicationMetricsReturnsMetrics(t *testing.T) {
	builder := NewBuilder()
	app := builder.Build()

	metrics := app.Metrics()
	if metrics == nil {
		t.Fatal("expected Metrics() to return non-nil metrics")
	}
}

func TestDefaultMetricsIncrementAndGet(t *testing.T) {
	metrics := newDefaultMetrics()

	metrics.Increment("test.counter")
	metrics.Increment("test.counter")
	metrics.Increment("other.counter")

	if val := metrics.GetCounter("test.counter"); val != 2 {
		t.Fatalf("expected counter 'test.counter' to be 2, got %d", val)
	}
	if val := metrics.GetCounter("other.counter"); val != 1 {
		t.Fatalf("expected counter 'other.counter' to be 1, got %d", val)
	}
	if val := metrics.GetCounter("nonexistent"); val != 0 {
		t.Fatalf("expected nonexistent counter to be 0, got %d", val)
	}
}

func TestAppendContainerAddsOptions(t *testing.T) {
	builder := NewBuilder()
	initialLen := len(builder.options)

	builder.AddServices()
	if len(builder.options) != initialLen {
		t.Fatalf("expected options length %d, got %d", initialLen, len(builder.options))
	}
}
