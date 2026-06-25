package webapp

import "github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"

// HealthChecksBuilder 表示健康检查注册器。
type HealthChecksBuilder struct {
	builder *WebApplicationBuilder
	options *observability.HealthOptions
}

// AddMetrics 注册指标配置。
func (b *WebApplicationBuilder) AddMetrics(configure ...func(options *observability.MetricsOptions)) *WebApplicationBuilder {
	options := b.ensureObservabilityOptions()
	b.metricsEnabled = true
	if len(configure) != 0 {
		configure[0](&options.Metrics)
	}
	return b
}

// AddOpenTelemetry 注册 OpenTelemetry 链路追踪配置。
func (b *WebApplicationBuilder) AddOpenTelemetry(configure ...func(options *observability.TracingOptions)) *WebApplicationBuilder {
	options := b.ensureObservabilityOptions()
	b.tracingEnabled = true
	if len(configure) != 0 {
		configure[0](&options.Tracing)
	}
	return b
}

// AddHealthChecks 注册健康检查配置。
func (b *WebApplicationBuilder) AddHealthChecks(configure ...func(options *observability.HealthOptions)) *HealthChecksBuilder {
	options := b.ensureObservabilityOptions()
	b.healthEnabled = true
	if len(configure) != 0 {
		configure[0](&options.Health)
	}
	return &HealthChecksBuilder{
		builder: b,
		options: &options.Health,
	}
}

// AddCheck 添加一个健康检查。
func (b *HealthChecksBuilder) AddCheck(name string, check observability.CheckFunc, tags ...string) *HealthChecksBuilder {
	b.options.AddCheck(name, check, tags...)
	return b
}

// Builder 返回应用构建器，方便结束 HealthChecks 链式配置后继续注册其他能力。
func (b *HealthChecksBuilder) Builder() *WebApplicationBuilder {
	return b.builder
}
