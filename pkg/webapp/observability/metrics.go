package observability

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

type httpMetricKey struct {
	Method string
	Route  string
	Status int
}

type httpMetricValue struct {
	Count   int64
	Sum     float64
	Buckets []int64
}

// Metrics 保存框架内置可观测性指标。
type Metrics struct {
	options  *Options
	start    time.Time
	mu       sync.RWMutex
	http     map[httpMetricKey]*httpMetricValue
	inFlight int64
	panics   int64
}

// NewMetrics 创建可观测性指标存储。
func NewMetrics(options *Options) *Metrics {
	if options == nil {
		options = NewOptions()
	}
	options.normalize()
	return &Metrics{
		options: options,
		start:   time.Now(),
		http:    make(map[httpMetricKey]*httpMetricValue),
	}
}

// Options 返回指标配置。
func (m *Metrics) Options() *Options {
	return m.options
}

// IncInFlight 记录正在处理的请求数。
func (m *Metrics) IncInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inFlight++
}

// DecInFlight 释放正在处理的请求数。
func (m *Metrics) DecInFlight() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.inFlight > 0 {
		m.inFlight--
	}
}

// IncPanic 记录一次 panic。
func (m *Metrics) IncPanic() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.panics++
}

// RecordHTTPRequest 记录一次 HTTP 请求。
func (m *Metrics) RecordHTTPRequest(method, route string, status int, duration time.Duration) {
	if route == "" {
		route = "unknown"
	}
	key := httpMetricKey{
		Method: method,
		Route:  route,
		Status: status,
	}
	seconds := duration.Seconds()

	m.mu.Lock()
	defer m.mu.Unlock()

	value, ok := m.http[key]
	if !ok {
		value = &httpMetricValue{
			Buckets: make([]int64, len(m.options.Metrics.Buckets)),
		}
		m.http[key] = value
	}

	value.Count++
	value.Sum += seconds
	for i, bucket := range m.options.Metrics.Buckets {
		if duration <= bucket {
			value.Buckets[i]++
		}
	}
}

// PrometheusText 导出 Prometheus text exposition 格式。
func (m *Metrics) PrometheusText() string {
	m.mu.RLock()
	httpSnapshot := make(map[httpMetricKey]httpMetricValue, len(m.http))
	for k, v := range m.http {
		httpSnapshot[k] = httpMetricValue{
			Count:   v.Count,
			Sum:     v.Sum,
			Buckets: append([]int64(nil), v.Buckets...),
		}
	}
	inFlight := m.inFlight
	panics := m.panics
	uptime := time.Since(m.start).Seconds()
	m.mu.RUnlock()

	var b strings.Builder
	ns := m.options.Metrics.Namespace

	fmt.Fprintf(&b, "# HELP %s_uptime_seconds Application uptime in seconds.\n", ns)
	fmt.Fprintf(&b, "# TYPE %s_uptime_seconds gauge\n", ns)
	fmt.Fprintf(&b, "%s_uptime_seconds %.3f\n", ns, uptime)

	fmt.Fprintf(&b, "# HELP %s_http_in_flight_requests Current in-flight HTTP requests.\n", ns)
	fmt.Fprintf(&b, "# TYPE %s_http_in_flight_requests gauge\n", ns)
	fmt.Fprintf(&b, "%s_http_in_flight_requests %d\n", ns, inFlight)

	fmt.Fprintf(&b, "# HELP %s_panics_total Total recovered panics.\n", ns)
	fmt.Fprintf(&b, "# TYPE %s_panics_total counter\n", ns)
	fmt.Fprintf(&b, "%s_panics_total %d\n", ns, panics)

	keys := make([]httpMetricKey, 0, len(httpSnapshot))
	for key := range httpSnapshot {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Route != keys[j].Route {
			return keys[i].Route < keys[j].Route
		}
		if keys[i].Method != keys[j].Method {
			return keys[i].Method < keys[j].Method
		}
		return keys[i].Status < keys[j].Status
	})

	fmt.Fprintf(&b, "# HELP %s_http_requests_total Total HTTP requests.\n", ns)
	fmt.Fprintf(&b, "# TYPE %s_http_requests_total counter\n", ns)
	for _, key := range keys {
		value := httpSnapshot[key]
		fmt.Fprintf(&b, "%s_http_requests_total{%s} %d\n", ns, httpLabels(key), value.Count)
	}

	fmt.Fprintf(&b, "# HELP %s_http_request_duration_seconds HTTP request duration in seconds.\n", ns)
	fmt.Fprintf(&b, "# TYPE %s_http_request_duration_seconds histogram\n", ns)
	for _, key := range keys {
		value := httpSnapshot[key]
		for i, bucket := range m.options.Metrics.Buckets {
			fmt.Fprintf(&b, "%s_http_request_duration_seconds_bucket{%s,le=\"%.3f\"} %d\n",
				ns, httpLabels(key), bucket.Seconds(), value.Buckets[i])
		}
		fmt.Fprintf(&b, "%s_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n",
			ns, httpLabels(key), value.Count)
		fmt.Fprintf(&b, "%s_http_request_duration_seconds_sum{%s} %.6f\n", ns, httpLabels(key), value.Sum)
		fmt.Fprintf(&b, "%s_http_request_duration_seconds_count{%s} %d\n", ns, httpLabels(key), value.Count)
	}

	if m.options.Metrics.IncludeRuntime {
		appendRuntimeMetrics(&b, ns)
	}

	return b.String()
}

func httpLabels(key httpMetricKey) string {
	return fmt.Sprintf("method=\"%s\",route=\"%s\",status=\"%d\"",
		escapeLabelValue(key.Method),
		escapeLabelValue(key.Route),
		key.Status,
	)
}

func escapeLabelValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}

func appendRuntimeMetrics(b *strings.Builder, namespace string) {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	fmt.Fprintf(b, "# HELP %s_runtime_goroutines Current goroutine count.\n", namespace)
	fmt.Fprintf(b, "# TYPE %s_runtime_goroutines gauge\n", namespace)
	fmt.Fprintf(b, "%s_runtime_goroutines %d\n", namespace, runtime.NumGoroutine())

	fmt.Fprintf(b, "# HELP %s_runtime_alloc_bytes Current heap allocation bytes.\n", namespace)
	fmt.Fprintf(b, "# TYPE %s_runtime_alloc_bytes gauge\n", namespace)
	fmt.Fprintf(b, "%s_runtime_alloc_bytes %d\n", namespace, stats.Alloc)

	fmt.Fprintf(b, "# HELP %s_runtime_sys_bytes Total bytes obtained from the OS.\n", namespace)
	fmt.Fprintf(b, "# TYPE %s_runtime_sys_bytes gauge\n", namespace)
	fmt.Fprintf(b, "%s_runtime_sys_bytes %d\n", namespace, stats.Sys)

	fmt.Fprintf(b, "# HELP %s_runtime_gc_total Total completed GC cycles.\n", namespace)
	fmt.Fprintf(b, "# TYPE %s_runtime_gc_total counter\n", namespace)
	fmt.Fprintf(b, "%s_runtime_gc_total %d\n", namespace, stats.NumGC)
}
