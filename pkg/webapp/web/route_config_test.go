package web

import (
	"testing"
)

func TestNewRouteConfigDefaults(t *testing.T) {
	config := &RouteConfig{
		Path:   "/test",
		Method: GET,
	}

	if config.Path != "/test" {
		t.Fatalf("expected path '/test', got %q", config.Path)
	}
	if config.Method != GET {
		t.Fatalf("expected method GET, got %v", config.Method)
	}
	if config.AllowAnonymous {
		t.Fatal("expected AllowAnonymous to be false by default")
	}
	if len(config.Schemes) != 0 {
		t.Fatalf("expected no schemes, got %d", len(config.Schemes))
	}
}

func TestRouteConfigWithAuthenticationScheme(t *testing.T) {
	config := &RouteConfig{Path: "/test", Method: GET}
	config.WithAuthenticationScheme("jwt", "oauth2")

	if len(config.Schemes) != 2 {
		t.Fatalf("expected 2 schemes, got %d", len(config.Schemes))
	}
	if config.Schemes[0] != "jwt" {
		t.Fatalf("expected first scheme 'jwt', got %q", config.Schemes[0])
	}
}

func TestRouteConfigWithAuthorizationPolicy(t *testing.T) {
	config := &RouteConfig{Path: "/test", Method: GET}
	config.WithAuthorizationPolicy("admin", "editor")

	if len(config.Policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(config.Policies))
	}
}

func TestRouteConfigWithRateLimiter(t *testing.T) {
	config := &RouteConfig{Path: "/test", Method: GET}
	config.WithRateLimiter("fixed")

	if len(config.RateLimiter) != 1 {
		t.Fatalf("expected 1 rate limiter, got %d", len(config.RateLimiter))
	}
}

func TestRouteConfigWithAllowAnonymous(t *testing.T) {
	config := &RouteConfig{Path: "/test", Method: GET}
	config.WithAllowAnonymous()

	if !config.AllowAnonymous {
		t.Fatal("expected AllowAnonymous to be true")
	}
}

func TestRouteConfigChainedConfiguration(t *testing.T) {
	config := &RouteConfig{Path: "/test", Method: GET}
	result := config.
		WithAuthenticationScheme("jwt").
		WithAuthorizationPolicy("admin").
		WithRateLimiter("fixed").
		WithAllowAnonymous()

	if result != config {
		t.Fatal("expected chained methods to return the same config")
	}
	if len(config.Schemes) != 1 {
		t.Fatalf("expected 1 scheme, got %d", len(config.Schemes))
	}
	if len(config.Policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(config.Policies))
	}
	if len(config.RateLimiter) != 1 {
		t.Fatalf("expected 1 rate limiter, got %d", len(config.RateLimiter))
	}
	if !config.AllowAnonymous {
		t.Fatal("expected AllowAnonymous to be true")
	}
}

func TestGroupRouteConfigCreatesRoutes(t *testing.T) {
	group := &GroupRouteConfig{
		Prefix:  "/api",
		Schemes: []string{"jwt"},
	}

	route := group.MapGet("/users", "handler")
	if route == nil {
		t.Fatal("expected MapGet to return a route")
	}
	if route.Path != "/users" {
		t.Fatalf("expected path '/users', got %q", route.Path)
	}
	if route.Method != GET {
		t.Fatalf("expected method GET, got %v", route.Method)
	}
	if route.Handler != "handler" {
		t.Fatalf("expected handler 'handler', got %v", route.Handler)
	}
}

func TestGroupRouteConfigInheritsGroupSettings(t *testing.T) {
	group := &GroupRouteConfig{
		Prefix:      "/api",
		Schemes:     []string{"jwt"},
		Policies:    []string{"admin"},
		RateLimiter: []string{"fixed"},
	}

	route := group.MapPost("/items", "handler")
	if len(route.Schemes) != 1 || route.Schemes[0] != "jwt" {
		t.Fatal("expected route to inherit group schemes")
	}
	if len(route.Policies) != 1 || route.Policies[0] != "admin" {
		t.Fatal("expected route to inherit group policies")
	}
	if len(route.RateLimiter) != 1 || route.RateLimiter[0] != "fixed" {
		t.Fatal("expected route to inherit group rate limiter")
	}
}

func TestGroupRouteConfigRegistersMultipleMethods(t *testing.T) {
	group := &GroupRouteConfig{Prefix: "/api"}

	routes := []*RouteConfig{
		group.MapGet("/get", "handler"),
		group.MapPost("/post", "handler"),
		group.MapPut("/put", "handler"),
		group.MapDelete("/delete", "handler"),
		group.MapPatch("/patch", "handler"),
	}

	if len(group.Routes) != 5 {
		t.Fatalf("expected 5 routes, got %d", len(group.Routes))
	}

	expectedMethods := []RequestMethod{GET, POST, PUT, DELETE, PATCH}
	for i, route := range routes {
		if route.Method != expectedMethods[i] {
			t.Fatalf("route %d: expected method %v, got %v", i, expectedMethods[i], route.Method)
		}
	}
}
