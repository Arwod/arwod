# PocketBase 钩子脚本管理系统技术实现

## 1. 技术架构

### 1.1 系统组件

![系统架构图](https://placeholder-for-architecture-diagram.png)

钩子脚本管理系统由以下主要组件组成：

1. **数据存储层**：使用PocketBase内置的SQLite数据库
2. **后端服务层**：基于Go语言的PocketBase核心扩展
3. **前端界面层**：基于Svelte的管理界面
4. **脚本执行引擎**：基于Goja的JavaScript运行时

### 1.2 技术栈

- **后端**：Go语言
- **前端**：Svelte、TypeScript、Monaco Editor
- **数据库**：SQLite
- **脚本运行时**：Goja (Go实现的JavaScript引擎)

## 2. 数据库设计详情

### 2.1 数据表结构

#### 2.1.1 `pb_hooks_scripts` 表

```sql
CREATE TABLE "pb_hooks_scripts" (
  "id" TEXT PRIMARY KEY,
  "name" TEXT NOT NULL,
  "description" TEXT,
  "code" TEXT NOT NULL,
  "type" TEXT NOT NULL,
  "event" TEXT NOT NULL,
  "collection_id" TEXT,
  "enabled" BOOLEAN NOT NULL DEFAULT TRUE,
  "order" INTEGER NOT NULL DEFAULT 0,
  "created" TEXT NOT NULL,
  "updated" TEXT NOT NULL,
  "created_by" TEXT,
  "updated_by" TEXT,
  FOREIGN KEY ("collection_id") REFERENCES "_collections" ("id") ON DELETE SET NULL,
  FOREIGN KEY ("created_by") REFERENCES "_superusers" ("id") ON DELETE SET NULL,
  FOREIGN KEY ("updated_by") REFERENCES "_superusers" ("id") ON DELETE SET NULL
);
```

#### 2.1.2 `pb_hooks_versions` 表（可选）

```sql
CREATE TABLE "pb_hooks_versions" (
  "id" TEXT PRIMARY KEY,
  "script_id" TEXT NOT NULL,
  "code" TEXT NOT NULL,
  "comment" TEXT,
  "created" TEXT NOT NULL,
  "created_by" TEXT,
  FOREIGN KEY ("script_id") REFERENCES "pb_hooks_scripts" ("id") ON DELETE CASCADE,
  FOREIGN KEY ("created_by") REFERENCES "_superusers" ("id") ON DELETE SET NULL
);
```

### 2.2 索引设计

```sql
CREATE INDEX "idx_hooks_scripts_type_event" ON "pb_hooks_scripts" ("type", "event");
CREATE INDEX "idx_hooks_scripts_collection" ON "pb_hooks_scripts" ("collection_id");
CREATE INDEX "idx_hooks_versions_script" ON "pb_hooks_versions" ("script_id");
```

## 3. 后端实现

### 3.1 核心数据模型

```go
// HookScript 表示存储在数据库中的钩子脚本
type HookScript struct {
    Id           string         `db:"id" json:"id"`
    Name         string         `db:"name" json:"name"`
    Description  string         `db:"description" json:"description"`
    Code         string         `db:"code" json:"code"`
    Type         string         `db:"type" json:"type"`
    Event        string         `db:"event" json:"event"`
    CollectionId *string        `db:"collection_id" json:"collection_id"`
    Enabled      bool           `db:"enabled" json:"enabled"`
    Order        int            `db:"order" json:"order"`
    Created      types.DateTime `db:"created" json:"created"`
    Updated      types.DateTime `db:"updated" json:"updated"`
    CreatedBy    *string        `db:"created_by" json:"created_by"`
    UpdatedBy    *string        `db:"updated_by" json:"updated_by"`
}

// HookScriptVersion 表示钩子脚本的历史版本
type HookScriptVersion struct {
    Id        string         `db:"id" json:"id"`
    ScriptId  string         `db:"script_id" json:"script_id"`
    Code      string         `db:"code" json:"code"`
    Comment   string         `db:"comment" json:"comment"`
    Created   types.DateTime `db:"created" json:"created"`
    CreatedBy *string        `db:"created_by" json:"created_by"`
}
```

### 3.2 钩子脚本加载器

```go
// HookScriptLoader 负责从数据库加载钩子脚本并注册到应用程序
type HookScriptLoader struct {
    app core.App
    vm  *goja.Runtime
}

// NewHookScriptLoader 创建一个新的钩子脚本加载器
func NewHookScriptLoader(app core.App) *HookScriptLoader {
    return &HookScriptLoader{
        app: app,
        vm:  goja.New(),
    }
}

// LoadAndRegisterScripts 从数据库加载所有启用的钩子脚本并注册到应用程序
func (l *HookScriptLoader) LoadAndRegisterScripts() error {
    // 查询所有启用的脚本
    scripts := []*HookScript{}
    err := l.app.DB().Select("*").From("pb_hooks_scripts").Where(dbx.HashExp{"enabled": true}).OrderBy("order").All(&scripts)
    if err != nil {
        return err
    }
    
    // 注册脚本到相应的钩子
    for _, script := range scripts {
        if err := l.registerScript(script); err != nil {
            l.app.Logger().Error("Failed to register hook script", "script", script.Name, "error", err)
        }
    }
    
    return nil
}

// registerScript 将单个脚本注册到相应的钩子
func (l *HookScriptLoader) registerScript(script *HookScript) error {
    // 根据脚本类型和事件类型，获取相应的钩子
    hook := l.getHookByTypeAndEvent(script.Type, script.Event)
    if hook == nil {
        return fmt.Errorf("unknown hook type %s or event %s", script.Type, script.Event)
    }
    
    // 编译脚本
    program, err := goja.Compile("", script.Code, true)
    if err != nil {
        return err
    }
    
    // 创建一个新的VM实例用于执行脚本
    vm := goja.New()
    
    // 注册钩子处理函数
    hook.BindFunc(func(e interface{}) error {
        // 设置VM上下文
        vm.Set("$app", l.app)
        vm.Set("$event", e)
        
        // 执行脚本
        _, err := vm.RunProgram(program)
        if err != nil {
            l.app.Logger().Error("Hook script execution error", "script", script.Name, "error", err)
            return err
        }
        
        return nil
    })
    
    return nil
}

// getHookByTypeAndEvent 根据类型和事件获取相应的钩子
func (l *HookScriptLoader) getHookByTypeAndEvent(typ, event string) interface{} {
    // 根据类型和事件返回相应的钩子
    // 这里需要根据PocketBase的钩子系统实现具体的映射逻辑
    // ...
}
```

### 3.3 API实现

```go
// 注册钩子脚本管理API
func RegisterHookScriptsApi(app core.App, router *echo.Group) {
    api := &HookScriptsApi{
        app: app,
    }
    
    subGroup := router.Group("/hooks")
    subGroup.GET("", api.list)
    subGroup.POST("", api.create, apis.RequireAdminAuth())
    subGroup.GET("/:id", api.view)
    subGroup.PATCH("/:id", api.update, apis.RequireAdminAuth())
    subGroup.DELETE("/:id", api.delete, apis.RequireAdminAuth())
    subGroup.POST("/:id/toggle", api.toggle, apis.RequireAdminAuth())
    subGroup.POST("/:id/test", api.test, apis.RequireAdminAuth())
    
    // 版本管理API
    versionsGroup := subGroup.Group("/:id/versions")
    versionsGroup.GET("", api.listVersions)
    versionsGroup.POST("", api.createVersion, apis.RequireAdminAuth())
    versionsGroup.GET("/:versionId", api.viewVersion)
    versionsGroup.POST("/:versionId/restore", api.restoreVersion, apis.RequireAdminAuth())
}

// HookScriptsApi 处理钩子脚本相关的API请求
type HookScriptsApi struct {
    app core.App
}

// list 返回所有钩子脚本
func (api *HookScriptsApi) list(c echo.Context) error {
    // 实现列表查询逻辑
    // ...
}

// create 创建新的钩子脚本
func (api *HookScriptsApi) create(c echo.Context) error {
    // 实现创建逻辑
    // ...
}

// view 查看单个钩子脚本
func (api *HookScriptsApi) view(c echo.Context) error {
    // 实现查看逻辑
    // ...
}

// update 更新钩子脚本
func (api *HookScriptsApi) update(c echo.Context) error {
    // 实现更新逻辑
    // ...
}

// delete 删除钩子脚本
func (api *HookScriptsApi) delete(c echo.Context) error {
    // 实现删除逻辑
    // ...
}

// toggle 启用/禁用钩子脚本
func (api *HookScriptsApi) toggle(c echo.Context) error {
    // 实现启用/禁用逻辑
    // ...
}

// test 测试钩子脚本
func (api *HookScriptsApi) test(c echo.Context) error {
    // 实现测试逻辑
    // ...
}

// 版本管理相关方法
// ...
```

## 4. 前端实现

### 4.1 路由配置

```javascript
// 在routes.js中添加钩子脚本管理相关路由
export default [
    // 现有路由...
    
    // 钩子脚本管理路由
    {
        path: "/hooks",
        component: () => import("@/components/hooks/HooksList"),
        auth: true,
        admin: true,
    },
    {
        path: "/hooks/create",
        component: () => import("@/components/hooks/HookUpsertPanel"),
        auth: true,
        admin: true,
    },
    {
        path: "/hooks/:id",
        component: () => import("@/components/hooks/HookUpsertPanel"),
        auth: true,
        admin: true,
    },
    {
        path: "/hooks/:id/versions",
        component: () => import("@/components/hooks/HookVersionsList"),
        auth: true,
        admin: true,
    },
];
```

### 4.2 脚本列表组件

```svelte
<!-- HooksList.svelte -->
<script>
    import { onMount } from "svelte";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import { confirm } from "@/components/confirmation";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import Loader from "@/components/base/Loader.svelte";
    import HookUpsertPanel from "@/components/hooks/HookUpsertPanel.svelte";
    
    let hooks = [];
    let isLoading = false;
    let upsertPanel;
    
    onMount(() => {
        loadHooks();
    });
    
    async function loadHooks() {
        isLoading = true;
        
        try {
            hooks = await ApiClient.hooks.getFullList();
        } catch (err) {
            addErrorToast("Failed to load hooks: " + CommonHelper.formatError(err));
        }
        
        isLoading = false;
    }
    
    function createHook() {
        upsertPanel.show({});
    }
    
    function editHook(hook) {
        upsertPanel.show(hook);
    }
    
    async function toggleHook(hook) {
        try {
            await ApiClient.hooks.toggle(hook.id);
            hook.enabled = !hook.enabled;
            addSuccessToast(`Hook ${hook.enabled ? "enabled" : "disabled"} successfully.`);
        } catch (err) {
            addErrorToast("Failed to toggle hook: " + CommonHelper.formatError(err));
        }
    }
    
    async function deleteHook(hook) {
        if (!await confirm("Do you really want to delete this hook script?")) {
            return;
        }
        
        try {
            await ApiClient.hooks.delete(hook.id);
            hooks = hooks.filter(h => h.id !== hook.id);
            addSuccessToast("Hook deleted successfully.");
        } catch (err) {
            addErrorToast("Failed to delete hook: " + CommonHelper.formatError(err));
        }
    }
    
    function onHookSaved(hook) {
        const index = hooks.findIndex(h => h.id === hook.id);
        
        if (index >= 0) {
            hooks[index] = hook;
            hooks = [...hooks]; // trigger reactivity
        } else {
            hooks = [...hooks, hook];
        }
    }
</script>

<div class="page-wrapper">
    <header class="page-header">
        <h1>Hook Scripts</h1>
        <button class="btn btn-primary" on:click={createHook}>
            <i class="ri-add-line"></i>
            <span class="txt">New hook script</span>
        </button>
    </header>
    
    {#if isLoading}
        <Loader />
    {:else if hooks.length === 0}
        <div class="block txt-center">
            <h2>No hook scripts found</h2>
            <button class="btn btn-primary" on:click={createHook}>
                <i class="ri-add-line"></i>
                <span class="txt">Create your first hook script</span>
            </button>
        </div>
    {:else}
        <div class="table-wrapper">
            <table class="table">
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Type</th>
                        <th>Event</th>
                        <th>Collection</th>
                        <th>Status</th>
                        <th>Order</th>
                        <th>Updated</th>
                        <th class="actions">Actions</th>
                    </tr>
                </thead>
                <tbody>
                    {#each hooks as hook (hook.id)}
                        <tr>
                            <td>{hook.name}</td>
                            <td>{hook.type}</td>
                            <td>{hook.event}</td>
                            <td>{hook.collection_id || "-"}</td>
                            <td>
                                <span class="label {hook.enabled ? 'success' : 'danger'}">
                                    {hook.enabled ? "Enabled" : "Disabled"}
                                </span>
                            </td>
                            <td>{hook.order}</td>
                            <td>{CommonHelper.formatDate(hook.updated)}</td>
                            <td class="actions">
                                <button class="btn btn-sm btn-outline" on:click={() => toggleHook(hook)}>
                                    <i class="ri-toggle-line"></i>
                                </button>
                                <button class="btn btn-sm btn-outline" on:click={() => editHook(hook)}>
                                    <i class="ri-pencil-line"></i>
                                </button>
                                <button class="btn btn-sm btn-outline btn-danger" on:click={() => deleteHook(hook)}>
                                    <i class="ri-delete-bin-line"></i>
                                </button>
                            </td>
                        </tr>
                    {/each}
                </tbody>
            </table>
        </div>
    {/if}
    
    <HookUpsertPanel bind:this={upsertPanel} on:save={e => onHookSaved(e.detail)} />
</div>
```

### 4.3 脚本编辑组件

```svelte
<!-- HookUpsertPanel.svelte -->
<script>
    import { createEventDispatcher } from "svelte";
    import { addSuccessToast, addErrorToast } from "@/stores/toasts";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import { collections } from "@/stores/collections";
    import Panel from "@/components/base/Panel.svelte";
    import Field from "@/components/base/Field.svelte";
    import CodeEditor from "@/components/base/CodeEditor.svelte";
    
    export let hook = {};
    
    let panel;
    let isLoading = false;
    let formData = {
        name: "",
        description: "",
        code: "",
        type: "record",
        event: "create",
        collection_id: null,
        enabled: true,
        order: 0,
    };
    
    const dispatch = createEventDispatcher();
    
    const hookTypes = [
        { value: "record", text: "Record" },
        { value: "collection", text: "Collection" },
        { value: "api", text: "API" },
        { value: "system", text: "System" },
    ];
    
    const hookEvents = {
        record: [
            { value: "create", text: "Create" },
            { value: "update", text: "Update" },
            { value: "delete", text: "Delete" },
        ],
        collection: [
            { value: "create", text: "Create" },
            { value: "update", text: "Update" },
            { value: "delete", text: "Delete" },
        ],
        api: [
            { value: "request", text: "Request" },
            { value: "response", text: "Response" },
        ],
        system: [
            { value: "bootstrap", text: "Bootstrap" },
            { value: "terminate", text: "Terminate" },
        ],
    };
    
    $: availableEvents = hookEvents[formData.type] || [];
    $: needsCollection = ["record", "collection"].includes(formData.type);
    
    export function show(data = {}) {
        hook = data;
        
        // Reset form
        formData = {
            name: hook.name || "",
            description: hook.description || "",
            code: hook.code || getDefaultCode(),
            type: hook.type || "record",
            event: hook.event || "create",
            collection_id: hook.collection_id || null,
            enabled: hook.enabled !== undefined ? hook.enabled : true,
            order: hook.order || 0,
        };
        
        panel.show();
    }
    
    function getDefaultCode() {
        return `// Hook script
// Available variables: $app, $event

// Make sure to call $event.next() to continue the event chain
$event.next();
`;
    }
    
    async function save() {
        if (!formData.name) {
            addErrorToast("Name is required");
            return;
        }
        
        if (!formData.code) {
            addErrorToast("Code is required");
            return;
        }
        
        isLoading = true;
        
        try {
            let result;
            
            if (hook.id) {
                // Update
                result = await ApiClient.hooks.update(hook.id, formData);
                addSuccessToast("Hook updated successfully");
            } else {
                // Create
                result = await ApiClient.hooks.create(formData);
                addSuccessToast("Hook created successfully");
            }
            
            dispatch("save", result);
            panel.hide();
        } catch (err) {
            addErrorToast("Failed to save hook: " + CommonHelper.formatError(err));
        }
        
        isLoading = false;
    }
    
    async function testHook() {
        if (!hook.id) {
            addErrorToast("Please save the hook first before testing");
            return;
        }
        
        isLoading = true;
        
        try {
            await ApiClient.hooks.test(hook.id);
            addSuccessToast("Hook test executed successfully");
        } catch (err) {
            addErrorToast("Hook test failed: " + CommonHelper.formatError(err));
        }
        
        isLoading = false;
    }
</script>

<Panel bind:this={panel} title={hook.id ? "Edit Hook Script" : "Create Hook Script"} on:hide>
    <div class="grid">
        <div class="col-6">
            <Field class="form-field required" name="name" let:uniqueId>
                <label for={uniqueId}>Name</label>
                <input
                    type="text"
                    id={uniqueId}
                    bind:value={formData.name}
                    required
                />
            </Field>
        </div>
        
        <div class="col-6">
            <Field class="form-field" name="description" let:uniqueId>
                <label for={uniqueId}>Description</label>
                <input
                    type="text"
                    id={uniqueId}
                    bind:value={formData.description}
                />
            </Field>
        </div>
        
        <div class="col-4">
            <Field class="form-field required" name="type" let:uniqueId>
                <label for={uniqueId}>Type</label>
                <select id={uniqueId} bind:value={formData.type}>
                    {#each hookTypes as type}
                        <option value={type.value}>{type.text}</option>
                    {/each}
                </select>
            </Field>
        </div>
        
        <div class="col-4">
            <Field class="form-field required" name="event" let:uniqueId>
                <label for={uniqueId}>Event</label>
                <select id={uniqueId} bind:value={formData.event}>
                    {#each availableEvents as event}
                        <option value={event.value}>{event.text}</option>
                    {/each}
                </select>
            </Field>
        </div>
        
        <div class="col-4">
            {#if needsCollection}
                <Field class="form-field" name="collection_id" let:uniqueId>
                    <label for={uniqueId}>Collection</label>
                    <select id={uniqueId} bind:value={formData.collection_id}>
                        <option value="">- Any -</option>
                        {#each $collections as collection}
                            <option value={collection.id}>{collection.name}</option>
                        {/each}
                    </select>
                </Field>
            {:else}
                <Field class="form-field" name="order" let:uniqueId>
                    <label for={uniqueId}>Execution Order</label>
                    <input
                        type="number"
                        id={uniqueId}
                        bind:value={formData.order}
                        min="0"
                    />
                </Field>
            {/if}
        </div>
        
        <div class="col-12">
            <Field class="form-field required" name="code" let:uniqueId>
                <label for={uniqueId}>Code</label>
                <CodeEditor
                    id={uniqueId}
                    bind:value={formData.code}
                    language="javascript"
                    height="400px"
                />
            </Field>
        </div>
        
        <div class="col-12">
            <Field class="form-field form-field-toggle" name="enabled" let:uniqueId>
                <input
                    type="checkbox"
                    id={uniqueId}
                    bind:checked={formData.enabled}
                />
                <label for={uniqueId}>Enabled</label>
            </Field>
        </div>
    </div>
    
    <svelte:fragment slot="footer">
        <button type="button" class="btn btn-secondary" on:click={() => panel.hide()} disabled={isLoading}>
            Cancel
        </button>
        
        {#if hook.id}
            <button type="button" class="btn btn-outline" on:click={testHook} disabled={isLoading}>
                <i class="ri-play-line"></i>
                <span class="txt">Test</span>
            </button>
        {/if}
        
        <button type="button" class="btn btn-primary" on:click={save} disabled={isLoading}>
            <i class="ri-save-line"></i>
            <span class="txt">Save</span>
        </button>
    </svelte:fragment>
</Panel>
```

### 4.4 API客户端扩展

```javascript
// 在ApiClient.js中添加钩子脚本相关API

// 现有代码...

ApiClient.hooks = {
    /**
     * @returns {Promise<Array<Object>>}
     */
    getFullList() {
        return ApiClient.send({
            path: "/api/hooks",
            method: "GET",
        });
    },

    /**
     * @param {String} id
     * @returns {Promise<Object>}
     */
    getOne(id) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}`,
            method: "GET",
        });
    },

    /**
     * @param {Object} bodyData
     * @returns {Promise<Object>}
     */
    create(bodyData) {
        return ApiClient.send({
            path: "/api/hooks",
            method: "POST",
            body: bodyData,
        });
    },

    /**
     * @param {String} id
     * @param {Object} bodyData
     * @returns {Promise<Object>}
     */
    update(id, bodyData) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}`,
            method: "PATCH",
            body: bodyData,
        });
    },

    /**
     * @param {String} id
     * @returns {Promise<Boolean>}
     */
    delete(id) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}`,
            method: "DELETE",
        });
    },

    /**
     * @param {String} id
     * @returns {Promise<Object>}
     */
    toggle(id) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/toggle`,
            method: "POST",
        });
    },

    /**
     * @param {String} id
     * @returns {Promise<Object>}
     */
    test(id) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/test`,
            method: "POST",
        });
    },

    /**
     * @param {String} id
     * @returns {Promise<Array<Object>>}
     */
    getVersions(id) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/versions`,
            method: "GET",
        });
    },

    /**
     * @param {String} id
     * @param {Object} bodyData
     * @returns {Promise<Object>}
     */
    createVersion(id, bodyData) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/versions`,
            method: "POST",
            body: bodyData,
        });
    },

    /**
     * @param {String} id
     * @param {String} versionId
     * @returns {Promise<Object>}
     */
    getVersion(id, versionId) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/versions/${encodeURIComponent(versionId)}`,
            method: "GET",
        });
    },

    /**
     * @param {String} id
     * @param {String} versionId
     * @returns {Promise<Object>}
     */
    restoreVersion(id, versionId) {
        return ApiClient.send({
            path: `/api/hooks/${encodeURIComponent(id)}/versions/${encodeURIComponent(versionId)}/restore`,
            method: "POST",
        });
    },
};
```

