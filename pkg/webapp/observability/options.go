package observability

import (
	"context"
	"time"
)

type TraceExporter string

const (
	TraceExporterNone     TraceExporter = "none"
	TraceExporterStdout   TraceExporter = "stdout"
	TraceExporterOTLPGRPC TraceExporter = "otlp_grpc"
)

type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// CheckFunc 表示一个健康检查函数。
type CheckFunc func(ctx context.Context) CheckResult

// Options 可观测性配置。
type Options struct {
	ServiceName string
	Environment string

	Metrics MetricsOptions
	Tracing TracingOptions
	Health  HealthOptions
}

type MetricsOptions struct {
	Namespace      string
	Path           string
	IncludeRuntime bool
	Buckets        []time.Duration
}

type TracingOptions struct {
	Exporter    TraceExporter
	Endpoint    string
	Insecure    bool
	SampleRatio float64
	Headers     map[string]string
}

type HealthOptions struct {
	LivenessPath  string
	ReadinessPath string
	HealthPath    string
	Timeout       time.Duration
	Checks        []CheckRegistration
}

type CheckRegistration struct {
	Name  string
	Tags  []string
	Check CheckFunc
}

type CheckResult struct {
	Status      HealthStatus `json:"status"`
	Description string       `json:"description,omitempty"`
	Error       string       `json:"error,omitempty"`
}

// Healthy 返回健康检查成功结果。
func Healthy(description ...string) CheckResult {
	return checkResult(HealthStatusHealthy, description...)
}

// Degraded 返回健康检查降级结果。
func Degraded(description ...string) CheckResult {
	return checkResult(HealthStatusDegraded, description...)
}

// Unhealthy 返回健康检查失败结果。
func Unhealthy(description ...string) CheckResult {
	return checkResult(HealthStatusUnhealthy, description...)
}

func checkResult(status HealthStatus, description ...string) CheckResult {
	result := CheckResult{Status: status}
	if len(description) > 0 {
		result.Description = description[0]
	}
	return result
}

// AddCheck 注册健康检查。未指定 tags 时默认加入 ready。
func (o *HealthOptions) AddCheck(name string, check CheckFunc, tags ...string) {
	if len(tags) == 0 {
		tags = []string{"ready"}
	}
	o.Checks = append(o.Checks, CheckRegistration{
		Name:  name,
		Tags:  tags,
		Check: check,
	})
}

// NewOptions 创建默认可观测性配置。
func NewOptions() *Options {
	buckets := defaultBuckets()
	return &Options{
		ServiceName: "go-workit",
		Environment: "development",
		Metrics: MetricsOptions{
			Namespace:      "go_workit",
			Path:           "/metrics",
			IncludeRuntime: true,
			Buckets:        buckets,
		},
		Tracing: TracingOptions{
			Exporter:    TraceExporterNone,
			Endpoint:    "localhost:4317",
			Insecure:    true,
			SampleRatio: 1,
		},
		Health: HealthOptions{
			LivenessPath:  "/live",
			ReadinessPath: "/ready",
			HealthPath:    "/health",
			Timeout:       3 * time.Second,
		},
	}
}

func (o *Options) normalize() {
	defaults := NewOptions()

	if o.ServiceName == "" {
		o.ServiceName = defaults.ServiceName
	}
	if o.Environment == "" {
		o.Environment = defaults.Environment
	}

	if o.Metrics.Namespace == "" {
		o.Metrics.Namespace = defaults.Metrics.Namespace
	}
	if o.Metrics.Path == "" {
		o.Metrics.Path = defaults.Metrics.Path
	}
	if len(o.Metrics.Buckets) == 0 {
		o.Metrics.Buckets = defaultBuckets()
	}

	if o.Tracing.Exporter == "" {
		o.Tracing.Exporter = defaults.Tracing.Exporter
	}
	if o.Tracing.Endpoint == "" {
		o.Tracing.Endpoint = defaults.Tracing.Endpoint
	}
	if o.Tracing.SampleRatio <= 0 || o.Tracing.SampleRatio > 1 {
		o.Tracing.SampleRatio = defaults.Tracing.SampleRatio
	}

	if o.Health.LivenessPath == "" {
		o.Health.LivenessPath = defaults.Health.LivenessPath
	}
	if o.Health.ReadinessPath == "" {
		o.Health.ReadinessPath = defaults.Health.ReadinessPath
	}
	if o.Health.HealthPath == "" {
		o.Health.HealthPath = defaults.Health.HealthPath
	}
	if o.Health.Timeout <= 0 {
		o.Health.Timeout = defaults.Health.Timeout
	}
}

func defaultBuckets() []time.Duration {
	return []time.Duration{
		5 * time.Millisecond,
		10 * time.Millisecond,
		25 * time.Millisecond,
		50 * time.Millisecond,
		100 * time.Millisecond,
		250 * time.Millisecond,
		500 * time.Millisecond,
		time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
	}
}
