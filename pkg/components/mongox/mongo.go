package mongox

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	options.ClientOptions
}

func NewClient(lc fx.Lifecycle, cfg *Options, logger *zap.Logger) *mongo.Client {
	client, err := mongo.Connect(context.Background(), &cfg.ClientOptions)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", zap.Error(err))
		return nil
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing MongoDB connection")
			return client.Disconnect(ctx)
		},
	})

	return client
}
