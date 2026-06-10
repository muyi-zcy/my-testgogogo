# Examples

每个 **一体化示例** 包含：

```
{example}/
├── backend/      # 可运行的 mock 后端
├── configs/      # 测试配置
├── apistep/      # API 封装
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

## 新建示例

1. 复制 `demo/` 或 `library/` 目录
2. 修改 `backend/` 实现案例接口
3. 更新 `configs/local.yaml` 中的 `base_url` 与端口
4. 编写 `apistep/`、`api/`、`flow/` 测试
5. 选择内置 `login` 或自定义 `auth.Provider`
