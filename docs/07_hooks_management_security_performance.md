# PocketBase 钩子脚本管理系统安全与性能优化

## 1. 概述

本文档详细说明了 PocketBase 钩子脚本管理系统的安全措施和性能优化策略。由于钩子脚本具有执行自定义代码的能力，它们可能成为安全风险的来源，同时也可能影响系统的整体性能。因此，实施适当的安全措施和性能优化策略至关重要。

## 2. 安全考虑

### 2.1 脚本执行安全

#### 2.1.1 JavaScript 沙箱隔离

PocketBase 使用 Goja 作为 JavaScript 运行时，它提供了一定程度的沙箱隔离，但仍需加强安全措施：

```go
func createSandboxedRuntime() *goja.Runtime {
    vm := goja.New()
    vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))
    
    // 限制可用的全局对象和函数
    restrictedGlobals := map[string]interface{}{
        "console":  console.NewConsole(vm),
        "setTimeout": func(call goja.FunctionCall) goja.Value {
            // 实现安全的 setTimeout
            // ...
            return goja.Undefined()
        },
        // 其他安全的全局函数...
    }
    
    // 注入受限的全局对象
    for name, obj := range restrictedGlobals {
        vm.Set(name, obj)
    }
    
    // 禁用不安全的全局对象和函数
    unsafeGlobals := []string{"Proxy", "WebAssembly", "eval", "Function"}
    for _, name := range unsafeGlobals {
        vm.Set(name, goja.Undefined())
    }
    
    return vm
}
```

#### 2.1.2 执行超时限制

为防止无限循环或资源耗尽攻击，应对脚本执行时间进行限制：

```go
func executeScriptWithTimeout(vm *goja.Runtime, script string, timeout time.Duration) (interface{}, error) {
    resultCh := make(chan interface{}, 1)
    errCh := make(chan error, 1)
    
    go func() {
        defer func() {
            if r := recover(); r != nil {
                errCh <- fmt.Errorf("script execution panicked: %v", r)
            }
        }())
        
        result, err := vm.RunString(script)
        if err != nil {
            errCh <- err
            return
        }
        
        resultCh <- result.Export()
    }()
    
    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return nil, err
    case <-time.After(timeout):
        // 中断执行
        vm.Interrupt("execution timeout")
        return nil, fmt.Errorf("script execution timed out after %v", timeout)
    }
}
```

#### 2.1.3 内存使用限制

限制脚本可以使用的内存量，防止内存耗尽攻击：

```go
func executeScriptWithMemoryLimit(vm *goja.Runtime, script string, memoryLimit int64) (interface{}, error) {
    // 设置内存限制
    runtime.GC()
    var memStats runtime.MemStats
    runtime.ReadMemStats(&memStats)
    startAlloc := memStats.Alloc
    
    // 监控内存使用的 goroutine
    stopCh := make(chan struct{})
    errCh := make(chan error, 1)
    
    go func() {
        ticker := time.NewTicker(100 * time.Millisecond)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                runtime.ReadMemStats(&memStats)
                if memStats.Alloc - startAlloc > uint64(memoryLimit) {
                    vm.Interrupt("memory limit exceeded")
                    errCh <- fmt.Errorf("script exceeded memory limit of %d bytes", memoryLimit)
                    return
                }
            case <-stopCh:
                return
            }
        }
    }()
    
    result, err := vm.RunString(script)
    close(stopCh)
    
    select {
    case memErr := <-errCh:
        return nil, memErr
    default:
        if err != nil {
            return nil, err
        }
        return result.Export(), nil
    }
}
```

#### 2.1.4 API 访问限制

限制脚本可以访问的 API 和功能，只暴露必要的功能：

