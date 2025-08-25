package scriptengine

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// HookManager manages script engine hook integrations
type HookManager struct {
	engine *ScriptEngine
	app    core.App
	logger *slog.Logger

	// Hook handler IDs for cleanup
	hookHandlers map[string][]string
}

// NewHookManager creates a new hook manager
func NewHookManager(engine *ScriptEngine, app core.App) *HookManager {
	return &HookManager{
		engine:       engine,
		app:          app,
		logger:       app.Logger(),
		hookHandlers: make(map[string][]string),
	}
}

// RegisterAllHooks registers script engine hooks for all PocketBase lifecycle events
func (hm *HookManager) RegisterAllHooks() error {
	// Application lifecycle hooks
	if err := hm.registerAppHooks(); err != nil {
		return fmt.Errorf("failed to register app hooks: %w", err)
	}

	// Model hooks
	if err := hm.registerModelHooks(); err != nil {
		return fmt.Errorf("failed to register model hooks: %w", err)
	}

	// Record hooks
	if err := hm.registerRecordHooks(); err != nil {
		return fmt.Errorf("failed to register record hooks: %w", err)
	}

	// Collection hooks
	if err := hm.registerCollectionHooks(); err != nil {
		return fmt.Errorf("failed to register collection hooks: %w", err)
	}

	// API request hooks
	if err := hm.registerAPIHooks(); err != nil {
		return fmt.Errorf("failed to register API hooks: %w", err)
	}

	// Mailer hooks
	if err := hm.registerMailerHooks(); err != nil {
		return fmt.Errorf("failed to register mailer hooks: %w", err)
	}

	// Realtime hooks
	if err := hm.registerRealtimeHooks(); err != nil {
		return fmt.Errorf("failed to register realtime hooks: %w", err)
	}

	// File hooks
	if err := hm.registerFileHooks(); err != nil {
		return fmt.Errorf("failed to register file hooks: %w", err)
	}

	hm.logger.Info("All script engine hooks registered successfully")
	return nil
}

// UnregisterAllHooks removes all registered hooks
func (hm *HookManager) UnregisterAllHooks() {
	for hookName, handlerIDs := range hm.hookHandlers {
		for _, handlerID := range handlerIDs {
			hm.removeHookHandler(hookName, handlerID)
		}
	}
	hm.hookHandlers = make(map[string][]string)
	hm.logger.Info("All script engine hooks unregistered")
}

// registerAppHooks registers application lifecycle hooks
func (hm *HookManager) registerAppHooks() error {
	// OnBootstrap hook
	bootstrapID := hm.app.OnBootstrap().BindFunc(func(e *core.BootstrapEvent) error {
		return hm.executeHookScripts("onBootstrap", map[string]interface{}{
			"app": e.App,
		})
	})
	hm.addHookHandler("onBootstrap", bootstrapID)

	// OnServe hook
	serveID := hm.app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		return hm.executeHookScripts("onServe", map[string]interface{}{
			"app":    e.App,
			"server": e.Server,
			"router": e.Router,
		})
	})
	hm.addHookHandler("onServe", serveID)

	// OnTerminate hook
	terminateID := hm.app.OnTerminate().BindFunc(func(e *core.TerminateEvent) error {
		return hm.executeHookScripts("onTerminate", map[string]interface{}{
			"app": e.App,
		})
	})
	hm.addHookHandler("onTerminate", terminateID)

	// OnBackupCreate hook
	backupCreateID := hm.app.OnBackupCreate().BindFunc(func(e *core.BackupEvent) error {
		return hm.executeHookScripts("onBackupCreate", map[string]interface{}{
			"app":  e.App,
			"name": e.Name,
		})
	})
	hm.addHookHandler("onBackupCreate", backupCreateID)

	// OnBackupRestore hook
	backupRestoreID := hm.app.OnBackupRestore().BindFunc(func(e *core.BackupEvent) error {
		return hm.executeHookScripts("onBackupRestore", map[string]interface{}{
			"app":  e.App,
			"name": e.Name,
		})
	})
	hm.addHookHandler("onBackupRestore", backupRestoreID)

	return nil
}

