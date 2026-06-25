package observability

import (
	"strings"
	"testing"
	"time"
)

func TestPrometheusTextIncludesHTTPRequestMetrics(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			Path:           "/metrics",
			IncludeRuntime: false,
			Buckets: []time.Duration{
				100 * time.Millisecond,
				time.Second,
			},
		},
	})

	metrics.RecordHTTPRequest("GET", "/hello/:id", 200, 75*time.Millisecond)
	output := metrics.PrometheusText()

	expected := []string{
		`test_app_http_requests_total{method="GET",route="/hello/:id",status="200"} 1`,
		`test_app_http_request_duration_seconds_count{method="GET",route="/hello/:id",status="200"} 1`,
		`test_app_http_request_duration_seconds_sum{method="GET",route="/hello/:id",status="200"}`,
	}

	for _, value := range expected {
		if !strings.Contains(output, value) {
			t.Fatalf("expected output to contain %q, got:\n%s", value, output)
		}
	}
}

func TestPrometheusTextEscapesLabels(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
			Buckets:        []time.Duration{time.Second},
		},
	})

	metrics.RecordHTTPRequest("GET", `/quoted/"path"`, 404, time.Millisecond)
	output := metrics.PrometheusText()

	if !strings.Contains(output, `route="/quoted/\"path\""`) {
		t.Fatalf("expected escaped route label, got:\n%s", output)
	}
}

func TestPrometheusTextIncludesUptime(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
		},
	})

	output := metrics.PrometheusText()
	if !strings.Contains(output, "test_app_uptime_seconds ") {
		t.Fatalf("expected uptime metric in output, got:\n%s", output)
	}
}

func TestPrometheusTextIncludesPanics(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
		},
	})

	metrics.IncPanic()
	output := metrics.PrometheusText()
	if !strings.Contains(output, "test_app_panics_total 1") {
		t.Fatalf("expected panic counter in output, got:\n%s", output)
	}
}

func TestPrometheusTextIncludesInFlight(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
		},
	})

	metrics.IncInFlight()
	metrics.IncInFlight()
	metrics.DecInFlight()

	output := metrics.PrometheusText()
	if !strings.Contains(output, "test_app_http_in_flight_requests 1") {
		t.Fatalf("expected in-flight gauge in output, got:\n%s", output)
	}
}

func TestPrometheusTextIncludesGoRuntimeWhenEnabled(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: true,
		},
	})

	output := metrics.PrometheusText()
	if !strings.Contains(output, "test_app_go_goroutines") && !strings.Contains(output, "go_goroutines") {
		t.Fatalf("expected Go runtime metrics in output, got:\n%s", output)
	}
}

func TestRecordHTTPRequestRecordsMultipleStatusCodes(t *testing.T) {
	metrics := NewMetrics(&Options{
		Metrics: MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
			Buckets:        []time.Duration{time.Second},
		},
	})

	metrics.RecordHTTPRequest("GET", "/api", 200, 50*time.Millisecond)
	metrics.RecordHTTPRequest("GET", "/api", 500, 100*time.Millisecond)

	output := metrics.PrometheusText()

	if !strings.Contains(output, `status="200"`) {
		t.Fatalf("expected status=200 in output, got:\n%s", output)
	}
	if !strings.Contains(output, `status="500"`) {
		t.Fatalf("expected status=500 in output, got:\n%s", output)
	}
}