```go
func injectRestrictedAPI(vm *goja.Runtime, app core.App) {
    // 创建受限的 $app 对象
    restrictedApp := map[string]interface{}{
        "dao": createRestrictedDAO(app.Dao()),
        "settings": createRestrictedSettings(app.Settings()),
        // 其他受限 API...
    }
    
    vm.Set("$app", restrictedApp)
}

func createRestrictedDAO(dao *daos.Dao) map[string]interface{} {
    return map[string]interface{}{
        "findRecordById": func(collection, id string) (map[string]interface{}, error) {
            // 实现安全的记录查询
            // ...
        },
        // 其他安全的 DAO 方法...
    }
}
```

#### 2.1.5 文件系统访问限制

限制或禁止脚本访问文件系统，或仅允许在特定目录中进行操作：

```go
func createRestrictedFS(vm *goja.Runtime) {
    // 创建受限的文件系统对象
    restrictedFS := map[string]interface{}{
        "readFile": func(path string) (string, error) {
            // 验证路径是否在允许的目录中
            if !isPathAllowed(path) {
                return "", fmt.Errorf("access denied to path: %s", path)
            }
            
            // 安全地读取文件
            // ...
        },
        // 其他受限的文件系统方法...
    }
    
    vm.Set("fs", restrictedFS)
}

func isPathAllowed(path string) bool {
    // 检查路径是否在允许的目录中
    allowedDirs := []string{"/tmp/scripts", "/var/data/public"}
    path = filepath.Clean(path)
    
    for _, dir := range allowedDirs {
        if strings.HasPrefix(path, dir) {
            return true
        }
    }
    
    return false
}
```

### 2.2 输入验证和清理

#### 2.2.1 脚本代码验证

在保存脚本之前验证其语法和潜在的恶意代码：

```go
func validateScript(code string) error {
    // 检查脚本大小
    if len(code) > maxScriptSize {
        return fmt.Errorf("script exceeds maximum size of %d bytes", maxScriptSize)
    }
    
    // 检查语法错误
    vm := goja.New()
    _, err := vm.Compile("", code)
    if err != nil {
        return fmt.Errorf("script contains syntax errors: %v", err)
    }
    
    // 检查潜在的恶意代码模式
    maliciousPatterns := []string{
        "process\.exit",
        "require\(\s*['\"]child_process['\"]\s*\)",
        "require\(\s*['\"]fs['\"]\s*\)",
        // 其他恶意模式...
    }
    
    for _, pattern := range maliciousPatterns {
        if regexp.MustCompile(pattern).MatchString(code) {
            return fmt.Errorf("script contains potentially malicious code pattern: %s", pattern)
        }
    }
    
    return nil
}
```

#### 2.2.2 API 输入验证

验证所有 API 输入，防止注入攻击和其他安全问题：

```go
func validateHookScriptInput(hook *models.HookScript) error {
    // 验证名称
    if hook.Name == "" {
        return errors.New("name is required")
    }
    if len(hook.Name) > 100 {
        return errors.New("name cannot exceed 100 characters")
    }
    
    // 验证描述
    if len(hook.Description) > 500 {
        return errors.New("description cannot exceed 500 characters")
    }
    
    // 验证类型
    validTypes := map[string]bool{"record": true, "collection": true, "api": true, "system": true}
    if !validTypes[hook.Type] {
        return fmt.Errorf("invalid type: %s", hook.Type)
    }
    
    // 验证事件
    if err := validateEvent(hook.Type, hook.Event); err != nil {
        return err
    }
    
    // 验证集合
    if (hook.Type == "record" || hook.Type == "collection") && hook.Collection == "" {
        return errors.New("collection is required for record and collection hooks")
    }
    
    // 验证脚本代码
    if err := validateScript(hook.Code); err != nil {
        return err
    }
    
    // 验证执行顺序
    if hook.Order < 0 {
        return errors.New("order must be a non-negative integer")
    }
    
    return nil
}

func validateEvent(hookType, event string) error {
    validEvents := map[string]map[string]bool{
        "record": {"create": true, "create.after": true, "update": true, "update.after": true, "delete": true, "delete.after": true},
        "collection": {"create": true, "create.after": true, "update": true, "update.after": true, "delete": true, "delete.after": true},
        "api": {"request": true, "response": true},
        "system": {"bootstrap": true, "serve": true, "terminate": true},
    }
    
    if events, ok := validEvents[hookType]; ok {
        if !events[event] {
            return fmt.Errorf("invalid event '%s' for hook type '%s'", event, hookType)
        }
    } else {
        return fmt.Errorf("unknown hook type: %s", hookType)
    }
    
    return nil
}
```

