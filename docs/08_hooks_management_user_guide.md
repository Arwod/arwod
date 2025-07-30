# PocketBase 钩子脚本管理系统用户指南

## 1. 概述

本用户指南详细说明了 PocketBase 钩子脚本管理系统的使用方法、功能和最佳实践。该系统允许管理员通过 PocketBase 管理界面直接创建、编辑、测试和管理钩子脚本，并将脚本存储在内置数据库中，而不是文件系统中。

### 1.1 什么是钩子脚本？

钩子脚本是一段 JavaScript 代码，可以在特定事件发生时自动执行，例如：

- 记录创建、更新或删除前后
- 集合创建、更新或删除前后
- API 请求处理前或响应发送前
- 系统启动、服务启动或终止时

通过钩子脚本，您可以扩展 PocketBase 的功能，实现自定义业务逻辑，而无需修改核心代码。

### 1.2 钩子脚本管理系统的优势

相比于传统的文件系统钩子脚本，数据库存储的钩子脚本管理系统具有以下优势：

- **集中管理**：所有脚本都可以在管理界面中集中管理，无需访问服务器文件系统
- **版本控制**：自动保存脚本的历史版本，方便回滚和比较
- **即时编辑**：直接在浏览器中编辑和测试脚本，无需重启服务器
- **状态管理**：轻松启用或禁用脚本，无需删除或重命名文件
- **执行顺序控制**：明确控制多个脚本的执行顺序
- **测试功能**：内置测试工具，方便验证脚本功能

## 2. 访问钩子脚本管理界面

### 2.1 导航到钩子脚本管理页面

1. 登录 PocketBase 管理界面
2. 在左侧导航菜单中，点击「钩子脚本」菜单项
3. 进入钩子脚本列表页面，显示所有现有的钩子脚本

