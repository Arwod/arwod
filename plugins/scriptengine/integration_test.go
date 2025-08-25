package scriptengine

import (
	"fmt"
	"testing"
	"time"
	"log/slog"

	"github.com/pocketbase/pocketbase/tests"
)

func TestIntegration_ErrorRecovery(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         2,
		MaxExecutionTime: 10 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 测试完整的脚本生命周期
	t.Run("script_lifecycle", func(t *testing.T) {
		// 1. 创建脚本
		script := &Script{
			ID:       "integration_test_script",
			Name:     "Integration Test Script",
			Content:  "console.log('Integration test script executed'); const result = $input.value * 2; console.log('Result:', result); return result;",
			Enabled:  true,
			Category: "hooks",
		}

		// 2. 加载脚本
		if err := engine.LoadScript(script); err != nil {
			t.Fatalf("Failed to load script: %v", err)
		}

		// 3. 执行脚本
		input := map[string]interface{}{"value": 21}
		result, err := engine.ExecuteScript(script.ID, input)
		if err != nil {
			t.Fatalf("Failed to execute script: %v", err)
		}

		// 4. 验证结果
		if result == nil {
			t.Fatal("Expected result but got nil")
		}

		t.Logf("Script execution result: %v", result)

		// 5. 更新脚本
		script.Content = "console.log('Updated script'); return $input.value * 3;"
		if err = engine.LoadScript(script); err != nil {
			t.Fatalf("Failed to reload updated script: %v", err)
		}

		// 6. 再次执行更新后的脚本
        result2, err := engine.ExecuteScript(script.ID, input)
        if err != nil {
            t.Fatalf("Failed to execute updated script: %v", err)
        }

        t.Logf("Updated script execution result: %v", result2)

		// 7. 卸载脚本
		engine.UnloadScript(script.ID)

		// 8. 验证脚本已卸载
		_, err = engine.ExecuteScript(script.ID, input)
		if err == nil {
			t.Error("Expected error when executing unloaded script")
		}
	})
}

func TestIntegration_ConcurrentExecution(t *testing.T) {
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

	// 创建多个不同类型的脚本
	scripts := []*Script{
		{
			ID:       "math_script",
			Name:     "Math Operations",
			Content:  "const result = Math.sqrt($input.number); console.log('Square root:', result); return result;",
			Enabled:  true,
			Category: "hooks",
		},
		{
			ID:       "string_script",
			Name:     "String Operations",
			Content:  "const result = $input.text.toUpperCase(); console.log('Uppercase:', result); return result;",
			Enabled:  true,
			Category: "hooks",
		},
		{
			ID:       "array_script",
			Name:     "Array Operations",
			Content:  "const result = $input.array.map(x => x * 2); console.log('Doubled array:', result); return result;",
			Enabled:  true,
			Category: "hooks",
		},
		{
			ID:       "date_script",
			Name:     "Date Operations",
			Content:  "const result = new Date().toISOString(); console.log('Current date:', result); return result;",
			Enabled:  true,
			Category: "hooks",
		},
	}

	// 加载所有脚本
	for _, script := range scripts {
		if err := engine.LoadScript(script); err != nil {
			t.Fatalf("Failed to load script %s: %v", script.ID, err)
		}
	}

	// 并发执行所有脚本
	done := make(chan error, len(scripts))

	for i, script := range scripts {
		go func(idx int, s *Script) {
			var input map[string]interface{}
			switch s.ID {
			case "math_script":
				input = map[string]interface{}{"number": 16}
			case "string_script":
				input = map[string]interface{}{"text": "hello world"}
			case "array_script":
				input = map[string]interface{}{"array": []int{1, 2, 3, 4, 5}}
			case "date_script":
				input = map[string]interface{}{}
			}

			result, err := engine.ExecuteScript(s.ID, input)
			if err != nil {
				done <- fmt.Errorf("script %s failed: %v", s.ID, err)
				return
			}

			t.Logf("Script %s result: %v", s.ID, result)
			done <- nil
		}(i, script)
	}

	// 等待所有脚本完成
	for i := 0; i < len(scripts); i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			t.Errorf("Timeout waiting for script %d", i)
		}
	}
}

func TestIntegration_HookInteraction(t *testing.T) {
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

	// 创建Hook脚本
	hookScript := &Script{
		ID:       "data_processor_hook",
		Name:     "Data Processor Hook",
		Content:  "console.log('Processing data in hook'); if ($record && $record.data) { $record.data.processed = true; $record.data.processedAt = new Date().toISOString(); }",
		Enabled:  true,
		Category: "hooks",
	}

	// 创建普通脚本
	normalScript := &Script{
		ID:       "data_validator",
		Name:     "Data Validator",
		Content:  "console.log('Validating data'); const isValid = $input.data && typeof $input.data === 'object'; console.log('Validation result:', isValid); return { valid: isValid, timestamp: new Date().toISOString() };",
		Enabled:  true,
		Category: "hooks",
	}

	// 加载脚本
	if err := engine.LoadScript(hookScript); err != nil {
		t.Fatalf("Failed to load hook script: %v", err)
	}

	if err := engine.LoadScript(normalScript); err != nil {
		t.Fatalf("Failed to load normal script: %v", err)
	}

	// 执行Hook
	hookData := map[string]interface{}{
		"record": map[string]interface{}{
			"id":   "test_record",
			"data": map[string]interface{}{"name": "Test Record"},
		},
	}

	if err := engine.ExecuteHook("dataProcessorHook", hookData); err != nil {
		t.Fatalf("Failed to execute hook: %v", err)
	}

	// 执行普通脚本
	validationInput := map[string]interface{}{
		"data": map[string]interface{}{"processed": true},
	}

	result, err := engine.ExecuteScript(normalScript.ID, validationInput)
	if err != nil {
		t.Fatalf("Failed to execute validation script: %v", err)
	}

	t.Logf("Validation result: %v", result)

	// 验证结果
	if result == nil {
		t.Fatal("Expected validation result but got nil")
	}
}