### 2.3 认证和授权

#### 2.3.1 管理员认证

确保只有经过认证的管理员可以访问钩子脚本管理功能：

```go
func adminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
        if admin == nil {
            return apis.NewForbiddenError("Only admins can access this resource", nil)
        }
        
        return next(c)
    }
}
```

#### 2.3.2 细粒度权限控制

实现细粒度的权限控制，限制哪些管理员可以创建、编辑或删除钩子脚本：

```go
func hookScriptPermissionMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        admin, _ := c.Get(apis.ContextAdminKey).(*models.Admin)
        if admin == nil {
            return apis.NewForbiddenError("Only admins can access this resource", nil)
        }
        
        // 检查管理员是否有权限管理钩子脚本
        if !admin.HasPermission("hooks.manage") {
            return apis.NewForbiddenError("You don't have permission to manage hook scripts", nil)
        }
        
        // 对于特定操作的额外权限检查
        action := c.Get("action").(string)
        switch action {
        case "create":
            if !admin.HasPermission("hooks.create") {
                return apis.NewForbiddenError("You don't have permission to create hook scripts", nil)
            }
        case "update":
            if !admin.HasPermission("hooks.update") {
                return apis.NewForbiddenError("You don't have permission to update hook scripts", nil)
            }
        case "delete":
            if !admin.HasPermission("hooks.delete") {
                return apis.NewForbiddenError("You don't have permission to delete hook scripts", nil)
            }
        }
        
        return next(c)
    }
}
```

### 2.4 审计和日志记录

#### 2.4.1 操作审计

记录所有对钩子脚本的操作，包括创建、更新、删除和执行：

```go
func auditHookScriptAction(app core.App, admin *models.Admin, action string, hookId string, details map[string]interface{}) {
    audit := &models.Audit{
        ID:        nanoid.New(),
        AdminID:   admin.ID,
        Action:    action,
        Resource:  "hook_script",
        ResourceID: hookId,
        Details:   details,
        Created:   time.Now().UTC(),
    }
    
    // 异步保存审计记录
    go func() {
        if err := app.Dao().SaveAudit(audit); err != nil {
            log.Printf("Failed to save audit log: %v", err)
        }
    }()
}
```

#### 2.4.2 执行日志

记录钩子脚本的执行情况，包括执行时间、错误和输出：

```go
func logHookExecution(app core.App, hook *models.HookScript, startTime time.Time, err error, output string) {
    executionTime := time.Since(startTime).Milliseconds()
    status := "success"
    errorMsg := ""
    
    if err != nil {
        status = "error"
        errorMsg = err.Error()
    }
    
    log := &models.HookLog{
        ID:            nanoid.New(),
        HookID:        hook.ID,
        Status:        status,
        ExecutionTime: int(executionTime),
        Error:         errorMsg,
        Context:       output,
        Created:       time.Now().UTC(),
    }
    
    // 异步保存日志
    go func() {
        if err := app.Dao().SaveHookLog(log); err != nil {
            log.Printf("Failed to save hook execution log: %v", err)
        }
    }()
}
```

### 2.5 版本控制和回滚

#### 2.5.1 自动版本控制

每次更新脚本时自动创建新版本，以便在出现问题时回滚：