![钩子脚本导航](https://placeholder-for-navigation-menu.png)

### 2.2 界面概览

钩子脚本管理界面主要包括以下几个部分：

- **钩子脚本列表**：显示所有钩子脚本的概览
- **脚本编辑器**：创建或编辑钩子脚本
- **版本历史**：查看和管理脚本的历史版本
- **测试工具**：测试脚本的执行效果

## 3. 管理钩子脚本

### 3.1 查看钩子脚本列表

钩子脚本列表页面显示所有现有的钩子脚本，包括以下信息：

- 脚本名称
- 脚本类型（记录、集合、API、系统）
- 事件类型（创建、更新、删除等）
- 关联的集合（如适用）
- 状态（启用/禁用）
- 执行顺序
- 最后更新时间

您可以使用以下功能来管理列表：

- **筛选**：按类型、事件、集合或状态筛选脚本
- **搜索**：按名称或描述搜索脚本
- **排序**：按任意列排序脚本
- **分页**：浏览大量脚本

### 3.2 创建新的钩子脚本

创建新的钩子脚本的步骤如下：

1. 在钩子脚本列表页面，点击右上角的「+ 新建钩子脚本」按钮
2. 在脚本编辑页面，填写以下信息：
   - **名称**：脚本的唯一名称（必填）
   - **描述**：脚本的简短描述（可选）
   - **类型**：选择脚本类型（记录、集合、API、系统）
   - **事件**：选择触发事件（根据所选类型显示不同选项）
   - **集合**：如果类型为记录或集合，选择关联的集合
   - **执行顺序**：设置脚本的执行顺序（数字越小越先执行）
   - **启用/禁用**：设置脚本的初始状态
3. 在代码编辑器中编写脚本代码
4. 点击「保存」按钮保存脚本

![创建钩子脚本](https://placeholder-for-create-hook.png)

### 3.3 编辑钩子脚本

编辑现有钩子脚本的步骤如下：

1. 在钩子脚本列表页面，点击要编辑的脚本的「编辑」按钮
2. 在脚本编辑页面，修改脚本信息和代码
3. 点击「保存」按钮保存更改，或点击「保存为新版本」按钮创建新版本

### 3.4 启用/禁用钩子脚本

您可以通过以下方式启用或禁用钩子脚本：

1. 在钩子脚本列表页面，点击脚本行中的启用/禁用开关
2. 在脚本编辑页面，切换启用/禁用开关，然后点击「保存」按钮

禁用的脚本不会在事件触发时执行，但仍然保留在系统中。

### 3.5 删除钩子脚本

删除钩子脚本的步骤如下：

1. 在钩子脚本列表页面，点击要删除的脚本的「删除」按钮
2. 在确认对话框中，点击「确认」按钮

**注意**：删除操作是永久性的，无法恢复。如果不确定是否需要删除脚本，建议先禁用它。

### 3.6 调整执行顺序

当多个钩子脚本注册到同一事件时，它们会按照执行顺序从小到大依次执行。您可以通过以下方式调整执行顺序：

1. 在钩子脚本列表页面，使用拖放功能调整脚本的顺序
2. 在脚本编辑页面，修改执行顺序字段的值，然后点击「保存」按钮

## 4. 编写钩子脚本

### 4.1 脚本类型和事件

PocketBase 支持以下类型的钩子脚本：

#### 4.1.1 记录钩子

记录钩子在记录操作前后触发，可用于验证、修改或扩展记录数据。

**可用事件**：
- `create`：记录创建前
- `create.after`：记录创建后
- `update`：记录更新前
- `update.after`：记录更新后
- `delete`：记录删除前
- `delete.after`：记录删除后

**示例**：在创建用户记录后发送欢迎邮件

```javascript
// 在用户创建后发送欢迎邮件
onRecordAfterCreateRequest(({ collection, record, data }) => {
    // 只处理用户集合
    if (collection.name !== 'users') {
        return;
    }
    
    // 发送欢迎邮件
    $app.dao().findRecordById('users', record.id).then((user) => {
        const html = `
            <h1>欢迎加入我们，${user.name}！</h1>
            <p>感谢您注册我们的服务。</p>
        `;
        
        $app.newMailClient().send({
            to: user.email,
            subject: '欢迎加入',
            html: html,
        });
    });
});
```

#### 4.1.2 集合钩子

集合钩子在集合操作前后触发，可用于验证、修改或扩展集合定义。

**可用事件**：
- `create`：集合创建前
- `create.after`：集合创建后
- `update`：集合更新前
- `update.after`：集合更新后
- `delete`：集合删除前
- `delete.after`：集合删除后

**示例**：在更新集合前添加审计字段

```javascript
// 在更新集合前添加审计字段
onCollectionBeforeUpdateRequest(({ collection, data }) => {
    // 检查是否已有审计字段
    const hasAuditFields = collection.schema.some(field => 
        field.name === 'updated_by' || field.name === 'updated_at'
    );
    
    // 如果没有审计字段，添加它们
    if (!hasAuditFields) {
        data.schema = data.schema || [];
        
        // 添加 updated_at 字段
        data.schema.push({
            name: 'updated_at',
            type: 'date',
            required: true,
            options: {
                default: '2022-01-01 00:00:00.000Z'
            }
        });
        
        // 添加 updated_by 字段
        data.schema.push({
            name: 'updated_by',
            type: 'text',
            required: false
        });
    }
});
```

#### 4.1.3 API 钩子

API 钩子在 API 请求处理前或响应发送前触发，可用于请求验证、修改响应或实现自定义 API 端点。

**可用事件**：
- `request`：API 请求处理前
- `response`：API 响应发送前

**示例**：添加自定义响应头

```javascript
// 添加自定义响应头
onApiBeforeResponseRequest(({ response }) => {
    // 添加自定义响应头
    response.header('X-Custom-Header', 'Hello from PocketBase');
    
    // 添加 CORS 头
    response.header('Access-Control-Allow-Origin', '*');
    response.header('Access-Control-Allow-Methods', 'GET, POST, PUT, DELETE, OPTIONS');
    response.header('Access-Control-Allow-Headers', 'Content-Type, Authorization');
});
```

#### 4.1.4 系统钩子

系统钩子在 PocketBase 系统事件发生时触发，可用于初始化、清理或监控系统状态。

**可用事件**：
- `bootstrap`：系统启动时
- `serve`：HTTP 服务启动时
- `terminate`：系统终止时

**示例**：在系统启动时初始化配置

```javascript
// 在系统启动时初始化配置
onBootstrap(() => {
    console.log('PocketBase 正在启动...');
    
    // 初始化配置
    const settings = $app.settings();
    if (!settings.smtp.enabled) {
        console.log('配置 SMTP 设置...');
        settings.smtp.enabled = true;
        settings.smtp.host = 'smtp.example.com';
        settings.smtp.port = 587;
        settings.smtp.username = 'noreply@example.com';
        settings.smtp.password = 'your-password';
        
        $app.dao().saveSettings(settings);
    }
    
    console.log('初始化完成！');
});
```

### 4.2 脚本上下文和 API

钩子脚本在执行时可以访问以下上下文和 API：

#### 4.2.1 全局对象

- `$app`：PocketBase 应用实例，提供对核心功能的访问
- `$template`：模板引擎，用于渲染 HTML 模板
- `console`：日志记录对象，用于输出调试信息

#### 4.2.2 事件处理函数

根据脚本类型和事件，可以使用以下事件处理函数：

**记录钩子**：
- `onRecordBeforeCreateRequest(handler)`
- `onRecordAfterCreateRequest(handler)`
- `onRecordBeforeUpdateRequest(handler)`
- `onRecordAfterUpdateRequest(handler)`
- `onRecordBeforeDeleteRequest(handler)`
- `onRecordAfterDeleteRequest(handler)`

**集合钩子**：
- `onCollectionBeforeCreateRequest(handler)`
- `onCollectionAfterCreateRequest(handler)`
- `onCollectionBeforeUpdateRequest(handler)`
- `onCollectionAfterUpdateRequest(handler)`
- `onCollectionBeforeDeleteRequest(handler)`
- `onCollectionAfterDeleteRequest(handler)`

**API 钩子**：
- `onApiBeforeRequestRequest(handler)`
- `onApiBeforeResponseRequest(handler)`

**系统钩子**：
- `onBootstrap(handler)`
- `onServe(handler)`
- `onTerminate(handler)`

#### 4.2.3 处理函数参数

处理函数接收一个包含上下文信息的对象，根据脚本类型和事件不同，可能包含以下属性：

**记录钩子**：
- `collection`：记录所属的集合
- `record`：操作的记录（对于 `create` 事件，这是一个空记录）
- `data`：请求数据（用于 `create` 和 `update` 事件）

**集合钩子**：
- `collection`：操作的集合
- `data`：请求数据（用于 `create` 和 `update` 事件）

**API 钩子**：
- `request`：HTTP 请求对象
- `response`：HTTP 响应对象

**系统钩子**：
- 无参数

### 4.3 常用 API 示例

#### 4.3.1 数据访问

```javascript
// 查询记录
const records = await $app.dao().findRecordsByFilter(
    'users',
    'created >= "2023-01-01 00:00:00" && verified = true',
    '+created',
    100,
    0
);

// 创建记录
const record = new Record($app.dao().findCollectionByNameOrId('posts'));
record.set({
    title: '新文章',
    content: '文章内容...',
    author: userId
});
await $app.dao().saveRecord(record);

// 更新记录
const user = await $app.dao().findRecordById('users', userId);
user.set('verified', true);
user.set('verifiedAt', new Date().toISOString());
await $app.dao().saveRecord(user);

// 删除记录
await $app.dao().deleteRecord(record);
```

#### 4.3.2 发送邮件

```javascript
// 发送简单邮件
$app.newMailClient().send({
    to: 'user@example.com',
    subject: '您好',
    html: '<h1>欢迎使用我们的服务</h1>',
});

// 使用模板发送邮件
const html = $template.render('welcome.html', {
    name: user.name,
    verifyUrl: `https://example.com/verify?token=${user.token}`,
});

