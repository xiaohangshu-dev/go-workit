// Package main 演示可观测性能力。
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
)

func main() {
	builder := webapp.NewBuilder()
	builder.AddMetrics(func(options *observability.MetricsOptions) {
		options.Namespace = "workit_observability_example"
		options.Path = "/metrics"
	})

	builder.AddOpenTelemetry(func(options *observability.TracingOptions) {
		options.Exporter = observability.TraceExporterStdout
		options.SampleRatio = 1
	})

	builder.AddHealthChecks().
		AddCheck("sample_dependency", func(ctx context.Context) observability.CheckResult {
			return observability.Healthy("sample dependency is reachable")
		}, "ready")

	app := builder.Build()

	app.MapMetrics()
	app.MapHealthChecks()

	app.MapRoute(func(router *gin.Engine) {
		router.GET("/hello", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Hello, observability",
			})
		})

		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(150 * time.Millisecond)
			c.JSON(http.StatusOK, gin.H{
				"message": "slow request completed",
			})
		})

		router.GET("/panic", func(c *gin.Context) {
			panic("observability example panic")
		})
	})

	app.Run()
}
