# PocketBase 钩子脚本管理系统设计总结

## 1. 项目概述

### 1.1 背景

PocketBase 是一个开源的后端解决方案，提供了数据库、认证、文件存储和 API 服务等功能。目前，PocketBase 支持通过在 `pb_hooks` 目录下创建 JavaScript 文件来实现钩子脚本功能，这些脚本可以在特定事件（如记录创建、更新、删除等）发生时自动执行。

然而，这种基于文件系统的钩子脚本管理方式存在以下问题：

- **开发体验差**：需要直接编辑服务器上的文件，无法在管理界面中直接创建和编辑
- **版本控制困难**：缺乏内置的版本历史记录功能
- **部署复杂**：需要手动将脚本文件部署到服务器
- **状态管理不便**：无法轻松启用/禁用脚本，需要通过重命名或移动文件实现
- **执行顺序控制困难**：难以明确控制多个脚本的执行顺序

为解决这些问题，我们设计了一个基于数据库存储的钩子脚本管理系统，通过 PocketBase 管理界面提供直观的脚本创建、编辑、测试和管理功能。

### 1.2 目标

本项目的主要目标是：

1. 设计并实现一个钩子脚本管理系统，将脚本存储在 PocketBase 内置数据库中
2. 开发用户友好的管理界面，支持在浏览器中创建、编辑、测试和管理钩子脚本
3. 提供版本控制功能，记录脚本的历史版本并支持版本比较和回滚
4. 实现脚本状态管理，支持轻松启用/禁用脚本
5. 支持控制脚本执行顺序
6. 提供脚本测试工具，方便验证脚本功能
7. 确保与现有基于文件系统的钩子脚本兼容

## 2. 系统设计概述

### 2.1 系统架构

钩子脚本管理系统采用分层架构，主要包括以下几个层次：

1. **数据存储层**：使用 SQLite 数据库存储钩子脚本信息和内容
2. **后端服务层**：提供钩子脚本的加载、执行和 API 服务
3. **前端界面层**：提供用户友好的管理界面
4. **脚本执行引擎**：负责执行 JavaScript 钩子脚本

### 2.2 核心组件

系统的核心组件包括：

1. **数据模型**：定义钩子脚本和版本历史的数据结构
2. **钩子脚本加载器**：从数据库加载钩子脚本并注册到系统中
3. **钩子脚本 API**：提供脚本的 CRUD、启用/禁用、测试等功能
4. **前端管理界面**：包括脚本列表、编辑器、版本历史和测试工具
5. **脚本执行引擎**：基于 Goja 的 JavaScript 执行环境

### 2.3 技术栈

系统使用以下技术栈：

1. **后端**：Go 语言（与 PocketBase 核心一致）
2. **前端**：Svelte（与 PocketBase 管理界面一致）
3. **数据库**：SQLite（PocketBase 内置数据库）
4. **JavaScript 引擎**：Goja（PocketBase 使用的 JavaScript 引擎）

## 3. 数据库设计

### 3.1 主要表结构

系统主要包含以下数据表：

#### 3.1.1 钩子脚本表 (pb_hooks_scripts)

存储钩子脚本的基本信息和内容：

```sql
CREATE TABLE `pb_hooks_scripts` (
    `id` TEXT PRIMARY KEY NOT NULL,
    `name` TEXT NOT NULL,
    `description` TEXT DEFAULT '',
    `type` TEXT NOT NULL, -- 'record', 'collection', 'api', 'system'
    `event` TEXT NOT NULL, -- 'create', 'update', 'delete', 'create.after', etc.
    `collection` TEXT DEFAULT '', -- 关联的集合，仅对 'record' 和 'collection' 类型有效
    `code` TEXT NOT NULL,
    `order` INTEGER DEFAULT 0, -- 执行顺序
    `enabled` BOOLEAN DEFAULT TRUE,
    `created` TEXT NOT NULL,
    `updated` TEXT NOT NULL
);

CREATE UNIQUE INDEX `idx_hooks_scripts_name` ON `pb_hooks_scripts` (`name`);
CREATE INDEX `idx_hooks_scripts_type_event` ON `pb_hooks_scripts` (`type`, `event`);
CREATE INDEX `idx_hooks_scripts_collection` ON `pb_hooks_scripts` (`collection`);
CREATE INDEX `idx_hooks_scripts_enabled` ON `pb_hooks_scripts` (`enabled`);
```

#### 3.1.2 钩子脚本版本表 (pb_hooks_versions)

存储钩子脚本的历史版本：

```sql
CREATE TABLE `pb_hooks_versions` (
    `id` TEXT PRIMARY KEY NOT NULL,
    `hook_id` TEXT NOT NULL,
    `version` INTEGER NOT NULL, -- 版本号
    `code` TEXT NOT NULL, -- 脚本代码
    `comment` TEXT DEFAULT '', -- 版本说明
    `created` TEXT NOT NULL,
    `created_by` TEXT DEFAULT '', -- 创建者
    FOREIGN KEY (`hook_id`) REFERENCES `pb_hooks_scripts` (`id`) ON DELETE CASCADE
);

CREATE INDEX `idx_hooks_versions_hook_id` ON `pb_hooks_versions` (`hook_id`);
CREATE UNIQUE INDEX `idx_hooks_versions_hook_id_version` ON `pb_hooks_versions` (`hook_id`, `version`);
```

### 3.2 数据关系

- 一个钩子脚本可以有多个历史版本（一对多关系）

