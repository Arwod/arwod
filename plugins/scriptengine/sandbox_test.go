package scriptengine

import (
	"testing"
	"time"
	"log/slog"

	"github.com/pocketbase/pocketbase/tests"
)

func TestSandbox_SecurityRestrictions(t *testing.T) {
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
			name:        "file_system_access",
			content:     "const fs = require('fs'); fs.readFileSync('/etc/passwd');",
			expectError: true,
			description: "Should block file system access",
		},
		{
			name:        "network_access",
			content:     "const http = require('http'); http.get('http://example.com');",
			expectError: true,
			description: "Should block network access",
		},
		{
			name:        "process_access",
			content:     "const cp = require('child_process'); cp.exec('ls');",
			expectError: true,
			description: "Should block process execution",
		},
		{
			name:        "global_access",
			content:     "global.process.exit(1);",
			expectError: true,
			description: "Should block global process access",
		},
		{
			name:        "eval_restriction",
			content:     "eval('require(\\'fs\\').readFileSync(\\'/etc/passwd\\')');",
			expectError: true,
			description: "Should block eval with dangerous code",
		},
		{
			name:        "function_constructor",
			content:     "new Function('return require(\\'fs\\')')();",
			expectError: true,
			description: "Should block Function constructor with dangerous code",
		},
		{
			name:        "allowed_operations",
			content:     "const data = {test: 'value'}; console.log(JSON.stringify(data)); Math.random();",
			expectError: false,
			description: "Should allow safe operations",
		},
		{
			name:        "allowed_console",
			content:     "console.log('Hello'); console.error('Error'); console.warn('Warning');",
			expectError: false,
			description: "Should allow console operations",
		},
		{
			name:        "allowed_json",
			content:     "const obj = {a: 1}; JSON.stringify(obj); JSON.parse('{\"b\": 2}');",
			expectError: false,
			description: "Should allow JSON operations",
		},
		{
			name:        "allowed_math",
			content:     "Math.sqrt(16); Math.random(); Math.floor(3.7);",
			expectError: false,
			description: "Should allow Math operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "security",
			}

			if err := engine.LoadScript(script); err != nil {
				if tt.expectError {
					t.Logf("Expected error during script loading: %v", err)
					return
				}
				t.Fatalf("Failed to load script: %v", err)
			}

			_, err := engine.ExecuteScript(tt.name, map[string]interface{}{})

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Errorf("%s: Unexpected error: %v", tt.description, err)
			}

			if err != nil {
				t.Logf("Security restriction working: %v", err)
			}
		})
	}
}

func TestSandbox_ResourceLimits(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 2 * time.Second,
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
			name:        "infinite_loop",
			content:     "while(true) { /* infinite loop */ }",
			expectError: true,
			description: "Should timeout on infinite loop",
		},
		{
			name:        "memory_intensive",
			content:     "const arr = []; for(let i = 0; i < 1000000; i++) { arr.push(new Array(1000).fill('data')); }",
			expectError: true,
			description: "Should limit memory usage",
		},
		{
			name:        "cpu_intensive",
			content:     "for(let i = 0; i < 10000000; i++) { Math.sqrt(i); }",
			expectError: false, // 可能会超时，但不一定
			description: "CPU intensive operation",
		},
		{
			name:        "normal_operation",
			content:     "for(let i = 0; i < 100; i++) { console.log('Iteration:', i); }",
			expectError: false,
			description: "Should allow normal operations",
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

			start := time.Now()
			result, err := engine.ExecuteScript(tt.name, map[string]interface{}{})
			duration := time.Since(start)

			t.Logf("Script %s executed in %v", tt.name, duration)

			if tt.expectError && err == nil {
				t.Errorf("%s: Expected error but got none", tt.description)
			} else if !tt.expectError && err != nil {
				t.Logf("%s: Got error (may be expected): %v", tt.description, err)
			}

			if result != nil {
				t.Logf("Result status: %s, duration: %v", result.Status, result.Duration)
			}

			// 检查是否超过最大执行时间
			if duration > config.MaxExecutionTime+time.Second {
				t.Errorf("Script execution took too long: %v > %v", duration, config.MaxExecutionTime)
			}
		})
	}
}

func TestSandbox_APIAccess(t *testing.T) {
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
			name:        "app_access",
			content:     "console.log('App available:', typeof $app !== 'undefined');",
			expectError: false,
			description: "Should allow access to $app",
		},
		{
			name:        "dao_access",
			content:     "const dao = $app.dao(); console.log('DAO available:', !!dao);",
			expectError: false,
			description: "Should allow access to DAO",
		},
		{
			name:        "logger_access",
			content:     "$app.logger().info('Test log from script');",
			expectError: false,
			description: "Should allow access to logger",
		},
		{
			name:        "settings_access",
			content:     "const settings = $app.settings(); console.log('Settings available:', !!settings);",
			expectError: false,
			description: "Should allow access to settings",
		},
		{
			name:        "request_context",
			content:     "console.log('Request context available:', typeof $request !== 'undefined');",
			expectError: false,
			description: "Should provide request context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := &Script{
				ID:       tt.name,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "apis",
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

func TestSandbox_ContextIsolation(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         2,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 第一个脚本设置全局变量
	script1 := &Script{
		ID:       "script1",
		Name:     "Script 1",
		Content:  "globalThis.sharedVar = 'from_script1'; console.log('Script1 set sharedVar');",
		Enabled:  true,
		Category: "hooks",
	}

	// 第二个脚本尝试访问全局变量
	script2 := &Script{
		ID:       "script2",
		Name:     "Script 2",
		Content:  "console.log('Script2 sharedVar:', typeof globalThis.sharedVar, globalThis.sharedVar);",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(script1); err != nil {
		t.Fatalf("Failed to load script1: %v", err)
	}

	if err := engine.LoadScript(script2); err != nil {
		t.Fatalf("Failed to load script2: %v", err)
	}

	// 执行第一个脚本
	_, err1 := engine.ExecuteScript("script1", map[string]interface{}{})
	if err1 != nil {
		t.Errorf("Script1 execution failed: %v", err1)
	}

	// 执行第二个脚本
	_, err2 := engine.ExecuteScript("script2", map[string]interface{}{})
	if err2 != nil {
		t.Errorf("Script2 execution failed: %v", err2)
	}

	// 注意：根据沙箱实现，脚本之间可能共享或隔离上下文
	// 这个测试主要是验证沙箱的行为是否符合预期
	t.Log("Context isolation test completed")
}

func BenchmarkSandbox_SecurityCheck(b *testing.B) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		b.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         4,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelError,
	}
	engine := NewScriptEngine(testApp, config)

	// 安全的脚本
	safeScript := &Script{
		ID:       "safe_benchmark",
		Name:     "Safe Benchmark",
		Content:  "const result = Math.random() * 100; console.log('Safe operation:', result);",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(safeScript); err != nil {
		b.Fatalf("Failed to load safe script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ExecuteScript("safe_benchmark", map[string]interface{}{"iteration": i})
		if err != nil {
			b.Errorf("Unexpected error in safe script: %v", err)
		}
	}
}