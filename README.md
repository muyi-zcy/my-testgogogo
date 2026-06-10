<p align="center">
  <img src="logo.png" alt="testgogogo" width="120" height="120" />
</p>

# my-testgogogo

Go API 集成测试框架：单接口契约验证、Flow 流程编排、可插拔认证、Markdown 测试报告。

适合对接真实后端或 mock 服务，用标准 `go test` 编写、运行和 CI 集成，无需额外测试运行器。

## 特性

- **单接口测试（API）** — 封装 HTTP 调用与断言，验证单个 REST 接口的契约与响应结构
- **流程测试（Flow）** — 多步骤串联，跨步骤变量传递（`flow.Vars`）与条件分支
- **可插拔认证** — 静态 Token、YAML 配置登录、自定义 `auth.Provider` 三种方式
- **Token 缓存** — 本地缓存登录凭证，避免每个用例重复登录
- **Markdown 报告** — 采集步骤、结构化结果与变量，生成可读测试报告
- **多环境配置** — YAML 分层配置，支持 `local.override.yaml` 存放敏感信息
- **一体化示例** — 自带 mock 后端，clone 即可运行完整演示

## 环境要求

- Go 1.22+
- Make（可选，用于一键命令）

## 快速开始

```bash
git clone https://github.com/muyi-zcy/my-testgogogo.git
cd my-testgogogo

make tidy          # 整理根模块与示例依赖
make demo-run      # demo 示例（端口 18080）
make library-run   # library 示例（端口 18081）
```

或进入示例目录单独运行：

```bash
cd examples/demo && make run
cd examples/library && make run
```

## 架构

```
my-testgogogo/
├── client/            通用 HTTP 客户端（Token 注入、DoJSON）
├── auth/              认证 Provider 注册 + Token 缓存
├── config/            多环境 YAML 配置加载
├── assert/            常用断言（HTTP 状态码、JSON 路径、分页）
├── report/            报告采集、Fragment 暂存、Markdown 生成
├── flow/              Flow 编排（Step 子测试、Vars 变量）
├── testkit/           测试便捷入口（配置、客户端、认证、报告）
├── cmd/my-testgogogo/ CLI 工具（report 子命令）
└── examples/
    ├── demo/          商品商城：自定义 demoauth 认证
    └── library/       图书管理：内置 login 认证
```

### 模块职责

| 包 | 职责 |
|----|------|
| `testkit` | 测试入口：`SkipIfDisabled`、`NewAuthenticatedClient`、`EnableReport` |
| `client` | HTTP 请求封装，自动拼接 baseURL、注入 Authorization |
| `auth` | Provider 插件模式，负责登录与 Token 缓存校验 |
| `config` | 读取 `configs/config.yaml` + `configs/<env>.yaml` 并合并 |
| `apistep`（示例内） | 业务 API 封装，供 `api/` 与 `flow/` 复用 |
| `report` | 用例级步骤采集，配合 CLI 合并为批次 Markdown 报告 |

## 两个示例对比

| | demo | library |
|---|------|---------|
| 业务 | 商品商城 | 图书管理 |
| 端口 | 18080 | 18081 |
| 账号 | `demo` / `demo123` | `librarian` / `lib123` |
| 认证 | 自定义 `demoauth` Provider | 内置 `login`（纯 YAML） |
| 适用场景 | 验证码、Bearer、非标准 JSON | 标准 POST 登录 + JSON token |

详见 [examples/README.md](examples/README.md)、[demo 说明](examples/demo/README.md)、[library 说明](examples/library/README.md)。

## 核心概念

### 单接口测试（API）

每个测试函数验证一个接口的行为。推荐将 HTTP 调用封装在 `apistep/`，测试文件只负责编排与断言。

```go
func TestBookList(t *testing.T) {
    testkit.SkipIfDisabled(t)
    r := testkit.EnableAPIReport(t, "图书列表查询", "GET /api/books 分页查询")

    c := testkit.NewAuthenticatedClient(t)
    ctx, cancel := testkit.TestContext(t)
    defer cancel()

    r.Step("list books", func(t *testing.T) {
        page, err := apistep.ListBooks(ctx, c, apistep.ListParams{PageNum: 1, PageSize: 10})
        require.NoError(t, err)
        assert.PageNotEmpty(t, page.Total, page.Records)
        r.SetResult(map[string]any{"total": page.Total, "count": len(page.Records)})
    })
}
```

### 流程测试（Flow）

多个步骤按业务顺序执行，前序步骤通过 `flow.Vars` 向后序传递数据，支持 `switch` 条件分支。

典型模式：

1. `flow.NewVars(flow.DefaultSeed())` 初始化变量
2. `r.Step(...)` 包装每个步骤（对应报告中的一行）
3. `vars.Set("key", value)` 写入，`vars.MustString("key")` 读取
4. 根据 `vars.Get("branch")` 走不同分支

完整示例见 `examples/demo/flow/example/item_query_test.go`。

### 推荐项目结构

在自己的项目中引用本框架时，建议目录如下：

```
your-project/
├── configs/
│   ├── config.yaml           # active 环境、auth、report
│   ├── local.yaml            # base_url、账号
│   └── local.override.yaml   # 敏感配置（.gitignore）
├── apistep/                  # API 封装（可被 api/ 与 flow/ 复用）
├── api/                      # 单接口测试
├── flow/                     # 流程测试
├── internal/myauth/          # 自定义 Provider（可选）
└── go.mod                    # require github.com/muyi-zcy/my-testgogogo
```