// registerModelHooks registers model operation hooks
func (hm *HookManager) registerModelHooks() error {
	// OnModelValidate hook
	modelValidateID := hm.app.OnModelValidate().BindFunc(func(e *core.ModelEvent) error {
		return hm.executeHookScripts("onModelValidate", map[string]interface{}{
			"app":   e.App,
			"model": e.Model,
		})
	})
	hm.addHookHandler("onModelValidate", modelValidateID)

	// OnModelCreate hook
	modelCreateID := hm.app.OnModelCreate().BindFunc(func(e *core.ModelEvent) error {
		return hm.executeHookScripts("onModelCreate", map[string]interface{}{
			"app":   e.App,
			"model": e.Model,
		})
	})
	hm.addHookHandler("onModelCreate", modelCreateID)

	// OnModelUpdate hook
	modelUpdateID := hm.app.OnModelUpdate().BindFunc(func(e *core.ModelEvent) error {
		return hm.executeHookScripts("onModelUpdate", map[string]interface{}{
			"app":   e.App,
			"model": e.Model,
		})
	})
	hm.addHookHandler("onModelUpdate", modelUpdateID)

	// OnModelDelete hook
	modelDeleteID := hm.app.OnModelDelete().BindFunc(func(e *core.ModelEvent) error {
		return hm.executeHookScripts("onModelDelete", map[string]interface{}{
			"app":   e.App,
			"model": e.Model,
		})
	})
	hm.addHookHandler("onModelDelete", modelDeleteID)

	return nil
}

// registerRecordHooks registers record operation hooks
func (hm *HookManager) registerRecordHooks() error {
	// OnRecordValidate hook
	recordValidateID := hm.app.OnRecordValidate().BindFunc(func(e *core.RecordEvent) error {
		return hm.executeHookScripts("onRecordValidate", map[string]interface{}{
			"app":    e.App,
			"record": e.Record,
		})
	})
	hm.addHookHandler("onRecordValidate", recordValidateID)

	// OnRecordCreate hook
	recordCreateID := hm.app.OnRecordCreate().BindFunc(func(e *core.RecordEvent) error {
		return hm.executeHookScripts("onRecordCreate", map[string]interface{}{
			"app":    e.App,
			"record": e.Record,
		})
	})
	hm.addHookHandler("onRecordCreate", recordCreateID)

	// OnRecordUpdate hook
	recordUpdateID := hm.app.OnRecordUpdate().BindFunc(func(e *core.RecordEvent) error {
		return hm.executeHookScripts("onRecordUpdate", map[string]interface{}{
			"app":    e.App,
			"record": e.Record,
		})
	})
	hm.addHookHandler("onRecordUpdate", recordUpdateID)

	// OnRecordDelete hook
	recordDeleteID := hm.app.OnRecordDelete().BindFunc(func(e *core.RecordEvent) error {
		return hm.executeHookScripts("onRecordDelete", map[string]interface{}{
			"app":    e.App,
			"record": e.Record,
		})
	})
	hm.addHookHandler("onRecordDelete", recordDeleteID)

	// OnRecordEnrich hook
	recordEnrichID := hm.app.OnRecordEnrich().BindFunc(func(e *core.RecordEnrichEvent) error {
		return hm.executeHookScripts("onRecordEnrich", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"collection": e.Record.Collection(),
			})
	})
	hm.addHookHandler("onRecordEnrich", recordEnrichID)

	return nil
}

// registerCollectionHooks registers collection operation hooks
func (hm *HookManager) registerCollectionHooks() error {
	// OnCollectionValidate hook
	collectionValidateID := hm.app.OnCollectionValidate().BindFunc(func(e *core.CollectionEvent) error {
		return hm.executeHookScripts("onCollectionValidate", map[string]interface{}{
			"app":        e.App,
			"collection": e.Collection,
		})
	})
	hm.addHookHandler("onCollectionValidate", collectionValidateID)

	// OnCollectionCreate hook
	collectionCreateID := hm.app.OnCollectionCreate().BindFunc(func(e *core.CollectionEvent) error {
		return hm.executeHookScripts("onCollectionCreate", map[string]interface{}{
			"app":        e.App,
			"collection": e.Collection,
		})
	})
	hm.addHookHandler("onCollectionCreate", collectionCreateID)

	// OnCollectionUpdate hook
	collectionUpdateID := hm.app.OnCollectionUpdate().BindFunc(func(e *core.CollectionEvent) error {
		return hm.executeHookScripts("onCollectionUpdate", map[string]interface{}{
			"app":        e.App,
			"collection": e.Collection,
		})
	})
	hm.addHookHandler("onCollectionUpdate", collectionUpdateID)

	// OnCollectionDelete hook
	collectionDeleteID := hm.app.OnCollectionDelete().BindFunc(func(e *core.CollectionEvent) error {
		return hm.executeHookScripts("onCollectionDelete", map[string]interface{}{
			"app":        e.App,
			"collection": e.Collection,
		})
	})
	hm.addHookHandler("onCollectionDelete", collectionDeleteID)

	return nil
}

