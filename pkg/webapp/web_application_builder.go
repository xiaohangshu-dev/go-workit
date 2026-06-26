package webapp

import (
	"context"
	"fmt"

	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"github.com/xiaohangshu-dev/go-workit/pkg/tools/httpclient"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/authz"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/dbctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/elasticctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/esctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/gormctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/kafkactx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/minioctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/mongoctx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/redisctx"

	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/ginx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/localiza"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/reqdecp"

	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/ratelimit"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/router"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"
	"go.uber.org/fx"
)

// WebApplicationBuilder 构建web应用
type WebApplicationBuilder struct {
	*app.ApplicationBuilder
	app               *app.Application
	authOpts          *auth.Options
	authzOpts         *authz.Options
	localizaOpts      *localiza.Options
	rateLimitOpts     *ratelimit.Options
	reqdecpOpts       *reqdecp.Options
	observability     *observability.Metrics
	health            *observability.HealthRegistry
	telemetry         *observability.Telemetry
	observabilityOpts *observability.Options
	metricsEnabled    bool
	tracingEnabled    bool
	healthEnabled     bool
	router            *router.Router
}

// NewWebAppBuilder 创建WebApplicationBuilder
func NewBuilder() *WebApplicationBuilder {

	hostBuild := app.NewBuilder()

	return &WebApplicationBuilder{
		ApplicationBuilder: hostBuild,
	}
}

// AddAuthentication 添加鉴权方案
func (b *WebApplicationBuilder) AddAuthentication(fn func(options *auth.Options)) *WebApplicationBuilder {

	if b.authOpts == nil {

		b.authOpts = auth.NewOptions()
	}

	fn(b.authOpts)

	if b.authOpts.DefaultScheme == "" {
		panic("default scheme is required")
	}

	return b
}

// AddAuthorization 添加授权策略
func (b *WebApplicationBuilder) AddAuthorization(fn func(options *authz.Options)) *WebApplicationBuilder {

	if b.authzOpts == nil {

		b.authzOpts = authz.NewOptions()
	}

	fn(b.authzOpts)

	return b
}

