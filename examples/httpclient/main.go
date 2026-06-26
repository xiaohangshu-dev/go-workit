// Package main 演示 HTTP 客户端上下文的使用方式。
//
//	builder.AddHttpClientContext(func(opts *httpclient.ContextOptions) {
//	    opts.UseHttpClient("default", func(o *httpclient.Options) { ... })
//	    opts.UseHttpClient("placeholder", func(o *httpclient.Options) { ... })
//	})
//	// default 直接注入 *httpclient.Client，命名客户端通过 fx.In + name 注入。
//	type Clients struct {
//	    fx.In
//	    Placeholder *httpclient.Client `name:"placeholder"`
//	}
package main

import (
	"time"

	"github.com/xiaohangshu-dev/go-workit/examples/httpclient/service"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"github.com/xiaohangshu-dev/go-workit/pkg/tools/httpclient"
	"go.uber.org/fx"
)

func main() {
	builder := app.NewBuilder()

	builder.AddHttpClientContext(func(opts *httpclient.ContextOptions) {
		opts.UseHttpClient("default", func(options *httpclient.Options) {
			options.BaseURL = "https://api.github.com"
			options.Timeout = 10 * time.Second
		})
		opts.UseHttpClient("placeholder", func(options *httpclient.Options) {
			options.BaseURL = "https://jsonplaceholder.typicode.com"
			options.Timeout = 10 * time.Second
		})
	})

	builder.AddServices(fx.Provide(service.NewUserService))
	builder.AddBackgroundService(service.NewDemoService)

	builder.Build().Run()
}
