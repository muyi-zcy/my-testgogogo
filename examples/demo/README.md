# demo 示例

**一体化示例**：包含 mock 后端接口 + 完整 API/Flow 测试。

## 目录结构

```
demo/
├── backend/           # mock HTTP 后端（案例接口实现）
├── configs/           # 测试配置（base_url 指向 backend）
├── internal/demoauth/ # 自定义认证 Provider
├── apistep/           # 可复用 API 封装
├── api/               # 单接口测试
├── flow/              # 流程测试（Vars 传递 + 条件分支，见 flow/example/）
└── scripts/run.sh     # 一键启动 backend 并跑测试
```

## 快速开始

### 方式 1：一键运行（推荐）

```bash
cd examples/demo
make run
```

自动启动 backend → 跑全部测试 → 停止 backend。

### 方式 2：手动分步

```bash
# 终端 1
make backend    # http://localhost:18080

# 终端 2
make test
```

## 后端接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/system/info` | 系统信息（无需登录） |
| POST | `/api/auth/v2/login` | 登录（demoauth 使用） |
| GET | `/api/auth/me` | 当前用户 |
| GET | `/api/auth/logout` | 登出 |
| GET | `/api/items` | 商品分页列表 |

默认账号：`demo` / `demo123`

## 配置

- `configs/config.yaml` — 主配置，`auth.provider: demoauth`
- `configs/local.yaml` — `base_url: http://localhost:18080`

## 自定义认证

见 `internal/demoauth/provider.go`，文档：[认证配置](../../docs/auth-config.md)
