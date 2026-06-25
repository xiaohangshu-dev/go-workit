package kafkax

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ReaderOptions struct {
	kafka.ReaderConfig
}

type WriterOptions struct {
	AllowAutoTopicCreation bool
	kafka.WriterConfig
}

func NewReader(lc fx.Lifecycle, cfg *ReaderOptions, logger *zap.Logger) *kafka.Reader {
	r := kafka.NewReader(cfg.ReaderConfig)

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Kafka reader")
			return r.Close()
		},
	})

	return r
}

func NewWriter(lc fx.Lifecycle, cfg *WriterOptions, logger *zap.Logger) *kafka.Writer {
	w := kafka.NewWriter(cfg.WriterConfig)
	w.AllowAutoTopicCreation = cfg.AllowAutoTopicCreation

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Kafka writer")
			return w.Close()
		},
	})

	return w
}
