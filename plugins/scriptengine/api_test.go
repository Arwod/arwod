package scriptengine

import (
	"testing"
	"time"
	"log/slog"

	"github.com/pocketbase/pocketbase/tests"
)

func TestAPIBindings_DatabaseOperations(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 10 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "dao_access",
			content:     "const dao = $app.dao(); console.log('DAO type:', typeof dao);",
			expectError: false,
			description: "Should provide access to DAO",
		},
		{
			name:        "find_collection",
			content:     "const collection = $app.dao().findCollectionByNameOrId('users'); console.log('Collection found:', !!collection);",
			expectError: false,
			description: "Should allow finding collections",
		},
		{
			name:        "list_collections",
			content:     "const collections = $app.dao().findCollectionsByType('base'); console.log('Collections count:', collections.length);",
			expectError: false,
			description: "Should allow listing collections",
		},
		{
			name:        "query_records",
			content:     "try { const records = $app.dao().findRecordsByFilter('users', 'id != \"\"', '-created', 10); console.log('Records found:', records.length); } catch(e) { console.log('Query error (expected):', e.message); }",
			expectError: false,
			description: "Should allow querying records",
		},
		{
			name:        "settings_access",
			content:     "const settings = $app.settings(); console.log('Settings available:', !!settings);",
			expectError: false,
			description: "Should provide access to settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "dbx",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			_, err := engine.ExecuteScript(tt.name, map[string]interface{}{})

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestAPIBindings_LoggingOperations(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "logger_info",
			content:     "$app.logger().info('Test info message from script');",
			expectError: false,
			description: "Should allow info logging",
		},
		{
			name:        "logger_error",
			content:     "$app.logger().error('Test error message from script');",
			expectError: false,
			description: "Should allow error logging",
		},
		{
			name:        "logger_warn",
			content:     "$app.logger().warn('Test warning message from script');",
			expectError: false,
			description: "Should allow warning logging",
		},
		{
			name:        "logger_debug",
			content:     "$app.logger().debug('Test debug message from script');",
			expectError: false,
			description: "Should allow debug logging",
		},
		{
			name:        "console_log",
			content:     "console.log('Console log message'); console.error('Console error'); console.warn('Console warn');",
			expectError: false,
			description: "Should allow console logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			_, err := engine.ExecuteScript(tt.name, map[string]interface{}{})

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestAPIBindings_RequestContext(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "request_available",
			content:     "console.log('Request available:', typeof $request !== 'undefined');",
			expectError: false,
			description: "Should provide request context",
		},
		{
			name:        "request_method",
			content:     "if ($request) { console.log('Request method:', $request.method || 'undefined'); }",
			expectError: false,
			description: "Should provide request method",
		},
		{
			name:        "request_headers",
			content:     "if ($request && $request.headers) { console.log('Headers available:', typeof $request.headers); }",
			expectError: false,
			description: "Should provide request headers",
		},
		{
			name:        "request_query",
			content:     "if ($request && $request.query) { console.log('Query available:', typeof $request.query); }",
			expectError: false,
			description: "Should provide request query parameters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "http",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟请求上下文
			requestContext := map[string]interface{}{
				"method":  "GET",
				"path":    "/api/test",
				"headers": map[string]string{"Content-Type": "application/json"},
				"query":   map[string]string{"test": "value"},
			}

			_, err := engine.ExecuteScript(tt.name, requestContext)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestAPIBindings_HookContext(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "record_available",
			content:     "console.log('Record available:', typeof $record !== 'undefined');",
			expectError: false,
			description: "Should provide record context in hooks",
		},
		{
			name:        "record_id",
			content:     "if ($record) { console.log('Record ID:', $record.id || 'undefined'); }",
			expectError: false,
			description: "Should provide record ID",
		},
		{
			name:        "record_collection",
			content:     "if ($record) { console.log('Record collection:', $record.collection || 'undefined'); }",
			expectError: false,
			description: "Should provide record collection",
		},
		{
			name:        "record_data",
			content:     "if ($record && $record.data) { console.log('Record data available:', typeof $record.data); }",
			expectError: false,
			description: "Should provide record data",
		},
		{
			name:        "hook_type",
			content:     "console.log('Hook type:', $hookType || 'undefined');",
			expectError: false,
			description: "Should provide hook type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟Hook上下文
			hookContext := map[string]interface{}{
				"record": map[string]interface{}{
					"id":         "test_record_id",
					"collection": "users",
					"data": map[string]interface{}{
						"name":  "Test User",
						"email": "test@example.com",
					},
				},
				"hookType": "beforeCreate",
			}

			_, err := engine.ExecuteScript(tt.name, hookContext)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestAPIBindings_UtilityFunctions(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	tests := []struct {
		name        string
		content     string
		expectError bool
		description string
	}{
		{
            name:        "json_operations",
            content:     "const obj = {test: 'value'}; const jsonStr = JSON.stringify(obj); const parsed = JSON.parse(jsonStr); console.log('JSON works:', parsed.test === 'value');",
            expectError: false,
            description: "Should provide JSON utilities",
        },
		{
			name:        "math_operations",
			content:     "const result = Math.sqrt(16) + Math.random() + Math.floor(3.7); console.log('Math result:', result);",
			expectError: false,
			description: "Should provide Math utilities",
		},
		{
			name:        "date_operations",
			content:     "const now = new Date(); const timestamp = now.getTime(); console.log('Date works:', timestamp > 0);",
			expectError: false,
			description: "Should provide Date utilities",
		},
		{
			name:        "string_operations",
			content:     "const str = 'Hello World'; const upper = str.toUpperCase(); const lower = str.toLowerCase(); console.log('String works:', upper.length === str.length);",
			expectError: false,
			description: "Should provide String utilities",
		},
		{
			name:        "array_operations",
			content:     "const arr = [1, 2, 3]; const mapped = arr.map(x => x * 2); const filtered = arr.filter(x => x > 1); console.log('Array works:', mapped.length === 3 && filtered.length === 2);",
			expectError: false,
			description: "Should provide Array utilities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			_, err := engine.ExecuteScript(tt.name, map[string]interface{}{})

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func BenchmarkAPIBindings_DatabaseAccess(b *testing.B) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		b.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         4,
		MaxExecutionTime: 10 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelError,
	}
	engine := NewScriptEngine(testApp, config)

	// 数据库访问基准测试脚本
	dbScript := &Script{
		ID:       "db_benchmark",
		Name:     "Database Benchmark",
		Content:  "const dao = $app.dao(); const collections = dao.findCollectionsByType('base'); console.log('Found collections:', collections.length);",
		Enabled:  true,
		Category: "benchmark",
	}

	if err := engine.LoadScript(dbScript); err != nil {
		b.Fatalf("Failed to load database script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ExecuteScript("db_benchmark", map[string]interface{}{"iteration": i})
		if err != nil {
			b.Errorf("Unexpected error in database access: %v", err)
		}
	}
}