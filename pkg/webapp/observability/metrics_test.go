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
		`test_app_http_request_duration_seconds_bucket{method="GET",route="/hello/:id",status="200",le="0.100"} 1`,
		`test_app_http_request_duration_seconds_count{method="GET",route="/hello/:id",status="200"} 1`,
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
