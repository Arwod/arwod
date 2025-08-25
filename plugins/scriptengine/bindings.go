package scriptengine

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/pocketbase/pocketbase/core"
)

// APIBindings provides safe API access for scripts
type APIBindings struct {
	app    core.App
	vm     *goja.Runtime
	config *Config
}

// NewAPIBindings creates a new API bindings instance
func NewAPIBindings(app core.App, vm *goja.Runtime, config *Config) *APIBindings {
	return &APIBindings{
		app:    app,
		vm:     vm,
		config: config,
	}
}

// BindAPIs binds all available APIs to the JavaScript runtime
func (b *APIBindings) BindAPIs() error {
	// Database APIs
	if err := b.bindDatabaseAPIs(); err != nil {
		return fmt.Errorf("failed to bind database APIs: %w", err)
	}

	// Collection APIs
	if err := b.bindCollectionAPIs(); err != nil {
		return fmt.Errorf("failed to bind collection APIs: %w", err)
	}

	// Record APIs
	if err := b.bindRecordAPIs(); err != nil {
		return fmt.Errorf("failed to bind record APIs: %w", err)
	}

	// Utility APIs
	if err := b.bindUtilityAPIs(); err != nil {
		return fmt.Errorf("failed to bind utility APIs: %w", err)
	}

	// HTTP APIs
	if err := b.bindHTTPAPIs(); err != nil {
		return fmt.Errorf("failed to bind HTTP APIs: %w", err)
	}

	// Validation APIs
	if err := b.bindValidationAPIs(); err != nil {
		return fmt.Errorf("failed to bind validation APIs: %w", err)
	}

	return nil
}

// bindDatabaseAPIs binds database-related APIs
func (b *APIBindings) bindDatabaseAPIs() error {
	dbAPI := map[string]interface{}{
		"findRecordById": b.findRecordById,
		"findRecords":    b.findRecords,
		"createRecord":   b.createRecord,
		"updateRecord":   b.updateRecord,
		"deleteRecord":   b.deleteRecord,
		"countRecords":   b.countRecords,
	}

	b.vm.Set("$db", dbAPI)
	return nil
}

// bindCollectionAPIs binds collection-related APIs
func (b *APIBindings) bindCollectionAPIs() error {
	collectionAPI := map[string]interface{}{
		"findById":   b.findCollectionById,
		"findByName": b.findCollectionByName,
		"list":       b.listCollections,
		"create":     b.createCollection,
		"update":     b.updateCollection,
		"delete":     b.deleteCollection,
	}

	b.vm.Set("$collection", collectionAPI)
	return nil
}

// bindRecordAPIs binds record-related APIs
func (b *APIBindings) bindRecordAPIs() error {
	recordAPI := map[string]interface{}{
		"new":      b.newRecord,
		"validate": b.validateRecord,
		"save":     b.saveRecord,
		"refresh":  b.refreshRecord,
		"expand":   b.expandRecord,
	}

	b.vm.Set("$record", recordAPI)
	return nil
}

// bindUtilityAPIs binds utility APIs
func (b *APIBindings) bindUtilityAPIs() error {
	utilAPI := map[string]interface{}{
		"log":        b.logMessage,
		"sleep":      b.sleep,
		"now":        b.now,
		"uuid":       b.generateUUID,
		"hash":       b.hashString,
		"encrypt":    b.encryptString,
		"decrypt":    b.decryptString,
		"base64":     b.base64Encode,
		"parseJSON":  b.parseJSON,
		"stringify":  b.stringifyJSON,
	}

	b.vm.Set("$util", utilAPI)
	return nil
}

// bindHTTPAPIs binds HTTP-related APIs
func (b *APIBindings) bindHTTPAPIs() error {
	httpAPI := map[string]interface{}{
		"get":    b.httpGet,
		"post":   b.httpPost,
		"put":    b.httpPut,
		"delete": b.httpDelete,
		"patch":  b.httpPatch,
	}

	b.vm.Set("$http", httpAPI)
	return nil
}

// bindValidationAPIs binds validation APIs
func (b *APIBindings) bindValidationAPIs() error {
	validationAPI := map[string]interface{}{
		"isEmail":    b.isEmail,
		"isURL":      b.isURL,
		"isUUID":     b.isUUID,
		"isNumeric":  b.isNumeric,
		"isAlpha":    b.isAlpha,
		"minLength": b.minLength,
		"maxLength": b.maxLength,
		"required":   b.required,
	}

	b.vm.Set("$validate", validationAPI)
	return nil
}

