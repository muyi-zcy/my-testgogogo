# library 示例

**一体化示例**：图书管理 mock 后端 + API/Flow 测试。

与 [demo](../demo/) 的区别：

| | demo | library |
|---|------|---------|
| 业务 | 商品商城 | 图书管理 |
| 端口 | 18080 | 18081 |
| 认证 | 自定义 `demoauth` Provider | 内置 `login` Provider（YAML 配置） |

## 目录结构

```
library/
├── backend/       # mock HTTP 后端
├── configs/       # 测试配置
├── apistep/       # API 封装
├── api/           # 单接口测试
├── flow/          # 流程测试（Vars 传递 + 条件分支，见 flow/example/）
└── scripts/run.sh # 一键启动 backend 并跑测试
```

## 快速开始

```bash
cd examples/library
make run
```

## 后端接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/system/info` | 系统信息 |
| POST | `/api/auth/login` | 登录，返回 `{token}` |
| GET | `/api/auth/me` | 当前用户 |
| GET | `/api/auth/logout` | 登出 |
| GET | `/api/books` | 图书分页列表 |

默认账号：`librarian` / `lib123`

## 认证

使用框架内置 `login` Provider，配置见 `configs/config.yaml`，无需编写 Provider 代码。

自定义认证请参考 [demo 示例](../demo/internal/demoauth/) 或 [认证文档](../../docs/auth-config.md)。
