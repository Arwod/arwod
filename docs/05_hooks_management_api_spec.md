# PocketBase 钩子脚本管理系统 API 规范

## 1. 概述

本文档定义了 PocketBase 钩子脚本管理系统的 API 规范，包括所有端点、请求/响应格式、错误处理和安全要求。这些 API 将用于在 PocketBase 管理界面中创建、读取、更新、删除和管理钩子脚本。

### 1.1 基本信息

- **基础路径**: `/api/hooks`
- **认证**: 所有 API 都需要管理员认证
- **响应格式**: JSON
- **版本控制**: 通过 URL 路径（如 `/api/v1/hooks`）

### 1.2 通用状态码

| 状态码 | 描述 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 400 | 请求错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 422 | 验证错误 |
| 500 | 服务器错误 |

### 1.3 通用错误响应格式

```json
{
  "code": 400,
  "message": "错误描述",
  "data": {
    "field1": ["错误原因1", "错误原因2"],
    "field2": ["错误原因"]
  }
}
```

## 2. 钩子脚本 API

### 2.1 获取钩子脚本列表

获取所有钩子脚本的列表，支持分页、排序和筛选。

**请求**:

```
GET /api/hooks
```

**查询参数**:

| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| page | 整数 | 页码 | 1 |
| perPage | 整数 | 每页数量 | 30 |
| sort | 字符串 | 排序字段和方向，如 "-created,name" | "order" |
| filter | 字符串 | 筛选条件，如 "type='record' && event='create'" | 无 |
| expand | 字符串 | 扩展关联字段 | 无 |

**响应** (200):

```json
{
  "page": 1,
  "perPage": 30,
  "totalItems": 50,
  "totalPages": 2,
  "items": [
    {
      "id": "abc123",
      "name": "用户创建后发送欢迎邮件",
      "description": "当新用户注册后自动发送欢迎邮件",
      "type": "record",
      "event": "create",
      "collection": "users",
      "code": "// 脚本代码...",
      "order": 1,
      "enabled": true,
      "created": "2023-01-01T12:00:00.000Z",
      "updated": "2023-01-02T14:30:00.000Z"
    },
    // 更多脚本...
  ]
}
```

### 2.2 获取单个钩子脚本

获取指定 ID 的钩子脚本详细信息。

**请求**:

```
GET /api/hooks/{id}
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**响应** (200):

```json
{
  "id": "abc123",
  "name": "用户创建后发送欢迎邮件",
  "description": "当新用户注册后自动发送欢迎邮件",
  "type": "record",
  "event": "create",
  "collection": "users",
  "code": "// 脚本代码...",
  "order": 1,
  "enabled": true,
  "created": "2023-01-01T12:00:00.000Z",
  "updated": "2023-01-02T14:30:00.000Z"
}
```

### 2.3 创建钩子脚本

创建新的钩子脚本。

**请求**:

```
POST /api/hooks
```

**请求体**:

```json
{
  "name": "用户创建后发送欢迎邮件",
  "description": "当新用户注册后自动发送欢迎邮件",
  "type": "record",
  "event": "create",
  "collection": "users",
  "code": "// 脚本代码...",
  "order": 1,
  "enabled": true
}
```

**字段说明**:

| 字段 | 类型 | 必填 | 描述 |
|------|------|------|------|
| name | 字符串 | 是 | 脚本名称，最大长度 100 |
| description | 字符串 | 否 | 脚本描述，最大长度 500 |
| type | 字符串 | 是 | 脚本类型，可选值：record, collection, api, system |
| event | 字符串 | 是 | 事件类型，根据 type 不同有不同的可选值 |
| collection | 字符串 | 条件 | 当 type 为 record 或 collection 时必填，指定关联的集合名称 |
| code | 字符串 | 是 | 脚本代码 |
| order | 整数 | 否 | 执行顺序，默认为 0 |
| enabled | 布尔 | 否 | 是否启用，默认为 true |

**响应** (201):

```json
{
  "id": "abc123",
  "name": "用户创建后发送欢迎邮件",
  "description": "当新用户注册后自动发送欢迎邮件",
  "type": "record",
  "event": "create",
  "collection": "users",
  "code": "// 脚本代码...",
  "order": 1,
  "enabled": true,
  "created": "2023-01-01T12:00:00.000Z",
  "updated": "2023-01-01T12:00:00.000Z"
}
```

### 2.4 更新钩子脚本

更新指定 ID 的钩子脚本。

**请求**:

```
PATCH /api/hooks/{id}
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**请求体**:

```json
{
  "name": "更新后的名称",
  "description": "更新后的描述",
  "code": "// 更新后的代码...",
  "enabled": false
}
```

**注意**: 请求体中只需包含需要更新的字段。

