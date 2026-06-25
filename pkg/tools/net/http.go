// Package net 提供可注入的 HTTP 客户端。
//
// ## 使用方式
//
// ### 方式一：单个客户端（最常见）
//
//	// 注册
//	builder.AddHttpClient(func(options *net.Options) {
//	    options.BaseURL = "https://api.github.com"
//	})
//
//	// 注入（直接声明字段即可）
//	type UserService struct {
//	    Client *net.Client  // 框架自动注入
//	}
//	func (s *UserService) GetUser(ctx context.Context, id int) (*User, error) {
//	    var user User
//	    err := s.Client.Get(ctx, "/users/1", &user)
//	    return &user, err
//	}
//
// ### 方式二：多个命名客户端
//
//	// 注册
//	builder.AddHttpClient("github", func(options *net.Options) {
//	    options.BaseURL = "https://api.github.com"
//	})
//	builder.AddHttpClient("openai", func(options *net.Options) {
//	    options.BaseURL = "https://api.openai.com"
//	})
//
//	// 注入 Provider，通过名称获取
//	type ApiService struct {
//	    Clients *net.Provider  // 框架自动注入
//	}
//	func (s *ApiService) DoWork(ctx context.Context) error {
//	    github := s.Clients.Get("github")
//	    openai := s.Clients.Get("openai")
//	    github.Get(ctx, "/users", &result)
//	    openai.Post(ctx, "/chat", body, &result)
//	}
package net

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// ---------- Options ----------

// Options HTTP 客户端配置
type Options struct {
	// BaseURL 请求基础地址，所有请求会拼接此前缀
	BaseURL string
	// Timeout 请求超时时间，默认 30 秒
	Timeout time.Duration
	// DefaultHeaders 默认请求头，每次请求会自动添加
	DefaultHeaders map[string]string
}

func (o *Options) normalize() {
	if o.Timeout <= 0 {
		o.Timeout = 30 * time.Second
	}
	if o.DefaultHeaders == nil {
		o.DefaultHeaders = make(map[string]string)
	}
}

// NewOptions 创建默认配置
func NewOptions() *Options {
	return &Options{
		Timeout:        30 * time.Second,
		DefaultHeaders: make(map[string]string),
	}
}

// ---------- Client ----------

// Client 可注入的 HTTP 客户端，封装了请求构建、序列化、反序列化。
type Client struct {
	options *Options
	client  *http.Client
}

// NewClient 创建 HTTP 客户端
func NewClient(options *Options) *Client {
	if options == nil {
		options = NewOptions()
	}
	options.normalize()

	return &Client{
		options: options,
		client: &http.Client{
			Timeout: options.Timeout,
		},
	}
}

// resolveURL 拼接 BaseURL 和请求路径
func (c *Client) resolveURL(path string) string {
	if c.options.BaseURL == "" {
		return path
	}
	base, _ := url.JoinPath(c.options.BaseURL, path)
	return base
}

// newRequest 创建通用请求
func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolveURL(path), body)
	if err != nil {
		return nil, fmt.Errorf("http: create request %s %s failed: %w", method, path, err)
	}
	for k, v := range c.options.DefaultHeaders {
		req.Header.Set(k, v)
	}
	return req, nil
}

// doRequest 发送请求并解析响应
func (c *Client) doRequest(req *http.Request, result any, allowedStatuses ...int) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http: request %s %s failed: %w", req.Method, req.URL.String(), err)
	}
	defer resp.Body.Close()

	if len(allowedStatuses) > 0 {
		statusOK := false
		for _, s := range allowedStatuses {
			if resp.StatusCode == s {
				statusOK = true
				break
			}
		}
		if !statusOK {
			body, _ := io.ReadAll(resp.Body)
			return resp, fmt.Errorf("http: %s %s returned %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(body))
		}
	} else if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return resp, fmt.Errorf("http: %s %s returned %d: %s", req.Method, req.URL.String(), resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp, fmt.Errorf("http: decode response from %s %s failed: %w", req.Method, req.URL.String(), err)
		}
	}
	return resp, nil
}

// ---------- 公开方法 ----------

// Get 发送 GET 请求，将 JSON 响应解析到 result 中。
// allowedStatuses 可选，指定期望的状态码，为空时 4xx/5xx 视为错误。
//
//	err := client.Get(ctx, "/api/users", &users)
//	err := client.Get(ctx, "/api/users", &users, http.StatusOK) // 只接受 200
func (c *Client) Get(ctx context.Context, path string, result any, allowedStatuses ...int) error {
	req, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, result, allowedStatuses...)
	return err
}

// Post 发送 POST 请求，body 自动序列化为 JSON，响应解析到 result 中。
func (c *Client) Post(ctx context.Context, path string, body any, result any, allowedStatuses ...int) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("http: encode request body failed: %w", err)
		}
	}
	req, err := c.newRequest(ctx, http.MethodPost, path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.doRequest(req, result, allowedStatuses...)
	return err
}

// Put 发送 PUT 请求。
func (c *Client) Put(ctx context.Context, path string, body any, result any, allowedStatuses ...int) error {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return fmt.Errorf("http: encode request body failed: %w", err)
		}
	}
	req, err := c.newRequest(ctx, http.MethodPut, path, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = c.doRequest(req, result, allowedStatuses...)
	return err
}

// Delete 发送 DELETE 请求。
func (c *Client) Delete(ctx context.Context, path string, result any, allowedStatuses ...int) error {
	req, err := c.newRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	_, err = c.doRequest(req, result, allowedStatuses...)
	return err
}

// Raw 发送原始请求，返回 *http.Response，由调用方自行处理。
func (c *Client) Raw(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	return c.client.Do(req)
}

// ---------- Provider（支持多客户端） ----------

// Provider 管理多个命名 HTTP 客户端。
// 当注册了多个命名客户端时，注入 Provider 通过名称获取对应客户端。
type Provider struct {
	mu      sync.RWMutex
	clients map[string]*Client
}

// NewProvider 创建 Provider。
func NewProvider() *Provider {
	return &Provider{
		clients: make(map[string]*Client),
	}
}

// Add 添加一个命名客户端。
func (p *Provider) Add(name string, client *Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.clients[name] = client
}

// Get 根据名称获取客户端。
// 如果只有一个默认客户端（未命名），可以通过空字符串 "" 获取。
func (p *Provider) Get(name string) *Client {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.clients[name]
}