```go
func saveHookScriptWithVersion(dao HookScriptDAO, hook *models.HookScript, admin *models.Admin, comment string) error {
    // 开始事务
    tx, err := dao.BeginTx()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // 保存钩子脚本
    if err := tx.SaveHookScript(hook); err != nil {
        return err
    }
    
    // 获取最新版本号
    latestVersion, err := tx.GetLatestVersionNumber(hook.ID)
    if err != nil && !errors.Is(err, sql.ErrNoRows) {
        return err
    }
    
    // 创建新版本
    version := &models.HookScriptVersion{
        ID:        nanoid.New(),
        HookID:    hook.ID,
        Version:   latestVersion + 1,
        Code:      hook.Code,
        Comment:   comment,
        Created:   time.Now().UTC(),
        CreatedBy: admin.ID,
    }
    
    if err := tx.SaveHookScriptVersion(version); err != nil {
        return err
    }
    
    // 提交事务
    return tx.Commit()
}
```

#### 2.5.2 回滚机制

提供回滚到之前版本的功能：

```go
func rollbackToVersion(dao HookScriptDAO, hookId string, versionId string, admin *models.Admin, comment string) error {
    // 获取指定版本
    version, err := dao.GetHookScriptVersionById(versionId)
    if err != nil {
        return err
    }
    
    // 获取当前钩子脚本
    hook, err := dao.GetHookScriptById(hookId)
    if err != nil {
        return err
    }
    
    // 更新脚本代码
    hook.Code = version.Code
    hook.Updated = time.Now().UTC()
    
    // 保存更新并创建新版本
    return saveHookScriptWithVersion(dao, hook, admin, fmt.Sprintf("Rolled back to version %d: %s", version.Version, comment))
}
```

## 3. 性能优化

### 3.1 脚本执行优化

#### 3.1.1 JavaScript 运行时池

使用运行时池来减少创建新 JavaScript 运行时的开销：

```go
type RuntimePool struct {
    pool    chan *goja.Runtime
    factory func() *goja.Runtime
    size    int
}

func NewRuntimePool(size int, factory func() *goja.Runtime) *RuntimePool {
    pool := &RuntimePool{
        pool:    make(chan *goja.Runtime, size),
        factory: factory,
        size:    size,
    }
    
    // 预热池
    for i := 0; i < size; i++ {
        pool.pool <- factory()
    }
    
    return pool
}

func (p *RuntimePool) Get() *goja.Runtime {
    select {
    case vm := <-p.pool:
        return vm
    default:
        // 池已耗尽，创建新实例
        return p.factory()
    }
}

func (p *RuntimePool) Put(vm *goja.Runtime) {
    // 重置运行时状态
    vm.ClearInterrupt()
    vm.Set("console", console.NewConsole(vm))
    
    select {
    case p.pool <- vm:
        // 成功返回池
    default:
        // 池已满，丢弃
    }
}

func (p *RuntimePool) Execute(script string, timeout time.Duration) (interface{}, error) {
    vm := p.Get()
    defer p.Put(vm)
    
    return executeScriptWithTimeout(vm, script, timeout)
}
```

#### 3.1.2 脚本预编译

预编译脚本以提高执行效率：

```go
type CompiledScript struct {
    program *goja.Program
    code    string
}

func compileScript(code string) (*CompiledScript, error) {
    program, err := goja.Compile("", code, false)
    if err != nil {
        return nil, err
    }
    
    return &CompiledScript{
        program: program,
        code:    code,
    }, nil
}

func (cs *CompiledScript) Execute(vm *goja.Runtime, timeout time.Duration) (interface{}, error) {
    resultCh := make(chan interface{}, 1)
    errCh := make(chan error, 1)
    
    go func() {
        defer func() {
            if r := recover(); r != nil {
                errCh <- fmt.Errorf("script execution panicked: %v", r)
            }
        }()
        
        result, err := vm.RunProgram(cs.program)
        if err != nil {
            errCh <- err
            return
        }
        
        resultCh <- result.Export()
    }()
    
    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return nil, err
    case <-time.After(timeout):
        vm.Interrupt("execution timeout")
        return nil, fmt.Errorf("script execution timed out after %v", timeout)
    }
}
```

