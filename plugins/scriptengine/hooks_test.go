package scriptengine

import (
	"fmt"
	"testing"
	"time"
	"log/slog"

	"github.com/pocketbase/pocketbase/tests"
)

func TestHooks_RecordHooks(t *testing.T) {
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

	// 测试不同的记录Hook
	tests := []struct {
		name        string
		hookName    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "before_create_hook",
			hookName:    "beforeRecordCreate",
			content:     "console.log('Before create hook triggered for:', $record.collection); if ($record.data) { $record.data.created_by_script = true; }",
			expectError: false,
			description: "Should handle before create hook",
		},
		{
			name:        "after_create_hook",
			hookName:    "afterRecordCreate",
			content:     "console.log('After create hook triggered for record:', $record.id); $app.logger().info('Record created successfully');",
			expectError: false,
			description: "Should handle after create hook",
		},
		{
			name:        "before_update_hook",
			hookName:    "beforeRecordUpdate",
			content:     "console.log('Before update hook triggered'); if ($record.data) { $record.data.updated_by_script = new Date().toISOString(); }",
			expectError: false,
			description: "Should handle before update hook",
		},
		{
			name:        "after_update_hook",
			hookName:    "afterRecordUpdate",
			content:     "console.log('After update hook triggered'); $app.logger().info('Record updated:', $record.id);",
			expectError: false,
			description: "Should handle after update hook",
		},
		{
			name:        "before_delete_hook",
			hookName:    "beforeRecordDelete",
			content:     "console.log('Before delete hook triggered for:', $record.id); $app.logger().warn('Record about to be deleted');",
			expectError: false,
			description: "Should handle before delete hook",
		},
		{
			name:        "after_delete_hook",
			hookName:    "afterRecordDelete",
			content:     "console.log('After delete hook triggered'); $app.logger().info('Record deleted successfully');",
			expectError: false,
			description: "Should handle after delete hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hook",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟Hook事件数据
			eventData := map[string]interface{}{
				"record": map[string]interface{}{
					"id": "test_record_123",
					"collection": "test_collection",
					"data": map[string]interface{}{},
				},
			}

			err := engine.ExecuteHook(tt.hookName, eventData)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestHooks_AuthHooks(t *testing.T) {
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
		hookName    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "before_auth_hook",
			hookName:    "beforeAuth",
			content:     "console.log('Before auth hook triggered'); $app.logger().info('Authentication attempt');",
			expectError: false,
			description: "Should handle before auth hook",
		},
		{
			name:        "after_auth_hook",
			hookName:    "afterAuth",
			content:     "console.log('After auth hook triggered'); $app.logger().info('Authentication successful');",
			expectError: false,
			description: "Should handle after auth hook",
		},
		{
			name:        "before_logout_hook",
			hookName:    "beforeLogout",
			content:     "console.log('Before logout hook triggered'); $app.logger().info('User logging out');",
			expectError: false,
			description: "Should handle before logout hook",
		},
		{
			name:        "after_logout_hook",
			hookName:    "afterLogout",
			content:     "console.log('After logout hook triggered'); $app.logger().info('User logged out');",
			expectError: false,
			description: "Should handle after logout hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "auth_hook",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟认证事件数据
			eventData := map[string]interface{}{
				"record": map[string]interface{}{
					"id": "user_123",
					"collection": "users",
				},
				"token": "test_token",
			}

			err := engine.ExecuteHook(tt.hookName, eventData)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestHooks_FileHooks(t *testing.T) {
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
		hookName    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "before_file_upload_hook",
			hookName:    "beforeFileUpload",
			content:     "console.log('Before file upload hook triggered'); $app.logger().info('File upload starting');",
			expectError: false,
			description: "Should handle before file upload hook",
		},
		{
			name:        "after_file_upload_hook",
			hookName:    "afterFileUpload",
			content:     "console.log('After file upload hook triggered'); $app.logger().info('File uploaded successfully');",
			expectError: false,
			description: "Should handle after file upload hook",
		},
		{
			name:        "before_file_delete_hook",
			hookName:    "beforeFileDelete",
			content:     "console.log('Before file delete hook triggered'); $app.logger().warn('File about to be deleted');",
			expectError: false,
			description: "Should handle before file delete hook",
		},
		{
			name:        "after_file_delete_hook",
			hookName:    "afterFileDelete",
			content:     "console.log('After file delete hook triggered'); $app.logger().info('File deleted successfully');",
			expectError: false,
			description: "Should handle after file delete hook",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "file_hook",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟文件事件数据
			eventData := map[string]interface{}{
				"record": map[string]interface{}{
					"id": "file_record_123",
					"collection": "files",
				},
				"uploadedFiles": []string{"test_file.jpg"},
			}

			err := engine.ExecuteHook(tt.hookName, eventData)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestHooks_CustomHooks(t *testing.T) {
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
		hookName    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "custom_business_logic",
			hookName:    "customBusinessLogic",
			content:     "console.log('Custom business logic executed'); const result = $input.value * 2; console.log('Calculated result:', result);",
			expectError: false,
			description: "Should handle custom business logic",
		},
		{
			name:        "data_validation",
			hookName:    "dataValidation",
			content:     "console.log('Data validation hook'); if (!$input.email || !$input.email.includes('@')) { throw new Error('Invalid email'); }",
			expectError: false,
			description: "Should handle data validation",
		},
		{
			name:        "notification_sender",
			hookName:    "sendNotification",
			content:     "console.log('Sending notification to:', $input.recipient); $app.logger().info('Notification sent');",
			expectError: false,
			description: "Should handle notification sending",
		},
		{
			name:        "audit_logger",
			hookName:    "auditLog",
			content:     "console.log('Audit log entry'); $app.logger().info('Action performed:', JSON.stringify($input));",
			expectError: false,
			description: "Should handle audit logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "custom_hook",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Fatalf("Failed to load script: %v", err)
			}

			// 模拟自定义Hook数据
			eventData := map[string]interface{}{
				"value":     42,
				"email":     "test@example.com",
				"recipient": "user@example.com",
				"action":    "user_login",
				"timestamp": time.Now().Unix(),
			}

			err := engine.ExecuteHook(tt.hookName, eventData)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}
		})
	}
}