// AddMongo 添加数据库配置
func (b *WebApplicationBuilder) AddMinioContext(fn func(options *minioctx.Options)) *WebApplicationBuilder {

	opts := minioctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddMongo 添加数据库配置
func (b *WebApplicationBuilder) AddMongoContext(fn func(options *mongoctx.Options)) *WebApplicationBuilder {

	opts := mongoctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddGorm 添加Gorm数据库
func (b *WebApplicationBuilder) AddGormContext(fn func(options *gormctx.Options)) *WebApplicationBuilder {

	opts := gormctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddDbContext 添加数据库上下文
func (b *WebApplicationBuilder) AddDbContext(fn func(options *dbctx.Options)) *WebApplicationBuilder {

	opts := dbctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddHttpClientContext 注册 HTTP 客户端上下文。
func (b *WebApplicationBuilder) AddHttpClientContext(fn func(options *httpclient.ContextOptions)) *WebApplicationBuilder {
	b.ApplicationBuilder.AddHttpClientContext(fn)
	return b
}

// AddRedis 添加Redis配置
func (b *WebApplicationBuilder) AddRedisContext(fn func(options *redisctx.Options)) *WebApplicationBuilder {

	opts := redisctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddEafka 添加Kafka配置
func (b *WebApplicationBuilder) AddKafkaContext(fn func(options *kafkactx.Options)) *WebApplicationBuilder {

	opts := kafkactx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddEs 添加ES配置
func (b *WebApplicationBuilder) AddElasticSearchContext(fn func(options *elasticctx.Options)) *WebApplicationBuilder {

	opts := elasticctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}
func (b *WebApplicationBuilder) AddEsContext(fn func(options *esctx.Options)) *WebApplicationBuilder {

	opts := esctx.NewOptions()

	fn(opts)

	b.ApplicationBuilder.AddServices(opts.Container()...)

	return b
}

// AddLocalization 添加国际化配置
func (b *WebApplicationBuilder) AddLocalization(fn func(options *localiza.Options)) *WebApplicationBuilder {
	opts := localiza.NewOptions()
	fn(opts)
	b.localizaOpts = opts
	return b
}

// AddRateLimiter 添加限流配置
func (b *WebApplicationBuilder) AddRateLimiter(fn func(options *ratelimit.Options)) *WebApplicationBuilder {
	opts := ratelimit.NewOptions()
	fn(opts)
	b.rateLimitOpts = opts
	return b
}

// AddReqDecomp 添加请求解压配置
func (b *WebApplicationBuilder) AddRequestDecompression(fn ...func(options *reqdecp.Options)) *WebApplicationBuilder {
	opts := reqdecp.NewOptions()
	if len(fn) != 0 {
		fn[0](opts)
	}
	b.reqdecpOpts = opts
	return b
}

// Build 构建应用
func (b *WebApplicationBuilder) Build(fn ...func(b *WebApplicationBuilder) web.Application) web.Application {
	// 构建应用主机
	b.app = b.ApplicationBuilder.Build()

	// 初始化默认选项
	if b.authOpts == nil {
		b.authOpts = auth.NewOptions()
	}
	if b.authzOpts == nil {
		b.authzOpts = authz.NewOptions()
	}
	if b.rateLimitOpts == nil {
		b.rateLimitOpts = ratelimit.NewOptions()
	}
	if b.reqdecpOpts == nil {
		b.reqdecpOpts = reqdecp.NewOptions()
	}
	if b.observabilityOpts == nil {
		b.observabilityOpts = observability.NewOptions()
	}

	if b.metricsEnabled {
		b.observability = observability.NewMetrics(b.observabilityOpts)
		b.app.AppendContainer(fx.Supply(b.observability))
	}
	if b.healthEnabled {
		b.health = observability.NewHealthRegistry(b.observabilityOpts)
		b.app.AppendContainer(fx.Supply(b.health))
	}
	if b.tracingEnabled {
		telemetry, err := observability.NewTelemetry(context.Background(), b.observabilityOpts)
		if err != nil {
			panic(fmt.Errorf("build telemetry error: %w", err))
		}
		b.telemetry = telemetry
		b.app.AppendContainer(
			fx.Supply(b.telemetry),
			fx.Invoke(func(lc fx.Lifecycle) {
				lc.Append(fx.Hook{
					OnStop: func(ctx context.Context) error {
						return b.telemetry.Shutdown(ctx)
					},
				})
			}),
		)
	}

	// 构建国际化
	if b.localizaOpts != nil {
		provider, err := localiza.NewBuilder(b.localizaOpts).Build()
		if err != nil {
			panic(fmt.Errorf("build localizer error: %w", err))
		}
		b.app.AppendContainer(fx.Provide(
			func() web.Localization {
				return provider
			}))
	}

	// 构建请求解压
	reqDecompressor := reqdecp.NewReqDecompressor(b.reqdecpOpts.Decompressions())
	b.app.AppendContainer(fx.Provide(func() web.ReqDecompressor {
		return reqDecompressor
	}))

	// 构建路由配置
	b.router = router.NewRouter(b.authOpts, b.authzOpts, b.rateLimitOpts)

	// 将路由配置注入容器供鉴权\授权\限流中间件使用
	b.app.AppendContainer(fx.Provide(func() web.Router {
		return b.router
	}))

	// 构建应用
	if len(fn) > 0 {
		return fn[0](b)
	}

	return ginx.NewWebApplication(b.app, b.observability, b.health, b.telemetry)
}

func (b *WebApplicationBuilder) ensureObservabilityOptions() *observability.Options {
	if b.observabilityOpts == nil {
		b.observabilityOpts = observability.NewOptions()
	}
	if serviceName := b.Config().GetString("app.name"); serviceName != "" {
		b.observabilityOpts.ServiceName = serviceName
	}
	if environment := b.Config().GetString("server.environment"); environment != "" {
		b.observabilityOpts.Environment = environment
	}
	return b.observabilityOpts
}

// App 获取应用实例
func (b *WebApplicationBuilder) App() *app.Application {
	return b.app
}

// Router 获取路由实例
func (b *WebApplicationBuilder) Router() *router.Router {
	return b.router
}

// Observability 获取可观测性指标存储。
func (b *WebApplicationBuilder) Observability() *observability.Metrics {
	return b.observability
}

// HealthChecks 获取健康检查注册表。
func (b *WebApplicationBuilder) HealthChecks() *observability.HealthRegistry {
	return b.health
}

// Telemetry 获取链路追踪上下文。
func (b *WebApplicationBuilder) Telemetry() *observability.Telemetry {
	return b.telemetry
}