#### 3.1.3 脚本缓存

缓存编译后的脚本，避免重复编译：

```go
type ScriptCache struct {
    cache  map[string]*CompiledScript
    mutex  sync.RWMutex
    maxAge time.Duration
}

func NewScriptCache(maxAge time.Duration) *ScriptCache {
    return &ScriptCache{
        cache:  make(map[string]*CompiledScript),
        maxAge: maxAge,
    }
}

func (sc *ScriptCache) Get(id string, code string) (*CompiledScript, error) {
    sc.mutex.RLock()
    script, ok := sc.cache[id]
    sc.mutex.RUnlock()
    
    if ok && script.code == code {
        return script, nil
    }
    
    // 编译脚本
    compiledScript, err := compileScript(code)
    if err != nil {
        return nil, err
    }
    
    // 更新缓存
    sc.mutex.Lock()
    sc.cache[id] = compiledScript
    sc.mutex.Unlock()
    
    return compiledScript, nil
}

func (sc *ScriptCache) Invalidate(id string) {
    sc.mutex.Lock()
    delete(sc.cache, id)
    sc.mutex.Unlock()
}

func (sc *ScriptCache) Clear() {
    sc.mutex.Lock()
    sc.cache = make(map[string]*CompiledScript)
    sc.mutex.Unlock()
}
```

### 3.2 数据库优化

#### 3.2.1 索引优化

为常用查询添加适当的索引：

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

#### 3.2.2 查询优化

优化数据库查询，减少不必要的查询和数据传输：

```go
func getHookScriptsOptimized(dao HookScriptDAO, filter string, sort string) ([]*models.HookScript, error) {
    // 使用单个查询获取所有需要的数据
    query := "SELECT id, name, type, event, collection, order, enabled, updated FROM pb_hooks_scripts"
    
    if filter != "" {
        query += " WHERE " + filter
    }
    
    if sort != "" {
        query += " ORDER BY " + sort
    } else {
        query += " ORDER BY order ASC"
    }
    
    // 执行查询
    // ...
    
    // 只有在需要时才加载完整的脚本代码
    // ...
    
    return hooks, nil
}
```

#### 3.2.3 连接池管理

优化数据库连接池配置：

```go
func configureDBPool(db *sql.DB) {
    // 设置最大打开连接数
    db.SetMaxOpenConns(25)
    
    // 设置最大空闲连接数
    db.SetMaxIdleConns(5)
    
    // 设置连接最大生存时间
    db.SetConnMaxLifetime(5 * time.Minute)
    
    // 设置连接最大空闲时间
    db.SetConnMaxIdleTime(5 * time.Minute)
}
```

### 3.3 API 性能优化

#### 3.3.1 分页和限制

实现分页和结果限制，避免返回过多数据：

```go
func listHookScripts(c echo.Context) error {
    // 获取分页参数
    page, _ := strconv.Atoi(c.QueryParam("page"))
    if page < 1 {
        page = 1
    }
    
    perPage, _ := strconv.Atoi(c.QueryParam("perPage"))
    if perPage < 1 || perPage > 100 {
        perPage = 30 // 默认每页 30 条
    }
    
    // 获取筛选和排序参数
    filter := c.QueryParam("filter")
    sort := c.QueryParam("sort")
    
    // 获取数据
    app := c.Get("app").(core.App)
    hooks, total, err := app.Dao().GetHookScriptsList(page, perPage, filter, sort)
    if err != nil {
        return apis.NewBadRequestError("Failed to list hook scripts", err)
    }
    
    // 返回结果
    return c.JSON(http.StatusOK, map[string]interface{}{
        "page":       page,
        "perPage":    perPage,
        "totalItems": total,
        "totalPages": int(math.Ceil(float64(total) / float64(perPage))),
        "items":      hooks,
    })
}
```

