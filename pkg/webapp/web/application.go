package web

import (
	"github.com/spf13/viper"
	"github.com/xiaohangshu-dev/go-workit/pkg/host"
	"go.uber.org/zap"
)

// Application  web 应用接口
type Application interface {
	host.Host
	Logger() *zap.Logger
	Config() *viper.Viper
	Env() *Environment
	Use(...any) Application
	UseSwagger() Application
	UseCORS(any) Application
	UseStaticFiles(urlPath, root string) Application
	UseHealthCheck() Application
	MapMetrics(path ...string) Application
	MapHealthChecks(path ...string) Application
	MapReadinessChecks(path ...string) Application
	MapLivenessChecks(path ...string) Application
	UseAuthentication() Application
	UseAuthorization() Application
	UseRecovery() Application
	UseLogger() Application
	UseLocalization() Application
	UseRateLimiter() Application
	UseRequestDecompression() Application
	MapRoute(...any) Application
	MapGrpcServices(...any) Application
}