## 4. 后端实现

### 4.1 核心数据模型

```go
// HookScript 表示一个钩子脚本
type HookScript struct {
    Id          string    `db:"id" json:"id"`
    Name        string    `db:"name" json:"name"`
    Description string    `db:"description" json:"description"`
    Type        string    `db:"type" json:"type"` // record, collection, api, system
    Event       string    `db:"event" json:"event"`
    Collection  string    `db:"collection" json:"collection"`
    Code        string    `db:"code" json:"code"`
    Order       int       `db:"order" json:"order"`
    Enabled     bool      `db:"enabled" json:"enabled"`
    Created     time.Time `db:"created" json:"created"`
    Updated     time.Time `db:"updated" json:"updated"`
}

// HookScriptVersion 表示一个钩子脚本的历史版本
type HookScriptVersion struct {
    Id        string    `db:"id" json:"id"`
    HookId    string    `db:"hook_id" json:"hook_id"`
    Version   int       `db:"version" json:"version"`
    Code      string    `db:"code" json:"code"`
    Comment   string    `db:"comment" json:"comment"`
    Created   time.Time `db:"created" json:"created"`
    CreatedBy string    `db:"created_by" json:"created_by"`
}
```

### 4.2 钩子脚本加载器

```go
// HookScriptLoader 负责从数据库加载钩子脚本并注册到系统中
type HookScriptLoader struct {
    app core.App
}

// LoadAndRegister 加载所有启用的钩子脚本并注册到系统中
func (l *HookScriptLoader) LoadAndRegister() error {
    // 从数据库加载所有启用的钩子脚本
    scripts, err := l.loadEnabledScripts()
    if err != nil {
        return result, err
    }

    return result, nil
}
```

### 9.2 兼容性策略

为了确保与现有基于文件系统的钩子脚本兼容，系统将同时支持从文件系统和数据库加载钩子脚本：

```go
// LoadHooks 从文件系统和数据库加载钩子脚本
func LoadHooks(app core.App) error {
    // 首先从文件系统加载钩子脚本
    if err := hooks.LoadFromFS(app); err != nil {
        return err
    }

    // 然后从数据库加载钩子脚本
    loader := NewHookScriptLoader(app)
    if err := loader.LoadAndRegister(); err != nil {
        return err
    }

    return nil
}
```

## 10. 性能和安全考虑

### 10.1 性能优化

为了确保钩子脚本管理系统的高性能，我们采取了以下优化措施：

1. **连接池**：使用数据库连接池减少连接开销
2. **缓存**：缓存已编译的脚本，避免重复编译
3. **批量操作**：支持批量导入和导出脚本
4. **分页**：列表页面使用分页加载，减少数据传输量
5. **执行时间限制**：限制脚本执行时间，防止长时间运行的脚本影响系统性能

### 10.2 安全措施

为了确保系统安全，我们实施了以下安全措施：

1. **沙箱隔离**：使用 Goja JavaScript 引擎的沙箱功能隔离脚本执行环境
2. **API 限制**：限制脚本可以访问的 API 和功能
3. **输入验证**：验证所有用户输入，防止注入攻击
4. **权限控制**：只允许管理员访问钩子脚本管理界面
5. **审计日志**：记录所有脚本的创建、修改和执行情况
6. **超时机制**：设置脚本执行超时，防止无限循环
7. **内存限制**：限制脚本可以使用的内存，防止内存泄漏

## 11. 结论

钩子脚本管理系统是 PocketBase 的重要扩展，它解决了基于文件系统的钩子脚本管理方式的诸多问题，提供了更好的开发体验和管理功能。通过将脚本存储在数据库中，并提供用户友好的管理界面，系统使得创建、编辑、测试和管理钩子脚本变得更加简单和高效。

系统的主要优势包括：

1. **集中管理**：所有脚本都可以在管理界面中集中管理
2. **版本控制**：自动保存脚本的历史版本，方便回滚和比较
3. **即时编辑**：直接在浏览器中编辑和测试脚本
4. **状态管理**：轻松启用或禁用脚本
5. **执行顺序控制**：明确控制多个脚本的执行顺序
6. **测试功能**：内置测试工具，方便验证脚本功能

通过这些功能，钩子脚本管理系统大大提高了 PocketBase 的可扩展性和易用性，使开发者能够更加高效地实现自定义业务逻辑。
### 4.3 API 实现

```go
// RegisterHookScriptsApi 注册钩子脚本 API
func RegisterHookScriptsApi(app core.App, router *echo.Group) {
    api := &HookScriptsApi{
        app: app,
    }

    subGroup := router.Group("/hooks", apis.RequireAdminAuth())

    // 钩子脚本 CRUD
    subGroup.GET("", api.list)
    subGroup.GET("/:id", api.view)
    subGroup.POST("", api.create)
    subGroup.PATCH("/:id", api.update)
    subGroup.DELETE("/:id", api.delete)

    // 启用/禁用钩子脚本
    subGroup.PATCH("/:id/enable", api.enable)
    subGroup.PATCH("/:id/disable", api.disable)

    // 更新执行顺序
    subGroup.PATCH("/reorder", api.reorder)

    // 测试钩子脚本
    subGroup.POST("/:id/test", api.test)

    // 版本管理
    subGroup.GET("/:id/versions", api.listVersions)
    subGroup.GET("/:id/versions/:version", api.viewVersion)
    subGroup.POST("/:id/versions", api.createVersion)
    subGroup.POST("/:id/versions/:version/restore", api.restoreVersion)

    // 导入/导出
    subGroup.POST("/import", api.importScripts)
    subGroup.POST("/export", api.exportScripts)

    // 从文件系统迁移
    subGroup.POST("/migrate-from-fs", api.migrateFromFs)

    // 元数据
    subGroup.GET("/meta", api.meta)
}
```

