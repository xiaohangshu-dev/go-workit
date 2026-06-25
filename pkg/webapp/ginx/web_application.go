package ginx

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/rpc"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/observability"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// WebApplication 实现 WebApplication 接口
type WebApplication struct {
	*app.Application
	routeRegistrations      []any
	grpcServiceConstructors []any
	handler                 http.Handler
	server                  *http.Server
	ServerOptions           *web.ServerConfig
	env                     *web.Environment
	observability           *observability.Metrics
	health                  *observability.HealthRegistry
	telemetry               *observability.Telemetry
}

// NewWebApplication 创建一个 WebApplication 实例
func NewWebApplication(app *app.Application, components ...any) web.Application {
	serverOptions := &web.ServerConfig{}
	var observabilityMetrics *observability.Metrics
	var healthRegistry *observability.HealthRegistry
	var telemetry *observability.Telemetry
	for _, component := range components {
		switch value := component.(type) {
		case *observability.Metrics:
			observabilityMetrics = value
		case *observability.HealthRegistry:
			healthRegistry = value
		case *observability.Telemetry:
			telemetry = value
		}
	}

	// 1. http_port 默认 8080
	httpPort := app.Config().GetInt("server.http_port")
	if httpPort == 0 {
		httpPort = 8080
	}
	if httpPort <= 0 || httpPort > 65535 {
		panic(fmt.Sprintf("invalid http_port: %d", httpPort))
	}
	serverOptions.HttpPort = strconv.Itoa(httpPort)

	// 2. grpc_port 默认 50051
	grpcPort := app.Config().GetInt("server.grpc_port")
	if grpcPort == 0 {
		grpcPort = 50051
	}
	if grpcPort <= 0 || grpcPort > 65535 {
		panic(fmt.Sprintf("invalid grpc_port: %d", grpcPort))
	}
	serverOptions.GrpcPort = strconv.Itoa(grpcPort)

	// 3. environment 默认 prod
	environment := strings.ToLower(app.Config().GetString("server.environment"))
	if environment == "" {
		environment = "prod"
	}
	switch environment {
	case web.Development, web.Testing, web.Production:
		serverOptions.Environment = environment
	default:
		panic("invalid environment: " + environment)
	}

	// 4. 设置 Gin 模式
	switch serverOptions.Environment {
	case web.Development:
		gin.SetMode(gin.DebugMode)
	case web.Testing:
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	env := &web.Environment{
		Env:           serverOptions.Environment,
		IsDevelopment: serverOptions.Environment == web.Development,
	}

	e := gin.New()

	// 5. recover 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultRecover = !app.Config().IsSet("server.use_default_recover") ||
		app.Config().GetBool("server.use_default_recover"); serverOptions.UseDefaultRecover {

		e.Use(newRecoveryWithZap(app.Logger(), observabilityMetrics))
	}

	// 6. logger 默认启用（除非明确配置为 false）
	if serverOptions.UseDefaultLogger = !app.Config().IsSet("server.use_default_logger") ||
		app.Config().GetBool("server.use_default_logger"); serverOptions.UseDefaultLogger {

		e.Use(newZapLogger(app.Logger()))
	}

	webapp := &WebApplication{
		handler:       e,
		ServerOptions: serverOptions,
		env:           env,
		Application:   app,
		observability: observabilityMetrics,
		health:        healthRegistry,
		telemetry:     telemetry,
	}
	webapp.useObservability()
	return webapp
}

// Run 启动 Web 应用程序
func (webapp *WebApplication) Run(params ...string) {
	// HTTP server
	webapp.server = &http.Server{
		Addr:         ":" + webapp.ServerOptions.HttpPort,
		Handler:      webapp.handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Fx 容器配置
	webapp.AppendContainer(
		fx.Supply(webapp.handler.(*gin.Engine)),

		// HTTP 生命周期管理
		fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						logger.Info("HTTP server starting",
							zap.String("port", webapp.ServerOptions.HttpPort))
						if err := webapp.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
							logger.Error("HTTP server error", zap.Error(err))
							_ = shutdowner.Shutdown()
						}
					}()
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Shutting down HTTP server")
					return webapp.server.Shutdown(ctx)
				},
			})
		}),
	)

	// gRPC server 生命周期管理（如果启用）
	if len(webapp.grpcServiceConstructors) > 0 {
		webapp.AppendContainer(
			fx.Provide(func() *grpc.Server {
				return grpc.NewServer()
			}),
			fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner, logger *zap.Logger, grpcSrv *grpc.Server) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						go func() {
							addr := ":" + webapp.ServerOptions.GrpcPort
							lis, err := net.Listen("tcp", addr)
							if err != nil {
								logger.Error("Failed to listen on GRPC port", zap.Error(err))
								_ = shutdowner.Shutdown()
								return
							}
							logger.Info("GRPC server starting", zap.String("port", webapp.ServerOptions.GrpcPort))
							if err := grpcSrv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
								logger.Error("GRPC server error", zap.Error(err))
								_ = shutdowner.Shutdown()
							}
						}()
						return nil
					},
					OnStop: func(ctx context.Context) error {
						logger.Info("Stopping GRPC server")
						stopped := make(chan struct{})
						go func() {
							grpcSrv.GracefulStop()
							close(stopped)
						}()
						select {
						case <-stopped:
							return nil
						case <-ctx.Done():
							grpcSrv.Stop() // 强制关闭
							return ctx.Err()
						}
					},
				})
			}),
		)

	}

	// 注册 HTTP 路由
	for _, r := range webapp.routeRegistrations {
		webapp.AppendContainer(fx.Invoke(r))
	}

	// 构建并运行 Fx 应用
	fxapp := webapp.FxApp(fx.New(webapp.Container()...))

	for _, group := range webapp.engine().RouterMap {

		webapp.engine().AddRoute(group.Method, group.Path, group.Handlers, group.AuthSchemes, group.AuthzPolicies, group.LimitersPolices, group.AllowAnonymous)

	}

	// 直接使用 Fx 的 Run 来管理生命周期和信号
	webapp.Logger().Info("Starting application...")
	fxapp.Run()
	webapp.Logger().Info("Application stopped gracefully")
}