// Database API implementations
func (b *APIBindings) findRecordById(collectionName, id string) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	record, err := b.app.FindRecordById(collection, id)
	if err != nil {
		return nil, fmt.Errorf("record not found: %w", err)
	}

	return b.recordToMap(record), nil
}

func (b *APIBindings) findRecords(collectionName string, filter string, sort string, limit int) ([]map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	records, err := b.app.FindRecordsByFilter(collection, filter, sort, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to find records: %w", err)
	}

	result := make([]map[string]interface{}, len(records))
	for i, record := range records {
		result[i] = b.recordToMap(record)
	}

	return result, nil
}

func (b *APIBindings) createRecord(collectionName string, data map[string]interface{}) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	for key, value := range data {
		record.Set(key, value)
	}

	if err := b.app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	return b.recordToMap(record), nil
}

func (b *APIBindings) updateRecord(collectionName, id string, data map[string]interface{}) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	record, err := b.app.FindRecordById(collection, id)
	if err != nil {
		return nil, fmt.Errorf("record not found: %w", err)
	}

	for key, value := range data {
		record.Set(key, value)
	}

	if err := b.app.Save(record); err != nil {
		return nil, fmt.Errorf("failed to update record: %w", err)
	}

	return b.recordToMap(record), nil
}

func (b *APIBindings) deleteRecord(collectionName, id string) error {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	record, err := b.app.FindRecordById(collection, id)
	if err != nil {
		return fmt.Errorf("record not found: %w", err)
	}

	if err := b.app.Delete(record); err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	return nil
}

func (b *APIBindings) countRecords(collectionName, filter string) (int, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return 0, fmt.Errorf("collection not found: %w", err)
	}

	records, err := b.app.FindRecordsByFilter(collection, filter, "", 0, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return len(records), nil
}

// Collection API implementations
func (b *APIBindings) findCollectionById(id string) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(id)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	return b.collectionToMap(collection), nil
}

func (b *APIBindings) findCollectionByName(name string) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(name)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	return b.collectionToMap(collection), nil
}

func (b *APIBindings) listCollections() ([]map[string]interface{}, error) {
	collections, err := b.app.FindAllCollections()
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	result := make([]map[string]interface{}, len(collections))
	for i, collection := range collections {
		result[i] = b.collectionToMap(collection)
	}

	return result, nil
}

func (b *APIBindings) createCollection(data map[string]interface{}) (map[string]interface{}, error) {
	collection := &core.Collection{}
	
	// Set basic fields
	if name, ok := data["name"].(string); ok {
		collection.Name = name
	}
	if ctype, ok := data["type"].(string); ok {
		collection.Type = ctype
	}

	if err := b.app.Save(collection); err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return b.collectionToMap(collection), nil
}

func (b *APIBindings) updateCollection(id string, data map[string]interface{}) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(id)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	// Update fields
	if name, ok := data["name"].(string); ok {
		collection.Name = name
	}
	if ctype, ok := data["type"].(string); ok {
		collection.Type = ctype
	}

	if err := b.app.Save(collection); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	return b.collectionToMap(collection), nil
}

func (b *APIBindings) deleteCollection(id string) error {
	collection, err := b.app.FindCollectionByNameOrId(id)
	if err != nil {
		return fmt.Errorf("collection not found: %w", err)
	}

	if err := b.app.Delete(collection); err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	return nil
}

// Record API implementations
func (b *APIBindings) newRecord(collectionName string) (map[string]interface{}, error) {
	collection, err := b.app.FindCollectionByNameOrId(collectionName)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %w", err)
	}

	record := core.NewRecord(collection)
	return b.recordToMap(record), nil
}

func (b *APIBindings) validateRecord(record map[string]interface{}) error {
	// Basic validation logic
	// This would need to be expanded based on collection schema
	return nil
}

func (b *APIBindings) saveRecord(record map[string]interface{}) (map[string]interface{}, error) {
	// This would need to convert the map back to a Record and save it
	// Implementation depends on how records are structured
	return record, nil
}

