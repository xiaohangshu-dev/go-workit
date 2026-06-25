package dbctx

import (
	"database/sql"

	"github.com/xiaohangshu-dev/go-workit/pkg/db"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/mysqlx"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/pgsqlx"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/sqlitex"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/sqlserverx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Options 数据库选项
type Options struct {
	container   []fx.Option         // 持有容器引用
	databaseMap map[string]struct{} // 数据库实例名称集合
}

// NewOptions 创建数据库选项
func NewOptions() *Options {
	return &Options{
		container:   make([]fx.Option, 0),
		databaseMap: make(map[string]struct{}),
	}
}

// UseMySQL 使用 MySQL 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseMySQL(instanceName string, fn func(*mysqlx.Options)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &mysqlx.Options{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
		},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
				return mysqlx.NewClient(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
						return mysqlx.NewClient(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UsePostgresSQL 使用 Postgres 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UsePostgresSQL(instanceName string, fn func(*pgsqlx.Options)) *Options {

	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &pgsqlx.Options{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
		},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
				return pgsqlx.NewClient(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
						return pgsqlx.NewClient(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UseSQLServer 使用 UseSQLServer 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseSQLServer(instanceName string, fn func(*sqlserverx.Options)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &sqlserverx.Options{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
		},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
				return sqlserverx.NewClient(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
						return sqlserverx.NewClient(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

// UseSQLServer 使用 UseSQLite 数据库实例
// 若 instanceName 为空，则默认注册一个单库的实例，可通过实例直接注入
// 若 instanceName 非空，则注册一个显式命名的实例，使用 name 标签注入
func (d *Options) UseSQLite(instanceName string, fn func(*sqlitex.Options)) *Options {
	if instanceName == "" {
		// 默认单库，无 name
		instanceName = "default"
	}

	if _, ok := d.databaseMap[instanceName]; ok {
		panic("database instance name already exists")
	}

	opts := &sqlitex.Options{
		DatabaseConfig: db.DatabaseConfig{
			MaxOpenConns:    db.MaxOpenConns,
			MaxIdleConns:    db.MaxIdleConns,
			ConnMaxLifetime: db.ConnMaxLifetime,
		},
	}

	fn(opts)

	if instanceName == "default" {
		// 单库，第一次注册 default，提供不带 name 的数据库
		d.container = append(d.container,
			fx.Provide(func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
				return sqlitex.NewClient(lc, opts, logger)
			}),
		)
	} else {
		// 多库，或显式传名字的数据库，使用 name 标签
		d.container = append(d.container,
			fx.Provide(
				fx.Annotate(
					func(lc fx.Lifecycle, logger *zap.Logger) *sql.DB {
						return sqlitex.NewClient(lc, opts, logger)
					},
					fx.ResultTags(`name:"`+instanceName+`"`),
				),
			),
		)
	}

	d.databaseMap[instanceName] = struct{}{}

	return d
}

func (o *Options) Container() []fx.Option {
	return o.container
}
