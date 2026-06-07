package sse

import (
	"net/http"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
)

// Stream 封装了 SSE 流式响应
type Stream struct {
	w       http.ResponseWriter
	flusher http.Flusher
	written bool // 标记是否已经发送过数据（自动设置头）
}

// NewStream 创建一个 SSE 流，自动设置响应头
func NewStream(c *gin.Context) *Stream {
	// 设置 SSE 必需的响应头（仅当尚未写入任何数据时）
	w := c.Writer
	// 注意：如果已经在别处写入了头部，这里不能重复设置，但通常不会
	// 为了安全，我们不在创建时立即设置头，而是在第一次 Publish 时设置
	return &Stream{
		w:       w,
		flusher: nil, // 延迟获取，避免不必要的类型断言
	}
}

// ensureHeaders 在第一次发送数据时设置响应头和 flusher
func (s *Stream) ensureHeaders(c *gin.Context) error {
	if s.written {
		return nil
	}
	// 设置 SSE 头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	// 允许跨域（可选）
	c.Header("Access-Control-Allow-Origin", "*")

	// 获取 flusher
	flusher, ok := s.w.(http.Flusher)
	if !ok {
		return http.ErrNotSupported
	}
	s.flusher = flusher
	s.written = true
	return nil
}

// Publish 发送一个 SSE 事件（自动处理 data 格式和 flush）
func (s *Stream) Publish(c *gin.Context, event *sse.Event) error {
	if err := s.ensureHeaders(c); err != nil {
		return err
	}
	if err := sse.Encode(s.w, *event); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// Close 可选，用于兼容 io.Closer（空实现，满足接口）
func (s *Stream) Close() error {
	return nil
}

// 快捷方法：直接发送字符串数据
func (s *Stream) Send(c *gin.Context, data any) error {
	return s.Publish(c, &sse.Event{
		Data: data,
	})
}
