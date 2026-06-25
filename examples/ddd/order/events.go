// Package order 演示领域事件、聚合根的定义。
package order

import "github.com/xiaohangshu-dev/go-workit/pkg/ddd"

// OrderCreated 订单创建事件。
type OrderCreated struct {
	ddd.BaseDomainEvent
	OrderID     string
	CustomerID  string
	TotalAmount float64
}

// NewOrderCreated 创建 OrderCreated 事件。
func NewOrderCreated(orderID, customerID string, totalAmount float64) OrderCreated {
	return OrderCreated{
		BaseDomainEvent: ddd.NewBaseDomainEvent("order.created"),
		OrderID:         orderID,
		CustomerID:      customerID,
		TotalAmount:     totalAmount,
	}
}

// OrderShipped 订单发货事件。
type OrderShipped struct {
	ddd.BaseDomainEvent
	OrderID    string
	TrackingNo string
}

// NewOrderShipped 创建 OrderShipped 事件。
func NewOrderShipped(orderID, trackingNo string) OrderShipped {
	return OrderShipped{
		BaseDomainEvent: ddd.NewBaseDomainEvent("order.shipped"),
		OrderID:         orderID,
		TrackingNo:      trackingNo,
	}
}

// OrderAggregate 订单聚合根。
type OrderAggregate struct {
	ddd.AggregateRoot[string]
	CustomerID  string
	TotalAmount float64
	Status      string
}

// NewOrderAggregate 创建订单聚合根，产生 OrderCreated 事件。
func NewOrderAggregate(orderID, customerID string, totalAmount float64) *OrderAggregate {
	agg := &OrderAggregate{
		AggregateRoot: ddd.NewAggregateRoot(orderID),
		CustomerID:    customerID,
		TotalAmount:   totalAmount,
		Status:        "pending",
	}
	agg.AddDomainEvent(NewOrderCreated(orderID, customerID, totalAmount))
	return agg
}

// Ship 发货操作，产生 OrderShipped 事件。
func (agg *OrderAggregate) Ship(trackingNo string) {
	agg.Status = "shipped"
	agg.AddDomainEvent(NewOrderShipped(agg.ID, trackingNo))
}