**响应** (200):

```json
{
  "id": "abc123",
  "name": "更新后的名称",
  "description": "更新后的描述",
  "type": "record",
  "event": "create",
  "collection": "users",
  "code": "// 更新后的代码...",
  "order": 1,
  "enabled": false,
  "created": "2023-01-01T12:00:00.000Z",
  "updated": "2023-01-03T15:45:00.000Z"
}
```

### 2.5 删除钩子脚本

删除指定 ID 的钩子脚本。

**请求**:

```
DELETE /api/hooks/{id}
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**响应** (204): 无内容

### 2.6 启用/禁用钩子脚本

快速切换钩子脚本的启用状态。

**请求**:

```
PATCH /api/hooks/{id}/toggle
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**响应** (200):

```json
{
  "id": "abc123",
  "enabled": true,
  "updated": "2023-01-03T16:20:00.000Z"
}
```

### 2.7 更新钩子脚本执行顺序

批量更新多个钩子脚本的执行顺序。

**请求**:

```
POST /api/hooks/reorder
```

**请求体**:

```json
{
  "orders": [
    {"id": "abc123", "order": 1},
    {"id": "def456", "order": 2},
    {"id": "ghi789", "order": 3}
  ]
}
```

**响应** (200):

```json
{
  "success": true
}
```

## 3. 钩子脚本测试 API

### 3.1 测试钩子脚本

测试钩子脚本的执行效果，不会影响实际数据。

**请求**:

```
POST /api/hooks/{id}/test
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**请求体**:

```json
{
  "testData": {
    // 测试数据，根据脚本类型和事件不同而不同
    "record": {
      "id": "test123",
      "name": "测试用户",
      "email": "test@example.com"
    },
    "collection": "users"
  }
}
```

**响应** (200):

```json
{
  "success": true,
  "executionTime": 45,  // 毫秒
  "logs": [
    {"type": "log", "message": "开始执行...", "timestamp": "2023-01-03T16:30:00.123Z"},
    {"type": "info", "message": "发送邮件到 test@example.com", "timestamp": "2023-01-03T16:30:00.145Z"},
    {"type": "error", "message": "邮件发送失败: 无效地址", "timestamp": "2023-01-03T16:30:00.156Z"}
  ],
  "result": {
    // 执行结果，根据脚本类型和事件不同而不同
    "record": {
      "id": "test123",
      "name": "测试用户",
      "email": "test@example.com",
      "welcomeEmailSent": false
    }
  },
  "error": "邮件发送失败: 无效地址"  // 如果执行出错
}
```

## 4. 钩子脚本版本 API

### 4.1 获取钩子脚本版本历史

获取指定钩子脚本的版本历史记录。

**请求**:

```
GET /api/hooks/{id}/versions
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**查询参数**:

| 参数 | 类型 | 描述 | 默认值 |
|------|------|------|--------|
| page | 整数 | 页码 | 1 |
| perPage | 整数 | 每页数量 | 30 |

**响应** (200):

```json
{
  "page": 1,
  "perPage": 30,
  "totalItems": 5,
  "totalPages": 1,
  "items": [
    {
      "id": "ver123",
      "hookId": "abc123",
      "version": 3,
      "code": "// 版本 3 的代码...",
      "comment": "修复邮件发送问题",
      "created": "2023-01-03T15:45:00.000Z",
      "createdBy": "admin123"
    },
    {
      "id": "ver456",
      "hookId": "abc123",
      "version": 2,
      "code": "// 版本 2 的代码...",
      "comment": "添加错误处理",
      "created": "2023-01-02T14:30:00.000Z",
      "createdBy": "admin123"
    },
    {
      "id": "ver789",
      "hookId": "abc123",
      "version": 1,
      "code": "// 版本 1 的代码...",
      "comment": "初始版本",
      "created": "2023-01-01T12:00:00.000Z",
      "createdBy": "admin123"
    }
  ]
}
```

### 4.2 获取特定版本的钩子脚本

获取钩子脚本的特定版本详情。

**请求**:

```
GET /api/hooks/{id}/versions/{versionId}
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |
| versionId | 字符串 | 版本 ID |

**响应** (200):

```json
{
  "id": "ver123",
  "hookId": "abc123",
  "version": 3,
  "code": "// 版本 3 的代码...",
  "comment": "修复邮件发送问题",
  "created": "2023-01-03T15:45:00.000Z",
  "createdBy": "admin123"
}
```

### 4.3 创建新版本

为钩子脚本创建新版本。

**请求**:

```
POST /api/hooks/{id}/versions
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |

**请求体**:

```json
{
  "code": "// 新版本的代码...",
  "comment": "版本说明"
}
```

**响应** (201):

