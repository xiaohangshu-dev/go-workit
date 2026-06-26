# ASP.NET Core 常用功能 vs go-workit 缺失分析

> 对比 ASP.NET Core，梳理 go-workit 框架尚未实现但较为常用的功能。

---

## 一、中间件与请求处理

### 1.1 结构化中间件管道（已有部分）
ASP.NET Core 的 `IApplicationBuilder.Use()` 中间件管道可清晰控制请求流程：
```csharp
app.UseAuthentication()
   .UseAuthorization()
   .UseExceptionHandler()
   .UseResponseCaching();
```

go-workit 已有：`UseAuthentication()`、`UseAuthorization()`、`UseRecovery()` 等，但缺少统一的中间件管道注册和管理（目前是通过 gin 的 `Use()` 直接注册）。

### 1.2 全局异常处理中间件 ❌
ASP.NET Core 内置 `UseExceptionHandler()` / `UseDeveloperExceptionPage()`，开发环境显示详细错误，生产环境显示友好页面。

go-workit 只有 `recovery` 中间件（捕获 panic），缺少：
- 开发环境返回详细错误堆栈
- 生产环境返回 `ProblemDetails`（RFC 7807）格式
- 自定义异常处理回调

### 1.3 响应压缩 ❌
ASP.NET Core 内置 `UseResponseCompression()`，支持 Brotli/Gzip/Deflate 压缩响应内容。

### 1.4 Problem Details（RFC 7807）❌
ASP.NET Core 的 `InvalidModelStateResult`、`ProblemDetails` 提供标准化的错误响应格式：
```json
{
    "type": "https://errors.example.com/validation",
    "title": "Validation Error",
    "status": 400,
    "detail": "...",
    "errors": { "field": ["error msg"] }
}
```

---

## 二、模型验证与绑定

### 2.1 声明式模型验证 ❌
ASP.NET Core 使用 Data Annotations：
```csharp
public class CreateUserRequest {
    [Required]
    [StringLength(100)]
    public string Name { get; set; }
    
    [EmailAddress]
    public string Email { get; set; }
}
```

go-workit 的 `pkg/tools/validate` 只有基础函数（`IsEmail`、`IsPhone`），缺少：
- 结构体标签验证（`validate:"required,min=3,max=100"`）
- 嵌套结构体验证
- 自定义验证规则
- 多语言错误消息

### 2.2 自动请求绑定 + 验证 ❌
ASP.NET Core 通过 `[FromBody]`、`[FromQuery]` 自动绑定请求到模型，并自动执行验证。

go-workit 需用户手动调用 `c.ShouldBindJSON()`。

---

## 三、弹性与容错

### 3.1 重试策略（Polly）❌
ASP.NET Core 通过 Polly 集成：
```csharp
builder.Services.AddHttpClient("github")
    .AddTransientHttpErrorPolicy(p => p.RetryAsync(3))
    .AddCircuitBreakerHandler(...);
```

go-workit 缺少：
- HTTP 调用自动重试
- 熔断器（Circuit Breaker）
- 超时策略
- 舱壁隔离（Bulkhead）

go-workit 当前通过 `builder.AddHttpClientContext(...)` 统一注册 HTTP 客户端。

---

## 四、缓存

### 4.1 响应缓存 ❌
ASP.NET Core 的 `[ResponseCache]` 属性控制客户端和服务端缓存：
```csharp
[ResponseCache(Duration = 60, Location = ResponseCacheLocation.Client)]
```

### 4.2 输出缓存中间件 ❌
ASP.NET Core 7+ 的 `UseOutputCache()` 中间件自动缓存响应。

---

## 五、API 治理

### 5.1 API 版本控制 ❌
ASP.NET Core 通过 `AddApiVersioning()` 支持 URL Path、Query String、Header 等多种版本策略。

### 5.2 请求速率限制（已有基础）
ASP.NET Core 内置 `AddRateLimiter()`。go-workit 已有 4 种限流策略，但缺少：
- 全局限流策略配置
- 限流响应头（`RateLimit-Remaining` 等）

### 5.3 CORS（已有）
go-workit 已有 `UseCORS()`。

---

## 六、安全

### 6.1 CSRF 防护 ❌
ASP.NET Core 内置 Antiforgery Token 防跨站请求伪造。

### 6.2 安全响应头 ❌
ASP.NET Core 可配置安全头：`X-Content-Type-Options`、`X-Frame-Options`、`Content-Security-Policy` 等。

---

## 七、其他

### 7.1 多环境配置（已有基础）
ASP.NET Core 根据 `ASPNETCORE_ENVIRONMENT` 自动加载 `appsettings.{Environment}.json`。

go-workit 已有 `server.environment` 配置，但未实现按环境加载不同配置文件。

### 7.2 后台任务（已有）
go-workit 已有 `BackgroundService`，与 ASP.NET Core 的 `IHostedService` 类似。

### 7.3 健康检查（已有）
go-workit 已有 `AddHealthChecks()`。

---

## 优先级建议

| 功能 | 优先级 | 说明 |
|------|--------|------|
| 模型验证（struct tags） | **P0** | 开发中最常用的功能之一 |
| 全局异常处理 + ProblemDetails | **P0** | 提升 API 一致性 |
| 弹性策略（重试/熔断） | **P1** | 微服务场景关键 |
| 响应压缩 | **P1** | 性能优化 |
| API 版本控制 | **P1** | API 演进必备 |
| CSRF 防护 | **P2** | Web 应用安全 |
| 响应缓存 | **P2** | 性能优化 |
