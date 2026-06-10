# 认证配置与自定义 Provider

my-testgogogo 采用 **Provider 插件模式**：框架负责 Token 缓存与 HTTP Client 注入，具体登录逻辑由内置配置或你自行实现。

## 三种方式（由简到繁）

### 方式 1：静态 Token（无需写代码）

适合 CI、本地调试、已有长期 Token 的场景。

`configs/local.yaml`：

```yaml
token: "your-fixed-token"
```

`configs/config.yaml`：

```yaml
auth:
  provider: static_token
```

测试中使用：

```go
c := testkit.NewAuthenticatedClient(t)
```

框架会将 Token 写入 `Authorization` 请求头。

---

### 方式 2：YAML 配置登录（无需写代码）

适合标准 REST 登录：POST 用户名密码 → JSON 返回 token。

完整示例见 `examples/library/configs/config.yaml`：

```yaml
auth:
  provider: login
  cache_enabled: true
  cache_dir: .cache/auth
  cache_ttl_hours: 168
  login:
    url: /api/auth/login
    method: POST
    without_auth: true          # 登录请求不带 Authorization
    body:
      username: "{{user.username}}"
      password: "{{user.password}}"
    token_path: token           # 从响应 JSON 提取 token，支持 data.token 点分路径
    validate:
      url: /api/auth/me         # 可选：缓存 token 校验接口
      method: GET
```

`configs/local.yaml`：

```yaml
base_url: http://localhost:18081
user:
  username: librarian
  password: lib123
```

Body 模板目前支持：

| 占位符 | 来源 |
|--------|------|
| `{{user.username}}` | `configs/{active}.yaml` → `user.username` |
| `{{user.password}}` | `configs/{active}.yaml` → `user.password` |
| `{{user.type}}` | `configs/{active}.yaml` → `user.type` |

---

### 方式 3：自定义 Provider（复杂场景）

适合以下情况：

- 验证码 / 多步登录
- 非标准响应格式（如 AjaxResult）
- 特殊 Header（`Bearer xxx`、`X-API-Key`）
- OAuth2 / 签名登录

参考实现：

- **Demo 示例**：`examples/demo/internal/demoauth/provider.go`（自定义 Provider + Bearer Token）
- **Library 示例**：`examples/library/configs/config.yaml`（内置 `login` Provider，YAML 配置即可）

---

## Provider 接口

定义位于 `auth/provider.go`：

```go
type Provider interface {
    Name() string
    Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (Credential, error)
    Validate(ctx context.Context, c *client.Client, cred Credential) (bool, error)
}

type Credential struct {
    Token  string  // 凭证值
    Header string  // 请求头名，默认 "Authorization"
    Prefix string  // 前缀，如 "Bearer "
}
```

| 方法 | 职责 |
|------|------|
| `Name()` | 与配置 `auth.provider` 对应 |
| `Authenticate()` | 执行登录，返回凭证，并通过 `c.SetToken()` 设置到 Client |
| `Validate()` | 校验缓存 Token 是否仍有效 |

`Credential.HeaderValue()` 会自动拼接前缀。例如 `Prefix: "Bearer "` 时，请求头值为 `Bearer xxx`。

若 Provider 需要 Bearer 等前缀，可实现可选接口 `CredentialRehydrator`，用于 Token 缓存命中时恢复完整凭证：

```go
type CredentialRehydrator interface {
    Rehydrate(rawToken string) Credential
}
```

Demo 示例见 `examples/demo/internal/demoauth/provider.go` 中的 `Rehydrate` 方法。

---

## 实现步骤

### 1. 实现 Provider

```go
package myauth

import (
    "context"

    "github.com/muyi-zcy/my-testgogogo/auth"
    "github.com/muyi-zcy/my-testgogogo/client"
    "github.com/muyi-zcy/my-testgogogo/config"
)

type Provider struct{}

func (Provider) Name() string { return "myauth" }

func (Provider) Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (auth.Credential, error) {
    // 1. 若配置了 token，直接使用
    if cfg.Token != "" {
        c.SetToken(cfg.Token)
        return auth.Credential{Token: cfg.Token, Header: "Authorization"}, nil
    }

    // 2. 调用业务登录接口（建议封装在 internal/xxxclient）
    token, err := login(ctx, c, cfg.User)
    if err != nil {
        return auth.Credential{}, err
    }

    c.SetToken(token)
    return auth.Credential{Token: token, Header: "Authorization"}, nil
}

func (Provider) Validate(ctx context.Context, c *client.Client, cred auth.Credential) (bool, error) {
    // 调用 getInfo 等接口验证 token 是否过期
    _, err := getProfile(ctx, c)
    return err == nil, err
}
```

### 2. 注册 Provider

在测试会 import 的包中注册（参考 `examples/demo/apistep/register.go`）：

```go
func init() {
    auth.RegisterProvider(myauth.Provider{})
}
```

`init()` 必须位于测试实际会加载的包中，否则 Provider 不会被注册。推荐单独建 `mytestkit/` 包，所有测试 import 它。