// MapRoute 注册路由
func (a *WebApplication) MapRoute(routeFuncList ...any) web.Application {

	for _, routeFunc := range routeFuncList {

		t := reflect.TypeOf(routeFunc)

		if t.Kind() != reflect.Func {
			panic("registerFunc must be a function")
		}

		ginType := reflect.TypeOf(&gin.Engine{})
		hasGin := false

		for i := 0; i < t.NumIn(); i++ {
			if t.In(i) == ginType {
				hasGin = true
				break
			}
		}

		if !hasGin {
			panic("registerFunc must have at least one parameter of type *gin.Engine")
		}

		a.routeRegistrations = append(a.routeRegistrations, routeFunc)
	}

	return a
}

// UseSwagger 配置Swagger
func (a *WebApplication) UseSwagger() web.Application {
	a.engine().GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	return a
}

// UseCORS 配置跨域
func (a *WebApplication) UseCORS(fn any) web.Application {
	exec, ok := fn.(func(*cors.Config))
	if !ok {
		panic("UseCORS: argument must be func(*cors.Config)")
	}

	cfg := cors.DefaultConfig()
	exec(&cfg)

	a.engine().Use(cors.New(cfg))
	return a
}

// UseStaticFiles 配置静态文件
func (a *WebApplication) UseStaticFiles(urlPath, root string) web.Application {
	a.engine().Static(urlPath, root)
	return a
}

// UseHealthCheck 配置健康检查
func (a *WebApplication) UseHealthCheck() web.Application {
	a.engine().GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return a
}

func (a *WebApplication) useObservability() {
	if a.telemetry != nil && a.telemetry.Enabled() {
		a.engine().Use(otelgin.Middleware(a.telemetry.Options().ServiceName))
	}

	if a.observability != nil {
		a.engine().Use(newObservability(a.observability).Handle())
	}
}

// MapMetrics 映射 Prometheus 指标端点，使用标准 promhttp handler。
func (a *WebApplication) MapMetrics(path ...string) web.Application {
	metrics := a.observability
	if metrics == nil {
		return a
	}

	routePath := firstPath(metrics.Options().Metrics.Path, path...)
	handler := metrics.PrometheusHandler()
	a.engine().GET(routePath, func(c *gin.Context) {
		handler.ServeHTTP(c.Writer, c.Request)
	}).WithAllowAnonymous()
	return a
}

// MapHealthChecks 映射健康检查端点，包含完整检查、就绪检查和存活检查。
func (a *WebApplication) MapHealthChecks(path ...string) web.Application {
	health := a.health
	if health == nil {
		return a
	}

	a.mapHealthEndpoint(firstPath(health.Options().Health.HealthPath, path...), func(c *gin.Context) {
		report := health.Run(c.Request.Context())
		c.JSON(healthStatusCode(report.Status), report)
	})
	a.mapHealthEndpoint(health.Options().Health.ReadinessPath, func(c *gin.Context) {
		report := health.Readiness(c.Request.Context())
		c.JSON(healthStatusCode(report.Status), report)
	})
	a.mapHealthEndpoint(health.Options().Health.LivenessPath, func(c *gin.Context) {
		report := health.Liveness(c.Request.Context())
		c.JSON(healthStatusCode(report.Status), report)
	})
	return a
}

// MapReadinessChecks 映射就绪检查端点。
func (a *WebApplication) MapReadinessChecks(path ...string) web.Application {
	health := a.health
	if health == nil {
		return a
	}

	a.mapHealthEndpoint(firstPath(health.Options().Health.ReadinessPath, path...), func(c *gin.Context) {
		report := health.Readiness(c.Request.Context())
		c.JSON(healthStatusCode(report.Status), report)
	})
	return a
}

// MapLivenessChecks 映射存活检查端点。
func (a *WebApplication) MapLivenessChecks(path ...string) web.Application {
	health := a.health
	if health == nil {
		return a
	}

	a.mapHealthEndpoint(firstPath(health.Options().Health.LivenessPath, path...), func(c *gin.Context) {
		report := health.Liveness(c.Request.Context())
		c.JSON(healthStatusCode(report.Status), report)
	})
	return a
}