## 5. 前端实现

### 5.1 路由配置

```javascript
// 在 src/components/routes.js 中添加钩子脚本管理路由
export default [
    // 其他路由...
    {
        path: "/hooks",
        component: () => import("./hooks/HooksList.svelte"),
        title: "钩子脚本",
        icon: "hook",
        requireAuth: true,
    },
    {
        path: "/hooks/create",
        component: () => import("./hooks/HookUpsertPanel.svelte"),
        title: "创建钩子脚本",
        requireAuth: true,
        hide: true,
    },
    {
        path: "/hooks/:id",
        component: () => import("./hooks/HookUpsertPanel.svelte"),
        title: "编辑钩子脚本",
        requireAuth: true,
        hide: true,
    },
    {
        path: "/hooks/:id/versions",
        component: () => import("./hooks/HookVersionsList.svelte"),
        title: "钩子脚本版本历史",
        requireAuth: true,
        hide: true,
    },
    {
        path: "/hooks/:id/test",
        component: () => import("./hooks/HookTestPanel.svelte"),
        title: "测试钩子脚本",
        requireAuth: true,
        hide: true,
    },
];
```

### 5.2 脚本列表组件

```html
<!-- src/components/hooks/HooksList.svelte -->
<script>
    import { onMount } from "svelte";
    import { link } from "svelte-spa-router";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addSuccessToast } from "@/stores/toasts";
    import { confirm } from "@/components/overlays/ConfirmDialog.svelte";
    import Pagination from "@/components/base/Pagination.svelte";
    import Field from "@/components/base/Field.svelte";
    import Loader from "@/components/base/Loader.svelte";
    import HookIcon from "@/components/icons/HookIcon.svelte";

    let hooks = [];
    let isLoading = true;
    let totalHooks = 0;
    let page = 1;
    let perPage = 30;
    let search = "";
    let filter = "";

    // 加载钩子脚本列表
    async function loadHooks() {
        isLoading = true;

        try {
            const result = await ApiClient.hooks.getList(page, perPage, filter);
            hooks = result.items;
            totalHooks = result.totalItems;
        } catch (err) {
            console.error("Failed to load hooks:", err);
        }

        isLoading = false;
    }

    // 启用/禁用钩子脚本
    async function toggleHookStatus(hook) {
        try {
            if (hook.enabled) {
                await ApiClient.hooks.disable(hook.id);
                hook.enabled = false;
            } else {
                await ApiClient.hooks.enable(hook.id);
                hook.enabled = true;
            }
            addSuccessToast(`钩子脚本 "${hook.name}" 已${hook.enabled ? "启用" : "禁用"}`);
        } catch (err) {
            console.error(`Failed to ${hook.enabled ? "disable" : "enable"} hook:`, err);
        }
    }

    // 删除钩子脚本
    async function deleteHook(hook) {
        if (!await confirm(`确定要删除钩子脚本 "${hook.name}" 吗？此操作不可撤销。`)) {
            return;
        }

        try {
            await ApiClient.hooks.delete(hook.id);
            hooks = hooks.filter(h => h.id !== hook.id);
            totalHooks--;
            addSuccessToast(`钩子脚本 "${hook.name}" 已删除`);
        } catch (err) {
            console.error("Failed to delete hook:", err);
        }
    }

    // 处理搜索
    function handleSearch() {
        filter = search ? `name ~ "${search}" || description ~ "${search}"` : "";
        page = 1;
        loadHooks();
    }

    // 处理分页
    function handlePageChange(e) {
        page = e.detail;
        loadHooks();
    }

    onMount(loadHooks);
</script>

<div class="page-wrapper">
    <header class="page-header">
        <h1 class="page-title">
            <HookIcon size="24" class="icon" />
            <span class="txt">钩子脚本</span>
        </h1>
        <div class="page-header-actions">
            <a href="/hooks/create" use:link class="btn btn-primary">
                <span class="txt">新建钩子脚本</span>
            </a>
            <button class="btn btn-outline" on:click={() => loadHooks()}>
                <span class="txt">刷新</span>
            </button>
            <div class="dropdown">
                <button class="btn btn-outline dropdown-toggle">
                    <span class="txt">更多</span>
                </button>
                <div class="dropdown-menu">
                    <button class="dropdown-item" on:click={() => window.location.href = "/hooks/import"}>
                        <span class="txt">导入</span>
                    </button>
                    <button class="dropdown-item" on:click={() => window.location.href = "/hooks/export"}>
                        <span class="txt">导出</span>
                    </button>
                    <button class="dropdown-item" on:click={() => window.location.href = "/hooks/migrate-from-fs"}>
                        <span class="txt">从文件系统迁移</span>
                    </button>
                </div>
            </div>
        </div>
    </header>

    <div class="page-body">
        <div class="filters">
            <Field class="form-field search-field" name="search" let:uniqueId>
                <label for={uniqueId}>搜索</label>
                <div class="form-field-addon">
                    <input
                        type="text"
                        id={uniqueId}
                        placeholder="按名称或描述搜索..."
                        bind:value={search}
                        on:keydown={(e) => e.key === "Enter" && handleSearch()}
                    />
                    <button class="btn btn-sm btn-secondary" on:click={handleSearch}>
                        <span class="txt">搜索</span>
                    </button>
                </div>
            </Field>
        </div>

        {#if isLoading}
            <Loader />
        {:else if hooks.length === 0}
            <div class="block txt-center m-b-lg">
                <h6>没有找到钩子脚本</h6>
                {#if filter}
                    <button class="btn btn-sm btn-secondary m-t-sm" on:click={() => { search = ""; filter = ""; loadHooks(); }}>
                        <span class="txt">清除筛选器</span>
                    </button>
                {:else}
                    <div class="m-t-sm">
                        <a href="/hooks/create" use:link class="btn btn-sm btn-secondary">
                            <span class="txt">创建第一个钩子脚本</span>
                        </a>
                    </div>
                {/if}
            </div>
        {:else}
            <div class="table-wrapper">
                <table class="table">
                    <thead>
                        <tr>
                            <th>名称</th>
                            <th>类型</th>
                            <th>事件</th>
                            <th>集合</th>
                            <th>状态</th>
                            <th>顺序</th>
                            <th>更新时间</th>
                            <th class="actions">操作</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each hooks as hook (hook.id)}
                            <tr>
                                <td>
                                    <div class="flex gap-5 align-center">
                                        <div class="hook-name">
                                            <a href={`/hooks/${hook.id}`} use:link>
                                                {hook.name}
                                            </a>
                                            {#if hook.description}
                                                <div class="hook-description">{hook.description}</div>
                                            {/if}
                                        </div>
                                    </div>
                                </td>
                                <td>{hook.type}</td>
                                <td>{hook.event}</td>
                                <td>{hook.collection || "-"}</td>
                                <td>
                                    <div class="flex">
                                        <span class={`label ${hook.enabled ? "success" : "gray"}`}>
                                            {hook.enabled ? "启用" : "禁用"}
                                        </span>
                                    </div>
                                </td>
                                <td>{hook.order}</td>
                                <td>{CommonHelper.formatToLocalDateTime(hook.updated)}</td>
                                <td class="actions">
                                    <div class="flex gap-5 justify-end">
                                        <button
                                            class="btn btn-sm btn-outline"
                                            title={hook.enabled ? "禁用" : "启用"}
                                            on:click={() => toggleHookStatus(hook)}
                                        >
                                            <span class="txt">{hook.enabled ? "禁用" : "启用"}</span>
                                        </button>
                                        <a
                                            href={`/hooks/${hook.id}/test`}
                                            use:link
                                            class="btn btn-sm btn-outline"
                                            title="测试"
                                        >
                                            <span class="txt">测试</span>
                                        </a>
                                        <a
                                            href={`/hooks/${hook.id}/versions`}
                                            use:link
                                            class="btn btn-sm btn-outline"
                                            title="版本历史"
                                        >
                                            <span class="txt">版本</span>
                                        </a>
                                        <a
                                            href={`/hooks/${hook.id}`}
                                            use:link
                                            class="btn btn-sm btn-outline"
                                            title="编辑"
                                        >
                                            <span class="txt">编辑</span>
                                        </a>
                                        <button
                                            class="btn btn-sm btn-outline btn-danger"
                                            title="删除"
                                            on:click={() => deleteHook(hook)}
                                        >
                                            <span class="txt">删除</span>
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        {/each}
                    </tbody>
                </table>
            </div>

            <Pagination
                page={page}
                totalItems={totalHooks}
                perPage={perPage}
                on:change={handlePageChange}
            />
        {/if}
    </div>
</div>

<style>
    .hook-name {
        font-weight: 600;
    }
    .hook-description {
        font-size: 0.85rem;
        color: var(--txtHintColor);
        margin-top: 3px;
    }
    .filters {
        display: flex;
        gap: 15px;
        margin-bottom: 15px;
    }
    .search-field {
        max-width: 320px;
    }
</style>
```

