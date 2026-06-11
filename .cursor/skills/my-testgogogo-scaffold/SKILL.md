---
name: my-testgogogo-scaffold
description: >-
  Scaffolds my-testgogogo integration test projects: new API tests, Flow tests,
  apistep wrappers, configs, auth providers, and example directories. Use when
  the user asks to create, add, or bootstrap tests, examples, or test project
  structure for my-testgogogo.
---

# my-testgogogo Scaffold

按用户意图生成测试项目文件。生成前先看 [my-testgogogo SKILL](../my-testgogogo/SKILL.md) 的决策树。

## 工作流

### 新增压测场景

```
进度：
- [ ] 1. 在 scenario/ 写 (ctx, *load.Env) error，内部调 apistep
- [ ] 2. 在 scenario/register.go 注册
- [ ] 3. 重构 api/ 或 flow/ 测试共用 scenario
- [ ] 4. configs/config.yaml 增加 load.scenarios
- [ ] 5. my-testgogogo load --scenario <name>
```

详见 [load.md](../my-testgogogo/load.md) 与 `docs/load-testing.md`。

### 新增单接口测试

```
进度：
- [ ] 1. 确认 apistep 是否已有封装；无则先建 apistep
- [ ] 2. 在 api/<module>/ 创建 *_test.go
- [ ] 3. 确认 configs/ 中 base_url 与 auth 可用
- [ ] 4. 运行 go test ./api/<module>/... -v -count=1
```

**文件约定**：

- 包名与目录一致（如 `api/book/` → `package book`）
- 首行 `testkit.SkipIfDisabled(t)`
- 需要报告：`testkit.EnableAPIReport(t, title, desc)`
- 模板见 [templates.md](../my-testgogogo/templates.md#api-测试模板)

### 新增 Flow 测试

```
进度：
- [ ] 1. 确认 apistep 覆盖流程所需接口
- [ ] 2. 在 flow/<name>/ 创建测试，使用 flow.Vars + r.Step
- [ ] 3. 可选：configs/testdata/flow.yaml 配置 page_size
- [ ] 4. 运行 go test ./flow/... -v -count=1
```

模板见 [templates.md](../my-testgogogo/templates.md#flow-测试模板)。参考 `examples/demo/flow/example/item_query_test.go`。

### 新增 apistep 封装

- 放 `apistep/` 包，函数签名 `(ctx, *client.Client, params) (result, error)`
- 用 `c.DoJSON` 发请求，错误用 `fmt.Errorf("action: %w", err)` 包装
- 无需认证：`client.WithoutAuth()`
- 查询参数：`client.WithQuery(url.Values{})`
- 请求体：`client.WithBody(map[string]any{...})`

### 接入框架到新项目

```
进度：
- [ ] 1. go get github.com/muyi-zcy/my-testgogogo
- [ ] 2. 创建 configs/config.yaml + configs/local.yaml
- [ ] 3. 创建 apistep/、api/ 目录
- [ ] 4. 选认证方式（见 auth.md）
- [ ] 5. 写一个 Smoke 测试验证连通性
- [ ] 6. go test ./... -v -count=1
```

目录结构见 [my-testgogogo SKILL](../my-testgogogo/SKILL.md#接入新项目)。

### 新建一体化示例

复制 `examples/library`（标准 login）或 `examples/demo`（自定义 auth）：

1. 重命名并修改 `go.mod` module path
2. 实现 `backend/main.go` mock 接口
3. 更新 `configs/local.yaml` 端口与 `base_url`
4. 编写 `apistep/`、`api/`、`flow/`
5. 更新 `scripts/run.sh` 端口
6. 根 `Makefile` 添加 `make <name>-run`（可选）

### 新增自定义认证

```
进度：
- [ ] 1. 创建 internal/<name>auth/provider.go
- [ ] 2. init() 注册 auth.RegisterProvider
- [ ] 3. apistep/register.go blank import
- [ ] 4. config.yaml 设置 auth.provider
- [ ] 5. 实现 Validate 防止过期缓存
```

详见 [auth.md](../my-testgogogo/auth.md)。

## 生成代码原则

1. **apistep 与测试分离** — 不在 `*_test.go` 里写原始 HTTP 调用
2. **复用 testkit** — 不直接 new `client.Client` 除非 Flow 需要先未认证再认证
3. **最小断言** — 用 `assert.HTTPStatus` / `assert.JSONFieldEqual` / `assert.PageNotEmpty`
4. **报告可选** — 用户未要求报告时，`EnableAPIReport` 可省略
5. **匹配现有风格** — 先读同目录已有文件再生成

## 验证命令

```bash
# 单包
go test ./api/book/... -v -count=1

# 全量（需 backend 运行或 make run）
go test ./... -v -count=1

# 带报告
make test-report
```

## 参考

- 模板：[templates.md](../my-testgogogo/templates.md)
- 认证：[auth.md](../my-testgogogo/auth.md)
- 示例：`examples/library/`（简单）、`examples/demo/`（Flow + 自定义 auth）