func TestHooks_ErrorHandling(t *testing.T) {
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
		hookName    string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "hook_with_error",
			hookName:    "errorHook",
			content:     "throw new Error('Intentional hook error');",
			expectError: true,
			description: "Should handle hook errors gracefully",
		},
		{
			name:        "hook_with_syntax_error",
			hookName:    "syntaxErrorHook",
			content:     "console.log('unclosed string",
			expectError: true,
			description: "Should handle syntax errors in hooks",
		},
		{
			name:        "hook_with_timeout",
			hookName:    "timeoutHook",
			content:     "while(true) { /* infinite loop */ }",
			expectError: true,
			description: "Should handle hook timeouts",
		},
		{
			name:        "successful_hook",
			hookName:    "successHook",
			content:     "console.log('Hook executed successfully'); return 'success';",
			expectError: false,
			description: "Should execute successful hooks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "error_test",
			}

			if err := engine.LoadScript(script); err != nil {
				if tt.expectError {
					t.Logf("Expected error during script loading: %v", err)
					return
				}
				t.Fatalf("Failed to load script: %v", err)
			}

			err := engine.ExecuteHook(tt.hookName, map[string]interface{}{"test": "data"})

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}

			if err != nil {
				t.Logf("Hook error (expected): %v", err)
			}
		})
	}
}

func TestHooks_ConcurrentExecution(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         4,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 创建多个并发Hook脚本
	for i := 0; i < 5; i++ {
		script := &Script{
			ID:       fmt.Sprintf("concurrent_hook_%d", i),
			Name:     fmt.Sprintf("Concurrent Hook %d", i),
			Content:  fmt.Sprintf("console.log('Concurrent hook %d executed'); const start = Date.now(); while(Date.now() - start < 100) {} console.log('Hook %d completed');", i, i),
			Enabled:  true,
			Category: "concurrent_test",
		}

		if err := engine.LoadScript(script); err != nil {
			t.Fatalf("Failed to load concurrent script %d: %v", i, err)
		}
	}

	// 并发执行Hook
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(hookID int) {
			hookName := fmt.Sprintf("concurrentHook%d", hookID)
			eventData := map[string]interface{}{"hookID": hookID, "timestamp": time.Now().Unix()}
			err := engine.ExecuteHook(hookName, eventData)
			done <- err
		}(i)
	}

	// 等待所有Hook完成
	for i := 0; i < 5; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Logf("Concurrent hook %d error: %v", i, err)
			}
		case <-time.After(10 * time.Second):
			t.Errorf("Timeout waiting for concurrent hook %d", i)
		}
	}

	t.Log("Concurrent hook execution test completed")
}

func BenchmarkHooks_Execution(b *testing.B) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		b.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         8,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelError,
	}
	engine := NewScriptEngine(testApp, config)

	// 基准测试Hook脚本
	benchScript := &Script{
		ID:       "benchmark_hook",
		Name:     "Benchmark Hook",
		Content:  "const result = Math.sqrt($input.value || 100); console.log('Hook result:', result);",
		Enabled:  true,
		Category: "benchmark",
	}

	if err := engine.LoadScript(benchScript); err != nil {
		b.Fatalf("Failed to load benchmark hook script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		eventData := map[string]interface{}{"value": i, "iteration": i}
		err := engine.ExecuteHook("benchmarkHook", eventData)
		if err != nil {
			b.Errorf("Unexpected error in benchmark hook: %v", err)
		}
	}
}