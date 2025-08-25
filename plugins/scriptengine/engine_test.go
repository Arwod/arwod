package scriptengine

import (
	"fmt"
	"testing"
	"time"
	"log/slog"

	"github.com/pocketbase/pocketbase/tests"
)

func TestScriptEngine_ExecuteScript(t *testing.T) {
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

	// 首先加载测试脚本
	testScripts := []*Script{
		{
			ID:       "test1",
			Name:     "Simple Console Log",
			Content:  "console.log('Hello, World!');",
			Enabled:  true,
			Category: "hooks",
		},
		{
			ID:       "test2",
			Name:     "Syntax Error",
			Content:  "console.log('unclosed string",
			Enabled:  true,
			Category: "hooks",
		},
		{
			ID:       "test3",
			Name:     "Runtime Error",
			Content:  "throw new Error('Runtime error');",
			Enabled:  true,
			Category: "hooks",
		},
	}

	// 加载脚本到引擎
	for _, script := range testScripts {
		if err := engine.LoadScript(script); err != nil {
			t.Logf("Failed to load script %s: %v", script.ID, err)
		}
	}

	tests := []struct {
		name     string
		scriptID string
		input    map[string]interface{}
		wantErr  bool
	}{
		{
			name:     "simple console.log",
			scriptID: "test1",
			input:    map[string]interface{}{"test": "data"},
			wantErr:  false,
		},
		{
			name:     "syntax error",
			scriptID: "test2",
			input:    map[string]interface{}{},
			wantErr:  true,
		},
		{
			name:     "runtime error",
			scriptID: "test3",
			input:    map[string]interface{}{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.ExecuteScript(tt.scriptID, tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScript() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil {
				t.Logf("Script %s executed with status: %s, duration: %v", tt.scriptID, result.Status, result.Duration)
			}
		})
	}
}

func TestScriptEngine_ExecuteWithTimeout(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         1,
		MaxExecutionTime: 1 * time.Second, // 短超时时间
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 加载一个会超时的脚本
	timeoutScript := &Script{
		ID:       "timeout_test",
		Name:     "Timeout Test",
		Content:  "while(true) { /* infinite loop */ }",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(timeoutScript); err != nil {
		t.Logf("Failed to load timeout script: %v", err)
	}

	result, err := engine.ExecuteScript("timeout_test", map[string]interface{}{})

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}

	if result != nil && result.Status != "timeout" {
		t.Errorf("Expected timeout status, got %s", result.Status)
	}
}

func TestScriptEngine_SecurityRestrictions(t *testing.T) {
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

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "file system access",
			content: "require('fs').readFileSync('/etc/passwd');",
			wantErr: true,
		},
		{
			name:    "network access",
			content: "require('http').get('http://example.com');",
			wantErr: true,
		},
		{
			name:    "process access",
			content: "require('child_process').exec('ls');",
			wantErr: true,
		},
		{
			name:    "allowed operations",
			content: "const data = {test: 'value'}; console.log(JSON.stringify(data));",
			wantErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptID := fmt.Sprintf("security_test_%d", i)
			script := &Script{
				ID:       scriptID,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Logf("Failed to load script %s: %v", scriptID, err)
			}

			_, err := engine.ExecuteScript(scriptID, map[string]interface{}{})

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScriptEngine_MemoryLimits(t *testing.T) {
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

	// 尝试分配大量内存的脚本
	memoryScript := &Script{
		ID:       "memory_test",
		Name:     "Memory Test",
		Content:  "const arr = []; for(let i = 0; i < 100000; i++) { arr.push(new Array(100).fill('x')); } console.log('Memory allocated');",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(memoryScript); err != nil {
		t.Logf("Failed to load memory script: %v", err)
	}

	result, err := engine.ExecuteScript("memory_test", map[string]interface{}{})

	if err != nil {
		t.Logf("Script execution error (expected for memory limits): %v", err)
	}

	if result != nil {
		t.Logf("Script executed with status: %s, duration: %v", result.Status, result.Duration)
	}
}

func TestScriptEngine_APIBindings(t *testing.T) {
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

	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "access app instance",
			content: "console.log('App name:', $app.settings().meta.appName);",
			wantErr: false,
		},
		{
			name:    "access dao",
			content: "const dao = $app.dao(); console.log('DAO available:', !!dao);",
			wantErr: false,
		},
		{
			name:    "access logger",
			content: "$app.logger().info('Test log message');",
			wantErr: false,
		},
		{
			name:    "access request context",
			content: "console.log('Request available:', !!$request);",
			wantErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptID := fmt.Sprintf("api_test_%d", i)
			script := &Script{
				ID:       scriptID,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Logf("Failed to load script %s: %v", scriptID, err)
			}

			_, err := engine.ExecuteScript(scriptID, map[string]interface{}{})

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScript() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScriptEngine_HookExecution(t *testing.T) {
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

	// 测试不同类型的Hook脚本
	tests := []struct {
		name        string
		triggerType string
		content     string
		wantErr     bool
	}{
		{
			name:        "before create hook",
			triggerType: "hook",
			content:     "console.log('Before create:', $record.id); $record.set('created_by_script', true);",
			wantErr:     false,
		},
		{
			name:        "after update hook",
			triggerType: "hook",
			content:     "console.log('After update:', $record.id); $app.logger().info('Record updated');",
			wantErr:     false,
		},
		{
			name:        "cron job",
			triggerType: "cron",
			content:     "console.log('Cron job executed at:', new Date().toISOString());",
			wantErr:     false,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scriptID := fmt.Sprintf("hook_test_%d", i)
			script := &Script{
				ID:       scriptID,
				Name:     tt.name,
				Content:  tt.content,
				Enabled:  true,
				Category: "hooks",
			}

			if err := engine.LoadScript(script); err != nil {
				t.Logf("Failed to load script %s: %v", scriptID, err)
			}

			result, err := engine.ExecuteScript(scriptID, map[string]interface{}{"trigger_type": tt.triggerType})

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteScript() error = %v, wantErr %v", err, tt.wantErr)
			}

			if result != nil && result.Output == nil {
				t.Logf("Script executed but no output captured")
			}
		})
	}
}

func BenchmarkScriptEngine_Execute(b *testing.B) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		b.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         4,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelError, // 减少日志输出
	}
	engine := NewScriptEngine(testApp, config)

	// 加载基准测试脚本
	benchScript := &Script{
		ID:       "benchmark_test",
		Name:     "Benchmark Test",
		Content:  "const result = Math.sqrt(Math.random() * 1000); console.log('Result:', result);",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(benchScript); err != nil {
		b.Fatalf("Failed to load benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.ExecuteScript("benchmark_test", map[string]interface{}{"iteration": i})
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}