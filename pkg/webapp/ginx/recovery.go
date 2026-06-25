package ginx

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// newRecoveryWithZap returns a middleware that recovers from any panics and logs the error with zap.
func newRecoveryWithZap(logger *zap.Logger, metrics ...*observability.Metrics) gin.HandlerFunc {
	var observabilityMetrics *observability.Metrics
	if len(metrics) > 0 {
		observabilityMetrics = metrics[0]
	}

	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if observabilityMetrics != nil {
					observabilityMetrics.IncPanic()
				}
				fields := []zap.Field{
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
					zap.ByteString("stack", debug.Stack()),
				}
				if spanContext := trace.SpanContextFromContext(c.Request.Context()); spanContext.IsValid() {
					fields = append(fields,
						zap.String("trace_id", spanContext.TraceID().String()),
						zap.String("span_id", spanContext.SpanID().String()),
					)
				}
				logger.Error("panic recovered", fields...)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
