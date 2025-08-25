// Package scriptengine implements a JavaScript script management engine for PocketBase.
// It provides script execution, management, and monitoring capabilities.
//
// Example:
//
//	scriptengine.MustRegister(app, scriptengine.Config{
//		PoolSize: 10,
//		MaxExecutionTime: 30 * time.Second,
//	})
package scriptengine

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/router"
)

// ScriptEngine manages JavaScript script execution and lifecycle
type ScriptEngine struct {
	app       core.App
	config    Config
	vmPool    *VMPool
	cronJobs  *cron.Cron
	scripts   map[string]*Script
	scriptsMu sync.RWMutex
	logger    *slog.Logger
}

// Config defines the configuration for the script engine
type Config struct {
	// PoolSize specifies the number of pre-warmed JavaScript VMs
	PoolSize int

	// MaxExecutionTime specifies the maximum execution time for scripts
	MaxExecutionTime time.Duration

	// EnableSandbox enables security sandbox for script execution
	EnableSandbox bool

	// LogLevel specifies the logging level
	LogLevel slog.Level

	// OnScriptError is called when a script execution error occurs
	OnScriptError func(script *Script, err error)
}

// Script represents a JavaScript script with metadata
type Script struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Content     string                 `json:"content"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	SourceType  string                 `json:"source_type"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ExecutionResult represents the result of script execution
type ExecutionResult struct {
	ScriptID    string                 `json:"script_id"`
	Status      string                 `json:"status"` // success, error, timeout
	Duration    time.Duration          `json:"duration"`
	Output      interface{}            `json:"output"`
	Error       string                 `json:"error,omitempty"`
	StackTrace  string                 `json:"stack_trace,omitempty"`
	InputData   map[string]interface{} `json:"input_data,omitempty"`
	ExecutedAt  time.Time              `json:"executed_at"`
}

// NewScriptEngine creates a new script engine instance
func NewScriptEngine(app core.App, config Config) *ScriptEngine {
	// Set default values
	if config.PoolSize <= 0 {
		config.PoolSize = 5
	}
	if config.MaxExecutionTime <= 0 {
		config.MaxExecutionTime = 30 * time.Second
	}

	engine := &ScriptEngine{
		app:      app,
		config:   config,
		scripts:  make(map[string]*Script),
		cronJobs: cron.New(),
		logger:   app.Logger().With("component", "scriptengine"),
	}

	// Initialize VM pool
	engine.vmPool = NewVMPool(config.PoolSize, engine.createVM)

	return engine
}

// Start initializes and starts the script engine
func (e *ScriptEngine) Start() error {
	e.logger.Info("Starting script engine")

	// Load scripts from database
	if err := e.loadScripts(); err != nil {
		return fmt.Errorf("failed to load scripts: %w", err)
	}

	// Start cron scheduler
	e.cronJobs.Start()

	e.logger.Info("Script engine started successfully")
	return nil
}

// Stop gracefully stops the script engine
func (e *ScriptEngine) Stop() error {
	e.logger.Info("Stopping script engine")

	// Stop cron scheduler
	e.cronJobs.Stop()

	// Clear scripts
	e.scriptsMu.Lock()
	e.scripts = make(map[string]*Script)
	e.scriptsMu.Unlock()

	e.logger.Info("Script engine stopped")
	return nil
}

// ExecuteScript executes a script by ID with optional input data
func (e *ScriptEngine) ExecuteScript(scriptID string, inputData map[string]interface{}) (*ExecutionResult, error) {
	e.scriptsMu.RLock()
	script, exists := e.scripts[scriptID]
	e.scriptsMu.RUnlock()

	// If script not found in memory, try to load from database
	if !exists {
		if err := e.loadScriptFromDatabase(scriptID); err != nil {
			return nil, fmt.Errorf("script not found: %s", scriptID)
		}
		
		// Try again after loading
		e.scriptsMu.RLock()
		script, exists = e.scripts[scriptID]
		e.scriptsMu.RUnlock()
		
		if !exists {
			return nil, fmt.Errorf("script not found: %s", scriptID)
		}
	}

	if !script.Enabled {
		return nil, fmt.Errorf("script is disabled: %s", scriptID)
	}

	return e.executeScript(script, inputData)
}

// ExecuteHook executes scripts for a specific hook event
func (e *ScriptEngine) ExecuteHook(hookName string, eventData interface{}) error {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	for _, script := range e.scripts {
		if !script.Enabled || script.Category != "hooks" {
			continue
		}

		// Check if script is for this hook
		if hookType, ok := script.Metadata["hook_type"].(string); ok && hookType == hookName {
			go func(s *Script) {
				if _, err := e.executeScript(s, map[string]interface{}{"event": eventData}); err != nil {
					e.logger.Error("Hook script execution failed", "script", s.Name, "hook", hookName, "error", err)
					if e.config.OnScriptError != nil {
						e.config.OnScriptError(s, err)
					}
				}
			}(script)
		}
	}

	return nil
}

