package ddd

import (
	"context"
	"fmt"
	"reflect"

	"github.com/mehdihadeli/go-mediatr"
	"go.uber.org/fx"
)

// DomainEventBus 领域事件总线，负责发布聚合根（AggregateRoot）产生的领域事件。
type DomainEventBus struct{}

// NewDomainEventBus 创建一个新的领域事件总线实例。
func NewDomainEventBus() *DomainEventBus {
	return &DomainEventBus{}
}

// Publish 发布聚合根中所有待处理的领域事件，并在发布后清空事件列表。
func (d *DomainEventBus) Publish(ctx context.Context, agg Aggregate) error {
	events := agg.GetDomainEvents()
	agg.ClearDomainEvents()

	for _, evt := range events {
		if err := mediatr.Publish(ctx, evt); err != nil {
			return fmt.Errorf("domain event bus: publish %q failed: %w", evt.EventName(), err)
		}
	}
	return nil
}

// DomainEventBusModule 返回一个 fx.Option，用于在 fx 依赖注入容器中注册领域事件总线。
func DomainEventBusModule(eventHandlerRegistrations ...fx.Option) fx.Option {
	return fx.Options(
		fx.Provide(NewDomainEventBus),
		fx.Options(eventHandlerRegistrations...),
	)
}

func handlerTag[T DomainEvent](idx int) string {
	t := reflect.TypeOf(*new(T))
	return fmt.Sprintf(`name:"_%s_h%d"`, t.String(), idx)
}

// RegisterDomainEventHandlers 注册指定领域事件类型的处理器。
//
// 支持同一事件类型注册多个处理器：
//
//	ddd.RegisterDomainEventHandlers[OrderCreated](
//	    NewOrderCreatedHandler,   // func(*zap.Logger) *OrderCreatedHandler
//	    NewOrderAuditHandler,     // func(*zap.Logger) *OrderAuditHandler
//	)
func RegisterDomainEventHandlers[T DomainEvent](ctors ...any) fx.Option {
	var opts []fx.Option
	for i, ctor := range ctors {
		ctor := ctor
		i := i
		tag := handlerTag[T](i)
		opts = append(opts,
			fx.Provide(
				fx.Annotate(ctor, fx.As(new(mediatr.NotificationHandler[T])), fx.ResultTags(tag)),
			),
			fx.Invoke(
				fx.Annotate(
					func(h mediatr.NotificationHandler[T]) error {
						return mediatr.RegisterNotificationHandler(h)
					},
					fx.ParamTags(tag),
				),
			),
		)
	}
	return fx.Options(opts...)
}
