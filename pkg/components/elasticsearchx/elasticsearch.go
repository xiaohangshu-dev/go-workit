package elasticsearchx

import (
	"context"

	"github.com/elastic/go-elasticsearch/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	elasticsearch.Config
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *elasticsearch.Client {
	client, err := elasticsearch.NewClient(cfg.Config)
	if err != nil {
		logger.Error("Elasticsearch client creation failed", zap.Error(err))
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing Elasticsearch client")
			return nil // elasticsearch.Client 无 Close 方法，连接由 http.Client 管理
		},
	})

	return client
}
