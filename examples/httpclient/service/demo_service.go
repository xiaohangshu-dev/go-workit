package service

import (
	"context"
	"fmt"

	"github.com/xiaohangshu-dev/go-workit/pkg/app"
	"go.uber.org/zap"
)

// DemoService 在应用启动时演示 HTTP 客户端的使用。
type DemoService struct {
	userSvc *UserService
	logger  *zap.Logger
}

// NewDemoService 构造函数。
func NewDemoService(userSvc *UserService, logger *zap.Logger) app.BackgroundService {
	return &DemoService{userSvc: userSvc, logger: logger}
}

// Start 执行一次 HTTP 请求演示。
func (s *DemoService) Start(ctx context.Context) error {
	s.logger.Info("=== HTTP 客户端示例开始 ===")

	user, err := s.userSvc.GetUser(ctx, 1)
	if err != nil {
		s.logger.Warn("API 请求失败（可忽略，无网络环境）", zap.Error(err))
		fmt.Println("[Demo] 若在联网环境，此处会打印用户信息")
	} else {
		fmt.Printf("[Demo] 获取到用户: %+v\n", user)
	}

	s.logger.Info("=== HTTP 客户端示例结束 ===")
	return nil
}

func (s *DemoService) Stop(ctx context.Context) error { return nil }