### 5.3 脚本编辑组件

```html
<!-- src/components/hooks/HookUpsertPanel.svelte -->
<script>
    import { onMount } from "svelte";
    import { push } from "svelte-spa-router";
    import { params } from "svelte-spa-router/params";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import { confirm } from "@/components/overlays/ConfirmDialog.svelte";
    import Field from "@/components/base/Field.svelte";
    import Loader from "@/components/base/Loader.svelte";
    import CodeEditor from "@/components/base/CodeEditor.svelte";

    const hookId = $params.id;
    const isNew = !hookId;

    let isLoading = true;
    let isSaving = false;
    let hook = {
        name: "",
        description: "",
        type: "record",
        event: "create",
        collection: "",
        code: "",
        order: 0,
        enabled: true,
    };
    let meta = {
        types: [],
        events: {},
        collections: [],
    };
    let availableEvents = [];

    // 加载钩子脚本元数据
    async function loadMeta() {
        try {
            meta = await ApiClient.hooks.getMeta();
            updateAvailableEvents();
        } catch (err) {
            console.error("Failed to load hook meta:", err);
            addErrorToast("加载钩子脚本元数据失败");
        }
    }

    // 加载钩子脚本
    async function loadHook() {
        if (isNew) {
            isLoading = false;
            return;
        }

        try {
            hook = await ApiClient.hooks.getOne(hookId);
            updateAvailableEvents();
        } catch (err) {
            console.error("Failed to load hook:", err);
            addErrorToast("加载钩子脚本失败");
            push("/hooks");
        }

        isLoading = false;
    }

    // 更新可用事件列表
    function updateAvailableEvents() {
        if (!meta.events || !hook.type) {
            availableEvents = [];
            return;
        }

        availableEvents = meta.events[hook.type] || [];

        // 如果当前事件不在可用事件列表中，选择第一个可用事件
        if (availableEvents.length > 0 && !availableEvents.includes(hook.event)) {
            hook.event = availableEvents[0];
        }
    }

    // 保存钩子脚本
    async function saveHook() {
        if (!hook.name) {
            addErrorToast("请输入钩子脚本名称");
            return;
        }

        if (!hook.code) {
            addErrorToast("请输入钩子脚本代码");
            return;
        }

        isSaving = true;

        try {
            if (isNew) {
                const result = await ApiClient.hooks.create(hook);
                addSuccessToast("钩子脚本创建成功");
                push(`/hooks/${result.id}`);
            } else {
                await ApiClient.hooks.update(hookId, hook);
                addSuccessToast("钩子脚本更新成功");
            }
        } catch (err) {
            console.error("Failed to save hook:", err);
            addErrorToast("保存钩子脚本失败");
        }

        isSaving = false;
    }

    // 保存为新版本
    async function saveAsNewVersion() {
        if (!hookId) {
            return;
        }

        const comment = await CommonHelper.prompt("请输入版本说明", "");
        if (comment === null) {
            return;
        }

        isSaving = true;

        try {
            await ApiClient.hooks.createVersion(hookId, {
                code: hook.code,
                comment: comment,
            });
            addSuccessToast("新版本创建成功");
        } catch (err) {
            console.error("Failed to create version:", err);
            addErrorToast("创建新版本失败");
        }

        isSaving = false;
    }

    // 处理类型变更
    function handleTypeChange() {
        updateAvailableEvents();
    }

    onMount(async () => {
        await loadMeta();
        await loadHook();
    });
</script>

<div class="page-wrapper">
    <header class="page-header">
        <nav class="breadcrumbs">
            <div class="breadcrumb-item">
                <a href="/hooks" class="btn btn-sm btn-circle btn-back">
                    <i class="ri-arrow-left-s-line" />
                </a>
            </div>
            <div class="breadcrumb-item">
                <a href="/hooks">钩子脚本</a>
            </div>
            <div class="breadcrumb-item active">
                {isNew ? "创建" : "编辑"}
            </div>
        </nav>

        <div class="page-header-actions">
            {#if !isNew}
                <button class="btn btn-outline" on:click={saveAsNewVersion} disabled={isSaving}>
                    <span class="txt">保存为新版本</span>
                </button>
                <a href={`/hooks/${hookId}/test`} class="btn btn-outline">
                    <span class="txt">测试</span>
                </a>
                <a href={`/hooks/${hookId}/versions`} class="btn btn-outline">
                    <span class="txt">版本历史</span>
                </a>
            {/if}
            <button class="btn btn-primary" on:click={saveHook} disabled={isSaving}>
                <span class="txt">{isSaving ? "保存中..." : "保存"}</span>
            </button>
        </div>
    </header>

    <div class="page-body">
        {#if isLoading}
            <Loader />
        {:else}
            <form class="hook-form" on:submit|preventDefault={saveHook}>
                <div class="grid">
                    <div class="col-sm-12 col-md-6">
                        <Field class="form-field required" name="name" let:uniqueId>
                            <label for={uniqueId}>名称</label>
                            <input
                                type="text"
                                id={uniqueId}
                                bind:value={hook.name}
                                placeholder="钩子脚本名称"
                                required
                            />
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-6">
                        <Field class="form-field" name="description" let:uniqueId>
                            <label for={uniqueId}>描述</label>
                            <input
                                type="text"
                                id={uniqueId}
                                bind:value={hook.description}
                                placeholder="钩子脚本描述（可选）"
                            />
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-4">
                        <Field class="form-field required" name="type" let:uniqueId>
                            <label for={uniqueId}>类型</label>
                            <select
                                id={uniqueId}
                                bind:value={hook.type}
                                on:change={handleTypeChange}
                                required
                            >
                                {#each meta.types as type}
                                    <option value={type}>{type}</option>
                                {/each}
                            </select>
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-4">
                        <Field class="form-field required" name="event" let:uniqueId>
                            <label for={uniqueId}>事件</label>
                            <select
                                id={uniqueId}
                                bind:value={hook.event}
                                required
                            >
                                {#each availableEvents as event}
                                    <option value={event}>{event}</option>
                                {/each}
                            </select>
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-4">
                        <Field class="form-field" class:disabled={hook.type !== "record" && hook.type !== "collection"} name="collection" let:uniqueId>
                            <label for={uniqueId}>集合</label>
                            <select
                                id={uniqueId}
                                bind:value={hook.collection}
                                disabled={hook.type !== "record" && hook.type !== "collection"}
                            >
                                <option value="">所有集合</option>
                                {#each meta.collections as collection}
                                    <option value={collection}>{collection}</option>
                                {/each}
                            </select>
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-6">
                        <Field class="form-field" name="order" let:uniqueId>
                            <label for={uniqueId}>执行顺序</label>
                            <input
                                type="number"
                                id={uniqueId}
                                bind:value={hook.order}
                                min="0"
                                step="1"
                            />
                            <div class="help-block">
                                数字越小越先执行
                            </div>
                        </Field>
                    </div>

                    <div class="col-sm-12 col-md-6">
                        <Field class="form-field" name="enabled" let:uniqueId>
                            <label for={uniqueId}>状态</label>
                            <div class="form-field-addon">
                                <label class="switch">
                                    <input
                                        type="checkbox"
                                        id={uniqueId}
                                        bind:checked={hook.enabled}
                                    />
                                    <span class="slider" />
                                </label>
                                <div class="help-block">
                                    {hook.enabled ? "启用" : "禁用"}
                                </div>
                            </div>
                        </Field>
                    </div>

                    <div class="col-sm-12">
                        <Field class="form-field required" name="code" let:uniqueId>
                            <label for={uniqueId}>代码</label>
                            <CodeEditor
                                id={uniqueId}
                                bind:value={hook.code}
                                lang="javascript"
                                height="400px"
                            />
                        </Field>
                    </div>
                </div>
            </form>
        {/if}
    </div>
</div>

<style>
    .hook-form {
        max-width: 1000px;
    }
    .grid {
        display: grid;
        grid-template-columns: repeat(12, 1fr);
        gap: 15px;
    }
    .col-sm-12 {
        grid-column: span 12;
    }
    @media (min-width: 768px) {
        .col-md-4 {
            grid-column: span 4;
        }
        .col-md-6 {
            grid-column: span 6;
        }
    }
    .disabled {
        opacity: 0.6;
        pointer-events: none;
    }
</style>
```

