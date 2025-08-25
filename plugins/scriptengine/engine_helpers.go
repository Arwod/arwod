package scriptengine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dop251/goja"
	"github.com/pocketbase/pocketbase/core"
)

// createVM creates a new JavaScript VM instance with PocketBase bindings
func (e *ScriptEngine) createVM() *goja.Runtime {
	// Create sandbox with default secure configuration
	sandboxConfig := DefaultSandboxConfig()
	sandbox := NewSandbox(sandboxConfig)

	// Apply security restrictions
	if err := sandbox.ApplyRestrictions(); err != nil {
		// Log error but continue with basic VM
		vm := goja.New()
		e.addPocketBaseBindings(vm)
		e.applySandboxRestrictions(vm)
		return vm
	}

	vm := sandbox.GetVM()

	// Add PocketBase-specific bindings
	e.addPocketBaseBindings(vm)

	return vm
}

// addPocketBaseBindings adds PocketBase API bindings to the VM
func (e *ScriptEngine) addPocketBaseBindings(vm *goja.Runtime) {
	// Add console object for logging
	vm.Set("console", map[string]interface{}{
		"log": func(args ...interface{}) {
			e.logger.Info("Script console.log", "args", args)
		},
		"error": func(args ...interface{}) {
			e.logger.Error("Script console.error", "args", args)
		},
		"warn": func(args ...interface{}) {
			e.logger.Warn("Script console.warn", "args", args)
		},
		"info": func(args ...interface{}) {
			e.logger.Info("Script console.info", "args", args)
		},
	})

	// Add app instance
	vm.Set("$app", e.app)

	// Add database access
	vm.Set("$db", e.app.DB())

	// Add logger
	vm.Set("$logger", e.logger)

	// Add utility functions
	vm.Set("$utils", map[string]interface{}{
		"now": func() time.Time {
			return time.Now()
		},
		"uuid": func() string {
			return core.GenerateDefaultRandomId()
		},
		"hash": func(data string) string {
			return fmt.Sprintf("%x", data) // Simple hash for demo
		},
	})

	// Add HTTP client
	vm.Set("$http", map[string]interface{}{
		"get": func(url string) (map[string]interface{}, error) {
			resp, err := http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return nil, err
			}

			return result, nil
		},
	})
}

// applySandboxRestrictions applies security restrictions to the VM
func (e *ScriptEngine) applySandboxRestrictions(vm *goja.Runtime) {
	// Fallback basic restrictions when sandbox fails
	vm.Set("eval", goja.Undefined())
	vm.Set("Function", goja.Undefined())
	vm.Set("setTimeout", goja.Undefined())
	vm.Set("setInterval", goja.Undefined())
	vm.Set("require", goja.Undefined())
	vm.Set("import", goja.Undefined())
	vm.Set("process", goja.Undefined())
	vm.Set("global", goja.Undefined())
	vm.Set("globalThis", goja.Undefined())

	// Set call stack limit
	vm.SetMaxCallStackSize(1000)
}

// loadScripts loads scripts from the database
func (e *ScriptEngine) loadScripts() error {
	// Query scripts from database
	records, err := e.app.FindRecordsByFilter(
		"js_scripts",
		"status = 'active'",
		"-version",
		0,
		0,
	)
	if err != nil {
		// If collection doesn't exist yet, just log and continue
		e.logger.Warn("Scripts collection not found, skipping script loading", "error", err)
		return nil
	}

	e.logger.Info("Found scripts to load", "count", len(records))

	// Convert records to scripts
	for _, record := range records {
		script := &Script{
			ID:         record.Id,
			Name:       record.GetString("name"),
			Category:   record.GetString("trigger_type"),
			Content:    record.GetString("content"),
			Enabled:    record.GetString("status") == "active",
			Priority:   record.GetInt("version"),
			SourceType: "database",
			CreatedAt:  record.GetDateTime("created").Time(),
			UpdatedAt:  record.GetDateTime("updated").Time(),
		}

		// Parse metadata
		if metadataStr := record.GetString("metadata"); metadataStr != "" {
			var metadata map[string]interface{}
			if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
				script.Metadata = metadata
			}
		}

		e.logger.Info("Attempting to load script", "name", script.Name, "category", script.Category, "id", script.ID)
		if err := e.LoadScript(script); err != nil {
			e.logger.Error("Failed to load script", "name", script.Name, "id", script.ID, "category", script.Category, "error", err)
		} else {
			e.logger.Info("Successfully loaded script", "name", script.Name, "id", script.ID)
		}
	}

	return nil
}

// loadScriptFromDatabase loads a single script from database by ID
func (e *ScriptEngine) loadScriptFromDatabase(scriptID string) error {
	// Query specific script from database
	record, err := e.app.FindRecordById("js_scripts", scriptID)
	if err != nil {
		return fmt.Errorf("script not found in database: %s", scriptID)
	}

	// Convert record to script
	script := &Script{
		ID:         record.Id,
		Name:       record.GetString("name"),
		Category:   record.GetString("trigger_type"),
		Content:    record.GetString("content"),
		Enabled:    record.GetString("status") == "active",
		Priority:   record.GetInt("version"),
		SourceType: "database",
		CreatedAt:  record.GetDateTime("created").Time(),
		UpdatedAt:  record.GetDateTime("updated").Time(),
	}

	// Parse metadata
	if metadataStr := record.GetString("metadata"); metadataStr != "" {
		var metadata map[string]interface{}
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err == nil {
			script.Metadata = metadata
		}
	}

	// Load script into memory
	return e.LoadScript(script)
}