func (a *WebApplication) mapHealthEndpoint(path string, handler gin.HandlerFunc) {
	a.engine().GET(path, handler).WithAllowAnonymous()
}

func firstPath(defaultPath string, paths ...string) string {
	if len(paths) > 0 && paths[0] != "" {
		return paths[0]
	}
	return defaultPath
}

func healthStatusCode(status observability.HealthStatus) int {
	if status == observability.HealthStatusUnhealthy {
		return http.StatusServiceUnavailable
	}
	return http.StatusOK
}

func (a *WebApplication) engine() *gin.Engine {
	return a.handler.(*gin.Engine)
}

// MapGrpcServices 注册 gRPC 服务
func (webapp *WebApplication) MapGrpcServices(constructors ...any) web.Application {
	for _, constructor := range constructors {
		webapp.grpcServiceConstructors = append(webapp.grpcServiceConstructors, constructor)
		webapp.AppendContainer(fx.Provide(constructor))

		// 推断构造函数的返回类型
		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("MapGrpcServices: constructor must be a function with at least one return value")
		}

		serviceType := constructorType.Out(0)

		// 对每个具体服务构造出一个 fx.Invoke
		invokeFn := makeGrpcInvoke(serviceType, webapp.Logger())
		webapp.AppendContainer(fx.Invoke(invokeFn))
	}

	return webapp
}

func makeGrpcInvoke(serviceType reflect.Type, logger *zap.Logger) any {
	// 构造函数类型：func(*grpc.Server, <YourServiceType>)
	fnType := reflect.FuncOf(
		[]reflect.Type{reflect.TypeOf((*grpc.Server)(nil)), serviceType}, // 入参类型
		[]reflect.Type{}, // 返回值类型为空
		false,            // 非变长参数
	)

	// 构造函数实现
	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		server := args[0].Interface().(*grpc.Server)
		svc := args[1].Interface()

		grpcSvc, ok := svc.(rpc.GrpcService)
		if !ok {
			panic(fmt.Sprintf("MapGrpcServices: %s does not implement GrpcService", reflect.TypeOf(svc)))
		}

		grpcSvc.Register(server)
		logger.Info("Registered gRPC service", zap.String("type", reflect.TypeOf(svc).String()))

		return nil
	})

	return fn.Interface()
}

// Use 注册中间件
func (b *WebApplication) Use(middleware ...any) web.Application {
	for _, constructor := range middleware {
		b.AppendContainer(fx.Provide(constructor))

		constructorType := reflect.TypeOf(constructor)
		if constructorType.Kind() != reflect.Func || constructorType.NumOut() == 0 {
			panic("UseMiddlewareDI: constructor must be a function that returns Middleware")
		}

		middlewareType := constructorType.Out(0)

		// 生成 fx.Invoke(fn(mwType, *gin.Engine))
		b.AppendContainer(fx.Invoke(makeMiddlewareInvoke(middlewareType)))
	}
	return b
}

func makeMiddlewareInvoke(middlewareType reflect.Type) any {
	fnType := reflect.FuncOf(
		[]reflect.Type{middlewareType, reflect.TypeOf((*gin.Engine)(nil))},
		[]reflect.Type{},
		false,
	)

	fn := reflect.MakeFunc(fnType, func(args []reflect.Value) []reflect.Value {
		mwVal := args[0]
		engine := args[1].Interface().(*gin.Engine)

		mw, ok := mwVal.Interface().(Middleware)
		if !ok {
			panic(fmt.Sprintf("type %v does not implement Middleware", mwVal.Type()))
		}

		engine.Use(mw.Handle())

		return nil
	})

	return fn.Interface()
}

// UseAuthentication 鉴权中间件
func (a *WebApplication) UseAuthentication() web.Application {
	a.Use(newAuthenticate)
	return a
}

// UseAuthorization 授权中间件
func (a *WebApplication) UseAuthorization() web.Application {
	a.Use(newAuthorize)
	return a
}

// Env 获取环境实例
func (a *WebApplication) Env() *web.Environment {
	return a.env
}

// UseRecovery 注册异常恢复中间件, 用于捕获 panic
func (a *WebApplication) UseRecovery() web.Application {
	a.engine().Use(newRecoveryWithZap(a.Logger()))
	return a
}

// UseLogger 注册日志中间件, 用于记录请求日志
func (a *WebApplication) UseLogger() web.Application {
	a.engine().Use(newZapLogger(a.Logger()))
	return a
}

// UseLocalization 配置国际化功能
func (a *WebApplication) UseLocalization() web.Application {
	a.Use(newLocalization)
	return a
}

// UseRateLimiter 配置限流功能
func (a *WebApplication) UseRateLimiter() web.Application {
	a.Use(newRateLimiter)
	return a
}

// UseReqDecomp 配置请求解压
func (a *WebApplication) UseRequestDecompression() web.Application {
	a.Use(newDecompression)
	return a
}
