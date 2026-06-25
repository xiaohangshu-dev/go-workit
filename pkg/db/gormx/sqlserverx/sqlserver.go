package sqlserverx

import (
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type Options struct {
	db.DatabaseConfig
	SQLServerCfg sqlserver.Config
	*gorm.Config
}

// NewSQLServer
func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = gormx.NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	conn, err := gorm.Open(sqlserver.New(opts.SQLServerCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLServer", zap.Error(err))
		return nil, err
	}

	sqldb, err := conn.DB()

	db.ConfigureConnectionPool(sqldb, opts.DatabaseConfig, logger, lc)

	return conn, nil
}