// executeScript executes a script with the given input data
func (e *ScriptEngine) executeScript(script *Script, inputData map[string]interface{}) (*ExecutionResult, error) {
	result := &ExecutionResult{
		ScriptID:   script.ID,
		InputData:  inputData,
		ExecutedAt: time.Now(),
	}

	start := time.Now()
	defer func() {
		result.Duration = time.Since(start)
	}()

	// Execute script in VM pool
	err := e.vmPool.Run(func(vm *goja.Runtime) error {
		// Set input data
		if inputData != nil {
			vm.Set("$input", inputData)
		}

		// Create execution context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), e.config.MaxExecutionTime)
		defer cancel()

		// Execute script with timeout
		done := make(chan struct{})
		var execErr error
		var output goja.Value

		go func() {
			defer close(done)
			output, execErr = vm.RunString(script.Content)
		}()

		select {
		case <-done:
			if execErr != nil {
				return execErr
			}
			result.Output = output.Export()
			return nil
		case <-ctx.Done():
			return errors.New("script execution timeout")
		}
	})

	if err != nil {
		result.Status = "error"
		result.Error = err.Error()
		// TODO: Extract stack trace from goja error
		result.StackTrace = ""
	} else {
		result.Status = "success"
	}

	// Log execution result
	e.logExecution(result)

	return result, err
}

// validateScript validates the script syntax and security
func (e *ScriptEngine) validateScript(script *Script) error {
	if script == nil {
		return errors.New("script is nil")
	}

	if script.Content == "" {
		return errors.New("script content is empty")
	}

	if script.ID == "" {
		return errors.New("script ID is required")
	}
	if script.Name == "" {
		return errors.New("script name is required")
	}
	if script.Category == "" {
		return errors.New("script category is required")
	}

	// Validate category
	validCategories := []string{"manual", "hook", "router", "cron", "dbx", "mails", "security", "filesystem", "filepath", "os", "forms", "apis", "http"}
	validCategory := false
	for _, cat := range validCategories {
		if script.Category == cat {
			validCategory = true
			break
		}
	}
	if !validCategory {
		return fmt.Errorf("invalid script category: %s", script.Category)
	}

	// Security validation using sandbox
	sandboxConfig := DefaultSandboxConfig()
	sandbox := NewSandbox(sandboxConfig)
	if err := sandbox.ValidateScript(script.Content); err != nil {
		return fmt.Errorf("security validation failed: %w", err)
	}

	// Basic syntax validation
	vm := goja.New()
	// Add console object for syntax validation
	vm.Set("console", map[string]interface{}{
		"log":   func(args ...interface{}) {},
		"error": func(args ...interface{}) {},
		"warn":  func(args ...interface{}) {},
		"info":  func(args ...interface{}) {},
	})
	// Add $input mock object for syntax validation
	vm.Set("$input", map[string]interface{}{})
	// Add other common variables that might be used in scripts
	vm.Set("$app", map[string]interface{}{})
	vm.Set("$record", map[string]interface{}{})
	vm.Set("$admin", map[string]interface{}{})
	vm.Set("$authRecord", map[string]interface{}{})
	_, err := vm.RunString(script.Content)
	if err != nil {
		return fmt.Errorf("script syntax validation failed: %w", err)
	}

	return nil
}

// createRouteHandler creates an HTTP handler for a route script
func (e *ScriptEngine) createRouteHandler(script *Script) func(e *core.RequestEvent) error {
	return func(e2 *core.RequestEvent) error {
		// Prepare input data
		inputData := map[string]interface{}{
			"request": map[string]interface{}{
				"method":  e2.Request.Method,
				"url":     e2.Request.URL.String(),
				"headers": e2.Request.Header,
			},
			"response": map[string]interface{}{
				"status": 200,
				"body":   "",
			},
		}

		// Execute script
		result, err := e.executeScript(script, inputData)
		if err != nil {
			return err
		}

		// Handle response
		if result.Output != nil {
			if response, ok := result.Output.(map[string]interface{}); ok {
				if status, ok := response["status"].(int); ok {
					e2.Response.WriteHeader(status)
				}
				if body, ok := response["body"].(string); ok {
					e2.Response.Write([]byte(body))
				}
			}
		}

		return nil
	}
}

// logExecution logs script execution results
func (e *ScriptEngine) logExecution(result *ExecutionResult) {
	// TODO: Store execution log in database
	if result.Status == "success" {
		e.logger.Info("Script executed successfully",
			"script_id", result.ScriptID,
			"duration", result.Duration,
		)
	} else {
		e.logger.Error("Script execution failed",
			"script_id", result.ScriptID,
			"duration", result.Duration,
			"error", result.Error,
		)
	}
}
