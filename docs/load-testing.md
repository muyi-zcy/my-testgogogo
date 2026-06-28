# 压测方案

my-testgogogo 压测能力基于 **Go 原生 loadkit**，直接复用 `apistep`、`scenario`、`api`、`flow`，与功能测试共用配置与认证，输出独立 Markdown 压测报告。

**不依赖 Vegeta 或任何外部 HTTP 压测工具。**

## 目标

| 能力 | 说明 |
|------|------|
| API 压测 | 单次 `apistep` 调用 = 1 次请求 |
| Flow 压测 | 多步 `apistep` 串联 = 1 次 scenario，报告含 per-step 延迟 |
| 复用 | HTTP 层只用 `apistep`；编排抽成 `scenario/`，功能测试与压测共用 |
| 配置 | `configs/config.yaml` 的 `load` 段 + CLI 覆盖 |
| 报告 | `reports/load/<日期>/load-report-<run-id>-<scenario>.md` |

## 架构

```
configs/config.yaml (load.*)
        │
        ▼
cmd/my-testgogogo load ──► loadkit.Runner
        │                      │
        │                      ├── auth.Authenticate（复用框架认证）
        │                      ├── worker pool + rate limiter
        │                      └── 指标采集（histogram）
        │
        ▼
scenario/（项目代码）
        │
        ├── apistep/（HTTP 封装，不变）
        │
        ├── api/*_test.go   → scenario + 断言 + 功能报告
        └── flow/*_test.go  → scenario + 断言 + 功能报告
        │
        ▼
load/report → Markdown 压测报告
```

### 模块职责

| 包 / 目录 | 职责 |
|-----------|------|
| `load/` | Runner、速率控制、指标、Markdown 报告生成 |
| `loadkit/` | 对外 API（类似 `testkit`） |
| `scenario/`（项目内） | 纯函数编排，签名 `(ctx, *runtime.Env, ...) (result, error)` |
| `runtime` | 统一运行时 Env（Client、AuthClient、Vars、Context） |
| `apistep/` | HTTP 调用（功能测试与压测共用） |
| `api/`、`flow/` | 调用 scenario + 断言 + 功能报告 |

## 核心抽象

```go
// runtime.Env 功能测试与压测共用的运行时环境
type Env struct {
    CTX        context.Context
    Config     *config.Config
    Client     *client.Client      // 匿名
    AuthClient *client.Client      // 已认证
    Vars       *flow.Vars
}

// load.Env 嵌入 runtime.Env，额外支持压测指标 env.Record()

// Scenario 可被功能测试与压测共用
type Scenario func(ctx context.Context, env *Env) error

// ScenarioMeta 注册信息
type ScenarioMeta struct {
    Name  string
    Type  ScenarioType // api | flow
    Title string
    Fn    Scenario
    Steps []string     // flow 类型：步骤名，用于 per-step 指标
}
```

## 目录结构（示例项目）

```
your-project/
├── configs/
│   ├── config.yaml      # 增加 load 段
│   └── local.yaml
├── apistep/             # HTTP 封装（不变）
├── scenario/            # 从 api/flow 提取的纯函数
│   ├── received.go
│   └── register.go      # Registry map[name]ScenarioMeta
├── api/                 # 调用 scenario + 断言
├── flow/                # 调用 scenario + 断言
└── go.mod
```

## 配置

`configs/config.yaml`：

```yaml
load:
  enabled: true
  require_confirm: false   # true 时必须 CLI 传 --confirm
  allowed_active:          # 非空时仅允许 listed active 环境压测
    - local
  output_dir: reports/load
  defaults:
    duration: 30s       # 压测时长
    rate: 20            # 目标 QPS（API）或 scenario/s（Flow）
    concurrency: 10     # 最大并发 worker
    warmup: 5s          # 预热，不计入报告
    timeout: 30s        # 单次 scenario 超时
  scenarios:
    - name: received-list
      type: api
      enabled: true
    - name: received-query-by-id
      type: api
      enabled: false
      rate: 10          # 覆盖 defaults
    - name: item-query-flow
      type: flow
      enabled: true
      duration: 60s
      concurrency: 5
```

CLI 覆盖示例：

```bash
my-testgogogo load --scenario received-list --duration 30s --rate 50 --concurrency 20
my-testgogogo load --all
```

## 复用模式

### apistep — 完全复用

压测与功能测试调用同一套 `apistep.*` 函数，不维护第二套 URL。

### API — 提取 scenario

