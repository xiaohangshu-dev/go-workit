// Package main 演示 HTTP 客户端的两种使用方式。
//
// ## 方式一：单个客户端
//
//	builder.AddHttpClient(func(options *httpclient.Options) {
//	    options.BaseURL = "https://api.github.com"
//	})
//	// 直接注入 *httpclient.Client
//	type Service struct { Client *httpclient.Client }
//
// ## 方式二：多个命名客户端
//
//	builder.AddNamedHttpClient("github", func(o *httpclient.Options) { ... })
//	builder.AddNamedHttpClient("openai", func(o *httpclient.Options) { ... })
//	// 注入 *httpclient.Provider，通过 Get("名称") 获取
//	type Service struct { Clients *httpclient.Provider }
//	func (s *Service) DoWork() {
//	    s.Clients.Get("github").Get(ctx, "/users", &result)
//	}
package main

import (
	"time"

	"github.com/xiaohangshu-dev/go-workit/examples/httpclient/service"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	httpclient "github.com/xiaohangshu-dev/go-workit/pkg/tools/net"
	"go.uber.org/fx"
)

func main() {
	builder := app.NewBuilder()

	// 方式一：注册默认客户端（直接注入 *httpclient.Client）
	// builder.AddHttpClient(func(options *httpclient.Options) {
	//     options.BaseURL = "https://api.github.com"
	// })

	// 方式二：注册多个命名客户端（通过 Provider 获取）
	builder.AddNamedHttpClient("github", func(options *httpclient.Options) {
		options.BaseURL = "https://api.github.com"
		options.Timeout = 10 * time.Second
	})
	builder.AddNamedHttpClient("placeholder", func(options *httpclient.Options) {
		options.BaseURL = "https://jsonplaceholder.typicode.com"
		options.Timeout = 10 * time.Second
	})

	builder.AddServices(fx.Provide(service.NewUserService))
	builder.AddBackgroundService(service.NewDemoService)

	builder.Build().Run()
}