```json
{
  "id": "ver123",
  "hookId": "abc123",
  "version": 4,
  "code": "// 新版本的代码...",
  "comment": "版本说明",
  "created": "2023-01-04T10:15:00.000Z",
  "createdBy": "admin123"
}
```

### 4.4 恢复到特定版本

将钩子脚本恢复到特定版本。

**请求**:

```
POST /api/hooks/{id}/versions/{versionId}/restore
```

**路径参数**:

| 参数 | 类型 | 描述 |
|------|------|------|
| id | 字符串 | 钩子脚本 ID |
| versionId | 字符串 | 版本 ID |

**请求体**:

```json
{
  "comment": "恢复到版本 2"
}
```

**响应** (200):

```json
{
  "id": "abc123",
  "name": "用户创建后发送欢迎邮件",
  "description": "当新用户注册后自动发送欢迎邮件",
  "type": "record",
  "event": "create",
  "collection": "users",
  "code": "// 恢复的代码...",
  "order": 1,
  "enabled": true,
  "created": "2023-01-01T12:00:00.000Z",
  "updated": "2023-01-04T11:30:00.000Z",
  "currentVersion": {
    "id": "ver999",
    "version": 5,
    "comment": "恢复到版本 2",
    "created": "2023-01-04T11:30:00.000Z",
    "createdBy": "admin123"
  }
}
```

## 5. 钩子脚本导入/导出 API

### 5.1 导出钩子脚本

导出一个或多个钩子脚本。

**请求**:

```
POST /api/hooks/export
```

**请求体**:

```json
{
  "ids": ["abc123", "def456"],  // 可选，不提供则导出所有脚本
  "format": "json"  // 可选，默认为 json，可选值：json, js
}
```

**响应** (200):

如果 format 为 json:

```json
{
  "hooks": [
    {
      "name": "用户创建后发送欢迎邮件",
      "description": "当新用户注册后自动发送欢迎邮件",
      "type": "record",
      "event": "create",
      "collection": "users",
      "code": "// 脚本代码...",
      "order": 1,
      "enabled": true
    },
    // 更多脚本...
  ]
}
```

如果 format 为 js，则返回 JavaScript 文件内容（Content-Type: application/javascript）。

### 5.2 导入钩子脚本

导入一个或多个钩子脚本。

**请求**:

```
POST /api/hooks/import
```

**请求体**:

如果是 JSON 格式:

```json
{
  "hooks": [
    {
      "name": "导入的脚本 1",
      "description": "从其他系统导入的脚本",
      "type": "record",
      "event": "create",
      "collection": "users",
      "code": "// 脚本代码...",
      "order": 1,
      "enabled": true
    },
    // 更多脚本...
  ],
  "overwrite": false  // 可选，是否覆盖同名脚本，默认为 false
}
```

如果是 JavaScript 文件，则使用 multipart/form-data 格式上传文件。

**响应** (200):

```json
{
  "imported": 2,  // 成功导入的脚本数量
  "skipped": 1,  // 跳过的脚本数量（如果 overwrite 为 false 且存在同名脚本）
  "failed": 0,   // 导入失败的脚本数量
  "hooks": [     // 导入的脚本列表
    {
      "id": "imp123",
      "name": "导入的脚本 1",
      "status": "success"
    },
    {
      "id": "imp456",
      "name": "导入的脚本 2",
      "status": "success"
    },
    {
      "name": "已存在的脚本",
      "status": "skipped",
      "reason": "同名脚本已存在"
    }
  ]
}
```

## 6. 钩子脚本迁移 API

### 6.1 从文件系统迁移脚本

将文件系统中的钩子脚本迁移到数据库。

**请求**:

```
POST /api/hooks/migrate-from-fs
```

**请求体**:

```json
{
  "hooksDir": "path/to/pb_hooks",  // 可选，默认使用配置中的 HooksDir
  "deleteOriginal": false,        // 可选，是否删除原始文件，默认为 false
  "overwrite": false             // 可选，是否覆盖同名脚本，默认为 false
}
```

**响应** (200):

```json
{
  "migrated": 5,  // 成功迁移的脚本数量
  "skipped": 1,   // 跳过的脚本数量
  "failed": 0,    // 迁移失败的脚本数量
  "hooks": [      // 迁移的脚本列表
    {
      "id": "mig123",
      "name": "users_create.pb.js",
      "status": "success",
      "originalPath": "pb_hooks/users_create.pb.js"
    },
    // 更多脚本...
  ]
}
```

## 7. 钩子脚本元数据 API

### 7.1 获取钩子类型和事件

获取所有可用的钩子类型和对应的事件。

**请求**:

```
GET /api/hooks/meta/types
```

**响应** (200):

