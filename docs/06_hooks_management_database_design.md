# PocketBase 钩子脚本管理系统数据库设计

## 1. 概述

本文档详细说明了 PocketBase 钩子脚本管理系统的数据库设计，包括表结构、字段定义、索引、关系和约束。该数据库设计旨在支持钩子脚本的存储、版本控制、状态管理和执行顺序控制，以及与 PocketBase 现有系统的集成。

## 2. 数据库引擎

PocketBase 使用 SQLite 作为其内置数据库引擎，因此钩子脚本管理系统的数据库设计也基于 SQLite。SQLite 是一个轻量级的、嵌入式的关系型数据库管理系统，具有以下特点：

- 零配置 - 无需安装和管理
- 单文件数据库 - 整个数据库存储在一个文件中
- 跨平台 - 可在各种操作系统上运行
- 自包含 - 无外部依赖
- 小型 - 库大小小于 600KB
- 支持 ACID 事务
- 支持 SQL 标准的大部分功能

## 3. 表结构设计

### 3.1 钩子脚本表 (pb_hooks_scripts)

该表存储所有钩子脚本的基本信息和内容。

#### 表结构

```sql
CREATE TABLE "pb_hooks_scripts" (
    "id" TEXT PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "type" TEXT NOT NULL,
    "event" TEXT NOT NULL,
    "collection" TEXT,
    "code" TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,
    "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "created" TEXT NOT NULL,
    "updated" TEXT NOT NULL
);
```

#### 字段说明

| 字段名 | 数据类型 | 是否必填 | 默认值 | 描述 |
|--------|----------|----------|--------|------|
| id | TEXT | 是 | - | 主键，唯一标识符，使用 UUID 或 nanoid |
| name | TEXT | 是 | - | 脚本名称，最大长度 100 字符 |
| description | TEXT | 否 | NULL | 脚本描述，最大长度 500 字符 |
| type | TEXT | 是 | - | 脚本类型，可选值：record, collection, api, system |
| event | TEXT | 是 | - | 事件类型，根据 type 不同有不同的可选值 |
| collection | TEXT | 条件 | NULL | 当 type 为 record 或 collection 时必填，指定关联的集合名称 |
| code | TEXT | 是 | - | 脚本代码内容 |
| order | INTEGER | 是 | 0 | 执行顺序，数字越小越先执行 |
| enabled | BOOLEAN | 是 | TRUE | 是否启用脚本 |
| created | TEXT | 是 | - | 创建时间，ISO 8601 格式 |
| updated | TEXT | 是 | - | 最后更新时间，ISO 8601 格式 |

#### 索引

```sql
-- 名称索引（用于快速查找和确保名称唯一性）
CREATE UNIQUE INDEX "idx_hooks_scripts_name" ON "pb_hooks_scripts" ("name");

-- 类型和事件索引（用于按类型和事件筛选脚本）
CREATE INDEX "idx_hooks_scripts_type_event" ON "pb_hooks_scripts" ("type", "event");

-- 集合索引（用于按集合筛选脚本）
CREATE INDEX "idx_hooks_scripts_collection" ON "pb_hooks_scripts" ("collection");

-- 启用状态索引（用于快速筛选启用/禁用的脚本）
CREATE INDEX "idx_hooks_scripts_enabled" ON "pb_hooks_scripts" ("enabled");

-- 执行顺序索引（用于按执行顺序排序脚本）
CREATE INDEX "idx_hooks_scripts_order" ON "pb_hooks_scripts" ("order");
```

### 3.2 钩子脚本版本表 (pb_hooks_versions)

该表存储钩子脚本的历史版本，用于版本控制和回滚。

#### 表结构

```sql
CREATE TABLE "pb_hooks_versions" (
    "id" TEXT PRIMARY KEY,
    "hook_id" TEXT NOT NULL,
    "version" INTEGER NOT NULL,
    "code" TEXT NOT NULL,
    "comment" TEXT,
    "created" TEXT NOT NULL,
    "created_by" TEXT,
    FOREIGN KEY ("hook_id") REFERENCES "pb_hooks_scripts" ("id") ON DELETE CASCADE
);
```

#### 字段说明

