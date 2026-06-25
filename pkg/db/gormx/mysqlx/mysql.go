package mysqlx

import (
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLConfigOptions  公共字段 + 扩展字段
type Options struct {
	db.DatabaseConfig
	MySQLCfg mysql.Config
	*gorm.Config
}

// NewMySQLDB
func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = gormx.NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	conn, err := gorm.Open(mysql.New(opts.MySQLCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM MySQL", zap.Error(err))
		return nil, err
	}

	sqldb, err := conn.DB()

	db.ConfigureConnectionPool(sqldb, opts.DatabaseConfig, logger, lc)

	return conn, nil
}
