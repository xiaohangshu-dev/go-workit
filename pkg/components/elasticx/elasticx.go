package elasticx

import (
	"context"

	"github.com/olivere/elastic/v7"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	Func []elastic.ClientOptionFunc
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *elastic.Client {
	client, err := elastic.NewClient(cfg.Func...)
	if err != nil {
		logger.Error("Elasticsearch client creation failed", zap.Error(err))
		panic(err)
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Stopping Elasticsearch client")
			client.Stop()
			return nil
		},
	})

	return client
}
