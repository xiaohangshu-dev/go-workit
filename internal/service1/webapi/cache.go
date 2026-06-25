package webapi

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func Cache(router *gin.Engine, rc *redis.Client, logger *zap.Logger) {

	// 创建路由组
	group := router.Group("/cache")

	// 创建路由

	group.GET("/redis", hello2(rc, logger))
}

// hello godoc
// @Summary hello workit
// @Description 返回 "你好,小航"
// @Tags Hello
// @Accept json
// @Produce json
// @Success 200 {object} api.Response[string]
// @Router /hello [get]
// @Security BearerAuth
func hello2(rc *redis.Client, log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		rc.Set(c, "hello", "你好,小航", 0)

		c.JSON(200, gin.H{
			"message": rc.Get(c, "hello").Val(),
		})
	}
}
