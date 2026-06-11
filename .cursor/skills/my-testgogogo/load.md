# 压测速查

完整方案见仓库 [docs/load-testing.md](../../../docs/load-testing.md)。

## 原则

- **不用 Vegeta**：Go loadkit 直接调 `apistep`
- **复用**：`scenario/` 纯函数，功能测试与压测共用
- **报告**：`reports/load/` 独立目录，与功能报告分开

## 决策

```
用户要什么？
├── 压单个 API 接口     → scenario（api 类型）+ my-testgogogo load
├── 压多步 Flow         → scenario（flow 类型）+ per-step 指标
├── 配置压测参数        → configs/config.yaml load 段 或 CLI flags
└── 看压测报告          → reports/load/<日期>/load-report-*.md
```

## 新增压测场景

```
进度：
- [ ] 1. 在 scenario/ 写纯函数 (ctx, *load.Env) error，内部调 apistep
- [ ] 2. 在 scenario/register.go 注册 name / type / title
- [ ] 3. 重构 api/ 或 flow/ 测试调用同一 scenario
- [ ] 4. configs/config.yaml 增加 load.scenarios 项
- [ ] 5. my-testgogogo load --scenario <name>
```

## scenario 模板（API）

```go
func ReceivedList(ctx context.Context, env *load.Env) error {
    params := apistep.WmsReceivedListParams{
        PageNum: 1, PageSize: env.Vars.MustInt("pageSize"),
    }
    resp, err := apistep.ListWmsReceived(ctx, env.Client, params)
    if err != nil {
        return err
    }
    if !resp.Success || resp.Code != 200 {
        return fmt.Errorf("bad response: %d", resp.Code)
    }
    return nil
}
```

## 配置片段

```yaml
load:
  enabled: true
  output_dir: reports/load
  defaults:
    duration: 30s
    rate: 20
    concurrency: 10
    warmup: 5s
  scenarios:
    - name: received-list
      type: api
      enabled: true
```

## CLI

项目需提供 `cmd/load/main.go`（调用 `loadkit.Main(scenario.Registry)`）：

```bash
go run ./cmd/load --scenario received-list --duration 30s --rate 50
# 或在项目目录：
go run ../../cmd/my-testgogogo load --scenario received-list
```

`my-testgogogo load` 会自动委托执行 `./cmd/load`。

## 注意

- 写接口默认 `enabled: false`
- Flow 分支拆成多个 scenario
- 长时压测注意 Token 刷新（Runner 内置）
- 过程图表依赖 Markdown 预览器的 Mermaid 支持（GitHub、VS Code 等）