| 字段名 | 数据类型 | 是否必填 | 默认值 | 描述 |
|--------|----------|----------|--------|------|
| id | TEXT | 是 | - | 主键，唯一标识符，使用 UUID 或 nanoid |
| hook_id | TEXT | 是 | - | 关联的钩子脚本 ID，外键引用 pb_hooks_scripts.id |
| version | INTEGER | 是 | - | 版本号，从 1 开始递增 |
| code | TEXT | 是 | - | 该版本的脚本代码内容 |
| comment | TEXT | 否 | NULL | 版本说明，最大长度 200 字符 |
| created | TEXT | 是 | - | 版本创建时间，ISO 8601 格式 |
| created_by | TEXT | 否 | NULL | 创建者 ID，关联到管理员用户 |

#### 索引

```sql
-- 钩子 ID 和版本号索引（用于快速查找特定钩子的特定版本）
CREATE UNIQUE INDEX "idx_hooks_versions_hook_version" ON "pb_hooks_versions" ("hook_id", "version");

-- 钩子 ID 索引（用于查找特定钩子的所有版本）
CREATE INDEX "idx_hooks_versions_hook_id" ON "pb_hooks_versions" ("hook_id");

-- 创建时间索引（用于按时间排序版本）
CREATE INDEX "idx_hooks_versions_created" ON "pb_hooks_versions" ("created");
```

### 3.3 钩子脚本执行日志表 (pb_hooks_logs) [可选]

该表存储钩子脚本的执行日志，用于调试和监控。

#### 表结构

```sql
CREATE TABLE "pb_hooks_logs" (
    "id" TEXT PRIMARY KEY,
    "hook_id" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "execution_time" INTEGER NOT NULL,
    "error" TEXT,
    "context" TEXT,
    "created" TEXT NOT NULL,
    FOREIGN KEY ("hook_id") REFERENCES "pb_hooks_scripts" ("id") ON DELETE CASCADE
);
```

#### 字段说明

| 字段名 | 数据类型 | 是否必填 | 默认值 | 描述 |
|--------|----------|----------|--------|------|
| id | TEXT | 是 | - | 主键，唯一标识符，使用 UUID 或 nanoid |
| hook_id | TEXT | 是 | - | 关联的钩子脚本 ID，外键引用 pb_hooks_scripts.id |
| status | TEXT | 是 | - | 执行状态，可选值：success, error |
| execution_time | INTEGER | 是 | - | 执行时间，单位为毫秒 |
| error | TEXT | 否 | NULL | 错误信息，如果执行失败 |
| context | TEXT | 否 | NULL | 执行上下文，JSON 格式 |
| created | TEXT | 是 | - | 日志创建时间，ISO 8601 格式 |

#### 索引

```sql
-- 钩子 ID 索引（用于查找特定钩子的所有日志）
CREATE INDEX "idx_hooks_logs_hook_id" ON "pb_hooks_logs" ("hook_id");

-- 状态索引（用于按状态筛选日志）
CREATE INDEX "idx_hooks_logs_status" ON "pb_hooks_logs" ("status");

-- 创建时间索引（用于按时间排序日志）
CREATE INDEX "idx_hooks_logs_created" ON "pb_hooks_logs" ("created");
```

## 4. 数据关系

### 4.1 钩子脚本与版本的关系

- 一个钩子脚本可以有多个版本（一对多关系）
- 通过 `pb_hooks_versions.hook_id` 外键引用 `pb_hooks_scripts.id`
- 当删除钩子脚本时，级联删除所有相关版本

### 4.2 钩子脚本与执行日志的关系 [可选]

- 一个钩子脚本可以有多个执行日志（一对多关系）
- 通过 `pb_hooks_logs.hook_id` 外键引用 `pb_hooks_scripts.id`
- 当删除钩子脚本时，级联删除所有相关日志

## 5. 数据约束

### 5.1 唯一性约束

- `pb_hooks_scripts.name` 必须唯一，确保脚本名称不重复
- `pb_hooks_versions.hook_id` 和 `pb_hooks_versions.version` 的组合必须唯一，确保每个钩子脚本的版本号不重复

### 5.2 引用完整性约束

- `pb_hooks_versions.hook_id` 必须引用有效的 `pb_hooks_scripts.id`
- `pb_hooks_logs.hook_id` 必须引用有效的 `pb_hooks_scripts.id`

### 5.3 字段约束

