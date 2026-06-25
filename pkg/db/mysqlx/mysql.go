package mysqlx

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	db.DatabaseConfig
	DSN string
}

func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) *sql.DB {

	conn, err := sql.Open("mysql", opts.DSN)
	if err != nil {

		logger.Error("Failed to open MySQL", zap.Error(err))
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		logger.Error("Failed to ping MySQL", zap.Error(err))
		panic(err)
	}

	db.ConfigureConnectionPool(conn, opts.DatabaseConfig, logger, lc)

	return conn
}
