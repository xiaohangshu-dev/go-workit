package ginx

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
)

type Observability struct {
	metrics *observability.Metrics
}

func newObservability(metrics *observability.Metrics) Middleware {
	return &Observability{
		metrics: metrics,
	}
}

func (o *Observability) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		options := o.metrics.Options()
		if c.Request.URL.Path == options.Metrics.Path {
			c.Next()
			return
		}

		start := time.Now()
		o.metrics.IncInFlight()
		defer o.metrics.DecInFlight()

		defer func() {
			status := c.Writer.Status()
			if status == 0 {
				status = http.StatusOK
			}

			if err := recover(); err != nil {
				o.metrics.RecordHTTPRequest(c.Request.Method, routePattern(c), http.StatusInternalServerError, time.Since(start))
				panic(err)
			}

			o.metrics.RecordHTTPRequest(c.Request.Method, routePattern(c), status, time.Since(start))
		}()

		c.Next()
	}
}

func routePattern(c *gin.Context) string {
	route := c.FullPath()
	if route == "" {
		route = c.Request.URL.Path
	}
	return route
}
