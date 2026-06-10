# my-testgogogo 认证速查

完整文档：`docs/auth-config.md`

## 选型

| 场景 | provider | 需要写代码 |
|------|----------|-----------|
| CI / 已有长期 Token | `static_token` | 否 |
| POST 登录 → JSON token | `login` | 否 |
| Bearer / 验证码 / 嵌套 JSON / 多步登录 | 自定义名（如 `demoauth`） | 是 |

参考实现：

- YAML login：`examples/library/configs/config.yaml`
- 自定义 Provider：`examples/demo/internal/demoauth/provider.go`

## static_token

`configs/config.yaml`：

```yaml
auth:
  provider: static_token
```

`configs/local.yaml`：

```yaml
base_url: http://localhost:8080
token: "your-fixed-token"
```

## login（YAML）

`configs/config.yaml`：

```yaml
auth:
  provider: login
  cache_enabled: true
  cache_dir: .cache/auth
  cache_ttl_hours: 168
  login:
    url: /api/auth/login
    method: POST
    without_auth: true
    body:
      username: "{{user.username}}"
      password: "{{user.password}}"
    token_path: token          # 支持 data.token 点分路径
    validate:
      url: /api/auth/me
      method: GET
```

Body 模板变量：`{{user.username}}` / `{{user.password}}` / `{{user.type}}`

## 自定义 Provider

### 接口

```go
type Provider interface {
    Name() string
    Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (Credential, error)
    Validate(ctx context.Context, c *client.Client, cred Credential) (bool, error)
}
```

### 检查清单

- [ ] `Name()` 与 `auth.provider` 一致
- [ ] `init()` 中 `auth.RegisterProvider`，且测试会 import 该包
- [ ] `Authenticate()` 内调用 `c.SetToken(...)`
- [ ] `Validate()` 能识别过期 token
- [ ] Bearer 前缀：返回 `Credential{Prefix: "Bearer "}` 或实现 `CredentialRehydrator`
- [ ] 密码放 `local.override.yaml`，不入库

### 注册方式

在 `apistep/register.go` blank import：

```go
import _ "your-project/internal/myauth"
```

### 认证流程

```
NewAuthenticatedClient()
  → auth.Authenticate()
  → 缓存命中？→ Validate 有效？→ 使用缓存
  → 否则 Provider.Authenticate() → 写缓存 → c.SetToken()
```

缓存 Key：`{active}_{base_url}_{username}`

清除缓存：`testkit.ClearAuthCache(t, cfg)`

## 测试中用法

```go
// 自动认证（推荐）
c := testkit.NewAuthenticatedClient(t)

// 手动控制
cfg := testkit.LoadConfig(t)
c := testkit.NewClient(t, cfg)
cred, err := auth.Authenticate(ctx, c, cfg)
```

## 常见变体

**Bearer Token**：

```go
return auth.Credential{Token: token, Header: "Authorization", Prefix: "Bearer "}, nil
```

**API Key Header**：

```go
c.SetAuthHeader("X-API-Key")
c.SetToken(apiKey)
return auth.Credential{Token: apiKey, Header: "X-API-Key"}, nil
```

**登出**：

```go
_ = c.Logout(ctx)              // 业务登出
testkit.ClearAuthCache(t, cfg) // 清本地缓存
```
