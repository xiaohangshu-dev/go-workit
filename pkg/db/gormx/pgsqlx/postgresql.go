package pgsqlx

import (
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/gormx"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfigOptions
type Options struct {
	db.DatabaseConfig `mapstructure:",squash"`
	PgSQLCfg          postgres.Config
	*gorm.Config
}

// NewPostgresDB
func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) (*gorm.DB, error) {

	if opts.Config.Logger == nil {
		opts.Config.Logger = gormx.NewGormZapLogger(logger, opts.LogLevel, opts.SlowThreshold)
	}

	conn, err := gorm.Open(postgres.New(opts.PgSQLCfg), opts.Config)

	if err != nil {
		logger.Error("Failed to open GORM PostgreSQL", zap.Error(err))
		return nil, err
	}
	sqldb, err := conn.DB()

	db.ConfigureConnectionPool(sqldb, opts.DatabaseConfig, logger, lc)
	return conn, nil
}
