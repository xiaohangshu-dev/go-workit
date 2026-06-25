package web

import (
	"testing"
	"time"
)

func TestNewClaimsPrincipalHasDefaults(t *testing.T) {
	cp := &ClaimsPrincipal{}

	if cp.Subject != "" {
		t.Fatalf("expected empty subject, got %q", cp.Subject)
	}
	if len(cp.Roles) != 0 {
		t.Fatalf("expected no roles, got %d", len(cp.Roles))
	}
	if len(cp.Claims) != 0 {
		t.Fatalf("expected no claims, got %d", len(cp.Claims))
	}
}

func TestAddRole(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddRole("admin")

	if len(cp.Roles) != 1 {
		t.Fatalf("expected 1 role, got %d", len(cp.Roles))
	}
	if cp.Roles[0] != "admin" {
		t.Fatalf("expected role 'admin', got %q", cp.Roles[0])
	}
}

func TestAddRoleDoesNotDuplicate(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddRole("admin")
	cp.AddRole("admin")

	if len(cp.Roles) != 1 {
		t.Fatalf("expected 1 role (no duplicates), got %d", len(cp.Roles))
	}
}

func TestIsInRole(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddRole("admin")

	if !cp.IsInRole("admin") {
		t.Fatal("expected IsInRole('admin') to return true")
	}
	if cp.IsInRole("user") {
		t.Fatal("expected IsInRole('user') to return false")
	}
}

func TestAddClaim(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddClaim("email", "user@example.com")

	if len(cp.Claims) != 1 {
		t.Fatalf("expected 1 claim, got %d", len(cp.Claims))
	}
	if cp.Claims[0].Type != "email" {
		t.Fatalf("expected claim type 'email', got %q", cp.Claims[0].Type)
	}
	if cp.Claims[0].Value != "user@example.com" {
		t.Fatalf("expected claim value 'user@example.com', got %v", cp.Claims[0].Value)
	}
}

func TestFindFirstReturnsExistingClaim(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddClaim("email", "user@example.com")

	value, ok := cp.FindFirst("email")
	if !ok {
		t.Fatal("expected FindFirst to find 'email'")
	}
	if value != "user@example.com" {
		t.Fatalf("expected value 'user@example.com', got %v", value)
	}
}

func TestFindFirstReturnsFalseForMissingClaim(t *testing.T) {
	cp := &ClaimsPrincipal{}

	_, ok := cp.FindFirst("nonexistent")
	if ok {
		t.Fatal("expected FindFirst to return false for missing claim")
	}
}

func TestHasClaim(t *testing.T) {
	cp := &ClaimsPrincipal{}
	cp.AddClaim("role", "admin")

	if !cp.HasClaim("role", "admin") {
		t.Fatal("expected HasClaim('role', 'admin') to return true")
	}
	if cp.HasClaim("role", "user") {
		t.Fatal("expected HasClaim('role', 'user') to return false")
	}
}

func TestCloneReturnsDeepCopy(t *testing.T) {
	cp := &ClaimsPrincipal{
		Subject:              "user-1",
		Name:                 "Alice",
		Roles:                []string{"admin"},
		IdentityProvider:     "internal",
		AuthenticationMethod: "password",
		AuthenticatedAt:      time.Now(),
	}
	cp.AddClaim("email", "alice@example.com")

	clone := cp.Clone()

	// Modify original
	cp.Subject = "modified"
	cp.Roles[0] = "user"
	cp.Claims[0].Value = "modified@example.com"

	if clone.Subject != "user-1" {
		t.Fatalf("expected clone subject 'user-1', got %q", clone.Subject)
	}
	if clone.Roles[0] != "admin" {
		t.Fatalf("expected clone role 'admin', got %q", clone.Roles[0])
	}
	if clone.Claims[0].Value != "alice@example.com" {
		t.Fatalf("expected clone claim value 'alice@example.com', got %v", clone.Claims[0].Value)
	}
}

func TestEnvironmentDefaults(t *testing.T) {
	env := &Environment{
		Env: "development",
	}

	if env.Env != "development" {
		t.Fatalf("expected env 'development', got %q", env.Env)
	}
	if env.IsDevelopment {
		t.Fatal("expected IsDevelopment to be false by default")
	}

	env.IsDevelopment = true
	if !env.IsDevelopment {
		t.Fatal("expected IsDevelopment to be true after setting")
	}
}
