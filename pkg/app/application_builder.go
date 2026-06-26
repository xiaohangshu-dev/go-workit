package app

import (
	"os"

	"github.com/spf13/viper"
	"github.com/xiaohangshu-dev/go-workit/pkg/config"
	"github.com/xiaohangshu-dev/go-workit/pkg/ddd"
	"github.com/xiaohangshu-dev/go-workit/pkg/tools/httpclient"
	"go.uber.org/fx"
)

// ApplicationBuilder 应用构建器
type ApplicationBuilder struct {
	config        *viper.Viper   // 配置管理
	options       []fx.Option    // 容器管理
	configBuilder config.Builder // 配置构建
}

// NewBuilder 创建一个新的应用构建器
func NewBuilder() *ApplicationBuilder {

	// 创建一个新的 viper 实例
	viper := viper.New()

	// 设置 app 默认值
	viper.SetDefault("app.name", "go-workit")

	// 设置 logger 默认值
	viper.SetDefault("log.level", "info")              // 默认日志级别为 info
	viper.SetDefault("log.filename", "./logs/app.log") // 默认不输出到文件
	viper.SetDefault("log.max_size", 100)              // 默认单文件最大 100 MB
	viper.SetDefault("log.max_backups", 3)             // 默认保留 3 个备份
	viper.SetDefault("log.max_age", 7)                 // 默认日志保留 7 天
	viper.SetDefault("log.compress", true)             // 默认不压缩旧日志
	viper.SetDefault("log.console", true)              // 默认输出到控制台

	// 创建配置构建器
	configBuilder := config.NewBuilder(viper)

	// 当前目录下存在 application.yaml 文件，则加载该文件
	if _, err := os.Stat("./application.yaml"); err == nil {
		configBuilder.AddYamlFile("./application.yaml")
	}

	// 加载环境变量
	configBuilder.AddEnvironmentVariables()

	// 加载命令行参数
	configBuilder.AddCommandLine()

	return &ApplicationBuilder{
		config:        viper,
		options:       make([]fx.Option, 0),
		configBuilder: configBuilder,
	}
}

// AddConfig 用户加载配置文件、环境变量、命令行参数。
// 配置添加后即生效,priority: 命令行 > 环境变量 > 配置文件
func (b *ApplicationBuilder) AddConfig(fn func(options *config.Options)) *ApplicationBuilder {

	opts := config.NewOptions(b.configBuilder)
	// 加载文件配置
	fn(opts)

	return b
}

// AddServices 服务注册
func (b *ApplicationBuilder) AddServices(opts ...fx.Option) *ApplicationBuilder {
	b.options = append(b.options, opts...)
	return b
}

// Build 构建应用实例
func (b *ApplicationBuilder) Build() *Application {

	// 配置 logger
	logger := newLogger(&Config{
		Level:      b.config.GetString("log.level"),    // 从配置文件/env/命令行拿
		Filename:   b.config.GetString("log.filename"), // 如果有配置文件路径则输出到文件
		MaxSize:    b.config.GetInt("log.max_size"),    // 单文件最大多少 MB，默认100
		MaxBackups: b.config.GetInt("log.max_backups"), // 保留几份备份，默认3
		MaxAge:     b.config.GetInt("log.max_age"),     // 最老的日志保留多少天，默认7
		Compress:   b.config.GetBool("log.compress"),   // 旧日志是否压缩，默认不开
		Console:    b.config.GetBool("log.console"),    // 是否同时输出到控制台，开发环境一般要 true
	})

	return NewApplication(b.options, b.config, logger)
}

// AddBackgroundService 添加后台服务
func (b *ApplicationBuilder) AddBackgroundService(ctor any) *ApplicationBuilder {
	b.options = append(b.options, fx.Provide(
		fx.Annotate(
			ctor,
			fx.ResultTags(`group:"background_services"`),
		),
	))
	return b
}

// ConfigureOptions 配置选项(暂未实现)
func (b *ApplicationBuilder) ConfigureOptions(provider any) *ApplicationBuilder {
	b.options = append(b.options, fx.Provide(provider))
	return b
}

// Config 返回配置实例,配置阶段也可读取配置
func (b *ApplicationBuilder) Config() *viper.Viper {
	return b.config
}

// AddDomainEventBus 注册领域事件总线。
func (b *ApplicationBuilder) AddDomainEventBus(eventHandlerRegistrations ...fx.Option) *ApplicationBuilder {
	return b.AddServices(ddd.DomainEventBusModule(eventHandlerRegistrations...))
}

// AddHttpClientContext 注册 HTTP 客户端上下文。
//
// 适用于统一注册一个或多个 HTTP 客户端，使用方式与 AddGormContext 保持一致。
//
//	builder.AddHttpClientContext(func(opts *httpclient.ContextOptions) {
//	    opts.UseHttpClient("default", func(options *httpclient.Options) {
//	        options.BaseURL = "https://api.github.com"
//	    })
//	    opts.UseHttpClient("placeholder", func(options *httpclient.Options) {
//	        options.BaseURL = "https://jsonplaceholder.typicode.com"
//	    })
//	})
func (b *ApplicationBuilder) AddHttpClientContext(configure func(options *httpclient.ContextOptions)) *ApplicationBuilder {
	opts := httpclient.NewContextOptions()
	if configure != nil {
		configure(opts)
	}
	b.AddServices(opts.Container()...)
	return b
}