## 5. 集成与启动

### 5.1 插件注册

```go
// 在main.go中注册钩子脚本管理插件
func main() {
    app := pocketbase.New()
    
    // 注册钩子脚本管理插件
    app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
        // 创建钩子脚本加载器
        loader := NewHookScriptLoader(app)
        
        // 加载并注册钩子脚本
        if err := loader.LoadAndRegisterScripts(); err != nil {
            app.Logger().Error("Failed to load hook scripts", "error", err)
        }
        
        // 注册API
        RegisterHookScriptsApi(app, e.Router.Group("/api"))
        
        return nil
    })
    
    // 启动应用
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 5.2 数据库迁移

```go
// 创建钩子脚本相关表的迁移
func init() {
    core.AppMigrations.Register(func(db dbx.Builder) error {
        // 创建钩子脚本表
        _, err := db.NewQuery("CREATE TABLE IF NOT EXISTS `pb_hooks_scripts` (
            `id` TEXT PRIMARY KEY,
            `name` TEXT NOT NULL,
            `description` TEXT,
            `code` TEXT NOT NULL,
            `type` TEXT NOT NULL,
            `event` TEXT NOT NULL,
            `collection_id` TEXT,
            `enabled` BOOLEAN NOT NULL DEFAULT TRUE,
            `order` INTEGER NOT NULL DEFAULT 0,
            `created` TEXT NOT NULL,
            `updated` TEXT NOT NULL,
            `created_by` TEXT,
            `updated_by` TEXT,
            FOREIGN KEY (`collection_id`) REFERENCES `_collections` (`id`) ON DELETE SET NULL,
            FOREIGN KEY (`created_by`) REFERENCES `_superusers` (`id`) ON DELETE SET NULL,
            FOREIGN KEY (`updated_by`) REFERENCES `_superusers` (`id`) ON DELETE SET NULL
        )").Execute()
        if err != nil {
            return err
        }
        
        // 创建钩子脚本版本表
        _, err = db.NewQuery("CREATE TABLE IF NOT EXISTS `pb_hooks_versions` (
            `id` TEXT PRIMARY KEY,
            `script_id` TEXT NOT NULL,
            `code` TEXT NOT NULL,
            `comment` TEXT,
            `created` TEXT NOT NULL,
            `created_by` TEXT,
            FOREIGN KEY (`script_id`) REFERENCES `pb_hooks_scripts` (`id`) ON DELETE CASCADE,
            FOREIGN KEY (`created_by`) REFERENCES `_superusers` (`id`) ON DELETE SET NULL
        )").Execute()
        if err != nil {
            return err
        }
        
        // 创建索引
        _, err = db.NewQuery("CREATE INDEX IF NOT EXISTS `idx_hooks_scripts_type_event` ON `pb_hooks_scripts` (`type`, `event`)").Execute()
        if err != nil {
            return err
        }
        
        _, err = db.NewQuery("CREATE INDEX IF NOT EXISTS `idx_hooks_scripts_collection` ON `pb_hooks_scripts` (`collection_id`)").Execute()
        if err != nil {
            return err
        }
        
        _, err = db.NewQuery("CREATE INDEX IF NOT EXISTS `idx_hooks_versions_script` ON `pb_hooks_versions` (`script_id`)").Execute()
        if err != nil {
            return err
        }
        
        return nil
    }, nil)
}
```

## 6. 测试计划

### 6.1 单元测试

```go
// 钩子脚本加载器测试
func TestHookScriptLoader(t *testing.T) {
    app, _ := tests.NewTestApp()
    defer app.Cleanup()
    
    // 创建测试脚本
    script := &HookScript{
        Id:      "test123",
        Name:    "Test Hook",
        Code:    "console.log('Test hook executed');",
        Type:    "record",
        Event:   "create",
        Enabled: true,
        Order:   0,
        Created: types.NowDateTime(),
        Updated: types.NowDateTime(),
    }
    
    // 保存测试脚本到数据库
    err := app.DB().Save("pb_hooks_scripts", script)
    if err != nil {
        t.Fatal(err)
    }
    
    // 创建加载器并加载脚本
    loader := NewHookScriptLoader(app)
    err = loader.LoadAndRegisterScripts()
    if err != nil {
        t.Fatal(err)
    }
    
    // 验证钩子是否被正确注册
    // ...
}
```

### 6.2 集成测试

```go
// API测试
func TestHookScriptsApi(t *testing.T) {
    app, e := tests.NewTestApp()
    defer app.Cleanup()
    
    // 注册API
    RegisterHookScriptsApi(app, e.Group("/api"))
    
    // 测试创建脚本
    res, err := tests.ApiRequest(e, "POST", "/api/hooks", map[string]any{
        "name":    "Test Hook",
        "code":    "console.log('Test');",
        "type":    "record",
        "event":   "create",
        "enabled": true,
        "order":   0,
    })
    if err != nil {
        t.Fatal(err)
    }
    if res.StatusCode != 200 {
        t.Fatalf("Expected status code 200, got %d", res.StatusCode)
    }
    
    // 解析响应
    result := map[string]any{}
    if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
        t.Fatal(err)
    }
    
    // 验证响应
    id, ok := result["id"].(string)
    if !ok || id == "" {
        t.Fatal("Expected non-empty id in response")
    }
    
    // 测试获取脚本
    // ...
    
    // 测试更新脚本
    // ...
    
    // 测试删除脚本
    // ...
}
```

### 6.3 端到端测试

```javascript
// 使用Playwright或Cypress进行前端测试
// ...
```

## 7. 部署和迁移策略

### 7.1 数据迁移

对于已有的钩子脚本文件，我们需要提供一个迁移工具，将文件系统中的脚本导入到数据库中：

```go
// 导入现有钩子脚本文件到数据库
func ImportHookScriptsFromFiles(app core.App, hooksDir string) error {
    // 获取所有钩子脚本文件
    files, err := os.ReadDir(hooksDir)
    if err != nil {
        return err
    }
    
    for _, file := range files {
        if file.IsDir() || !strings.HasSuffix(file.Name(), ".js") {
            continue
        }
        
        // 读取文件内容
        content, err := os.ReadFile(filepath.Join(hooksDir, file.Name()))
        if err != nil {
            app.Logger().Error("Failed to read hook script file", "file", file.Name(), "error", err)
            continue
        }
        
        // 解析文件名以获取类型和事件
        name := strings.TrimSuffix(file.Name(), ".js")
        parts := strings.Split(name, "_")
        
        scriptType := "system"
        scriptEvent := "bootstrap"
        var collectionId *string
        
        if len(parts) > 1 {
            scriptType = parts[0]
            scriptEvent = parts[1]
            
            // 如果文件名包含集合ID，则解析它
            if len(parts) > 2 && (scriptType == "record" || scriptType == "collection") {
                // 查找匹配的集合
                collection, _ := app.Dao().FindCollectionByNameOrId(parts[2])
                if collection != nil {
                    id := collection.Id
                    collectionId = &id
                }
            }
        }
        
        // 创建脚本记录
        script := &HookScript{
            Id:           security.RandomString(15),
            Name:         name,
            Description:  "Imported from file: " + file.Name(),
            Code:         string(content),
            Type:         scriptType,
            Event:        scriptEvent,
            CollectionId: collectionId,
            Enabled:      true,
            Order:        0,
            Created:      types.NowDateTime(),
            Updated:      types.NowDateTime(),
        }
        
        // 保存到数据库
        if err := app.DB().Save("pb_hooks_scripts", script); err != nil {
            app.Logger().Error("Failed to save imported hook script", "file", file.Name(), "error", err)
            continue
        }
        
        app.Logger().Info("Imported hook script", "file", file.Name(), "id", script.Id)
    }
    
    return nil
}
```

### 7.2 兼容性策略

为了保持向后兼容性，我们将同时支持两种钩子加载方式：

```go
// 在应用启动时加载钩子脚本
app.OnBootstrap().Add(func(e *core.BootstrapEvent) error {
    // 1. 从文件系统加载钩子脚本（现有方式）
    if app.Settings().HooksFromFiles {
        // 现有的文件加载逻辑
        // ...
    }
    
    // 2. 从数据库加载钩子脚本（新方式）
    if app.Settings().HooksFromDatabase {
        loader := NewHookScriptLoader(app)
        if err := loader.LoadAndRegisterScripts(); err != nil {
            app.Logger().Error("Failed to load hook scripts from database", "error", err)
        }
    }
    
    return e.Next()
})
```

## 8. 性能和安全考虑

### 8.1 性能优化

- 使用连接池管理Goja运行时实例
- 缓存编译后的脚本
- 限制单个脚本的执行时间
- 监控脚本执行性能

### 8.2 安全措施

- 沙箱隔离脚本执行环境
- 限制脚本访问的API和资源
- 记录脚本执行日志
- 实现脚本执行超时机制
- 只允许超级用户管理钩子脚本

## 9. 结论

通过实现这个钩子脚本管理系统，我们可以大大提高PocketBase钩子系统的可用性和维护性。该系统将允许开发者在Web界面上直接编辑和管理钩子脚本，无需直接访问服务器文件系统，同时也为多环境部署提供更好的支持。