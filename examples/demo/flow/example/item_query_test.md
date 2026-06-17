# 商品查询流程

> 用例：`TestFlowItemQueryChain` · 包路径：`examples/demo/flow/example`

## 流程概述

演示 Flow 测试的多步骤串联、变量传递与条件分支。通过 `t.Run` 分别覆盖「有数据」与「空列表」两条互斥路径。

## 子用例

| 子用例 | 说明 |
|--------|------|
| `has_data` | 商品列表有数据时，按编码精确过滤 |
| `empty_list` | 商品列表为空时，验证空结果分页结构 |

---

## 子用例：has_data（有数据分支）

**流程步骤：** 系统信息 → 用户信息 → 商品列表 → 按 vars 编码过滤

### 1. get system info

**操作：** 调用 `GET /api/system/info`

**验证：** 返回系统名称非空，写入 `systemName` 变量

### 2. get user info

**操作：** 调用 `GET /api/auth/me`

**验证：** 用户名非空；记录用户名、角色数、权限数

### 3. list items

**操作：** 分页查询商品列表（默认条件）

**验证：** 列表有数据时记录首条商品编码/名称，分支标记为 `has_data`

### 4. query item by code from vars

**操作：** 用首条商品编码作为过滤条件再次查询

**验证：** 返回记录非空，且编码与 vars 中 `firstItemCode` 一致

---

## 子用例：empty_list（空列表分支）

**流程步骤：** 系统信息 → 用户信息 → 空列表 → 验证空结果分页

### 1–2. get system info / get user info

同有数据分支。

### 3. list items

**操作：** 用不存在编码 `__NONEXISTENT__` 查询商品列表

**验证：** 列表为空，分支标记为 `empty`

### 4. verify empty list pagination

**操作：** 再次用不存在编码查询

**验证：** `records` 为空，`total` = 0

## 备注

- 用户有角色时流程完整执行；无角色时以有限断言结束
- 参考实现：`item_query_test.go`