### 5.4 API 客户端扩展

```javascript
// 在 src/utils/ApiClient.js 中添加钩子脚本 API

// 现有代码...

// 钩子脚本 API
ApiClient.hooks = {
    // 获取钩子脚本列表
    getList: async function(page = 1, perPage = 30, filter = "") {
        const query = {};
        if (page) {
            query.page = page;
        }
        if (perPage) {
            query.perPage = perPage;
        }
        if (filter) {
            query.filter = filter;
        }

        return ApiClient.send("GET", "/api/hooks", query);
    },

    // 获取单个钩子脚本
    getOne: async function(id) {
        return ApiClient.send("GET", `/api/hooks/${encodeURIComponent(id)}`);
    },

    // 创建钩子脚本
    create: async function(hookData) {
        return ApiClient.send("POST", "/api/hooks", {}, hookData);
    },

    // 更新钩子脚本
    update: async function(id, hookData) {
        return ApiClient.send("PATCH", `/api/hooks/${encodeURIComponent(id)}`, {}, hookData);
    },

    // 删除钩子脚本
    delete: async function(id) {
        return ApiClient.send("DELETE", `/api/hooks/${encodeURIComponent(id)}`);
    },

    // 启用钩子脚本
    enable: async function(id) {
        return ApiClient.send("PATCH", `/api/hooks/${encodeURIComponent(id)}/enable`);
    },

    // 禁用钩子脚本
    disable: async function(id) {
        return ApiClient.send("PATCH", `/api/hooks/${encodeURIComponent(id)}/disable`);
    },

    // 更新执行顺序
    reorder: async function(orderData) {
        return ApiClient.send("PATCH", "/api/hooks/reorder", {}, orderData);
    },

    // 测试钩子脚本
    test: async function(id, testData) {
        return ApiClient.send("POST", `/api/hooks/${encodeURIComponent(id)}/test`, {}, testData);
    },

    // 获取版本列表
    getVersions: async function(id, page = 1, perPage = 30) {
        const query = {};
        if (page) {
            query.page = page;
        }
        if (perPage) {
            query.perPage = perPage;
        }

        return ApiClient.send("GET", `/api/hooks/${encodeURIComponent(id)}/versions`, query);
    },

    // 获取单个版本
    getVersion: async function(id, version) {
        return ApiClient.send("GET", `/api/hooks/${encodeURIComponent(id)}/versions/${encodeURIComponent(version)}`);
    },

    // 创建新版本
    createVersion: async function(id, versionData) {
        return ApiClient.send("POST", `/api/hooks/${encodeURIComponent(id)}/versions`, {}, versionData);
    },

    // 恢复到特定版本
    restoreVersion: async function(id, version, comment = "") {
        return ApiClient.send("POST", `/api/hooks/${encodeURIComponent(id)}/versions/${encodeURIComponent(version)}/restore`, {}, { comment });
    },

    // 导入脚本
    import: async function(formData) {
        return ApiClient.send("POST", "/api/hooks/import", {}, formData, {
            "Content-Type": "multipart/form-data",
        });
    },

    // 导出脚本
    export: async function(ids, format = "json") {
        return ApiClient.send("POST", "/api/hooks/export", {}, { ids, format });
    },

    // 从文件系统迁移
    migrateFromFs: async function(options = {}) {
        return ApiClient.send("POST", "/api/hooks/migrate-from-fs", {}, options);
    },

    // 获取元数据
    getMeta: async function() {
        return ApiClient.send("GET", "/api/hooks/meta");
    },
};

// 导出 ApiClient
export default ApiClient;
```

