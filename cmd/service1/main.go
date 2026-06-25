// Package main API文档
//
// @title           我的服务 API
// @version         1.0
// @description     这是一个示例 API 文档
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description 输入格式: Bearer {token}
package main

import (
	_ "github.com/xiaohangshu-dev/go-workit/api/service1/docs" // swagger 一定要有这行,指向你的文档地址
	"github.com/xiaohangshu-dev/go-workit/internal/service1/webapi"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/auth"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/auth/scheme/jwt"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/authz"
)

func main() {

	builder := webapp.NewBuilder()

	builder.AddAuthentication(func(options *auth.Options) {

		options.DefaultScheme = "local_jwt_bearer"

		options.AddJwtBearer("local_jwt_bearer", func(options *jwt.Options) {

			options.TokenValidationParameters = jwt.TokenValidationParameters{
				ValidateIssuer:           true,
				ValidateAudience:         true,
				ValidateLifetime:         true,
				ValidateIssuerSigningKey: true,
				SigningKey:               []byte("secret"),
				ValidIssuer:              "sample",
				ValidAudience:            "sample",
				RequireExpiration:        true,
			}

		}).AddJwtBearer("oauth2_jwt_bearer", func(options *jwt.Options) {

			options.Authority = "http://localhost:8090"
			options.RequireHttpsMetadata = false
			options.TokenValidationParameters = jwt.TokenValidationParameters{
				ValidateIssuer: true,
				ValidIssuer:    "http://localhost:8090",
			}

		})

	})

	builder.AddAuthorization(func(options *authz.Options) {

		options.DefaultPolicy = "admin_role_policy"

		options.RequireRole("admin_role_policy", "admin", "super_admin")

	})

	app := builder.Build()

	if app.Env().IsDevelopment {
		app.UseSwagger()
	}

	app.UseAuthentication()

	app.UseAuthorization()

	app.MapMetrics()
	app.MapHealthChecks()

	app.MapRoute(webapi.Hello)

	app.Run()
}
