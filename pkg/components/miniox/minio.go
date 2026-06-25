package miniox

import (
	"context"

	"github.com/minio/minio-go/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	Endpoint string
	minio.Options
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *minio.Client {
	client, err := minio.New(cfg.Endpoint, &cfg.Options)
	if err != nil {
		logger.Error("Minio client creation failed", zap.Error(err))
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing MinIO client")
			return nil // minio.Client 没有 Close 方法，记录日志即可
		},
	})

	return client
}