- `pb_hooks_scripts.type` 必须是以下值之一：record, collection, api, system
- `pb_hooks_scripts.event` 必须是有效的事件类型，根据 type 不同而不同
- 当 `pb_hooks_scripts.type` 为 record 或 collection 时，`pb_hooks_scripts.collection` 不能为空
- `pb_hooks_scripts.order` 必须是非负整数

## 6. 数据迁移

### 6.1 初始迁移脚本

以下是创建钩子脚本管理系统所需表的 SQL 迁移脚本：

```sql
-- 创建钩子脚本表
CREATE TABLE "pb_hooks_scripts" (
    "id" TEXT PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT,
    "type" TEXT NOT NULL,
    "event" TEXT NOT NULL,
    "collection" TEXT,
    "code" TEXT NOT NULL,
    "order" INTEGER NOT NULL DEFAULT 0,
    "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
    "created" TEXT NOT NULL,
    "updated" TEXT NOT NULL
);

-- 创建钩子脚本表索引
CREATE UNIQUE INDEX "idx_hooks_scripts_name" ON "pb_hooks_scripts" ("name");
CREATE INDEX "idx_hooks_scripts_type_event" ON "pb_hooks_scripts" ("type", "event");
CREATE INDEX "idx_hooks_scripts_collection" ON "pb_hooks_scripts" ("collection");
CREATE INDEX "idx_hooks_scripts_enabled" ON "pb_hooks_scripts" ("enabled");
CREATE INDEX "idx_hooks_scripts_order" ON "pb_hooks_scripts" ("order");

-- 创建钩子脚本版本表
CREATE TABLE "pb_hooks_versions" (
    "id" TEXT PRIMARY KEY,
    "hook_id" TEXT NOT NULL,
    "version" INTEGER NOT NULL,
    "code" TEXT NOT NULL,
    "comment" TEXT,
    "created" TEXT NOT NULL,
    "created_by" TEXT,
    FOREIGN KEY ("hook_id") REFERENCES "pb_hooks_scripts" ("id") ON DELETE CASCADE
);

-- 创建钩子脚本版本表索引
CREATE UNIQUE INDEX "idx_hooks_versions_hook_version" ON "pb_hooks_versions" ("hook_id", "version");
CREATE INDEX "idx_hooks_versions_hook_id" ON "pb_hooks_versions" ("hook_id");
CREATE INDEX "idx_hooks_versions_created" ON "pb_hooks_versions" ("created");

-- 创建钩子脚本执行日志表 [可选]
CREATE TABLE "pb_hooks_logs" (
    "id" TEXT PRIMARY KEY,
    "hook_id" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "execution_time" INTEGER NOT NULL,
    "error" TEXT,
    "context" TEXT,
    "created" TEXT NOT NULL,
    FOREIGN KEY ("hook_id") REFERENCES "pb_hooks_scripts" ("id") ON DELETE CASCADE
);

-- 创建钩子脚本执行日志表索引 [可选]
CREATE INDEX "idx_hooks_logs_hook_id" ON "pb_hooks_logs" ("hook_id");
CREATE INDEX "idx_hooks_logs_status" ON "pb_hooks_logs" ("status");
CREATE INDEX "idx_hooks_logs_created" ON "pb_hooks_logs" ("created");
```

### 6.2 从文件系统迁移数据

以下是将现有文件系统中的钩子脚本迁移到数据库的 Go 代码示例：