func (b *APIBindings) refreshRecord(record map[string]interface{}) (map[string]interface{}, error) {
	// Refresh record from database
	return record, nil
}

func (b *APIBindings) expandRecord(record map[string]interface{}, relations []string) (map[string]interface{}, error) {
	// Expand record relations
	return record, nil
}

// Utility API implementations
func (b *APIBindings) logMessage(level, message string) {
	switch strings.ToLower(level) {
	case "debug":
		log.Printf("[SCRIPT DEBUG] %s", message)
	case "info":
		log.Printf("[SCRIPT INFO] %s", message)
	case "warn":
		log.Printf("[SCRIPT WARN] %s", message)
	case "error":
		log.Printf("[SCRIPT ERROR] %s", message)
	default:
		log.Printf("[SCRIPT] %s", message)
	}
}

func (b *APIBindings) sleep(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (b *APIBindings) now() int64 {
	return time.Now().Unix()
}

func (b *APIBindings) generateUUID() string {
	// Generate UUID - would need to import uuid package
	return "uuid-placeholder"
}

func (b *APIBindings) hashString(input string) string {
	// Hash string - would need to implement proper hashing
	return "hash-placeholder"
}

func (b *APIBindings) encryptString(input string) string {
	// Encrypt string - would need to implement encryption
	return "encrypted-placeholder"
}

func (b *APIBindings) decryptString(input string) string {
	// Decrypt string - would need to implement decryption
	return "decrypted-placeholder"
}

func (b *APIBindings) base64Encode(input string) string {
	// Base64 encode - would need to implement
	return "base64-placeholder"
}

func (b *APIBindings) parseJSON(input string) (interface{}, error) {
	var result interface{}
	err := json.Unmarshal([]byte(input), &result)
	return result, err
}

func (b *APIBindings) stringifyJSON(input interface{}) (string, error) {
	bytes, err := json.Marshal(input)
	return string(bytes), err
}

// HTTP API implementations (placeholder - would need proper HTTP client)
func (b *APIBindings) httpGet(url string, headers map[string]string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": 200, "body": "placeholder"}, nil
}

func (b *APIBindings) httpPost(url string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": 200, "body": "placeholder"}, nil
}

func (b *APIBindings) httpPut(url string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": 200, "body": "placeholder"}, nil
}

func (b *APIBindings) httpDelete(url string, headers map[string]string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": 200, "body": "placeholder"}, nil
}

func (b *APIBindings) httpPatch(url string, data interface{}, headers map[string]string) (map[string]interface{}, error) {
	return map[string]interface{}{"status": 200, "body": "placeholder"}, nil
}

// Validation API implementations
func (b *APIBindings) isEmail(input string) bool {
	// Email validation - would need proper regex
	return strings.Contains(input, "@")
}

func (b *APIBindings) isURL(input string) bool {
	// URL validation - would need proper validation
	return strings.HasPrefix(input, "http")
}

func (b *APIBindings) isUUID(input string) bool {
	// UUID validation - would need proper validation
	return len(input) == 36
}

func (b *APIBindings) isNumeric(input string) bool {
	// Numeric validation - would need proper validation
	return true
}

func (b *APIBindings) isAlpha(input string) bool {
	// Alpha validation - would need proper validation
	return true
}

func (b *APIBindings) minLength(input string, min int) bool {
	return len(input) >= min
}

func (b *APIBindings) maxLength(input string, max int) bool {
	return len(input) <= max
}

func (b *APIBindings) required(input interface{}) bool {
	return input != nil && input != ""
}

// Helper methods
func (b *APIBindings) recordToMap(record *core.Record) map[string]interface{} {
	result := make(map[string]interface{})
	
	// Basic fields
	result["id"] = record.Id
	result["created"] = record.GetDateTime("created")
	result["updated"] = record.GetDateTime("updated")
	
	// Data fields
	for key, value := range record.FieldsData() {
		result[key] = value
	}
	
	return result
}

func (b *APIBindings) collectionToMap(collection *core.Collection) map[string]interface{} {
	result := make(map[string]interface{})
	
	result["id"] = collection.Id
	result["name"] = collection.Name
	result["type"] = collection.Type
	result["created"] = collection.Created
	result["updated"] = collection.Updated
	
	// Use Fields instead of Schema
	result["fields"] = collection.Fields
	
	return result
}