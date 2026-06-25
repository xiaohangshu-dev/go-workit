package pgsqlx

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	db.DatabaseConfig
	DSN string
}

func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) *sql.DB {

	conn, err := sql.Open("postgres", opts.DSN)
	if err != nil {

		logger.Error("Failed to open PostgreSQL", zap.Error(err))
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		logger.Error("Failed to ping PostgreSQL", zap.Error(err))
		panic(err)
	}

	db.ConfigureConnectionPool(conn, opts.DatabaseConfig, logger, lc)

	return conn
}
