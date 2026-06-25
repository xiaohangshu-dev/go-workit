// Package service 演示 HTTP 客户端注入的业务服务。
package service

import (
	"context"
	"fmt"

	httpclient "github.com/xiaohangshu-dev/go-workit/pkg/tools/net"
	"go.uber.org/zap"
)

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UserService 调用外部 API 获取用户数据。
//
// 方式一：直接注入 *httpclient.Client（适用于单个客户端）
//
//	type UserService struct {
//	    Client *httpclient.Client
//	}
//
// 方式二：注入 *httpclient.Provider（适用于多个命名客户端）
type UserService struct {
	Clients *httpclient.Provider
	logger  *zap.Logger
}

// NewUserService 构造函数。
func NewUserService(clients *httpclient.Provider, logger *zap.Logger) *UserService {
	return &UserService{Clients: clients, logger: logger}
}

// GetUser 使用命名客户端 "placeholder" 获取用户信息。
func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	client := s.Clients.Get("placeholder")
	var user User
	err := client.Get(ctx, fmt.Sprintf("/users/%d", id), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