// registerAPIHooks registers API request hooks
func (hm *HookManager) registerAPIHooks() error {
	// OnRecordsListRequest hook
	recordsListID := hm.app.OnRecordsListRequest().BindFunc(func(e *core.RecordsListRequestEvent) error {
		return hm.executeHookScripts("onRecordsListRequest", map[string]interface{}{
				"app":        e.App,
				"collection": e.Collection,
				"records":    e.Records,
				"result":     e.Result,
			})
	})
	hm.addHookHandler("onRecordsListRequest", recordsListID)

	// OnRecordViewRequest hook
	recordViewID := hm.app.OnRecordViewRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		return hm.executeHookScripts("onRecordViewRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"requestInfo": e.RequestInfo,
			})
	})
	hm.addHookHandler("onRecordViewRequest", recordViewID)

	// OnRecordCreateRequest hook
	recordCreateRequestID := hm.app.OnRecordCreateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		return hm.executeHookScripts("onRecordCreateRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"requestInfo": e.RequestInfo,
			})
	})
	hm.addHookHandler("onRecordCreateRequest", recordCreateRequestID)

	// OnRecordUpdateRequest hook
	recordUpdateRequestID := hm.app.OnRecordUpdateRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		return hm.executeHookScripts("onRecordUpdateRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"requestInfo": e.RequestInfo,
			})
	})
	hm.addHookHandler("onRecordUpdateRequest", recordUpdateRequestID)

	// OnRecordDeleteRequest hook
	recordDeleteRequestID := hm.app.OnRecordDeleteRequest().BindFunc(func(e *core.RecordRequestEvent) error {
		return hm.executeHookScripts("onRecordDeleteRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"requestInfo": e.RequestInfo,
			})
	})
	hm.addHookHandler("onRecordDeleteRequest", recordDeleteRequestID)

	return nil
}

// registerMailerHooks registers mailer hooks
func (hm *HookManager) registerMailerHooks() error {
	// OnMailerSend hook
	mailerSendID := hm.app.OnMailerSend().BindFunc(func(e *core.MailerEvent) error {
		return hm.executeHookScripts("onMailerSend", map[string]interface{}{
			"app":     e.App,
			"message": e.Message,
			"mailer":  e.Mailer,
		})
	})
	hm.addHookHandler("onMailerSend", mailerSendID)

	// OnMailerRecordPasswordResetSend hook
	passwordResetID := hm.app.OnMailerRecordPasswordResetSend().BindFunc(func(e *core.MailerRecordEvent) error {
		return hm.executeHookScripts("onMailerRecordPasswordResetSend", map[string]interface{}{
			"app":        e.App,
			"record":     e.Record,
			"message":    e.Message,
			"meta":       e.Meta,
		})
	})
	hm.addHookHandler("onMailerRecordPasswordResetSend", passwordResetID)

	// OnMailerRecordVerificationSend hook
	verificationID := hm.app.OnMailerRecordVerificationSend().BindFunc(func(e *core.MailerRecordEvent) error {
		return hm.executeHookScripts("onMailerRecordVerificationSend", map[string]interface{}{
			"app":        e.App,
			"record":     e.Record,
			"message":    e.Message,
			"meta":       e.Meta,
		})
	})
	hm.addHookHandler("onMailerRecordVerificationSend", verificationID)

	return nil
}

// registerRealtimeHooks registers realtime hooks
func (hm *HookManager) registerRealtimeHooks() error {
	// OnRealtimeConnectRequest hook
	realtimeConnectID := hm.app.OnRealtimeConnectRequest().BindFunc(func(e *core.RealtimeConnectRequestEvent) error {
		return hm.executeHookScripts("onRealtimeConnectRequest", map[string]interface{}{
				"app":     e.App,
				"client":  e.Client,
			})
	})
	hm.addHookHandler("onRealtimeConnectRequest", realtimeConnectID)

	// OnRealtimeMessageSend hook
	realtimeMessageID := hm.app.OnRealtimeMessageSend().BindFunc(func(e *core.RealtimeMessageEvent) error {
		return hm.executeHookScripts("onRealtimeMessageSend", map[string]interface{}{
			"app":     e.App,
			"client":  e.Client,
			"message": e.Message,
		})
	})
	hm.addHookHandler("onRealtimeMessageSend", realtimeMessageID)

	return nil
}

