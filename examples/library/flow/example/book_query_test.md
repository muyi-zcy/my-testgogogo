# 图书查询流程

> 用例：`TestFlowBookQueryChain` · 包路径：`examples/library/flow/example`

## 流程概述

演示 library 示例的 Flow 测试：多步骤串联、变量传递与条件分支。通过 `t.Run` 分别覆盖「有数据」与「空列表」两条互斥路径。

## 子用例

| 子用例 | 说明 |
|--------|------|
| `has_data` | 图书列表有数据时，按 ISBN 精确过滤 |
| `empty_list` | 图书列表为空时，验证空结果分页结构 |

---

## 子用例：has_data（有数据分支）

**流程步骤：** 系统信息 → 管理员信息 → 图书列表 → 按 ISBN 过滤

### 1. get system info

**操作：** 获取系统信息

**验证：** 系统名称非空，写入 `systemName` 变量

### 2. get librarian info

**操作：** 获取当前管理员（登录用户）信息

**验证：** 用户名非空；记录用户名、角色数、权限数

### 3. list books

**操作：** 分页查询图书列表（默认条件）

**验证：** 列表有数据时记录首本图书 ISBN/书名，分支标记为 `has_data`

### 4. query book by isbn from vars

**操作：** 用首本图书 ISBN 作为过滤条件再次查询

**验证：** 返回记录非空，且 ISBN 与 vars 中 `firstISBN` 一致

---

## 子用例：empty_list（空列表分支）

**流程步骤：** 系统信息 → 管理员信息 → 空列表 → 验证空结果分页

### 1–2. get system info / get librarian info

同有数据分支。

### 3. list books

**操作：** 用不存在 ISBN `__NONEXISTENT__` 查询图书列表

**验证：** 列表为空，分支标记为 `empty`

### 4. verify empty list pagination

**操作：** 再次用不存在 ISBN 查询

**验证：** `records` 为空，`total` = 0

## 备注

- 管理员具备角色权限时流程完整执行
- 参考实现：`book_query_test.go`