```go
func MigrateHooksFromFileSystem(app core.App, hooksDir string) error {
    // 获取数据库连接
    db := app.DB()
    
    // 获取当前时间
    now := time.Now().UTC().Format(time.RFC3339)
    
    // 遍历钩子目录中的所有文件
    files, err := os.ReadDir(hooksDir)
    if err != nil {
        return err
    }
    
    // 开始事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    for _, file := range files {
        // 跳过目录和非 JS/TS 文件
        if file.IsDir() || !regexp.MustCompile(`^.*(\.pb\.js|\.pb\.ts)$`).MatchString(file.Name()) {
            continue
        }
        
        // 读取文件内容
        filePath := filepath.Join(hooksDir, file.Name())
        code, err := os.ReadFile(filePath)
        if err != nil {
            return err
        }
        
        // 解析文件名以确定类型、事件和集合
        // 假设文件名格式为：[collection]_[event].pb.js
        // 例如：users_create.pb.js, api_request.pb.js, bootstrap.pb.js
        parts := strings.Split(strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())), "_")
        
        var hookType, event, collection string
        
        if len(parts) == 1 {
            // 系统钩子，如 bootstrap.pb.js
            hookType = "system"
            event = parts[0]
        } else if len(parts) >= 2 {
            // 确定钩子类型和事件
            if parts[0] == "api" {
                hookType = "api"
                event = parts[1]
            } else {
                // 检查是否为集合钩子
                if strings.HasPrefix(parts[1], "collection") {
                    hookType = "collection"
                    event = strings.TrimPrefix(parts[1], "collection.")
                    collection = parts[0]
                } else {
                    // 默认为记录钩子
                    hookType = "record"
                    event = parts[1]
                    collection = parts[0]
                }
            }
        }
        
        // 生成唯一 ID
        id := nanoid.New()
        
        // 插入钩子脚本记录
        _, err = tx.Exec(
            "INSERT INTO pb_hooks_scripts (id, name, description, type, event, collection, code, order, enabled, created, updated) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
            id,
            file.Name(),
            "Migrated from file: " + filePath,
            hookType,
            event,
            collection,
            string(code),
            0, // 默认顺序
            true, // 默认启用
            now,
            now,
        )
        if err != nil {
            return err
        }
        
        // 插入初始版本记录
        _, err = tx.Exec(
            "INSERT INTO pb_hooks_versions (id, hook_id, version, code, comment, created, created_by) VALUES (?, ?, ?, ?, ?, ?, ?)",
            nanoid.New(),
            id,
            1, // 初始版本
            string(code),
            "Initial version migrated from file system",
            now,
            nil, // 无创建者
        )
        if err != nil {
            return err
        }
    }
    
    // 提交事务
    return tx.Commit()
}
```

## 7. 数据访问层

### 7.1 钩子脚本数据模型

以下是 Go 语言中钩子脚本的数据模型定义：

```go
package models

import (
    "time"
    
    "github.com/pocketbase/pocketbase/models"
)

// HookScript 定义钩子脚本模型
type HookScript struct {
    models.BaseModel
    
    Name        string    `db:"name" json:"name"`
    Description string    `db:"description" json:"description,omitempty"`
    Type        string    `db:"type" json:"type"`
    Event       string    `db:"event" json:"event"`
    Collection  string    `db:"collection" json:"collection,omitempty"`
    Code        string    `db:"code" json:"code"`
    Order       int       `db:"order" json:"order"`
    Enabled     bool      `db:"enabled" json:"enabled"`
    Created     time.Time `db:"created" json:"created"`
    Updated     time.Time `db:"updated" json:"updated"`
}

// TableName 返回模型对应的数据库表名
func (h *HookScript) TableName() string {
    return "pb_hooks_scripts"
}

// HookScriptVersion 定义钩子脚本版本模型
type HookScriptVersion struct {
    models.BaseModel
    
    HookID    string    `db:"hook_id" json:"hookId"`
    Version   int       `db:"version" json:"version"`
    Code      string    `db:"code" json:"code"`
    Comment   string    `db:"comment" json:"comment,omitempty"`
    Created   time.Time `db:"created" json:"created"`
    CreatedBy string    `db:"created_by" json:"createdBy,omitempty"`
}

// TableName 返回模型对应的数据库表名
func (v *HookScriptVersion) TableName() string {
    return "pb_hooks_versions"
}

// HookLog 定义钩子脚本执行日志模型 [可选]
type HookLog struct {
    models.BaseModel
    
    HookID        string    `db:"hook_id" json:"hookId"`
    Status        string    `db:"status" json:"status"`
    ExecutionTime int       `db:"execution_time" json:"executionTime"`
    Error         string    `db:"error" json:"error,omitempty"`
    Context       string    `db:"context" json:"context,omitempty"`
    Created       time.Time `db:"created" json:"created"`
}

// TableName 返回模型对应的数据库表名
func (l *HookLog) TableName() string {
    return "pb_hooks_logs"
}
```

### 7.2 钩子脚本数据访问接口

以下是 Go 语言中钩子脚本的数据访问接口定义：

