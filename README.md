# go-workit

go-workit 是一个现代化的 Go 开发框架，提供了完整的应用开发解决方案，包括依赖注入、配置管理、Web 服务、数据库连接、缓存管理等功能。

## 架构说明

### 核心模块

1. **app** - 应用核心模块，提供应用的初始化和生命周期管理
2. **webapp** - Web 应用模块，提供 HTTP 和 gRPC 服务支持
3. **ddd** - 领域驱动设计模块，提供聚合根、实体、领域事件等核心概念
4. **db** - 数据库模块，支持 MySQL、PostgreSQL、SQLite、SQL Server 等多种数据库
5. **cache** - 缓存模块，提供 Redis 缓存支持
6. **config** - 配置管理模块，支持 YAML、JSON 配置文件、环境变量和命令行参数
7. **eventbus** - 事件总线模块，提供事件发布和订阅功能
8. **host** - 主机模块，定义应用的基本接口
9. **tools** - 工具模块，提供各种实用工具函数

### 架构图

```
┌─────────────────────────────────────────────────────────┐
│                       go-workit                        │
├─────────────────┬─────────────────┬────────────────────┤
│                 │                 │                    │
│    ┌────────────▼───────┐  ┌─────▼──────────┐  ┌──────▼─────────┐
│    │      app           │  │   webapp       │  │      ddd       │
│    └────────────────────┘  └────────────────┘  └────────────────┘
│                 │                 │                    │
│    ┌────────────▼───────┐  ┌─────▼──────────┐  ┌──────▼─────────┐
│    │     config         │  │    auth        │  │   domain_event │
│    └────────────────────┘  └────────────────┘  └────────────────┘
│                 │                 │                    │
│    ┌────────────▼───────┐  ┌─────▼──────────┐  ┌──────▼─────────┐
│    │      db            │  │    router      │  │   entity       │
│    └────────────────────┘  └────────────────┘  └────────────────┘
│                 │                 │                    │
│    ┌────────────▼───────┐  ┌─────▼──────────┐  ┌──────▼─────────┐
│    │     cache          │  │    ratelimit   │  │   value_object │
│    └────────────────────┘  └────────────────┘  └────────────────┘
└─────────────────┴─────────────────┴────────────────────┘
```

## 快速开始

### 安装

```bash
go get github.com/xiaohangshu-dev/go-workit
```

> 当前 Web 路由扩展依赖 `github.com/xiaohangshu-dev/gin` fork。由于 Go module 的 `replace` 不会传递给下游项目，使用本框架的应用需要在自己的 `go.mod` 中显式添加：
>
> ```go
> replace github.com/gin-gonic/gin => github.com/xiaohangshu-dev/gin v0.0.0-20251127022746-130901f68014
> ```

### 基本使用

#### 创建一个简单的 Web 应用

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/db/mysqlx"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/dbctx"
)

func main() {
	// 创建 Web 应用构建器
	builder := webapp.NewBuilder()

	// 添加数据库配置
	builder.AddDbContext(func(options *dbctx.Options) {
		options.UseMySQL("default", func(config *mysqlx.Options) {
			config.DSN = "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
		})
	})

	// 构建并运行应用
	app := builder.Build()
	app.MapRoute(func(router *gin.Engine) {
		router.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello, World!",
			})
		})
	})

	app.Run()
}
```

#### 配置文件示例 (application.yaml)

```yaml
server:
  http_port: 8080
  grpc_port: 50051
  environment: development

log:
  level: info
  filename: ./logs/app.log
  max_size: 100
  max_backups: 3
  max_age: 7
  compress: true
  console: true

mysql:
  default:
    dsn: user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local

redis:
  default:
    addr: 127.0.0.1:6379
    password: ""
    db: 0
```

## 核心功能

### 1. 依赖注入

使用 Uber 的 fx 框架实现依赖注入，提高代码的可测试性和可维护性。

### 2. 配置管理

支持 YAML、JSON 配置文件、环境变量和命令行参数，提供统一的配置访问接口。

### 3. Web 服务

提供 HTTP 和 gRPC 服务支持，集成了路由、中间件、认证、授权、限流等功能。

### 4. 数据库连接

支持 MySQL、PostgreSQL、SQLite、SQL Server 等多种数据库，提供统一的连接管理。

### 5. 缓存管理

集成 Redis 缓存，提供统一的缓存访问接口。

### 6. 领域驱动设计

实现了领域驱动设计的核心概念，如聚合根、实体、领域事件等。

### 7. 事件总线

提供事件发布和订阅功能，支持领域事件的处理。

## 中间件

- **认证中间件** - 支持 JWT、OAuth2 等认证方式
- **授权中间件** - 基于策略的授权
- **限流中间件** - 支持多种限流策略
- **日志中间件** - 请求日志记录
- **恢复中间件** - 异常恢复
- **CORS 中间件** - 跨域资源共享
- **静态文件中间件** - 静态文件服务
- **健康检查中间件** - 健康检查端点

## 最佳实践

1. **模块化设计** - 将应用拆分为多个模块，每个模块负责特定的功能
2. **依赖注入** - 使用依赖注入管理组件之间的依赖关系
3. **配置外部化** - 将配置从代码中分离，使用配置文件、环境变量或命令行参数
4. **错误处理** - 统一错误处理方式，提供清晰的错误信息
5. **日志记录** - 合理使用日志，记录应用的运行状态和错误信息
6. **测试** - 编写单元测试和集成测试，确保代码的质量和可靠性

## 贡献

欢迎贡献代码、报告问题或提出建议。请提交 Pull Request 或 Issue。

## 许可证

MIT
