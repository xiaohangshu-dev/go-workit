package order

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// OrderCreatedHandler 处理 OrderCreated 事件。
type OrderCreatedHandler struct {
	logger *zap.Logger
}

// NewOrderCreatedHandler 构造函数。
func NewOrderCreatedHandler(logger *zap.Logger) *OrderCreatedHandler {
	return &OrderCreatedHandler{logger: logger}
}

// Handle 处理订单创建事件。
func (h *OrderCreatedHandler) Handle(ctx context.Context, event OrderCreated) error {
	h.logger.Info("收到领域事件",
		zap.String("event", event.EventName()),
		zap.String("order_id", event.OrderID),
		zap.Float64("amount", event.TotalAmount),
	)
	fmt.Printf("[订单服务] 订单 %s 已创建，金额: %.2f\n", event.OrderID, event.TotalAmount)
	return nil
}

// OrderAuditHandler 同一个 OrderCreated 事件的另一个处理器，用于审计日志。
type OrderAuditHandler struct {
	logger *zap.Logger
}

// NewOrderAuditHandler 构造函数。
func NewOrderAuditHandler(logger *zap.Logger) *OrderAuditHandler {
	return &OrderAuditHandler{logger: logger}
}

// Handle 处理订单创建事件的审计日志。
func (h *OrderAuditHandler) Handle(ctx context.Context, event OrderCreated) error {
	h.logger.Info("审计日志", zap.String("event", event.EventName()), zap.Any("data", event))
	fmt.Printf("[审计服务] 订单 %s 已记录审计日志\n", event.OrderID)
	return nil
}

// OrderShippedHandler 处理 OrderShipped 事件。
type OrderShippedHandler struct {
	logger *zap.Logger
}

// NewOrderShippedHandler 构造函数。
func NewOrderShippedHandler(logger *zap.Logger) *OrderShippedHandler {
	return &OrderShippedHandler{logger: logger}
}

// Handle 处理订单发货事件。
func (h *OrderShippedHandler) Handle(ctx context.Context, event OrderShipped) error {
	h.logger.Info("收到领域事件",
		zap.String("event", event.EventName()),
		zap.String("order_id", event.OrderID),
		zap.String("tracking_no", event.TrackingNo),
	)
	fmt.Printf("[物流服务] 订单 %s 已发货，运单号: %s\n", event.OrderID, event.TrackingNo)
	return nil
}
