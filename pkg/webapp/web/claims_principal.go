package web

import "slices"

import "time"

// ClaimsPrincipal 表示用户身份和声明
type ClaimsPrincipal struct {
	Subject              string    // 一般对应 nameidentifier 或 sub（JWT）
	Name                 string    // Display name
	Roles                []string  // 权限角色列表
	Claims               []Claim   // 所有的 Claim（通用扩展）
	IdentityProvider     string    // 身份提供者（如 AzureAD、Google、Internal 等）
	AuthenticationMethod string    // 认证方式（如 password、oauth、saml）
	AuthenticatedAt      time.Time // 认证时间
}

// AddRole 添加角色
func (cp *ClaimsPrincipal) AddRole(role string) {
	if slices.Contains(cp.Roles, role) {
			return
		}
	cp.Roles = append(cp.Roles, role)
}

// IsInRole 判断是否有指定角色, 返回 true 表示有指定角色
func (cp *ClaimsPrincipal) IsInRole(role string) bool {
	return slices.Contains(cp.Roles, role)
}

// AddClaim 添加 Claim
func (cp *ClaimsPrincipal) AddClaim(key string, value interface{}) {
	cp.Claims = append(cp.Claims, Claim{Type: key, Value: value})
}

// FindFirst 查找第一个 Claim, 返回 Claim 的值和是否存在
func (cp *ClaimsPrincipal) FindFirst(key string) (any, bool) {
	for _, c := range cp.Claims {
		if c.Type == key {
			return c.Value, true
		}
	}
	return nil, false
}

// HasClaim 判断是否有指定 Claim, 返回 true 表示有指定 Claim
func (cp *ClaimsPrincipal) HasClaim(key string, value any) bool {
	for _, c := range cp.Claims {
		if c.Type == key && c.Value == value {
			return true
		}
	}
	return false
}

// Clone 克隆 ClaimsPrincipal 对象
func (cp *ClaimsPrincipal) Clone() *ClaimsPrincipal {
	newClaims := make([]Claim, len(cp.Claims))
	copy(newClaims, cp.Claims)

	newRoles := make([]string, len(cp.Roles))
	copy(newRoles, cp.Roles)

	return &ClaimsPrincipal{
		Subject:              cp.Subject,
		Name:                 cp.Name,
		Roles:                newRoles,
		Claims:               newClaims,
		IdentityProvider:     cp.IdentityProvider,
		AuthenticationMethod: cp.AuthenticationMethod,
		AuthenticatedAt:      cp.AuthenticatedAt,
	}
}
