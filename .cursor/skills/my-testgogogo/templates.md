# my-testgogogo 代码模板

复制后替换包名、路径、业务字段。所有测试需 `testkit.SkipIfDisabled(t)`。

## configs/config.yaml

```yaml
active: local

test:
  skip_integration: false
  vars:              # 可选：跨环境共享变量
    page_size: 10

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
    token_path: token
    validate:
      url: /api/auth/me
      method: GET

report:
  enabled: true
  base_dir: reports
```

## configs/local.yaml

```yaml
base_url: http://localhost:8080
timeout_seconds: 15
user:
  username: testuser
  password: testpass

vars:                # 可选：环境级变量，覆盖 test.vars 同名项
  user_id: "1001"
```

测试内读取：

```go
// 推荐：自定义结构体，字段与 vars 键名通过 yaml tag 对应
type TestVars struct {
    UserID   string `yaml:"user_id"`
    PageSize int    `yaml:"page_size"`
}

cfg := testkit.LoadConfig(t)
vars := testkit.MustVars[TestVars](t, cfg)
_ = vars.UserID

// 或按 key 逐个读取
userID := testkit.MustVarString(t, cfg, "user_id")
```

## apistep 封装模板

```go
package apistep

import (
    "context"
    "fmt"
    "net/url"
    "strconv"

    "github.com/muyi-zcy/my-testgogogo/client"
)

type Widget struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type PageResult struct {
    Current int64     `json:"current"`
    Size    int64     `json:"size"`
    Total   int64     `json:"total"`
    Records []Widget  `json:"records"`
}

type ListParams struct {
    PageNum  int
    PageSize int
}

func ListWidgets(ctx context.Context, c *client.Client, params ListParams) (*PageResult, error) {
    if params.PageNum <= 0 {
        params.PageNum = 1
    }
    if params.PageSize <= 0 {
        params.PageSize = 10
    }
    q := url.Values{}
    q.Set("pageNum", strconv.Itoa(params.PageNum))
    q.Set("pageSize", strconv.Itoa(params.PageSize))

    var page PageResult
    if err := c.DoJSON(ctx, "GET", "/api/widgets", &page, client.WithQuery(q)); err != nil {
        return nil, fmt.Errorf("list widgets: %w", err)
    }
    return &page, nil
}
```

## API 测试模板

```go
package widget

import (
    "testing"

    "github.com/muyi-zcy/my-testgogogo/assert"
    "github.com/muyi-zcy/my-testgogogo/examples/yourproject/apistep"
    "github.com/muyi-zcy/my-testgogogo/testkit"
    "github.com/stretchr/testify/require"
)

func TestWidgetList(t *testing.T) {
    testkit.SkipIfDisabled(t)
    r := testkit.EnableAPIReport(t, "组件列表查询", "GET /api/widgets 分页")

    c := testkit.NewAuthenticatedClient(t)
    ctx, cancel := testkit.TestContext(t)
    defer cancel()

    r.Step("list widgets", func(t *testing.T) {
        page, err := apistep.ListWidgets(ctx, c, apistep.ListParams{PageNum: 1, PageSize: 10})
        require.NoError(t, err)
        r.SetResponse(page)
        assert.PageNotEmpty(t, page.Total, page.Records)
        r.SetResult(map[string]any{"total": page.Total, "count": len(page.Records)})
    })
}
```

## Flow 测试模板

```go
package example

import (
    "testing"

    "github.com/muyi-zcy/my-testgogogo/examples/yourproject/apistep"
    "github.com/muyi-zcy/my-testgogogo/flow"
    "github.com/muyi-zcy/my-testgogogo/report"
    "github.com/muyi-zcy/my-testgogogo/testkit"
    "github.com/stretchr/testify/require"
)

func TestFlowWidgetQuery(t *testing.T) {
    testkit.SkipIfDisabled(t)

    r := testkit.EnableReport(t, testkit.ReportMeta{
        Generate:    true,
        Title:       "组件查询流程",
        Category:    report.CategoryFlow,
        Description: "系统信息 → 用户信息 → 列表查询",
    })

    cfg := testkit.LoadConfig(t)
    vars := flow.NewVars(flow.DefaultSeed())
    c := testkit.NewClient(t, cfg)
    ctx, cancel := testkit.TestContext(t)
    defer cancel()

    r.Step("get system info", func(t *testing.T) {
        info, err := apistep.GetSystemInfo(ctx, c)
        require.NoError(t, err)
        vars.Set("systemName", info["name"])
        r.SetResult(map[string]any{"system": info})
    })

    authClient := testkit.NewAuthenticatedClient(t)

    r.Step("list widgets", func(t *testing.T) {
        page, err := apistep.ListWidgets(ctx, authClient, apistep.ListParams{
            PageNum: 1, PageSize: vars.MustInt("pageSize"),
        })
        require.NoError(t, err)
        vars.Set("total", int(page.Total))
        r.SetResult(map[string]any{"total": page.Total})
    })
}
```

## 自定义 Provider 注册

`apistep/register.go`：

```go
package apistep

import _ "your-project/internal/myauth"
```

`internal/myauth/provider.go`：

```go
package myauth

import (
    "context"

    "github.com/muyi-zcy/my-testgogogo/auth"
    "github.com/muyi-zcy/my-testgogogo/client"
    "github.com/muyi-zcy/my-testgogogo/config"
)

func init() { auth.RegisterProvider(Provider{}) }

type Provider struct{}

func (Provider) Name() string { return "myauth" }

func (Provider) Authenticate(ctx context.Context, c *client.Client, cfg *config.Config) (auth.Credential, error) {
    // 登录逻辑；最后 c.SetToken(...) 并返回 Credential
    return auth.Credential{Token: "...", Header: "Authorization"}, nil
}

func (Provider) Validate(ctx context.Context, c *client.Client, cred auth.Credential) (bool, error) {
    // 调用校验接口，token 无效返回 false
    return true, nil
}
```

## configs/testdata/flow.yaml

```yaml
page_size: 10
```

供 `flow.DefaultSeed()` 读取，键名为 `pageSize`。

## 示例 Makefile（测试子项目）

```makefile
.PHONY: backend test run test-api test-flow test-report clean

BACKEND_PORT ?= 8080
RUN_ID := $(shell date +%Y%m%d-%H%M%S)

backend:
    BACKEND_PORT=$(BACKEND_PORT) go run ./backend

test:
    go test ./... -v -count=1

test-api:
    go test ./api/... -v -count=1

test-flow:
    go test ./flow/... -v -count=1

run:
    @chmod +x ./scripts/run.sh && ./scripts/run.sh

test-report:
    @mkdir -p reports/api/staging reports/flow/staging
    @MY_TESTGOGOGO_REPORT_RUN_ID=$(RUN_ID) ./scripts/run.sh -json 2>&1 \
        | go run <path-to>/cmd/my-testgogogo report --run-id $(RUN_ID) \
        --command "go test ./... -json -count=1"
```