// registerFileHooks registers file operation hooks
func (hm *HookManager) registerFileHooks() error {
	// OnFileDownloadRequest hook
	fileDownloadID := hm.app.OnFileDownloadRequest().BindFunc(func(e *core.FileDownloadRequestEvent) error {
		return hm.executeHookScripts("onFileDownloadRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"filename":   e.ServedName,
			})
	})
	hm.addHookHandler("onFileDownloadRequest", fileDownloadID)

	// OnFileTokenRequest hook
	fileTokenID := hm.app.OnFileTokenRequest().BindFunc(func(e *core.FileTokenRequestEvent) error {
		return hm.executeHookScripts("onFileTokenRequest", map[string]interface{}{
				"app":        e.App,
				"record":     e.Record,
				"token":      e.Token,
			})
	})
	hm.addHookHandler("onFileTokenRequest", fileTokenID)

	return nil
}

// executeHookScripts executes all scripts registered for a specific hook
func (hm *HookManager) executeHookScripts(hookName string, eventData map[string]interface{}) error {
	// Find all scripts with the specified hook category
	scripts := hm.engine.ListScripts()
	hookScripts := make([]*Script, 0)

	for _, script := range scripts {
		if !script.Enabled {
			continue
		}

		// Check if script is registered for this hook
		if script.Category == "hooks" {
			// Check metadata for specific hook name
			if script.Metadata != nil {
				if hooks, ok := script.Metadata["hooks"].([]interface{}); ok {
					for _, h := range hooks {
						if hookStr, ok := h.(string); ok && hookStr == hookName {
							hookScripts = append(hookScripts, script)
							break
						}
					}
				}
				// Fallback: check if hook name is in script name
				if len(hookScripts) == 0 && strings.Contains(strings.ToLower(script.Name), strings.ToLower(hookName)) {
					hookScripts = append(hookScripts, script)
				}
			}
		}
	}

	// Execute hook scripts in priority order
	for _, script := range hookScripts {

		// Add hook-specific data to input
		inputData := map[string]interface{}{
			"hookName":  hookName,
			"eventData": eventData,
			"timestamp": time.Now(),
		}

		// Execute the script
		result, err := hm.engine.executeScript(script, inputData)
		if err != nil {
			hm.logger.Error("Hook script execution failed",
				"hook", hookName,
				"script", script.Name,
				"error", err)
			// Continue with other scripts even if one fails
			continue
		}

		// Log successful execution
		hm.logger.Debug("Hook script executed successfully",
			"hook", hookName,
			"script", script.Name,
			"duration", result.Duration)
	}

	return nil
}

// addHookHandler adds a hook handler ID for tracking
func (hm *HookManager) addHookHandler(hookName, handlerID string) {
	if hm.hookHandlers[hookName] == nil {
		hm.hookHandlers[hookName] = make([]string, 0)
	}
	hm.hookHandlers[hookName] = append(hm.hookHandlers[hookName], handlerID)
}

// removeHookHandler removes a specific hook handler
func (hm *HookManager) removeHookHandler(hookName, handlerID string) {
	// Note: PocketBase hooks don't have a direct Remove method by ID
	// This is a placeholder for future implementation
	// For now, we rely on UnregisterAllHooks to clean up
	hm.logger.Debug("Hook handler marked for removal",
		"hook", hookName,
		"handlerID", handlerID)
}

// GetRegisteredHooks returns a list of all registered hook names
func (hm *HookManager) GetRegisteredHooks() []string {
	hooks := make([]string, 0, len(hm.hookHandlers))
	for hookName := range hm.hookHandlers {
		hooks = append(hooks, hookName)
	}
	return hooks
}

// GetHookStats returns statistics about hook registrations
func (hm *HookManager) GetHookStats() map[string]interface{} {
	stats := map[string]interface{}{
		"totalHooks":        len(hm.hookHandlers),
		"registeredHooks":   hm.GetRegisteredHooks(),
		"hookHandlerCounts": make(map[string]int),
	}

	for hookName, handlers := range hm.hookHandlers {
		stats["hookHandlerCounts"].(map[string]int)[hookName] = len(handlers)
	}

	return stats
}