### 3. 配置指定 Provider

`configs/config.yaml`：

```yaml
auth:
  provider: myauth    # 与 Name() 返回值一致
  cache_enabled: true
  cache_dir: .cache/auth
  cache_ttl_hours: 168
```

### 4. 测试中使用

```go
c := testkit.NewAuthenticatedClient(t)
```

或手动控制：

```go
cfg := testkit.LoadConfig(t)
c := testkit.NewClient(t, cfg)
ctx, cancel := testkit.TestContext(t)
defer cancel()

cred, err := auth.Authenticate(ctx, c, cfg)
require.NoError(t, err)
// cred.Token / c.Token() 已就绪
```

---

## 认证流程

```
testkit.NewAuthenticatedClient()
        ↓
auth.Authenticate()
        ↓
ResolveProvider(cfg)          ← 按 auth.provider 查找 Provider
        ↓
Token 缓存命中？ ──是──→ Validate 有效？ ──是──→ 直接使用
        ↓ 否                      ↓ 否
Provider.Authenticate()       清除缓存
        ↓
保存 Token 到 .cache/auth/
        ↓
c.SetToken() → 后续请求自动带 Authorization
```

Token 缓存 Key 格式：`{active}_{base_url}_{username}`。

清除缓存：

```go
testkit.ClearAuthCache(t, cfg)
```

---

## 常见自定义场景

### Bearer Token

```go
return auth.Credential{
    Token:  token,
    Header: "Authorization",
    Prefix: "Bearer ",
}, nil
```

### API Key 放自定义 Header

```go
c.SetAuthHeader("X-API-Key")
c.SetToken(apiKey)
return auth.Credential{Token: apiKey, Header: "X-API-Key"}, nil
```

### 验证码 / 多步登录

在 `Authenticate()` 中编写完整流程。Demo 示例逻辑：

```
sys/login 失败 → 若配置了 captcha_code → captcha/login
```

`captcha_code` 可在 `configs/local.yaml` 中配置，对应 `config.Config.CaptchaCode`。

### 登出并清缓存

```go
err := c.Logout(ctx)           // 业务登出接口
testkit.ClearAuthCache(t, cfg) // 清除本地 token 缓存
```

---

## 推荐项目结构

```
your-project/
├── configs/
│   ├── config.yaml       # auth.provider: myauth
│   └── local.yaml
├── internal/
│   ├── myauth/
│   │   └── provider.go   # 实现 auth.Provider
│   └── myclient/
│       └── client.go     # 封装业务 API + 响应解码
├── mytestkit/
│   └── kit.go            # init() 注册 Provider + 便捷方法
├── apistep/
├── api/
└── flow/
```

---

## 内置 vs 自定义对照

| 场景 | 推荐方式 |
|------|----------|
| 固定 Token | `static_token` |
| POST 登录 + JSON token | `login`（YAML 配置） |
| 验证码 / 非标准 JSON / 特殊 Header | 自定义 Provider（参考 `examples/demo`） |
| 标准 REST 登录 | 内置 `login`（参考 `examples/library`） |

---

## 检查清单

- [ ] `Name()` 与 `configs/config.yaml` 中 `auth.provider` 一致
- [ ] `init()` 中调用 `auth.RegisterProvider(...)`，且测试包会 import 到
- [ ] `Authenticate()` 中调用 `c.SetToken(...)`
- [ ] `Validate()` 能识别过期 Token（否则缓存失效后仍会用旧 Token）
- [ ] 敏感信息放 `local.override.yaml`（已在 `.gitignore` 中忽略）

---

## 配置字段说明

### auth（config.yaml）

| 字段 | 说明 | 默认值 |
|------|------|--------|
| `provider` | 认证方式：`static_token` / `login` / 自定义名 | 有 token 时为 `static_token`，否则 `login` |
| `cache_enabled` | 是否启用 Token 本地缓存 | `true` |
| `cache_dir` | 缓存目录 | `.cache/auth` |
| `cache_ttl_hours` | 缓存有效期（小时） | `168` |

### auth.login（provider=login 时）

| 字段 | 说明 | 必填 |
|------|------|------|
| `url` | 登录接口路径 | 是 |
| `method` | HTTP 方法 | 否，默认 `POST` |
| `without_auth` | 登录请求是否不带 Authorization | 否 |
| `body` | 请求体字段（支持模板变量） | 否，默认 username/password |
| `token_path` | 响应 JSON 中 token 的点分路径 | 否，默认 `token` |
| `validate.url` | 缓存 token 校验接口 | 否 |
| `validate.method` | 校验接口 HTTP 方法 | 否，默认 `GET` |

### 环境配置（local.yaml）

| 字段 | 说明 |
|------|------|
| `base_url` | 后端地址 |
| `user.username` | 登录用户名 |
| `user.password` | 登录密码 |
| `token` | 静态 token（可选，优先于登录） |
| `captcha_code` | 验证码（自定义 Provider 可用，如 demoauth 扩展） |

敏感配置建议写入 `configs/local.override.yaml`，该文件不会提交到 Git。
