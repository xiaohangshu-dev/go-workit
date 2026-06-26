package httpclient

import "go.uber.org/fx"

// ContextOptions 管理多个 HTTP 客户端注册。
type ContextOptions struct {
	container []fx.Option
	clientMap map[string]struct{}
}

// NewContextOptions 创建 HTTP 客户端上下文选项。
func NewContextOptions() *ContextOptions {
	return &ContextOptions{
		container: make([]fx.Option, 0),
		clientMap: make(map[string]struct{}),
	}
}

// UseHttpClient 注册一个 HTTP 客户端。
// 若 instanceName 为空，则默认注册为 "default"，可通过 DI 直接注入 *Client。
// 若 instanceName 非空且不为 "default"，则注册为命名客户端，通过 fx.In + name 标签注入。
func (o *ContextOptions) UseHttpClient(instanceName string, fn func(*Options)) *ContextOptions {
	if instanceName == "" {
		instanceName = "default"
	}
	if _, ok := o.clientMap[instanceName]; ok {
		panic("http client instance name already exists")
	}

	opts := NewOptions()
	if fn != nil {
		fn(opts)
	}

	client := NewClient(opts)
	if instanceName == "default" {
		o.container = append(o.container,
			fx.Provide(func() *Client { return client }),
		)
	} else {
		o.container = append(o.container,
			fx.Provide(
				fx.Annotate(
					func() *Client { return client },
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}
	o.clientMap[instanceName] = struct{}{}

	return o
}

// Container 返回 HTTP 客户端容器注册。
func (o *ContextOptions) Container() []fx.Option {
	return o.container
}
