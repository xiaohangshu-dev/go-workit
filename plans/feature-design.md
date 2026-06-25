# 新增功能详细设计

> 基于 ASP.NET Core 常用功能，为 go-workit 设计三个新模块。

---

## 一、模型验证（Model Validation）

### 目标
类似 ASP.NET Core 的 Data Annotations，通过 struct tag 声明验证规则。

### 设计

```go
// 使用方式
type CreateUserRequest struct {
    Name  string `validate:"required,min=2,max=100"`
    Email string `validate:"required,email"`
    Age   int    `validate:"gte=0,lte=150"`
    Phone string `validate:"phone"`
}

// 验证
errs := validate.Struct(req)
// errs 格式: []validate.ErrorItem
// [{Field: "Name", Message: "名称不能为空"}, ...]
```

### 验证标签

| 标签 | 说明 | 示例 |
|------|------|------|
| `required` | 非空（字符串非零值，指针非 nil） | `validate:"required"` |
| `min=N` | 最小值（数字）或最小长度（字符串） | `validate:"min=1"` |
| `max=N` | 最大值或最大长度 | `validate:"max=100"` |
| `len=N` | 固定长度 | `validate:"len=11"` |
| `email` | 邮箱格式 | `validate:"email"` |
| `phone` | 手机号格式 | `validate:"phone"` |
| `gte=N` | 大于等于 N | `validate:"gte=0"` |
| `lte=N` | 小于等于 N | `validate:"lte=150"` |
| `oneof=a b c` | 枚举值 | `validate:"oneof=admin user"` |

### 集成方式

```go
// 验证 + 自动返回 ProblemDetails
type Handler struct {
    validator *validate.Validator
}

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    c.ShouldBindJSON(&req)
    
    if errs := h.validator.Struct(req); errs != nil {
        // 自动转为 400 ProblemDetails 响应
        c.JSON(400, errs.ToProblemDetails())
        return
    }
}
```

### 文件结构
```
pkg/tools/validate/
├── validate.go       // 核心验证逻辑
├── rules.go          // 内置验证规则
└── errors.go         // 错误类型
```

---

## 二、ProblemDetails 错误响应（RFC 7807）

### 目标
标准化的 API 错误格式，统一错误响应。

### 设计

```go
// 错误响应格式
type ProblemDetails struct {
    Type     string `json:"type,omitempty"`
    Title    string `json:"title"`
    Status   int    `json:"status"`
    Detail   string `json:"detail,omitempty"`
    Instance string `json:"instance,omitempty"`
    Errors   any    `json:"errors,omitempty"`  // 扩展字段，用于验证错误
}

// 使用方式
// 1. 全局中间件（自动捕获异常）
builder.AddProblemDetails()  // 注册中间件

// 2. 手动返回
return nil, problem.New().
    Status(400).
    Title("参数错误").
    Detail("邮箱格式不正确").
    Err()

// 3. 验证错误
return nil, problem.Validation().
    AddError("email", "邮箱格式不正确").
    Err()
```

### 中间件行为

| 场景 | 行为 |
|------|------|
| 验证失败 | 自动返回 400 + 字段级错误 |
| 404 路由 | 自动返回 404 ProblemDetails |
| 500 异常 | 开发环境返回详细堆栈，生产环境返回 500 |
| 手动抛出 | 通过 `problem.Error` |

### 集成方式

```go
builder.AddProblemDetails(func(options *problem.Options) {
    options.ShowStackTrace = gin.Mode() == gin.DebugMode
    options.IncludeStackTrace = false  // 生产环境不显示堆栈
})
```

### 文件结构
```
pkg/webapp/problem/
├── problem.go        // ProblemDetails 定义
├── builder.go        // Builder 扩展
├── middleware.go     // gin 中间件
└── examples_test.go  // 使用示例
```

---

## 三、重试与熔断（Resilience）

### 目标
类似 Polly 的 HTTP 调用弹性策略，集成到 HTTP 客户端。

### 设计

```go
// 使用方式
builder.AddNamedHttpClient("github", func(options *httpclient.Options) {
    options.BaseURL = "https://api.github.com"
    // 重试策略
    options.Retry = httpclient.RetryOptions{
        MaxRetries:  3,
        BaseDelay:   100 * time.Millisecond,
        MaxDelay:    2 * time.Second,
        RetryOn:     []int{408, 429, 500, 502, 503},
    }
    // 熔断器策略
    options.CircuitBreaker = httpclient.CircuitBreakerOptions{
        FailureThreshold:    5,     // 连续失败次数
        SuccessThreshold:    2,     // 恢复后需要成功次数
        HalfOpenMaxRequests: 1,     // 半开状态最大请求数
        OpenDuration:        30 * time.Second, // 断开持续时间
    }
})
```

### 重试策略

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `MaxRetries` | 最大重试次数 | 0（不重试）|
| `BaseDelay` | 初始延迟 | 100ms |
| `MaxDelay` | 最大延迟 | 2s |
| `RetryOn` | 触发重试的状态码 | `[408, 429, 500, 502, 503]` |

### 熔断器状态机

```
Closed ──(连续失败 N 次)──▶ Open ──(等待超时)──▶ HalfOpen
  ▲                                                  │
  └──────────────(请求成功 N 次)──────────────────────┘
  ▲                                                  │
  └──────────────(请求失败 1 次)──────▶ Open ─────────┘
```

### 集成方式
重试和熔断逻辑封装在 `httpclient.Client` 内部，对用户透明：

```go
// 用户无需关心重试和熔断
client.Get(ctx, "/api/users", &users)
// 内部自动：失败 → 重试 → 熔断器断开 → 快速失败
```

### 文件结构
```
pkg/tools/net/
├── http.go              // Client 核心（已有）
├── retry.go             // 重试策略
└── circuitbreaker.go    // 熔断器
```