$app.newMailClient().send({
    to: user.email,
    subject: '欢迎加入',
    html: html,
});
```

#### 4.3.3 文件操作

```javascript
// 读取文件
const fs = require('fs');
const content = fs.readFileSync('/tmp/data.json', 'utf8');
const data = JSON.parse(content);

// 写入文件
fs.writeFileSync('/tmp/output.json', JSON.stringify(data, null, 2));
```

#### 4.3.4 HTTP 请求

```javascript
// 发送 GET 请求
const http = require('http');
const response = await http.get('https://api.example.com/data');
const data = response.json();

// 发送 POST 请求
const response = await http.post('https://api.example.com/users', {
    json: { name: 'John', email: 'john@example.com' },
    headers: { 'Authorization': 'Bearer token' },
});
```

### 4.4 最佳实践

#### 4.4.1 错误处理

始终包含适当的错误处理，以防止脚本执行失败：

```javascript
onRecordBeforeCreateRequest(({ collection, record, data }) => {
    try {
        // 验证数据
        if (collection.name === 'orders' && !data.items) {
            throw new Error('订单必须包含至少一个商品');
        }
        
        // 处理数据
        // ...
    } catch (error) {
        console.error(`处理记录创建时出错: ${error.message}`);
        throw error; // 重新抛出错误以中止请求
    }
});
```

#### 4.4.2 性能优化

编写高效的脚本，避免不必要的操作：

```javascript
// 不好的做法：在循环中执行数据库查询
for (const id of userIds) {
    const user = await $app.dao().findRecordById('users', id); // 每次循环都查询数据库
    // 处理用户...
}

