package ginx

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xiaohangshu-dev/go-workit/pkg/webapp/web"
	"go.uber.org/zap"
)

type fakeRouter struct {
	defaultPolicy string
	policies      map[string]func(*web.ClaimsPrincipal) bool
}

func (r *fakeRouter) GlobalScheme() string {
	return ""
}

func (r *fakeRouter) GlobalPolicy() string {
	return r.defaultPolicy
}

func (r *fakeRouter) GlobalRatelimit() string {
	return ""
}

func (r *fakeRouter) Authenticate(string) (web.Authenticate, bool) {
	return nil, false
}

func (r *fakeRouter) Authorize(policy string) (func(*web.ClaimsPrincipal) bool, bool) {
	handler, ok := r.policies[policy]
	return handler, ok
}

func (r *fakeRouter) RateLimiter(string) (web.RateLimiter, bool) {
	return nil, false
}

func TestAuthorizeUsesDefaultPolicyWhenRouteHasNoPolicy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	router := &fakeRouter{
		defaultPolicy: "deny",
		policies: map[string]func(*web.ClaimsPrincipal) bool{
			"deny": func(*web.ClaimsPrincipal) bool {
				return false
			},
		},
	}

	authz := newAuthorize(engine, router, zap.NewNop())
	engine.Use(func(c *gin.Context) {
		c.Set(contextClaimsKey, &web.ClaimsPrincipal{
			Subject:         "user-1",
			AuthenticatedAt: time.Now(),
		})
		c.Next()
	})
	engine.Use(authz.Handle())
	engine.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/secure", nil)

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected default policy to forbid request, got status %d", recorder.Code)
	}
}

func TestAuthorizeForbidsWhenConfiguredPolicyIsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	router := &fakeRouter{
		defaultPolicy: "missing",
		policies:      map[string]func(*web.ClaimsPrincipal) bool{},
	}

	authz := newAuthorize(engine, router, zap.NewNop())
	engine.Use(func(c *gin.Context) {
		c.Set(contextClaimsKey, &web.ClaimsPrincipal{
			Subject:         "user-1",
			AuthenticatedAt: time.Now(),
		})
		c.Next()
	})
	engine.Use(authz.Handle())
	engine.GET("/secure", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/secure", nil)

	engine.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected missing policy to forbid request, got status %d", recorder.Code)
	}
}