// RegisterRoutes registers script-based routes
func (e *ScriptEngine) RegisterRoutes(r *router.RouterGroup[*core.RequestEvent]) error {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	for _, script := range e.scripts {
		if !script.Enabled || script.Category != "router" {
			continue
		}

		// Extract route information from metadata
		method, ok := script.Metadata["method"].(string)
		if !ok {
			method = "GET"
		}

		path, ok := script.Metadata["path"].(string)
		if !ok {
			continue
		}

		// Register route handler
		handler := e.createRouteHandler(script)
		r.Route(method, path, handler)
	}

	// Register management API routes
	e.RegisterScriptAPIs(r)

	return nil
}

// StartCronScheduler starts the cron scheduler for scheduled scripts
func (e *ScriptEngine) StartCronScheduler() error {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	for _, script := range e.scripts {
		if !script.Enabled || script.Category != "cron" {
			continue
		}

		// Extract cron expression from metadata
		cronExpr, ok := script.Metadata["cron_expression"].(string)
		if !ok {
			continue
		}

		// Add cron job
		if err := e.cronJobs.Add(script.ID, cronExpr, func() {
			if _, err := e.executeScript(script, nil); err != nil {
				e.logger.Error("Cron script execution failed", "script", script.Name, "error", err)
				if e.config.OnScriptError != nil {
					e.config.OnScriptError(script, err)
				}
			}
		}); err != nil {
			e.logger.Error("Failed to add cron job", "script", script.Name, "error", err)
		}
	}

	return nil
}

// LoadScript loads or reloads a script
func (e *ScriptEngine) LoadScript(script *Script) error {
	e.scriptsMu.Lock()
	defer e.scriptsMu.Unlock()

	// Validate script
	if err := e.validateScript(script); err != nil {
		return fmt.Errorf("script validation failed: %w", err)
	}

	// Store script
	e.scripts[script.ID] = script

	e.logger.Info("Script loaded", "name", script.Name, "category", script.Category)
	return nil
}

// UnloadScript removes a script from the engine
func (e *ScriptEngine) UnloadScript(scriptID string) error {
	e.scriptsMu.Lock()
	defer e.scriptsMu.Unlock()

	script, exists := e.scripts[scriptID]
	if !exists {
		return fmt.Errorf("script not found: %s", scriptID)
	}

	// Remove from cron if it's a cron script
	if script.Category == "cron" {
		e.cronJobs.Remove(scriptID)
	}

	// Remove from scripts map
	delete(e.scripts, scriptID)

	e.logger.Info("Script unloaded", "name", script.Name)
	return nil
}

// GetScript retrieves a script by ID
func (e *ScriptEngine) GetScript(scriptID string) (*Script, error) {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	script, exists := e.scripts[scriptID]
	if !exists {
		return nil, fmt.Errorf("script not found: %s", scriptID)
	}

	return script, nil
}

// ListScripts returns all loaded scripts
func (e *ScriptEngine) ListScripts() []*Script {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	scripts := make([]*Script, 0, len(e.scripts))
	for _, script := range e.scripts {
		scripts = append(scripts, script)
	}

	return scripts
}

// GetStats returns engine statistics
func (e *ScriptEngine) GetStats() map[string]interface{} {
	e.scriptsMu.RLock()
	defer e.scriptsMu.RUnlock()

	stats := map[string]interface{}{
		"total_scripts":   len(e.scripts),
		"enabled_scripts": 0,
		"pool_size":       e.config.PoolSize,
		"categories":      make(map[string]int),
	}

	categories := make(map[string]int)
	enabledCount := 0

	for _, script := range e.scripts {
		if script.Enabled {
			enabledCount++
		}
		categories[script.Category]++
	}

	stats["enabled_scripts"] = enabledCount
	stats["categories"] = categories

	return stats
}

// MustRegister registers the scriptengine plugin in the provided app instance.
// It panics if the registration fails.
func MustRegister(app core.App, config Config) {
	if err := Register(app, config); err != nil {
		panic(err)
	}
}

// Register registers the scriptengine plugin in the provided app instance.
func Register(app core.App, config Config) error {
	// Set default values
	if config.PoolSize <= 0 {
		config.PoolSize = 5
	}
	if config.MaxExecutionTime <= 0 {
		config.MaxExecutionTime = 30 * time.Second
	}
	if config.LogLevel == 0 {
		config.LogLevel = slog.LevelInfo
	}

	// Create script engine instance
	engine := NewScriptEngine(app, config)

	// Register on app bootstrap
	app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		err := e.Next()
		if err != nil {
			return err
		}

		// Start the script engine
		if err := engine.Start(); err != nil {
			return fmt.Errorf("failed to start script engine: %w", err)
		}

		return nil
	})

	// Register API routes on serve
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Register script engine API routes BEFORE calling Next()
		if err := engine.RegisterRoutes(e.Router.Group("/api/scripts")); err != nil {
			return fmt.Errorf("failed to register script engine routes: %w", err)
		}

		// Continue with the serve process
		return e.Next()
	})

	// Register cleanup on terminate
	app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		err := e.Next()
		if err != nil {
			return err
		}

		// Stop the script engine
		if err := engine.Stop(); err != nil {
			return fmt.Errorf("failed to stop script engine: %w", err)
		}

		return nil
	})

	return nil
}