## 配置

配置文件位于各测试项目的 `configs/` 目录（示例中为 `examples/demo/configs/` 等）。

### 文件说明

| 文件 | 用途 |
|------|------|
| `config.yaml` | 根配置：`active` 环境名、`test.skip_integration`、`auth`、`report` |
| `local.yaml` | 环境配置：`base_url`、测试账号、`token` |
| `local.override.yaml` | 本地覆盖（密码、Token 等），不入库 |

### 最小配置示例

`configs/config.yaml`：

```yaml
active: local

test:
  skip_integration: false   # true 时跳过所有集成测试

auth:
  provider: login           # static_token / login / 自定义名
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

report:
  enabled: true
  output_dir: reports
  staging_dir: reports/staging
```

`configs/local.yaml`：

```yaml
base_url: http://localhost:18081
timeout_seconds: 15
user:
  username: librarian
  password: lib123
```

### 跳过集成测试

CI 中若暂时无法连接后端，可在 `config.yaml` 设置：

```yaml
test:
  skip_integration: true
```

所有调用 `testkit.SkipIfDisabled(t)` 的用例将自动 `t.Skip`。

## 认证

框架采用 **Provider 插件模式**：Token 缓存与 HTTP 注入由框架负责，登录逻辑由配置或自定义代码实现。

| 方式 | 配置 | 适用场景 |
|------|------|----------|
| 静态 Token | `auth.provider: static_token` + `local.yaml` 中 `token` | CI、已有长期 Token |
| YAML 登录 | `auth.provider: login` + `auth.login` 段 | 标准 REST 登录 |
| 自定义 Provider | `auth.provider: myauth` + `auth.RegisterProvider` | 验证码、Bearer、OAuth 等 |

测试中获取已认证客户端：

```go
c := testkit.NewAuthenticatedClient(t)
```

完整说明见 [认证配置与自定义 Provider](docs/auth-config.md)。

## Markdown 报告

启用报告后，每个用例会采集步骤、备注、结构化结果，并通过 CLI 合并为批次 Markdown。

### 在测试中启用

```go
r := testkit.EnableAPIReport(t, "用例标题", "用例说明")
// 或 Flow 测试：
r := testkit.EnableReport(t, testkit.ReportMeta{
    Generate: true, Title: "流程标题", Category: report.CategoryFlow,
})

r.Step("步骤名", func(t *testing.T) { /* ... */ })
r.Note("备注信息")
r.SetResult(map[string]any{"key": "value"})
```

### 生成报告

在示例目录中：

```bash
cd examples/demo
make test-report    # 跑测试 + 生成 Markdown
```

报告输出至 `reports/<日期>/test-report-<run-id>-<用例名>.md`，包含环境信息、步骤耗时、JSON 结构化结果等。

CLI 也可单独使用：

```bash
go test -json ./... | go run ../../cmd/my-testgogogo report \
  --run-id 20260610-120000 \
  --command "go test ./... -json"
```

## 日常命令

根目录 `Makefile`：

```bash
make help           # 查看所有命令
make tidy           # go mod tidy（根模块 + 两个示例）
make fmt            # 格式化全部代码
make build          # 编译框架与 CLI（输出 bin/my-testgogogo）
make check          # fmt + build + demo-run + library-run
make clean          # 清理 .cache、bin 与示例报告
```

Demo 示例：

```bash
make demo-run       # 一键：启动 backend + 跑全部测试
make demo-backend   # 仅启动 mock 后端（:18080）
make demo-api       # 单接口测试
make demo-flow      # 流程测试
make demo-report    # 测试 + Markdown 报告
```

Library 示例：

```bash
make library-run    # 一键：启动 backend + 跑全部测试
make library-api    # 单接口测试
make library-flow   # 流程测试
make library-report # 测试 + Markdown 报告
```

## 在新项目中使用

1. 添加依赖：

```bash
go get github.com/muyi-zcy/my-testgogogo
```

2. 创建 `configs/` 目录，按上文配置 `base_url` 与认证方式
3. 编写 `apistep/` 封装 API，`api/` 编写单接口测试
4. 需要多步骤流程时，在 `flow/` 中使用 `flow.Vars` 与 `r.Step`
5. 复杂认证场景参考 [demo 自定义 Provider](examples/demo/internal/demoauth/provider.go)

## 文档

- [Examples 说明](examples/README.md) — 示例结构与新建指南
- [认证配置与自定义 Provider](docs/auth-config.md) — 三种认证方式详解
- [demo 示例](examples/demo/README.md) — 商品商城 + demoauth
- [library 示例](examples/library/README.md) — 图书管理 + 内置 login

## Cursor Skills

仓库内置 Agent Skills，在 Cursor 中可自动辅助编写测试、配置认证与脚手架代码：

| Skill | 用途 |
|-------|------|
| `my-testgogogo` | 框架总览：命令、决策树、API/Flow/报告/认证 |
| `my-testgogogo-scaffold` | 脚手架：新建测试、apistep、示例、自定义 Provider |

技能文件位于 `.cursor/skills/`，含代码模板（`templates.md`）与认证速查（`auth.md`）。

## License

MIT