```json
{
  "types": [
    {
      "name": "record",
      "label": "记录钩子",
      "events": [
        {"name": "create", "label": "创建前"},
        {"name": "create.after", "label": "创建后"},
        {"name": "update", "label": "更新前"},
        {"name": "update.after", "label": "更新后"},
        {"name": "delete", "label": "删除前"},
        {"name": "delete.after", "label": "删除后"}
      ]
    },
    {
      "name": "collection",
      "label": "集合钩子",
      "events": [
        {"name": "create", "label": "创建前"},
        {"name": "create.after", "label": "创建后"},
        {"name": "update", "label": "更新前"},
        {"name": "update.after", "label": "更新后"},
        {"name": "delete", "label": "删除前"},
        {"name": "delete.after", "label": "删除后"}
      ]
    },
    {
      "name": "api",
      "label": "API 钩子",
      "events": [
        {"name": "request", "label": "请求前"},
        {"name": "response", "label": "响应前"}
      ]
    },
    {
      "name": "system",
      "label": "系统钩子",
      "events": [
        {"name": "bootstrap", "label": "启动"},
        {"name": "serve", "label": "服务启动"},
        {"name": "terminate", "label": "终止"}
      ]
    }
  ]
}
```

### 7.2 获取集合列表

获取所有可用的集合，用于创建记录和集合钩子。

**请求**:

```
GET /api/hooks/meta/collections
```

**响应** (200):

```json
{
  "collections": [
    {"name": "users", "label": "用户"},
    {"name": "posts", "label": "文章"},
    {"name": "comments", "label": "评论"},
    // 更多集合...
  ]
}
```

## 8. 安全考虑

### 8.1 认证和授权

所有 API 端点都需要管理员认证。可以通过以下方式之一进行认证：

1. **Cookie 认证**：通过管理员登录页面获取的会话 Cookie
2. **Token 认证**：在请求头中包含 `Authorization: Admin TOKEN` 格式的令牌

### 8.2 输入验证

所有 API 输入都应进行严格验证，特别是：

1. 脚本代码应检查潜在的恶意代码
2. 所有字符串输入应验证长度和格式
3. 所有 ID 应验证格式和存在性

### 8.3 速率限制

为防止滥用，API 应实施速率限制：

1. 每个管理员每分钟最多 60 个请求
2. 测试 API 每个管理员每分钟最多 10 个请求

### 8.4 审计日志

所有关键操作都应记录审计日志，包括：

1. 脚本创建和修改
2. 脚本启用和禁用
3. 脚本删除
4. 版本创建和恢复

## 9. 错误代码和消息

| 错误代码 | 消息 | 描述 |
|---------|------|------|
| 400001 | 无效的请求参数 | 请求参数格式错误或缺失 |
| 400002 | 无效的脚本代码 | 脚本代码包含语法错误 |
| 400003 | 无效的脚本类型或事件 | 指定的脚本类型或事件不存在 |
| 400004 | 无效的集合 | 指定的集合不存在 |
| 400005 | 无效的测试数据 | 测试数据格式错误或不匹配脚本类型 |
| 403001 | 权限不足 | 当前用户没有执行该操作的权限 |
| 404001 | 脚本不存在 | 指定 ID 的脚本不存在 |
| 404002 | 版本不存在 | 指定 ID 的版本不存在 |
| 409001 | 脚本名称已存在 | 创建的脚本名称已被使用 |
| 500001 | 脚本执行错误 | 脚本执行过程中发生错误 |
| 500002 | 数据库错误 | 数据库操作失败 |

## 10. 版本控制和兼容性

### 10.1 API 版本控制

API 版本通过 URL 路径进行控制，如 `/api/v1/hooks`。当 API 发生不兼容的变更时，将增加版本号。

### 10.2 向后兼容性

1. 新增字段不会破坏现有客户端
2. 不会删除现有字段，而是将其标记为已弃用
3. 弃用的字段将在下一个主要版本中移除

## 11. 性能考虑

### 11.1 分页和限制

1. 所有列表 API 都支持分页
2. 默认每页返回 30 条记录
3. 最大每页记录数为 100

### 11.2 缓存

1. 获取单个脚本的 API 响应可以缓存 5 分钟
2. 元数据 API 响应可以缓存 1 小时
3. 使用 ETag 和条件请求减少带宽使用

## 12. 结论

本 API 规范定义了 PocketBase 钩子脚本管理系统的所有接口，包括钩子脚本的 CRUD 操作、测试、版本控制、导入/导出和迁移功能。通过这些 API，前端界面可以与后端系统进行交互，实现完整的钩子脚本管理功能。

在实现过程中，应特别注意安全性、性能和向后兼容性，确保系统能够安全、高效地运行，并且在未来的版本更新中不会破坏现有的集成。