package scriptengine

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// SandboxConfig defines the configuration for script sandbox
type SandboxConfig struct {
	// MaxExecutionTime limits script execution time
	MaxExecutionTime time.Duration

	// MaxCallStackSize limits the call stack depth
	MaxCallStackSize int

	// MaxMemoryUsage limits memory usage (in bytes)
	MaxMemoryUsage int64

	// AllowedModules defines which modules can be imported
	AllowedModules []string

	// BlockedGlobals defines which global objects should be blocked
	BlockedGlobals []string

	// AllowNetworkAccess controls network access
	AllowNetworkAccess bool

	// AllowFileSystemAccess controls file system access
	AllowFileSystemAccess bool

	// AllowProcessAccess controls process access
	AllowProcessAccess bool
}

// DefaultSandboxConfig returns a default secure sandbox configuration
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		MaxExecutionTime:      30 * time.Second,
		MaxCallStackSize:      1000,
		MaxMemoryUsage:        50 * 1024 * 1024, // 50MB
		AllowedModules:        []string{"console", "JSON", "Math", "Date"},
		BlockedGlobals:        []string{"eval", "Function", "setTimeout", "setInterval", "require", "import", "process", "global", "globalThis"},
		AllowNetworkAccess:    false,
		AllowFileSystemAccess: false,
		AllowProcessAccess:    false,
	}
}

// Sandbox provides security isolation for JavaScript execution
type Sandbox struct {
	config SandboxConfig
	vm     *goja.Runtime
}

// NewSandbox creates a new sandbox with the given configuration
func NewSandbox(config SandboxConfig) *Sandbox {
	return &Sandbox{
		config: config,
		vm:     goja.New(),
	}
}

// ApplyRestrictions applies security restrictions to the VM
func (s *Sandbox) ApplyRestrictions() error {
	// Set call stack size limit
	s.vm.SetMaxCallStackSize(s.config.MaxCallStackSize)

	// Block dangerous globals
	for _, global := range s.config.BlockedGlobals {
		s.vm.Set(global, goja.Undefined())
	}

	// Override console to prevent information leakage
	s.setupSecureConsole()

	// Setup module restrictions
	s.setupModuleRestrictions()

	// Setup network restrictions
	if !s.config.AllowNetworkAccess {
		s.blockNetworkAccess()
	}

	// Setup file system restrictions
	if !s.config.AllowFileSystemAccess {
		s.blockFileSystemAccess()
	}

	// Setup process restrictions
	if !s.config.AllowProcessAccess {
		s.blockProcessAccess()
	}

	return nil
}

// ExecuteWithTimeout executes script with timeout and resource limits
func (s *Sandbox) ExecuteWithTimeout(ctx context.Context, script string) (goja.Value, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, s.config.MaxExecutionTime)
	defer cancel()

	// Channel to receive execution result
	resultChan := make(chan executionResult, 1)

	// Execute script in goroutine
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- executionResult{
					err: fmt.Errorf("script panic: %v", r),
				}
			}
		}()

		value, err := s.vm.RunString(script)
		resultChan <- executionResult{
			value: value,
			err:   err,
		}
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result.value, result.err
	case <-ctx.Done():
		return goja.Undefined(), fmt.Errorf("script execution timeout after %v", s.config.MaxExecutionTime)
	}
}

// GetVM returns the underlying Goja runtime
func (s *Sandbox) GetVM() *goja.Runtime {
	return s.vm
}

// setupSecureConsole sets up a secure console object
func (s *Sandbox) setupSecureConsole() {
	console := s.vm.NewObject()
	console.Set("log", func(args ...interface{}) {
		// Log to internal logger instead of stdout
		// This prevents information leakage
	})
	console.Set("error", func(args ...interface{}) {
		// Log errors to internal logger
	})
	console.Set("warn", func(args ...interface{}) {
		// Log warnings to internal logger
	})
	console.Set("info", func(args ...interface{}) {
		// Log info to internal logger
	})
	s.vm.Set("console", console)
}

// setupModuleRestrictions sets up module import restrictions
func (s *Sandbox) setupModuleRestrictions() {
	// Block require function
	s.vm.Set("require", func(module string) interface{} {
		if !s.isModuleAllowed(module) {
			panic(fmt.Sprintf("Module '%s' is not allowed", module))
		}
		return goja.Undefined()
	})

	// Block dynamic imports
	s.vm.Set("import", goja.Undefined())
}

// blockNetworkAccess blocks network-related functions
func (s *Sandbox) blockNetworkAccess() {
	// Block fetch API
	s.vm.Set("fetch", goja.Undefined())
	s.vm.Set("XMLHttpRequest", goja.Undefined())
	s.vm.Set("WebSocket", goja.Undefined())
}

// blockFileSystemAccess blocks file system access
func (s *Sandbox) blockFileSystemAccess() {
	// Block file system related globals
	s.vm.Set("readFile", goja.Undefined())
	s.vm.Set("writeFile", goja.Undefined())
	s.vm.Set("fs", goja.Undefined())
}

// blockProcessAccess blocks process access
func (s *Sandbox) blockProcessAccess() {
	// Block process related globals
	s.vm.Set("process", goja.Undefined())
	s.vm.Set("child_process", goja.Undefined())
	s.vm.Set("exec", goja.Undefined())
	s.vm.Set("spawn", goja.Undefined())
}

// isModuleAllowed checks if a module is in the allowed list
func (s *Sandbox) isModuleAllowed(module string) bool {
	for _, allowed := range s.config.AllowedModules {
		if matched, _ := regexp.MatchString(allowed, module); matched {
			return true
		}
	}
	return false
}

// ValidateScript performs static analysis on script content
func (s *Sandbox) ValidateScript(script string) error {
	// Check for dangerous patterns
	dangerousPatterns := []string{
		`eval\s*\(`,
		`Function\s*\(`,
		`setTimeout\s*\(`,
		`setInterval\s*\(`,
		`require\s*\(`,
		`import\s*\(`,
		`process\s*\.`,
		`global\s*\.`,
		`globalThis\s*\.`,
		`__proto__`,
		`constructor\s*\.`,
		`prototype\s*\.`,
	}

	for _, pattern := range dangerousPatterns {
		if matched, _ := regexp.MatchString(pattern, script); matched {
			return fmt.Errorf("dangerous pattern detected: %s", pattern)
		}
	}

	// Check script length
	if len(script) > 100000 { // 100KB limit
		return fmt.Errorf("script too large: %d bytes (max 100KB)", len(script))
	}

	// Check for excessive nesting
	if strings.Count(script, "{") > 1000 {
		return fmt.Errorf("excessive nesting detected")
	}

	return nil
}

// executionResult holds the result of script execution
type executionResult struct {
	value goja.Value
	err   error
}

// SecurityViolation represents a security violation during script execution
type SecurityViolation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Script      string `json:"script,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// Error implements the error interface
func (sv SecurityViolation) Error() string {
	return fmt.Sprintf("Security violation [%s]: %s", sv.Type, sv.Description)
}

// NewSecurityViolation creates a new security violation
func NewSecurityViolation(violationType, description, script string) *SecurityViolation {
	return &SecurityViolation{
		Type:        violationType,
		Description: description,
		Script:      script,
		Timestamp:   time.Now(),
	}
}