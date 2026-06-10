# my-testgogogo 接口自动化测试报告

| 项目 | 值 |
|------|----|
| 生成时间 | 2026-06-10 22:58:26 |
| 报告编号 |  |
| Go 版本 | go1.22.12 |
| 运行环境 | local |
| 接口地址 | http://localhost:18080 |
| 测试账号 | demo |

## 总览

**结果：通过**

## 详细报告

### 商品查询流程验证（有数据分支）

- 用例：`TestFlowItemQueryChain/has_data`
- 分类：功能测试
- 包路径：`examples/demo/flow/example`
- 结果：**PASS**
- 耗时：4ms
- 说明：系统信息 → 用户信息 → 商品列表 → 按 vars 编码过滤

#### 步骤

| # | 步骤 | 结果 | 开始时间 | 结束时间 | 耗时 |
|---|------|------|----------|----------|------|
| 1 | get system info | PASS | 2026-06-10 22:58:26.746 | 2026-06-10 22:58:26.748 | 2ms |
| 2 | get user info | PASS | 2026-06-10 22:58:26.749 | 2026-06-10 22:58:26.749 | 0ms |
| 3 | list items | PASS | 2026-06-10 22:58:26.749 | 2026-06-10 22:58:26.750 | 0ms |
| 4 | query item by code from vars | PASS | 2026-06-10 22:58:26.750 | 2026-06-10 22:58:26.750 | 0ms |

##### 1. get system info

| 项目 | 值 |
|------|----|
| 结果 | PASS |
| 开始时间 | 2026-06-10 22:58:26.746 |
| 结束时间 | 2026-06-10 22:58:26.748 |
| 耗时 | 2ms |
| 耗时(ms) | 2 |

**结构化结果：**

| 字段 | 值 |
|------|----|
| system | map[name:my-testgogogo-demo version:1.0.0] |

```json
{
  "system": {
    "name": "my-testgogogo-demo",
    "version": "1.0.0"
  }
}
```


##### 2. get user info

| 项目 | 值 |
|------|----|
| 结果 | PASS |
| 开始时间 | 2026-06-10 22:58:26.749 |
| 结束时间 | 2026-06-10 22:58:26.749 |
| 耗时 | 0ms |

**结构化结果：**

| 字段 | 值 |
|------|----|
| nickName | Demo User |
| permissions | 2 |
| roleCount | 1 |
| username | demo |

```json
{
  "nickName": "Demo User",
  "permissions": 2,
  "roleCount": 1,
  "username": "demo"
}
```


##### 3. list items

| 项目 | 值 |
|------|----|
| 结果 | PASS |
| 开始时间 | 2026-06-10 22:58:26.749 |
| 结束时间 | 2026-06-10 22:58:26.750 |
| 耗时 | 0ms |

**结构化结果：**

| 字段 | 值 |
|------|----|
| current | 1 |
| firstItem.code | SKU-001 |
| firstItem.name | Demo Item Alpha |
| size | 10 |
| total | 3 |

```json
{
  "current": 1,
  "firstItem": {
    "code": "SKU-001",
    "name": "Demo Item Alpha"
  },
  "size": 10,
  "total": 3
}
```


##### 4. query item by code from vars

| 项目 | 值 |
|------|----|
| 结果 | PASS |
| 开始时间 | 2026-06-10 22:58:26.750 |
| 结束时间 | 2026-06-10 22:58:26.750 |
| 耗时 | 0ms |

**结构化结果：**

| 字段 | 值 |
|------|----|
| matched | {1 SKU-001 Demo Item Alpha} |
| queryCode | SKU-001 |
| total | 1 |

```json
{
  "matched": {
    "id": "1",
    "code": "SKU-001",
    "name": "Demo Item Alpha"
  },
  "queryCode": "SKU-001",
  "total": 1
}
```


#### 备注

- 走有数据分支：按编码精确过滤
- 用户具备角色权限，流程完整执行

