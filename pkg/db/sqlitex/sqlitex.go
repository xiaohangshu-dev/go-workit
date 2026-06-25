package sqlitex

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Options struct {
	db.DatabaseConfig
	DSN string
}

func NewClient(lc fx.Lifecycle, opts *Options, logger *zap.Logger) *sql.DB {

	conn, err := sql.Open("sqlite3", opts.DSN)
	if err != nil {

		logger.Error("Failed to open SQLite", zap.Error(err))
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		logger.Error("Failed to ping SQLite", zap.Error(err))
		panic(err)
	}

	db.ConfigureConnectionPool(conn, opts.DatabaseConfig, logger, lc)

	return conn
}
