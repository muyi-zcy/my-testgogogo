# Examples

每个 **一体化示例** 包含：

```
{example}/
├── backend/      # 可运行的 mock 后端（wms 无 mock，对接真实后端）
├── configs/      # 测试配置（含 load 段）
├── apistep/      # HTTP 封装 + DTO
├── scenario/     # 业务编排（api / flow / 压测共用）
├── cmd/load/     # 压测入口（可选）
├── api/          # 单接口测试
├── flow/         # 流程测试（可选）
├── scripts/      # 辅助脚本
├── Makefile
└── README.md
```

## 可用示例

| 示例 | 说明 | 认证方式 | 快速开始 |
|------|------|----------|----------|
| [demo](demo/) | 商品商城 mock 后端 + 测试 | 自定义 `demoauth` Provider | `cd demo && make run` |
| [library](library/) | 图书管理 mock 后端 + 测试 | 内置 `login` Provider（YAML） | `cd library && make run` |
| [wms](wms/) | 真实 WMS 后端 + 压测 | 内置 `login` Provider | `cd wms && make test` |

## 新建示例

1. 复制 `demo/` 或 `library/` 目录
2. 修改 `backend/` 实现案例接口
3. 更新 `configs/local.yaml` 中的 `base_url` 与端口
4. 编写 `apistep/`、`scenario/`、`api/`、`flow/` 测试
5. 选择内置 `login` 或自定义 `auth.Provider`
6. 需要压测时添加 `scenario/register.go` 与 `cmd/load/`