## 6. 集成与启动

### 6.1 在 main.go 中注册钩子脚本管理插件

```go
package main

import (
    "log"
    "os"

    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/plugins/hooks"
)

func main() {
    app := pocketbase.New()

    // 注册钩子脚本管理插件
    hooks.Register(app)

    // 启动应用
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 6.2 钩子脚本管理插件注册函数

```go
package hooks

import (
    "github.com/labstack/echo/v5"
    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/apis"
)

// Register 注册钩子脚本管理插件
func Register(app core.App) {
    // 创建钩子脚本加载器
    loader := NewHookScriptLoader(app)

    // 注册钩子脚本 API
    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        RegisterHookScriptsApi(app, e.Router.Group("/api"))
        return nil
    })

    // 在应用启动前加载和注册钩子脚本
    app.OnBeforeBootstrap().Add(func(e *core.BootstrapEvent) error {
        return loader.LoadAndRegister()
    })
}
```

## 7. 数据库迁移

### 7.1 创建钩子脚本表

```go
package migrations

import (
    "github.com/pocketbase/dbx"
    "github.com/pocketbase/pocketbase/daos"
    "github.com/pocketbase/pocketbase/migrations"
)

// 创建钩子脚本表的迁移
func init() {
    migrations.Register(func(db dbx.Builder) error {
        dao := daos.New(db);

        // 创建钩子脚本表
        err := dao.DB().NewQuery(`
            CREATE TABLE IF NOT EXISTS {{$prefix}}hooks_scripts (
                id TEXT PRIMARY KEY NOT NULL,
                name TEXT NOT NULL,
                description TEXT DEFAULT '',
                type TEXT NOT NULL,
                event TEXT NOT NULL,
                collection TEXT DEFAULT '',
                code TEXT NOT NULL,
                order INTEGER DEFAULT 0,
                enabled BOOLEAN DEFAULT TRUE,
                created TEXT NOT NULL,
                updated TEXT NOT NULL
            );

            CREATE UNIQUE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_scripts_name ON {{$prefix}}hooks_scripts (name);
            CREATE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_scripts_type_event ON {{$prefix}}hooks_scripts (type, event);
            CREATE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_scripts_collection ON {{$prefix}}hooks_scripts (collection);
            CREATE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_scripts_enabled ON {{$prefix}}hooks_scripts (enabled);
        `).Execute()

        if err != nil {
            return err
        }

        // 创建钩子脚本版本表
        return dao.DB().NewQuery(`
            CREATE TABLE IF NOT EXISTS {{$prefix}}hooks_versions (
                id TEXT PRIMARY KEY NOT NULL,
                hook_id TEXT NOT NULL,
                version INTEGER NOT NULL,
                code TEXT NOT NULL,
                comment TEXT DEFAULT '',
                created TEXT NOT NULL,
                created_by TEXT DEFAULT '',
                FOREIGN KEY (hook_id) REFERENCES {{$prefix}}hooks_scripts (id) ON DELETE CASCADE
            );

            CREATE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_versions_hook_id ON {{$prefix}}hooks_versions (hook_id);
            CREATE UNIQUE INDEX IF NOT EXISTS idx_{{$prefix}}hooks_versions_hook_id_version ON {{$prefix}}hooks_versions (hook_id, version);
        `).Execute()
    }, func(db dbx.Builder) error {
        dao := daos.New(db);

        // 删除钩子脚本版本表
        err := dao.DB().NewQuery(`DROP TABLE IF EXISTS {{$prefix}}hooks_versions`).Execute()
        if err != nil {
            return err
        }

        // 删除钩子脚本表
        return dao.DB().NewQuery(`DROP TABLE IF EXISTS {{$prefix}}hooks_scripts`).Execute()
    })
}
```

## 8. 测试计划

### 8.1 单元测试

```go
package hooks

