package db

import (
	"context"
	"database/sql"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// configureConnectionPool 配置连接池并注册 Fx 生命周期
func ConfigureConnectionPool(sqlDB *sql.DB, cfg DatabaseConfig, logger *zap.Logger, lc fx.Lifecycle) {

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 注册 Fx 生命周期
	lc.Append(fx.Hook{
		// OnStart 时检查数据库连接
		OnStart: func(ctx context.Context) error {
			if err := sqlDB.PingContext(ctx); err != nil {
				logger.Error("Failed to ping database on start", zap.Error(err))
				return err
			}
			logger.Info("Database connection successful")
			return nil
		},
		// OnStop 时关闭数据库连接
		OnStop: func(ctx context.Context) error {
			logger.Info("Closing database connection")
			return sqlDB.Close()
		},
	})
}