// 好的做法：批量查询数据库
const filter = `id IN ["${userIds.join('", "')}"]`;
const users = await $app.dao().findRecordsByFilter('users', filter);
for (const user of users) {
    // 处理用户...
}
```

#### 4.4.3 模块化和代码复用

使用模块化和函数封装来提高代码可读性和可维护性：

```javascript
// 定义可复用的函数
function validateEmail(email) {
    const regex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    return regex.test(email);
}

function sendWelcomeEmail(user) {
    // 发送欢迎邮件的逻辑
    // ...
}

// 在钩子中使用这些函数
onRecordBeforeCreateRequest(({ collection, data }) => {
    if (collection.name === 'users') {
        if (!validateEmail(data.email)) {
            throw new Error('无效的电子邮件地址');
        }
    }
});

onRecordAfterCreateRequest(({ collection, record }) => {
    if (collection.name === 'users') {
        sendWelcomeEmail(record);
    }
});
```

## 5. 测试钩子脚本

### 5.1 使用测试工具

钩子脚本管理系统提供了内置的测试工具，可以在不影响实际数据的情况下测试脚本的执行效果：

1. 在脚本编辑页面，点击「测试」按钮
2. 在测试面板中，选择测试事件类型
3. 输入测试数据（JSON 格式）
4. 点击「运行测试」按钮
5. 查看测试结果，包括执行状态、执行时间、控制台输出和错误信息

![测试钩子脚本](https://placeholder-for-test-hook.png)

### 5.2 测试数据示例

以下是不同类型钩子脚本的测试数据示例：

#### 5.2.1 记录钩子测试数据

```json
{
  "collection": {
    "id": "test_collection",
    "name": "users",
    "schema": [
      {
        "id": "field1",
        "name": "name",
        "type": "text",
        "required": true
      },
      {
        "id": "field2",
        "name": "email",
        "type": "email",
        "required": true
      }
    ]
  },
  "record": {
    "id": "test_record",
    "name": "测试用户",
    "email": "test@example.com"
  },
  "data": {
    "name": "更新的用户名",
    "email": "updated@example.com"
  }
}
```

#### 5.2.2 集合钩子测试数据

```json
{
  "collection": {
    "id": "test_collection",
    "name": "posts",
    "schema": [
      {
        "id": "field1",
        "name": "title",
        "type": "text",
        "required": true
      },
      {
        "id": "field2",
        "name": "content",
        "type": "editor",
        "required": true
      }
    ]
  },
  "data": {
    "name": "updated_posts",
    "schema": [
      {
        "id": "field1",
        "name": "title",
        "type": "text",
        "required": true
      },
      {
        "id": "field2",
        "name": "content",
        "type": "editor",
        "required": true
      },
      {
        "id": "field3",
        "name": "author",
        "type": "relation",
        "required": true,
        "options": {
          "collectionId": "users"
        }
      }
    ]
  }
}
```

#### 5.2.3 API 钩子测试数据

```json
{
  "request": {
    "method": "GET",
    "url": "/api/collections/users/records",
    "headers": {
      "Authorization": "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "Content-Type": "application/json"
    },
    "query": {
      "filter": "verified = true",
      "sort": "-created"
    },
    "data": {}
  },
  "response": {
    "status": 200,
    "headers": {
      "Content-Type": "application/json"
    },
    "data": {
      "page": 1,
      "perPage": 30,
      "totalItems": 2,
      "totalPages": 1,
      "items": [
        {
          "id": "test_user1",
          "name": "用户1",
          "email": "user1@example.com",
          "verified": true
        },
        {
          "id": "test_user2",
          "name": "用户2",
          "email": "user2@example.com",
          "verified": true
        }
      ]
    }
  }
}
```

## 6. 管理脚本版本

### 6.1 查看版本历史

查看钩子脚本的版本历史的步骤如下：

1. 在钩子脚本列表页面，点击要查看的脚本的「版本历史」按钮
2. 在版本历史页面，查看所有历史版本，包括版本号、创建时间、创建者和版本说明

![版本历史](https://placeholder-for-version-history.png)

### 6.2 比较版本

比较不同版本的钩子脚本的步骤如下：

1. 在版本历史页面，选择两个要比较的版本
2. 点击「比较」按钮
3. 在比较视图中，查看两个版本之间的差异，高亮显示的部分表示变更

### 6.3 恢复到历史版本

恢复到历史版本的步骤如下：

1. 在版本历史页面，找到要恢复的版本
2. 点击该版本的「恢复」按钮
3. 在确认对话框中，输入版本说明（可选）
4. 点击「确认」按钮

恢复操作会创建一个新的版本，而不是直接修改当前版本。

## 7. 导入和导出脚本

### 7.1 导出脚本

导出钩子脚本的步骤如下：

1. 在钩子脚本列表页面，选择要导出的脚本（可以选择多个）
2. 点击「导出」按钮
3. 选择导出格式（JSON 或 JavaScript）
4. 点击「确认」按钮
5. 下载导出的文件

### 7.2 导入脚本

导入钩子脚本的步骤如下：

1. 在钩子脚本列表页面，点击「导入」按钮
2. 选择导入文件（JSON 或 JavaScript 格式）
3. 选择导入选项（是否覆盖同名脚本）
4. 点击「确认」按钮
5. 查看导入结果，包括成功导入的脚本数量和跳过的脚本数量

## 8. 从文件系统迁移脚本

如果您之前使用文件系统存储钩子脚本，可以使用迁移工具将它们导入到数据库中：

1. 在钩子脚本列表页面，点击「从文件系统迁移」按钮
2. 选择钩子脚本目录（默认为配置中的 `HooksDir`）
3. 选择迁移选项（是否删除原始文件、是否覆盖同名脚本）
4. 点击「确认」按钮
5. 查看迁移结果，包括成功迁移的脚本数量和跳过的脚本数量

## 9. 故障排除

### 9.1 常见问题

#### 9.1.1 脚本不执行

如果钩子脚本没有执行，请检查以下几点：

1. 确认脚本已启用（在列表页面查看状态）
2. 确认脚本类型和事件正确（例如，如果要在记录创建后执行，应使用 `create.after` 事件）
3. 确认关联的集合正确（对于记录和集合钩子）
4. 检查脚本代码是否有语法错误（使用测试工具验证）
5. 查看服务器日志，检查是否有错误消息

#### 9.1.2 脚本执行错误

如果钩子脚本执行时出现错误，请检查以下几点：

1. 使用测试工具验证脚本逻辑
2. 检查脚本中的错误处理是否完善
3. 确认脚本使用的 API 和功能是否可用
4. 查看服务器日志，获取详细的错误信息

#### 9.1.3 性能问题

如果钩子脚本导致性能问题，请考虑以下优化措施：

1. 减少数据库查询次数，使用批量查询
2. 避免在循环中执行耗时操作
3. 使用异步处理耗时任务
4. 优化脚本逻辑，减少不必要的计算
5. 考虑禁用不必要的脚本

### 9.2 调试技巧

#### 9.2.1 使用控制台日志

在脚本中使用 `console.log()` 输出调试信息：

```javascript
console.log('开始执行脚本...');
console.log('数据:', JSON.stringify(data, null, 2));
console.log('完成处理');
```

#### 9.2.2 使用测试工具

使用内置的测试工具验证脚本逻辑，查看执行结果和错误信息。

#### 9.2.3 检查服务器日志

查看 PocketBase 服务器日志，获取详细的错误信息和调试输出。

## 10. 安全最佳实践

### 10.1 输入验证

始终验证用户输入，防止注入攻击和其他安全问题：

```javascript
onRecordBeforeCreateRequest(({ collection, data }) => {
    if (collection.name === 'users') {
        // 验证电子邮件
        if (data.email && !/^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/.test(data.email)) {
            throw new Error('无效的电子邮件地址');
        }
        
        // 验证密码强度
        if (data.password && data.password.length < 8) {
            throw new Error('密码必须至少包含 8 个字符');
        }
    }
});
```

### 10.2 权限控制

确保脚本遵循适当的权限控制，不要绕过系统的安全机制：

```javascript
onRecordBeforeUpdateRequest(({ collection, record, data, request }) => {
    if (collection.name === 'posts') {
        // 获取当前用户
        const authRecord = request.authRecord;
        
        // 检查用户是否有权限更新文章
        if (!authRecord || (authRecord.id !== record.author && authRecord.role !== 'admin')) {
            throw new Error('您没有权限更新此文章');
        }
    }
});
```

### 10.3 敏感信息处理

避免在脚本中硬编码敏感信息，使用环境变量或安全的配置管理：

```javascript
// 不好的做法：硬编码 API 密钥
const apiKey = 'sk_live_1234567890abcdef';

// 好的做法：从环境变量或设置中获取 API 密钥
const settings = $app.settings();
const apiKey = settings.meta.externalApiKey || process.env.EXTERNAL_API_KEY;
```

### 10.4 错误消息安全

避免在错误消息中泄露敏感信息：

```javascript
// 不好的做法：泄露敏感信息
try {
    // 数据库操作...
} catch (error) {
    throw new Error(`数据库错误: ${error.message}`); // 可能泄露数据库结构
}

// 好的做法：使用通用错误消息
try {
    // 数据库操作...
} catch (error) {
    console.error(`数据库错误: ${error.message}`); // 在日志中记录详细信息
    throw new Error('处理请求时出错'); // 向用户显示通用消息
}
```

## 11. 结论

本用户指南详细说明了 PocketBase 钩子脚本管理系统的使用方法、功能和最佳实践。通过使用这个系统，您可以更方便地创建、编辑、测试和管理钩子脚本，扩展 PocketBase 的功能，实现自定义业务逻辑。

钩子脚本是 PocketBase 的强大功能，可以帮助您实现各种自定义需求，从简单的数据验证到复杂的业务流程自动化。通过本指南中的示例和最佳实践，您可以充分利用这一功能，构建更强大、更灵活的应用程序。