import (
    "testing"
    "time"

    "github.com/pocketbase/pocketbase/tests"
)

func TestHookScriptLoader(t *testing.T) {
    app, _ := tests.NewTestApp()
    defer app.Cleanup()

    // 创建测试钩子脚本
    script := &HookScript{
        Id:          "test123",
        Name:        "test_hook",
        Description: "Test hook script",
        Type:        "record",
        Event:       "create",
        Collection:  "users",
        Code:        "console.log('Test hook script');",
        Order:       0,
        Enabled:     true,
        Created:     time.Now(),
        Updated:     time.Now(),
    }

    // 保存测试钩子脚本
    err := app.Dao().Save(script)
    if err != nil {
        t.Fatal(err)
    }

    // 创建钩子脚本加载器
    loader := NewHookScriptLoader(app)

    // 加载并注册钩子脚本
    err = loader.LoadAndRegister()
    if err != nil {
        t.Fatal(err)
    }

    // 验证钩子脚本是否已注册
    // ...
}
```

### 8.2 集成测试

```go
package hooks

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v5"
    "github.com/pocketbase/pocketbase/tests"
)

func TestHookScriptsApi(t *testing.T) {
    app, _ := tests.NewTestApp()
    defer app.Cleanup()

    // 创建 API 路由
    router := echo.New()
    RegisterHookScriptsApi(app, router.Group("/api"))

    // 测试获取钩子脚本列表
    req := httptest.NewRequest(http.MethodGet, "/api/hooks", nil)
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("Expected status code %d, got %d", http.StatusOK, rec.Code)
    }

    // 测试创建钩子脚本
    reqBody := `{
        "name": "test_hook",
        "description": "Test hook script",
        "type": "record",
        "event": "create",
        "collection": "users",
        "code": "console.log('Test hook script');",
        "order": 0,
        "enabled": true
    }`

    req = httptest.NewRequest(http.MethodPost, "/api/hooks", strings.NewReader(reqBody))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec = httptest.NewRecorder()

    router.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("Expected status code %d, got %d", http.StatusOK, rec.Code)
    }

    // 其他 API 测试...
}
```

### 8.3 端到端测试

```javascript
// 使用 Playwright 进行端到端测试

