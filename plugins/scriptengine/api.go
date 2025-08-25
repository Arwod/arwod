package scriptengine

import (
	"encoding/json"
	"strconv"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// RegisterScriptAPIs registers script management API routes
func (e *ScriptEngine) RegisterScriptAPIs(rg *router.RouterGroup[*core.RequestEvent]) {
	// Script listing API
	rg.GET("", e.handleListScripts)

	// Script execution API
	rg.POST("/execute", e.handleExecuteScript)
	rg.POST("/:id/test", e.handleTestScript)
	rg.POST("/:id/execute", e.handleExecuteScriptByID)

	// Script management APIs
	rg.PUT("/:id/enable", e.handleEnableScript)
	rg.PUT("/:id/disable", e.handleDisableScript)
	rg.POST("/:id/reload", e.handleReloadScript)

	// Script monitoring APIs
	rg.GET("/:id/logs", e.handleGetScriptLogs)
	rg.GET("/:id/metrics", e.handleGetScriptMetrics)
	rg.GET("/health", e.handleHealthCheck)
}

// handleExecuteScript handles script execution requests
func (e *ScriptEngine) handleExecuteScript(c *core.RequestEvent) error {
	var request struct {
		ScriptID  string                 `json:"script_id"`
		InputData map[string]interface{} `json:"input_data"`
	}

	if err := c.BindBody(&request); err != nil {
		return apis.NewBadRequestError("Invalid request body", err)
	}

	if request.ScriptID == "" {
		return apis.NewBadRequestError("script_id is required", nil)
	}

	// Execute script
	result, err := e.ExecuteScript(request.ScriptID, request.InputData)
	if err != nil {
		return apis.NewApiError(500, "Script execution failed", err)
	}

	return c.JSON(200, result)
}

// handleTestScript handles script test execution
func (e *ScriptEngine) handleTestScript(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	var testParams map[string]interface{}
	if err := c.BindBody(&testParams); err != nil {
		// If no body provided, use empty params
		testParams = make(map[string]interface{})
	}

	// Get script
	script, err := e.GetScript(scriptID)
	if err != nil {
		return apis.NewNotFoundError("Script not found", err)
	}

	// Execute script in test mode (add test flag to input)
	testParams["__test_mode__"] = true
	result, err := e.ExecuteScript(scriptID, testParams)
	if err != nil {
		return apis.NewApiError(500, "Script test execution failed", err)
	}

	return c.JSON(200, map[string]interface{}{
		"success": result.Status == "success",
		"result":  result,
		"script":  script,
	})
}

// handleExecuteScriptByID handles script execution by ID
func (e *ScriptEngine) handleExecuteScriptByID(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	var inputData map[string]interface{}
	if err := c.BindBody(&inputData); err != nil {
		// If no body provided, use empty input
		inputData = make(map[string]interface{})
	}

	// Execute script
	result, err := e.ExecuteScript(scriptID, inputData)
	if err != nil {
		return apis.NewApiError(500, "Script execution failed", err)
	}

	return c.JSON(200, result)
}

// handleEnableScript handles script enable requests
func (e *ScriptEngine) handleEnableScript(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	// Get script from database
	record, err := c.App.FindRecordById("js_scripts", scriptID)
	if err != nil {
		return apis.NewNotFoundError("Script not found", err)
	}

	// Update enabled status
	record.Set("enabled", true)
	if err := c.App.Save(record); err != nil {
		return apis.NewApiError(500, "Failed to enable script", err)
	}

	// Reload script in engine
	script := &Script{
		ID:       record.Id,
		Name:     record.GetString("name"),
		Category: record.GetString("category"),
		Content:  record.GetString("content"),
		Enabled:  record.GetBool("enabled"),
		Priority: record.GetInt("priority"),
	}

	if metadata := record.GetString("metadata"); metadata != "" {
		json.Unmarshal([]byte(metadata), &script.Metadata)
	}

	if err := e.LoadScript(script); err != nil {
		e.logger.Error("Failed to reload script after enabling", "script_id", scriptID, "error", err)
	}

	return c.JSON(200, map[string]string{"status": "enabled"})
}

// handleDisableScript handles script disable requests
func (e *ScriptEngine) handleDisableScript(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	// Get script from database
	record, err := c.App.FindRecordById("js_scripts", scriptID)
	if err != nil {
		return apis.NewNotFoundError("Script not found", err)
	}

	// Update enabled status
	record.Set("enabled", false)
	if err := c.App.Save(record); err != nil {
		return apis.NewApiError(500, "Failed to disable script", err)
	}

	// Unload script from engine
	if err := e.UnloadScript(scriptID); err != nil {
		e.logger.Error("Failed to unload script after disabling", "script_id", scriptID, "error", err)
	}

	return c.JSON(200, map[string]string{"status": "disabled"})
}

// handleReloadScript handles script reload requests
func (e *ScriptEngine) handleReloadScript(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	// Get script from database
	record, err := c.App.FindRecordById("js_scripts", scriptID)
	if err != nil {
		return apis.NewNotFoundError("Script not found", err)
	}

	// Create script object
	script := &Script{
		ID:       record.Id,
		Name:     record.GetString("name"),
		Category: record.GetString("category"),
		Content:  record.GetString("content"),
		Enabled:  record.GetBool("enabled"),
		Priority: record.GetInt("priority"),
	}

	if metadata := record.GetString("metadata"); metadata != "" {
		json.Unmarshal([]byte(metadata), &script.Metadata)
	}

	// Unload and reload script
	e.UnloadScript(scriptID)
	if err := e.LoadScript(script); err != nil {
		return apis.NewApiError(500, "Failed to reload script", err)
	}

	return c.JSON(200, map[string]string{"status": "reloaded"})
}

// handleGetScriptLogs handles script logs requests
func (e *ScriptEngine) handleGetScriptLogs(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	// Parse query parameters
	limitStr := c.Request.URL.Query().Get("limit")
	limit := 100 // default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offsetStr := c.Request.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Query execution logs from database
	records, err := c.App.FindRecordsByFilter(
		"js_execution_logs",
		"script_id = {:script_id}",
		"-created", // order by created desc
		limit,
		offset,
		dbx.Params{"script_id": scriptID},
	)
	if err != nil {
		return apis.NewApiError(500, "Failed to fetch logs", err)
	}

	// Convert records to log entries
	logs := make([]map[string]interface{}, len(records))
	for i, record := range records {
		logs[i] = map[string]interface{}{
			"id":         record.Id,
			"script_id":  record.GetString("script_id"),
			"status":     record.GetString("status"),
			"duration":   record.GetInt("duration"),
			"error":      record.GetString("error"),
			"output":     record.GetString("output"),
			"input_data": record.GetString("input_data"),
			"created":    record.GetDateTime("created"),
		}
	}

	return c.JSON(200, map[string]interface{}{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
		"total":  len(logs),
	})
}

// handleGetScriptMetrics handles script metrics requests
func (e *ScriptEngine) handleGetScriptMetrics(c *core.RequestEvent) error {
	scriptID := c.Request.PathValue("id")
	if scriptID == "" {
		return apis.NewBadRequestError("Script ID is required", nil)
	}

	// Get script
	script, err := e.GetScript(scriptID)
	if err != nil {
		return apis.NewNotFoundError("Script not found", err)
	}

	// Get execution statistics from logs
	totalCount, err := c.App.CountRecords("js_execution_logs", dbx.NewExp("script_id = {:script_id}", dbx.Params{"script_id": scriptID}))
	if err != nil {
		totalCount = 0
	}

	successCount, err := c.App.CountRecords("js_execution_logs", dbx.NewExp("script_id = {:script_id} AND status = 'success'", dbx.Params{"script_id": scriptID}))
	if err != nil {
		successCount = 0
	}

	errorCount, err := c.App.CountRecords("js_execution_logs", dbx.NewExp("script_id = {:script_id} AND status = 'error'", dbx.Params{"script_id": scriptID}))
	if err != nil {
		errorCount = 0
	}

	successRate := float64(0)
	if totalCount > 0 {
		successRate = float64(successCount) / float64(totalCount) * 100
	}

	metrics := map[string]interface{}{
		"script_id":    scriptID,
		"script_name":  script.Name,
		"enabled":      script.Enabled,
		"category":     script.Category,
		"total_runs":   totalCount,
		"success_runs": successCount,
		"error_runs":   errorCount,
		"success_rate": successRate,
		"last_updated": script.UpdatedAt,
	}

	return c.JSON(200, metrics)
}

// handleHealthCheck handles health check requests
func (e *ScriptEngine) handleHealthCheck(c *core.RequestEvent) error {
	stats := e.GetStats()

	health := map[string]interface{}{
		"status":       "healthy",
		"engine_stats": stats,
		"timestamp":    c.Request.Header.Get("X-Request-Time"),
	}

	return c.JSON(200, health)
}

// handleListScripts handles listing all scripts
func (e *ScriptEngine) handleListScripts(c *core.RequestEvent) error {
	scripts := e.ListScripts()

	return c.JSON(200, map[string]interface{}{
		"scripts": scripts,
		"total":   len(scripts),
	})
}