```go
// scenario/received.go
func ReceivedList(ctx context.Context, env *load.Env) error {
    params := apistep.WmsReceivedListParams{
        PageNum: 1, PageSize: env.Vars.MustInt("pageSize"),
    }
    resp, err := apistep.ListWmsReceived(ctx, env.Client, params)
    if err != nil {
        return err
    }
    if !resp.Success || resp.Code != 200 {
        return fmt.Errorf("unexpected response: code=%d", resp.Code)
    }
    return nil
}

// api/received/received_test.go
r.Step("list wms received", func(t *testing.T) {
    err := scenario.ReceivedList(ctx, loadkit.NewEnv(cfg, c, vars))
    require.NoError(t, err)
})
```

### Flow — 提取 Run 函数

将 flow 测试中的 apistep 调用链抽为 `scenario.ItemQueryFlow`：

- 功能测试：scenario 成功后做细粒度 `assert`、写 report
- 压测：只关心整体成功率与各步延迟

注意：

- 带分支的 flow 注册为多个 scenario（如 `item-query-has-data`、`item-query-empty`）
- 写接口默认 `enabled: false`，需独立 scenario 与更低 rate

## 执行引擎

```
Runner
├── 预热（warmup，丢弃指标）
├── 速率控制（token bucket / ticker）
├── Worker Pool（concurrency 上限）
├── 每 worker：Authenticate 一次 → 循环执行 Scenario
├── 长时压测：Token 过期时重新 Authenticate
├── 指标：总数、成功率、实际 QPS、延迟分位 + **时间桶过程数据**
├── scenario 可选 `env.Record(name, value)` 上报业务指标
└── 结束 → Markdown（含 Mermaid 图表）+ JSON staging
```

### 时间序列与图表

按 `bucket_interval`（默认 1s）自动采集每个时间桶的 QPS、成功率、p95 延迟。scenario 可通过 `env.Record` 上报业务指标：

```go
env.Record("list_total", float64(resp.Result.Total))
```

报告新增 **压测过程** 章节：数据表格 + Mermaid 折线图（QPS / 延迟 / 成功率 / 自定义指标）。

配置：

```yaml
load:
  defaults:
    bucket_interval: 1s
  report:
    charts: true
    chart_metrics:   # 要画图的 custom 指标，空则全部
      - list_total
```

## 压测报告

输出路径：

```
reports/load/<日期>/load-report-<run-id>-<scenario>.md
reports/load/staging/<run-id>/<scenario>.json
```

与功能测试报告**并列三套目录**（api / flow / load），不混用。

## 功能测试报告目录

```
reports/
├── api/
│   ├── staging/<run-id>/          # Fragment JSON
│   ├── <日期>/api-report-*.md
│   └── latest.md
├── flow/
│   ├── staging/<run-id>/
│   ├── <日期>/flow-report-*.md
│   └── latest.md
└── load/
    ├── staging/<run-id>/
    └── <日期>/load-report-*.md
```

## CLI 与 Make

每个压测项目需提供 `cmd/load/main.go`：

```go
os.Exit(loadkit.Main(scenario.Registry))
```

```bash
# 项目内直接运行
go run ./cmd/load --scenario received-list

# 框架 CLI（在项目目录下，自动委托 cmd/load）
go run ../../cmd/my-testgogogo load --scenario received-list --duration 30s --rate 50
my-testgogogo load --all

# 示例 Makefile
make load-api
make load-flow
```

## 实施阶段

| 阶段 | 内容 | 状态 |
|------|------|------|
| P0 | `load/` + `loadkit/` + Runner + 指标 | ✅ |
| P1 | CLI `load` 子命令 + YAML 配置 | ✅ |
| P2 | Markdown 压测报告 | ✅ |
| P3 | demo/library：`scenario/` + 重构 api/flow | ✅ |
| P4 | Flow scenario + per-step 指标 | ✅ |
| P5 | 示例文档同步 | ✅ |

## 安全与约束

1. **写接口**：默认不启用；启用时需数据隔离或专用测试账号
2. **目标环境**：`load.allowed_active` 限制可压测的 `active` 环境；`load.require_confirm: true` 时需 CLI `--confirm`
3. **Flow 有状态步骤**：每轮 scenario 独立取 ID，避免脏数据
4. **Token 过期**：每个压测 worker 执行前重新 `Authenticate`

## 与功能测试的关系

| | 功能测试 | 压测 |
|---|----------|------|
| 入口 | `go test` | `my-testgogogo load` |
| HTTP | `apistep` | `apistep`（同一套） |
| 编排 | `scenario` 或直接 apistep | `scenario` |
| 断言 | `require` / `assert` | 仅 error / 业务 code |
| 报告 | `reports/api/*.md` | `reports/load/*.md` |
| 配置 | `configs/` + auth | 同上 + `load.*` |
