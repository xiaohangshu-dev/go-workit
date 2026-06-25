// Package main 演示 DDD 领域事件总线在 go-workit 框架中的完整使用方式。
//
// 示例场景：订单创建与发货流程
//   - 多个领域事件（OrderCreated、OrderShipped）
//   - 同一事件多个处理器（订单服务 + 审计日志）
//   - 通过 Builder API 一行注册
//   - 通过 BackgroundService 自动发布事件
package main

import (
	"github.com/xiaohangshu-dev/go-workit/examples/ddd/order"
	"github.com/xiaohangshu-dev/go-workit/examples/ddd/service"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"github.com/xiaohangshu-dev/go-workit/pkg/ddd"
)

func main() {
	builder := app.NewBuilder()

	// 注册多个事件的处理器
	builder.AddDomainEventBus(
		ddd.RegisterDomainEventHandlers[order.OrderCreated](
			order.NewOrderCreatedHandler, // 业务处理
			order.NewOrderAuditHandler,   // 审计日志（同一事件另一个处理器）
		),
		ddd.RegisterDomainEventHandlers[order.OrderShipped](
			order.NewOrderShippedHandler, // 处理 OrderShipped 事件
		),
	)

	// 注册后台服务，启动时自动发布领域事件
	builder.AddBackgroundService(service.NewDemoService)

	// 构建并运行
	builder.Build().Run()
}
