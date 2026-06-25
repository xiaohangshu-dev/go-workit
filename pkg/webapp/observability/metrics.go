package observability

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/expfmt"
)

// httpLabels 返回 Prometheus 标签名列表（按固定顺序）
func httpLabelNames() []string {
	return []string{"method", "route", "status"}
}

func statusString(status int) string {
	return strconv.Itoa(status)
}

// Metrics 保存框架内置可观测性指标。
// 使用 Prometheus client_golang 作为标准指标库。
type Metrics struct {
	options   *Options
	start     time.Time
	namespace string

	// Prometheus 指标
	httpRequestsTotal *prometheus.CounterVec
	httpDuration      *prometheus.HistogramVec
	inFlightGauge     prometheus.Gauge
	panicsCounter     prometheus.Counter
	uptimeGauge       prometheus.GaugeFunc

	registry *prometheus.Registry
	handler  http.Handler
}

// NewMetrics 创建可观测性指标存储。
func NewMetrics(options *Options) *Metrics {
	if options == nil {
		options = NewOptions()
	}
	options.normalize()

	ns := sanitizeNamespace(options.Metrics.Namespace)
	reg := prometheus.NewRegistry()

	m := &Metrics{
		options:   options,
		start:     time.Now(),
		namespace: ns,
		registry:  reg,
	}

	// 注册 HTTP 请求计数器
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: ns,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests.",
		},
		httpLabelNames(),
	)
	reg.MustRegister(m.httpRequestsTotal)

	// 注册 HTTP 请求耗时直方图
	m.httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   durationsToFloat(options.Metrics.Buckets),
		},
		httpLabelNames(),
	)
	reg.MustRegister(m.httpDuration)

	// 注册处理中的请求数
	m.inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: ns,
		Name:      "http_in_flight_requests",
		Help:      "Current number of in-flight HTTP requests.",
	})
	reg.MustRegister(m.inFlightGauge)

	// 注册 Panic 计数器
	m.panicsCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: ns,
		Name:      "panics_total",
		Help:      "Total number of recovered panics.",
	})
	reg.MustRegister(m.panicsCounter)

	// 注册运行时长
	m.uptimeGauge = prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "uptime_seconds",
			Help:      "Application uptime in seconds.",
		},
		func() float64 { return time.Since(m.start).Seconds() },
	)
	reg.MustRegister(m.uptimeGauge)

	// 可选：添加 Go 运行时指标
	if options.Metrics.IncludeRuntime {
		reg.MustRegister(collectors.NewGoCollector())
		reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
			Namespace: ns,
		}))
	}

	// 创建 HTTP handler
	m.handler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{
		Registry: reg,
	})

	return m
}

// durationsToFloat 将 time.Duration 切片转换为 float64 切片（单位：秒）
func durationsToFloat(buckets []time.Duration) []float64 {
	result := make([]float64, len(buckets))
	for i, d := range buckets {
		result[i] = d.Seconds()
	}
	return result
}

// sanitizeNamespace 清理命名空间，确保符合 Prometheus 命名规范
func sanitizeNamespace(ns string) string {
	ns = strings.NewReplacer("-", "_", ".", "_", " ", "_").Replace(ns)
	if len(ns) > 0 && ns[len(ns)-1] == '_' {
		ns = ns[:len(ns)-1]
	}
	return ns
}

// Options 返回指标配置。
func (m *Metrics) Options() *Options {
	return m.options
}

// IncInFlight 记录正在处理的请求数。
func (m *Metrics) IncInFlight() {
	m.inFlightGauge.Inc()
}

// DecInFlight 释放正在处理的请求数。
func (m *Metrics) DecInFlight() {
	m.inFlightGauge.Dec()
}

// IncPanic 记录一次 panic。
func (m *Metrics) IncPanic() {
	m.panicsCounter.Inc()
}

// RecordHTTPRequest 记录一次 HTTP 请求。
func (m *Metrics) RecordHTTPRequest(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unknown"
	}
	labels := prometheus.Labels{
		"method": method,
		"route":  route,
		"status": statusString(status),
	}
	m.httpRequestsTotal.With(labels).Inc()
	m.httpDuration.With(labels).Observe(duration.Seconds())
}

// PrometheusText 返回 Prometheus text exposition 格式的指标数据。
func (m *Metrics) PrometheusText() string {
	families, err := m.registry.Gather()
	if err != nil {
		return "# error gathering metrics: " + err.Error() + "\n"
	}

	var b strings.Builder
	for _, f := range families {
		_, err := expfmt.MetricFamilyToText(&b, f)
		if err != nil {
			return "# error writing metric: " + err.Error() + "\n"
		}
	}
	return b.String()
}

// PrometheusHandler 返回标准的 Prometheus HTTP handler，可用于挂载到路由上。
func (m *Metrics) PrometheusHandler() http.Handler {
	return m.handler
}
