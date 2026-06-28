---
name: my-testgogogo
description: >-
  Guides API integration testing with the my-testgogogo Go framework: API/Flow
  tests, load testing, YAML config, auth providers, Markdown reports, and Make
  commands. Use when writing or running integration tests, load tests,
  scaffolding test projects, configuring auth, generating reports, or working in
  this repository.
---

# my-testgogogo

Go API 集成测试框架。用标准 `go test` 编写单接口测试与 Flow 流程测试，支持可插拔认证与 Markdown 报告。

## 决策树

```
用户要什么？
├── 跑通示例 / 验证环境     → 「快速验证」
├── 在新项目接入框架        → 「接入新项目」
├── 写单接口测试            → 「编写 API 测试」+ templates.md
├── 写多步骤流程测试        → 「编写 Flow 测试」+ templates.md
├── API / Flow 压测         → 「压测」+ load.md
├── 配置登录 / Token        → 「认证配置」+ auth.md
├── 生成 Markdown 报告      → 「生成报告」
└── 新建一体化示例          → 「新建示例」
```

## 快速验证

```bash
make tidy
make demo-run      # 商品商城，:18080，demoauth
make library-run   # 图书管理，:18081，内置 login
make check         # fmt + build + 两个示例全量测试
```

示例自带 mock backend，`make run` 会自动启停后端。

| 示例 | 端口 | 账号 | 认证 |
|------|------|------|------|
| `examples/demo` | 18080 | demo / demo123 | 自定义 `demoauth` |
| `examples/library` | 18081 | librarian / lib123 | YAML `login` |

## 接入新项目

1. `go get github.com/muyi-zcy/my-testgogogo`
2. 创建目录结构（见下方）
3. 编写 `configs/config.yaml` + `configs/local.yaml`（可在 `test.vars` / `vars` 中定义全局变量）
4. 在 `apistep/` 封装 HTTP 调用，在 `api/` 写单接口测试
5. 每个测试文件开头调用 `testkit.SkipIfDisabled(t)`

```
your-project/
├── configs/
│   ├── config.yaml
│   ├── local.yaml
│   └── local.override.yaml   # 敏感信息，不入库
├── apistep/                  # HTTP 封装 + DTO
├── scenario/                 # 业务编排（api / flow / 压测共用）
├── api/                      # 单接口测试（scenario + 断言 + 报告）
├── flow/                     # 流程测试（scenario + 分支断言 + 报告）
├── cmd/load/                 # 压测入口
├── configs/testdata/flow.yaml  # Flow 种子数据（可选）
└── go.mod
```

**认证选型**（详见 [auth.md](auth.md)）：

- 固定 Token → `static_token`
- 标准 POST 登录 → `login`（参考 `examples/library`）
- 验证码 / Bearer / 非标准 JSON → 自定义 Provider（参考 `examples/demo`）

## 编写 API 测试

**规则**：

- HTTP 调用放 `apistep/`，业务编排放 `scenario/`，测试文件只做断言与报告
- 使用 `testkit.NewScenarioEnv(t)` 获取运行时环境（含匿名 + 已认证客户端）
- 需要报告时调用 `testkit.EnableAPIReport(t, title, desc)`

模板见 [templates.md](templates.md#api-测试模板)。

参考：`examples/library/api/book/book_test.go`

## 编写 Flow 测试

**规则**：

- 多步编排放 `scenario/` 的 Flow 函数（如 `ItemQueryFlow`）
- Flow 内用 `env.RunStep("step-name", fn)` 包装各步（压测可采集 per-step 延迟）
- 测试文件调用 scenario，根据 `env.Vars.Get("branch")` 做分支断言
- 报告分类用 `report.CategoryFlow`

模板见 [templates.md](templates.md#flow-测试模板)。

参考：`examples/demo/flow/example/item_query_test.go`

## 认证配置

三种方式由简到繁：静态 Token → YAML login → 自定义 Provider。

自定义 Provider 必须：

1. 实现 `auth.Provider`（`Name` / `Authenticate` / `Validate`）
2. 在 `init()` 中 `auth.RegisterProvider(...)`
3. `init()` 所在包被测试 import（推荐 blank import 在 `apistep/register.go`）
4. `configs/config.yaml` 中 `auth.provider` 与 `Name()` 一致

详见 [auth.md](auth.md) 与仓库 `docs/auth-config.md`。

## 生成报告

`configs/config.yaml` 启用 report 后，测试中调用 `EnableAPIReport` 或 `EnableReport`。

```bash
cd examples/demo && make test-report
# API：reports/api/<日期>/api-report-<run-id>.md
# Flow：reports/flow/<日期>/flow-report-<run-id>.md
```

Reporter API：`r.Step` / `r.Note` / `r.SetResponse(resp)` / `r.SetResult(map[string]any{...})`

## 压测

基于 **loadkit**（Go 原生），直接复用 `apistep` 与 `scenario`，不依赖 Vegeta。

- HTTP 仍走 `apistep/`；编排抽到 `scenario/`，`api/` 与 `flow/` 测试共用
- 配置：`configs/config.yaml` 的 `load` 段；CLI：`my-testgogogo load`
- 功能报告：`reports/api/`、`reports/flow/`（按类型分目录）
- 压测报告：`reports/load/<日期>/load-report-<run-id>-<scenario>.md`

详见 [load.md](load.md) 与仓库 `docs/load-testing.md`。

## 日常命令

| 命令 | 作用 |
|------|------|
| `make demo-run` / `make library-run` | 一键启后端 + 全量测试 |
| `make demo-api` / `make library-api` | 仅单接口测试 |
| `make demo-flow` / `make library-flow` | 仅流程测试 |
| `make demo-report` / `make library-report` | 测试 + Markdown 报告 |
| `my-testgogogo load --scenario <name>` | 压测（loadkit，复用 scenario） |
| `make clean` | 清理 `.cache`、报告 |
| `test.skip_integration: true` | CI 跳过所有集成测试 |

## 新建示例

复制 `examples/demo` 或 `examples/library`：

1. 修改 `backend/` 实现 mock 接口
2. 更新 `configs/local.yaml` 的 `base_url` 与端口
3. 编写 `apistep/`、`scenario/`、`api/`、`flow/`
4. 选择 `login` 或自定义 Provider
5. 添加 `cmd/load/` 与 `scenario/register.go`（压测）
6. 在根 `Makefile` 添加对应 target（可选）

## 常见错误

| 现象 | 处理 |
|------|------|
| `authenticate before test` 失败 | 检查 `base_url`、账号、`auth.provider` |
| Provider not found | 确认 `RegisterProvider` 的包被测试 import |
| 缓存 Token 过期仍使用 | 实现 `Validate()` 返回 false，或 `testkit.ClearAuthCache` |
| 配置加载失败 | 确认从项目根运行测试（`path.ModuleRoot` 依赖 go.mod） |
| 报告未生成 | 确认 `report.enabled: true` 且 `Generate: true` |

## 参考文件

- 代码模板：[templates.md](templates.md)
- 认证速查：[auth.md](auth.md)
- 压测方案：[load.md](load.md)、`docs/load-testing.md`
- 仓库文档：`README.md`、`docs/auth-config.md`、`examples/README.md`
