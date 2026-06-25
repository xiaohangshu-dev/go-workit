// Package service 演示后台服务中发布领域事件。
package service

import (
	"context"
	"time"

	"github.com/mehdihadeli/go-mediatr"
	"github.com/xiaohangshu-dev/go-workit/examples/ddd/order"
	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"github.com/xiaohangshu-dev/go-workit/pkg/ddd"
	"go.uber.org/zap"
)

// DemoService 在应用启动时发布领域事件。
type DemoService struct {
	bus    *ddd.DomainEventBus
	logger *zap.Logger
}

// NewDemoService 构造函数。
func NewDemoService(bus *ddd.DomainEventBus, logger *zap.Logger) app.BackgroundService {
	return &DemoService{bus: bus, logger: logger}
}

// Start 创建订单并发布领域事件。
func (s *DemoService) Start(ctx context.Context) error {
	s.logger.Info("=== DDD 领域事件总线示例开始 ===")

	// 1. 创建订单 → 产生 OrderCreated 事件
	o := order.NewOrderAggregate("ORD-2024001", "CUST-9527", 299.99)
	s.logger.Info("订单已创建", zap.String("id", o.ID))

	// 2. 发货操作 → 产生 OrderShipped 事件
	o.Ship("SF-1234567890")
	s.logger.Info("订单已发货", zap.String("tracking", "SF-1234567890"))

	time.Sleep(10 * time.Millisecond)

	// 3. 逐个发布事件（go-mediatr 需要具体类型匹配 handler）
	s.logger.Info("正在发布领域事件...")
	for _, event := range o.GetDomainEvents() {
		switch e := event.(type) {
		case order.OrderCreated:
			mediatr.Publish(ctx, e)
		case order.OrderShipped:
			mediatr.Publish(ctx, e)
		}
	}
	o.ClearDomainEvents()
	s.logger.Info("领域事件发布完成")

	s.logger.Info("=== DDD 领域事件总线示例结束 ===")
	return nil
}

// Stop 应用停止时清理。
func (s *DemoService) Stop(ctx context.Context) error {
	return nil
}
