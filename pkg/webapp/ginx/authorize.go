package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

// Authorize 授权中间件
type Authorize struct {
	*gin.Engine
	web.Router
	logger *zap.Logger
}

// newAuthorize 初始化授权中间件
func newAuthorize(engine *gin.Engine, router web.Router, logger *zap.Logger) *Authorize {
	return &Authorize{
		Engine: engine,
		Router: router,
		logger: logger,
	}
}

// Handle 授权中间件处理函数
func (a *Authorize) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.Request.URL.Path

		// OPTIONS 请求直接放行
		if method == http.MethodOptions {
			c.Next()
			return
		}

		nodeValue := a.GetNodeValue(c)
		if nodeValue.AllowAnonymous {
			c.Next()
			return
		}

		claims := ginGetClaimsPrincipal(c)
		if claims == nil {
			a.logger.Error("authorization failed: ClaimsPrincipal is nil")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		policyNames := nodeValue.AuthzPolicies
		if len(policyNames) == 0 {
			if defaultPolicy := a.GlobalPolicy(); defaultPolicy != "" {
				policyNames = append(policyNames, defaultPolicy)
			}
		}

		for _, policyName := range policyNames {
			policyFunc, ok := a.Authorize(policyName)
			if !ok {
				a.logger.Warn("authorization failed: policy not found",
					zap.String("path", path),
					zap.String("policy", policyName))
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			if !policyFunc(claims) {
				a.logger.Warn("authorization failed",
					zap.String("path", path),
					zap.String("policy", policyName))
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// 继续执行后续中间件
		c.Next()
	}
}

// ginGetClaimsPrincipal 从gin.Context中获取ClaimsPrincipal
func ginGetClaimsPrincipal(c *gin.Context) *web.ClaimsPrincipal {
	claims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	principal, ok := claims.(*web.ClaimsPrincipal)
	if !ok {
		return nil
	}

	return principal

}
