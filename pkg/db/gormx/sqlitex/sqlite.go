package sqlitex

import (
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Options struct {
	db.DatabaseConfig
	SQLiteCfg sqlite.Config
	*gorm.Config
}

// NewSQLServer
func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = gormx.NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	conn, err := gorm.Open(sqlite.New(opts.SQLiteCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM SQLite", zap.Error(err))
		return nil, err
	}

	sqldb, err := conn.DB()

	db.ConfigureConnectionPool(sqldb, opts.DatabaseConfig, logger, lc)

	return conn, nil
}