#### 3.3.2 响应缓存

缓存不经常变化的 API 响应：

```go
func cacheMiddleware(duration time.Duration) echo.MiddlewareFunc {
    cache := make(map[string]cacheEntry)
    mutex := &sync.RWMutex{}
    
    type cacheEntry struct {
        response []byte
        expiry   time.Time
    }
    
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            // 只缓存 GET 请求
            if c.Request().Method != http.MethodGet {
                return next(c)
            }
            
            // 生成缓存键
            key := c.Request().URL.String()
            
            // 检查缓存
            mutex.RLock()
            entry, found := cache[key]
            mutex.RUnlock()
            
            if found && time.Now().Before(entry.expiry) {
                // 返回缓存的响应
                c.Response().Header().Set("Content-Type", "application/json")
                c.Response().Header().Set("X-Cache", "HIT")
                c.Response().WriteHeader(http.StatusOK)
                c.Response().Write(entry.response)
                return nil
            }
            
            // 创建响应记录器
            rec := httptest.NewRecorder()
            ctx := c.Request().Context()
            c.SetRequest(c.Request().WithContext(ctx))
            
            // 执行下一个处理器
            if err := next(c); err != nil {
                return err
            }
            
            // 缓存响应
            if c.Response().Status == http.StatusOK {
                mutex.Lock()
                cache[key] = cacheEntry{
                    response: rec.Body.Bytes(),
                    expiry:   time.Now().Add(duration),
                }
                mutex.Unlock()
            }
            
            return nil
        }
    }
}
```

#### 3.3.3 压缩响应

压缩 API 响应以减少带宽使用：

```go
func setupAPI(app core.App, e *echo.Echo) {
    // 启用 Gzip 压缩
    e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
        Level: 5,
        MinLength: 256,
    }))
    
    // 其他 API 设置...
}
```

### 3.4 前端性能优化

#### 3.4.1 代码编辑器优化

优化 Monaco Editor 的加载和使用：

```javascript
// 懒加载 Monaco Editor
const loadMonacoEditor = async () => {
  if (window.monaco) {
    return window.monaco;
  }
  
  // 动态导入 Monaco Editor
  const monaco = await import('monaco-editor');
  
  // 配置 Monaco Editor
  monaco.languages.typescript.javascriptDefaults.setDiagnosticsOptions({
    noSemanticValidation: false,
    noSyntaxValidation: false,
  });
  
  // 添加 PocketBase API 类型定义
  monaco.languages.typescript.javascriptDefaults.addExtraLib(
    pbTypeDefs,
    'pocketbase.d.ts'
  );
  
  window.monaco = monaco;
  return monaco;
};

// 创建编辑器实例时使用轻量级配置
const createEditor = async (container, code) => {
  const monaco = await loadMonacoEditor();
  
  return monaco.editor.create(container, {
    value: code,
    language: 'javascript',
    theme: 'vs',
    minimap: { enabled: false },  // 禁用小地图以提高性能
    automaticLayout: true,
    scrollBeyondLastLine: false,
    fontSize: 14,
    lineNumbers: 'on',
    folding: true,
    renderLineHighlight: 'line',
    scrollbar: {
      useShadows: false,
      verticalScrollbarSize: 10,
      horizontalScrollbarSize: 10,
      alwaysConsumeMouseWheel: false
    },
    overviewRulerLanes: 0,  // 禁用概览标尺以提高性能
  });
};
```

#### 3.4.2 虚拟滚动

对于长列表使用虚拟滚动，减少 DOM 元素数量：