func TestIntegration_ScriptLifecycle(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         2,
		MaxExecutionTime: 3 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 创建会出错的脚本
	errorScript := &Script{
		ID:       "error_script",
		Name:     "Error Script",
		Content:  "throw new Error('Intentional error for testing');",
		Enabled:  true,
		Category: "hooks",
	}

	// 创建正常的脚本
	normalScript := &Script{
		ID:       "normal_script",
		Name:     "Normal Script",
		Content:  "console.log('Normal script executed successfully'); return 'success';",
		Enabled:  true,
		Category: "hooks",
	}

	// 加载脚本
	if err := engine.LoadScript(errorScript); err != nil {
		t.Fatalf("Failed to load error script: %v", err)
	}

	if err := engine.LoadScript(normalScript); err != nil {
		t.Fatalf("Failed to load normal script: %v", err)
	}

	// 执行出错的脚本
	_, err = engine.ExecuteScript(errorScript.ID, map[string]interface{}{})
	if err == nil {
		t.Error("Expected error from error script but got none")
	} else {
		t.Logf("Expected error from error script: %v", err)
	}

	// 验证引擎仍然可以执行正常脚本
	result, err := engine.ExecuteScript(normalScript.ID, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute normal script after error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result from normal script but got nil")
	}

	t.Logf("Normal script result after error recovery: %v", result)

	// 多次执行正常脚本以确保引擎稳定
	for i := 0; i < 5; i++ {
		result, err := engine.ExecuteScript(normalScript.ID, map[string]interface{}{"iteration": i})
		if err != nil {
			t.Fatalf("Failed to execute normal script iteration %d: %v", i, err)
		}
		t.Logf("Iteration %d result: %v", i, result)
	}
}

func TestIntegration_ResourceManagement(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         3,
		MaxExecutionTime: 2 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelInfo,
	}
	engine := NewScriptEngine(testApp, config)

	// 创建资源密集型脚本
	resourceScript := &Script{
		ID:       "resource_script",
		Name:     "Resource Intensive Script",
		Content:  "console.log('Starting resource intensive task'); const arr = []; for(let i = 0; i < 1000; i++) { arr.push(Math.random()); } console.log('Generated array with', arr.length, 'elements'); return arr.length;",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(resourceScript); err != nil {
		t.Fatalf("Failed to load resource script: %v", err)
	}

	// 并发执行多个资源密集型任务
	done := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func(taskID int) {
			input := map[string]interface{}{"taskID": taskID}
			result, err := engine.ExecuteScript(resourceScript.ID, input)
			if err != nil {
				done <- fmt.Errorf("task %d failed: %v", taskID, err)
				return
			}
			t.Logf("Task %d completed with result: %v", taskID, result)
			done <- nil
		}(i)
	}

	// 等待所有任务完成
	for i := 0; i < 5; i++ {
		select {
		case err := <-done:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(10 * time.Second):
			t.Errorf("Timeout waiting for resource task %d", i)
		}
	}

	// 验证引擎仍然响应
	simpleScript := &Script{
		ID:       "simple_script",
		Name:     "Simple Script",
		Content:  "return 'engine_responsive';",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(simpleScript); err != nil {
		t.Fatalf("Failed to load simple script: %v", err)
	}

	result, err := engine.ExecuteScript(simpleScript.ID, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to execute simple script after resource test: %v", err)
	}

	if result == nil || result.Output != "engine_responsive" {
		t.Errorf("Expected 'engine_responsive' but got: %v", result)
	}

	t.Log("Resource management test completed successfully")
}

func BenchmarkIntegration_FullWorkflow(b *testing.B) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		b.Fatalf("Failed to create test app: %v", err)
	}
	defer testApp.Cleanup()

	config := Config{
		PoolSize:         4,
		MaxExecutionTime: 5 * time.Second,
		EnableSandbox:    true,
		LogLevel:         slog.LevelError, // 减少日志输出以提高基准测试性能
	}
	engine := NewScriptEngine(testApp, config)

	// 基准测试脚本
	benchScript := &Script{
		ID:       "benchmark_integration_script",
		Name:     "Benchmark Integration Script",
		Content:  "const result = Math.sqrt($input.value || 100) * 2; return { result: result, timestamp: Date.now() };",
		Enabled:  true,
		Category: "hooks",
	}

	if err := engine.LoadScript(benchScript); err != nil {
		b.Fatalf("Failed to load benchmark script: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := map[string]interface{}{"value": i, "iteration": i}
		_, err := engine.ExecuteScript(benchScript.ID, input)
		if err != nil {
			b.Errorf("Unexpected error in benchmark iteration %d: %v", i, err)
		}
	}
}