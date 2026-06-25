package ginx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
)

func TestObservabilityMiddlewareRecordsHTTPMetricsWhenMetricsWereAdded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	metrics := observability.NewMetrics(&observability.Options{
		Metrics: observability.MetricsOptions{
			Namespace:      "test_app",
			IncludeRuntime: false,
		},
	})
	app := &WebApplication{
		handler:       gin.New(),
		observability: metrics,
	}
	app.useObservability()
	app.engine().GET("/hello", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/hello", nil)

	app.engine().ServeHTTP(recorder, request)

	output := metrics.PrometheusText()
	if !strings.Contains(output, `test_app_http_requests_total{method="GET",route="/hello",status="404"} 1`) {
		t.Fatalf("expected HTTP request metrics to be recorded, got:\n%s", output)
	}
}

func TestMapHealthChecksDoesNothingWhenHealthChecksWereNotAdded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &WebApplication{
		handler: gin.New(),
	}

	app.MapHealthChecks()

	if routes := app.engine().RouterMap; len(routes) != 0 {
		t.Fatalf("expected no health routes when AddHealthChecks was not called, got %d", len(routes))
	}
}

func TestMapHealthChecksMapsAllHealthEndpointsWhenConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	options := observability.NewOptions()
	options.Health.HealthPath = "/healthz"
	options.Health.ReadinessPath = "/readyz"
	options.Health.LivenessPath = "/livez"

	app := &WebApplication{
		handler: gin.New(),
		health:  observability.NewHealthRegistry(options),
	}

	app.MapHealthChecks()

	for _, path := range []string{"/healthz", "/readyz", "/livez"} {
		if _, ok := app.engine().RouterMap["GET:"+path]; !ok {
			t.Fatalf("expected %s to be mapped by MapHealthChecks", path)
		}
	}
}