```go
package daos

import (
    "github.com/pocketbase/pocketbase/daos"
    "github.com/pocketbase/pocketbase/models"
    
    "your-project/models" // 替换为实际的模型包路径
)

// HookScriptDAO 定义钩子脚本数据访问接口
type HookScriptDAO interface {
    // 获取所有钩子脚本
    GetAll(filter string, sort string) ([]*models.HookScript, error)
    
    // 分页获取钩子脚本
    GetList(page, perPage int, filter string, sort string) ([]*models.HookScript, int, error)
    
    // 根据 ID 获取钩子脚本
    GetById(id string) (*models.HookScript, error)
    
    // 根据名称获取钩子脚本
    GetByName(name string) (*models.HookScript, error)
    
    // 根据类型和事件获取钩子脚本
    GetByTypeAndEvent(hookType, event string) ([]*models.HookScript, error)
    
    // 根据集合获取钩子脚本
    GetByCollection(collection string) ([]*models.HookScript, error)
    
    // 保存钩子脚本（创建或更新）
    Save(hook *models.HookScript) error
    
    // 删除钩子脚本
    Delete(hook *models.HookScript) error
    
    // 获取钩子脚本的所有版本
    GetVersions(hookId string, page, perPage int) ([]*models.HookScriptVersion, int, error)
    
    // 获取钩子脚本的特定版本
    GetVersion(hookId string, version int) (*models.HookScriptVersion, error)
    
    // 创建新版本
    SaveVersion(version *models.HookScriptVersion) error
    
    // 获取钩子脚本的最新版本号
    GetLatestVersionNumber(hookId string) (int, error)
    
    // 记录钩子脚本执行日志 [可选]
    SaveLog(log *models.HookLog) error
    
    // 获取钩子脚本的执行日志 [可选]
    GetLogs(hookId string, page, perPage int) ([]*models.HookLog, int, error)
}

// 实现 HookScriptDAO 接口
type HookScriptDAOImpl struct {
    dao *daos.Dao
}

// NewHookScriptDAO 创建新的钩子脚本数据访问对象
func NewHookScriptDAO(dao *daos.Dao) HookScriptDAO {
    return &HookScriptDAOImpl{dao: dao}
}

// 实现接口方法...
```

## 8. 性能考虑

### 8.1 索引优化

- 为常用查询条件创建索引，如名称、类型、事件、集合、启用状态和执行顺序
- 为外键创建索引，如版本表和日志表中的 hook_id
- 为排序字段创建索引，如执行顺序和创建时间

### 8.2 查询优化

- 使用参数化查询避免 SQL 注入
- 使用事务确保数据一致性
- 限制返回的记录数量，使用分页
- 只选择需要的字段，避免 SELECT *

### 8.3 缓存策略

- 缓存频繁访问的钩子脚本
- 当脚本更新时使缓存失效
- 使用内存缓存减少数据库访问

## 9. 安全考虑

### 9.1 数据验证

- 验证所有输入数据，特别是脚本代码
- 限制脚本代码的大小
- 验证类型、事件和集合的有效性

### 9.2 访问控制

- 只允许管理员访问钩子脚本管理功能
- 记录所有对钩子脚本的修改
- 实现版本控制，允许回滚恶意修改

### 9.3 脚本执行安全

- 限制脚本执行时间
- 限制脚本可以访问的资源
- 监控脚本执行，记录异常

## 10. 扩展性考虑

### 10.1 未来字段扩展

数据库设计应考虑未来可能的扩展，如：

- 添加标签或分类字段
- 添加更多元数据字段
- 添加权限控制字段

### 10.2 多语言支持

- 考虑添加语言字段，支持多语言脚本
- 考虑添加脚本模板功能

### 10.3 团队协作

- 添加创建者和修改者字段
- 添加审核和批准流程
- 添加评论和讨论功能

## 11. 结论

本文档详细说明了 PocketBase 钩子脚本管理系统的数据库设计，包括表结构、字段定义、索引、关系和约束。该设计旨在支持钩子脚本的存储、版本控制、状态管理和执行顺序控制，以及与 PocketBase 现有系统的集成。

通过实施本设计，可以实现一个功能完善、性能优良、安全可靠的钩子脚本管理系统，为 PocketBase 用户提供更好的开发体验和管理能力。