```svelte
<!-- HooksList.svelte -->
<script>
  import { onMount } from 'svelte';
  import VirtualList from './VirtualList.svelte';
  
  let hooks = [];
  let loading = true;
  
  onMount(async () => {
    try {
      const response = await fetch('/api/hooks?page=1&perPage=100');
      const data = await response.json();
      hooks = data.items;
    } catch (error) {
      console.error('Failed to load hooks:', error);
    } finally {
      loading = false;
    }
  });
  
  const rowHeight = 60; // 每行高度
</script>

{#if loading}
  <div class="loading">Loading...</div>
{:else}
  <VirtualList items={hooks} height={500} itemHeight={rowHeight} let:item>
    <div class="hook-item">
      <div class="hook-name">{item.name}</div>
      <div class="hook-type">{item.type} / {item.event}</div>
      <div class="hook-status" class:enabled={item.enabled}>
        {item.enabled ? 'Enabled' : 'Disabled'}
      </div>
      <div class="hook-actions">
        <button on:click={() => editHook(item)}>Edit</button>
        <button on:click={() => toggleHook(item)}>
          {item.enabled ? 'Disable' : 'Enable'}
        </button>
        <button on:click={() => deleteHook(item)}>Delete</button>
      </div>
    </div>
  </VirtualList>
{/if}
```

#### 3.4.3 延迟加载和代码分割

使用延迟加载和代码分割减少初始加载时间：

```javascript
// 路由配置
export const routes = [
  {
    path: '/hooks',
    component: () => import('./views/HooksList.svelte'),
  },
  {
    path: '/hooks/new',
    component: () => import('./views/HookEditor.svelte'),
  },
  {
    path: '/hooks/:id',
    component: () => import('./views/HookEditor.svelte'),
  },
  {
    path: '/hooks/:id/versions',
    component: () => import('./views/HookVersions.svelte'),
  },
];
```

## 4. 监控和故障排除

### 4.1 性能监控

#### 4.1.1 脚本执行监控

监控脚本执行时间和资源使用情况：

```go
func monitorScriptExecution(app core.App) {
    // 定期收集和报告脚本执行统计信息
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    
    for range ticker.C {
        // 查询执行日志
        logs, err := app.Dao().GetHookLogsStats()
        if err != nil {
            log.Printf("Failed to get hook logs stats: %v", err)
            continue
        }
        
        // 分析执行时间
        var totalTime int64
        var maxTime int64
        var errorCount int
        
        for _, l := range logs {
            totalTime += int64(l.ExecutionTime)
            if int64(l.ExecutionTime) > maxTime {
                maxTime = int64(l.ExecutionTime)
            }
            if l.Status == "error" {
                errorCount++
            }
        }
        
        avgTime := float64(totalTime) / float64(len(logs))
        
        // 记录统计信息
        log.Printf("Hook execution stats: count=%d, avg_time=%.2fms, max_time=%dms, errors=%d",
            len(logs), avgTime, maxTime, errorCount)
        
        // 检查是否有性能问题
        if avgTime > 500 || maxTime > 5000 || float64(errorCount)/float64(len(logs)) > 0.1 {
            log.Printf("WARNING: Hook execution performance issues detected!")
            // 可以发送警报或通知
        }
    }
}
```

#### 4.1.2 系统资源监控

监控系统资源使用情况，确保钩子脚本不会导致资源耗尽：

```go
func monitorSystemResources() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        var memStats runtime.MemStats
        runtime.ReadMemStats(&memStats)
        
        log.Printf("System resources: Alloc=%vMB, TotalAlloc=%vMB, Sys=%vMB, NumGC=%v",
            memStats.Alloc/1024/1024,
            memStats.TotalAlloc/1024/1024,
            memStats.Sys/1024/1024,
            memStats.NumGC)
        
        // 检查内存使用是否过高
        if memStats.Alloc > 1024*1024*1024 { // 1GB
            log.Printf("WARNING: High memory usage detected!")
            // 可以触发垃圾回收或发送警报
            runtime.GC()
        }
    }
}
```

### 4.2 错误处理和恢复

#### 4.2.1 脚本错误恢复

确保脚本错误不会导致整个系统崩溃：

