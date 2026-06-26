// Package service 演示 HTTP 客户端注入的业务服务。
package service

import (
	"context"
	"fmt"

	"github.com/xiaohangshu-dev/go-workit/pkg/tools/httpclient"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// HttpClients 命名 HTTP 客户端集合。
type HttpClients struct {
	fx.In

	Placeholder *httpclient.Client `name:"placeholder"`
}

// UserService 调用外部 API 获取用户数据。
type UserService struct {
	Client      *httpclient.Client
	Placeholder *httpclient.Client
	logger      *zap.Logger
}

// NewUserService 构造函数。
func NewUserService(client *httpclient.Client, clients HttpClients, logger *zap.Logger) *UserService {
	return &UserService{
		Client:      client,
		Placeholder: clients.Placeholder,
		logger:      logger,
	}
}

// GetUser 使用命名客户端 "placeholder" 获取用户信息。
func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
	var user User
	err := s.Placeholder.Get(ctx, fmt.Sprintf("/users/%d", id), &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