const { test, expect } = require('@playwright/test');

test('钩子脚本管理界面', async ({ page }) => {
    // 登录管理界面
    await page.goto('http://localhost:8090/_/');
    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'password123');
    await page.click('button[type="submit"]');

    // 导航到钩子脚本页面
    await page.click('a[href="/hooks"]');
    await expect(page).toHaveURL(/.*\/hooks/);

    // 创建新钩子脚本
    await page.click('text=新建钩子脚本');
    await expect(page).toHaveURL(/.*\/hooks\/create/);

    await page.fill('input[name="name"]', 'test_hook');
    await page.fill('input[name="description"]', 'Test hook script');
    await page.selectOption('select[name="type"]', 'record');
    await page.selectOption('select[name="event"]', 'create');
    await page.selectOption('select[name="collection"]', 'users');

    // 输入脚本代码
    await page.fill('.CodeMirror textarea', 'console.log("Test hook script");');

    // 保存脚本
    await page.click('text=保存');
    await expect(page.locator('.toast-success')).toBeVisible();

    // 验证脚本是否显示在列表中
    await page.goto('/hooks');
    await expect(page.locator('text=test_hook')).toBeVisible();

    // 编辑脚本
    await page.click('text=test_hook');
    await page.fill('input[name="description"]', 'Updated test hook script');
    await page.click('text=保存');
    await expect(page.locator('.toast-success')).toBeVisible();

    // 测试脚本
    await page.click('text=测试');
    await page.fill('textarea', JSON.stringify({
        collection: { name: 'users' },
        record: {},
        data: { username: 'test', email: 'test@example.com' }
    }));
    await page.click('text=运行测试');
    await expect(page.locator('.test-result')).toBeVisible();

    // 禁用脚本
    await page.goto('/hooks');
    await page.click('text=禁用');
    await expect(page.locator('.toast-success')).toBeVisible();

    // 删除脚本
    await page.click('text=删除');
    await page.click('text=确认');
    await expect(page.locator('.toast-success')).toBeVisible();
    await expect(page.locator('text=test_hook')).not.toBeVisible();
});
```

## 9. 部署和迁移策略

### 9.1 将文件系统脚本迁移到数据库

```go
package hooks

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"
    "time"

    "github.com/pocketbase/pocketbase/core"
    "github.com/pocketbase/pocketbase/tools/types"
)

// MigrateFromFs 将文件系统中的钩子脚本迁移到数据库
func MigrateFromFs(app core.App, options map[string]interface{}) (map[string]interface{}, error) {
    // 获取钩子脚本目录
    hooksDir := app.Settings().HooksDir
    if dir, ok := options["dir"].(string); ok && dir != "" {
        hooksDir = dir
    }

    // 是否删除原始文件
    deleteOriginal := false
    if del, ok := options["deleteOriginal"].(bool); ok {
        deleteOriginal = del
    }

    // 是否覆盖同名脚本
    overwrite := false
    if ow, ok := options["overwrite"].(bool); ok {
        overwrite = ow
    }

    // 结果统计
    result := map[string]interface{}{
        "total":   0,
        "success": 0,
        "skipped": 0,
        "failed":  0,
        "errors":  []string{},
    }

    // 遍历钩子脚本目录
    err := filepath.Walk(hooksDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        // 跳过目录和非 JS 文件
        if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".js") {
            return nil
        }

        result["total"] = result["total"].(int) + 1

        // 读取文件内容
        code, err := ioutil.ReadFile(path)
        if err != nil {
            result["failed"] = result["failed"].(int) + 1
            result["errors"] = append(result["errors"].([]string), err.Error())
            return nil
        }

        // 解析文件名
        filename := filepath.Base(path)
        name := strings.TrimSuffix(filename, filepath.Ext(filename))

        // 解析钩子类型和事件
        parts := strings.Split(name, "_")
        if len(parts) < 2 {
            result["failed"] = result["failed"].(int) + 1
            result["errors"] = append(result["errors"].([]string), "Invalid hook script name: "+name)
            return nil
        }

        hookType := parts[0]
        hookEvent := parts[1]
        hookCollection := ""

        // 解析集合名称（如果有）
        if len(parts) > 2 {
            hookCollection = parts[2]
        }

        // 检查是否已存在同名脚本
        existing, _ := app.Dao().FindFirstRecordByData("hooks_scripts", "name", name)
        if existing != nil && !overwrite {
            result["skipped"] = result["skipped"].(int) + 1
            return nil
        }

        // 创建或更新钩子脚本
        script := &HookScript{
            Name:        name,
            Description: "Migrated from file: " + filename,
            Type:        hookType,
            Event:       hookEvent,
            Collection:  hookCollection,
            Code:        string(code),
            Order:       0,
            Enabled:     true,
            Created:     types.NowDateTime(),
            Updated:     types.NowDateTime(),
        }

        if existing != nil {
            script.Id = existing.Id
        }

        // 保存钩子脚本
        err = app.Dao().SaveRecord(script)
        if err != nil {
            result["failed"] = result["failed"].(int) + 1
            result["errors"] = append(result["errors"].([]string), err.Error())
            return nil
        }

        result["success"] = result["success"].(int) + 1

        // 删除原始文件（如果需要）
        if deleteOriginal {
            os.Remove(path)
        }

        return nil
    })

    if err != nil {
        return result, err