```go
func executeHookSafely(app core.App, hook *models.HookScript, event string, data interface{}) (interface{}, error) {
    defer func() {
        if r := recover(); r != nil {
            stack := make([]byte, 4096)
            stack = stack[:runtime.Stack(stack, false)]
            log.Printf("PANIC in hook execution: %v\n%s", r, stack)
            
            // 记录错误
            logHookExecution(app, hook, time.Now(), fmt.Errorf("panic: %v", r), string(stack))
        }
    }()
    
    // 执行脚本
    // ...
    
    return result, nil
}
```

#### 4.2.2 自动重试机制

对于临时性错误，实现自动重试机制：

```go
func executeWithRetry(app core.App, hook *models.HookScript, event string, data interface{}, maxRetries int, retryDelay time.Duration) (interface{}, error) {
    var lastErr error
    
    for i := 0; i <= maxRetries; i++ {
        result, err := executeHookSafely(app, hook, event, data)
        if err == nil {
            return result, nil
        }
        
        lastErr = err
        
        // 检查是否是临时性错误
        if isTemporaryError(err) {
            log.Printf("Temporary error executing hook %s (retry %d/%d): %v", hook.ID, i, maxRetries, err)
            
            if i < maxRetries {
                // 等待一段时间后重试
                time.Sleep(retryDelay)
                continue
            }
        } else {
            // 非临时性错误，不再重试
            break
        }
    }
    
    return nil, fmt.Errorf("failed after %d retries: %v", maxRetries, lastErr)
}

func isTemporaryError(err error) bool {
    // 检查是否是临时性错误，如网络超时、资源暂时不可用等
    if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, io.ErrUnexpectedEOF) {
        return true
    }
    
    // 检查错误消息是否包含临时性错误的关键词
    errMsg := err.Error()
    temporaryKeywords := []string{"timeout", "temporary", "retriable", "retry", "again", "reset"}
    
    for _, keyword := range temporaryKeywords {
        if strings.Contains(strings.ToLower(errMsg), keyword) {
            return true
        }
    }
    
    return false
}
```

## 5. 最佳实践建议

### 5.1 安全最佳实践

1. **最小权限原则**：只给脚本提供完成任务所需的最小权限集。
2. **输入验证**：验证所有用户输入，包括脚本代码和 API 参数。
3. **超时限制**：为所有脚本执行设置合理的超时限制。
4. **资源限制**：限制脚本可以使用的内存和 CPU 资源。
5. **沙箱隔离**：在隔离的环境中执行脚本，限制对系统资源的访问。
6. **审计日志**：记录所有关键操作，包括脚本创建、修改和执行。
7. **版本控制**：保留脚本的历史版本，以便在出现问题时回滚。
8. **定期安全审查**：定期审查脚本代码和系统配置，查找潜在的安全问题。

### 5.2 性能最佳实践

1. **脚本优化**：编写高效的脚本，避免不必要的计算和 I/O 操作。
2. **缓存策略**：缓存编译后的脚本和频繁访问的数据。
3. **连接池管理**：优化数据库连接池配置，避免连接泄漏。
4. **异步处理**：对于耗时操作，使用异步处理避免阻塞。
5. **批量处理**：尽可能批量处理数据，减少数据库交互次数。
6. **索引优化**：为常用查询添加适当的索引。
7. **监控和调优**：持续监控系统性能，根据实际情况进行调优。
8. **负载测试**：在部署前进行负载测试，确保系统在高负载下仍能正常工作。

## 6. 结论

本文档详细说明了 PocketBase 钩子脚本管理系统的安全措施和性能优化策略。通过实施这些措施，可以确保钩子脚本在安全的环境中高效执行，同时保护系统免受潜在的安全威胁和性能问题的影响。

安全和性能是一个持续的过程，需要定期审查和更新。随着系统的发展和新威胁的出现，应当不断完善安全措施和优化性能策略，确保系统始终保持安全、稳定和高效。