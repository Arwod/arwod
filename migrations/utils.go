package migrations

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// 表名常量定义
const (
	TableSysDept         = "sys_department"
	TableSysUser         = "sys_user"
	TableSysPost         = "sys_position"
	TableSysRole         = "sys_role"
	TableSysMenu         = "sys_menu"
	TableSysOperationLog = "sys_operation_log"
	TableSysDictType     = "sys_dict_type"
	TableSysDictData     = "sys_dict_data"
	TableSysConfig       = "sys_config"
	TableSysLoginInfo    = "sys_login_info"
	TableSysUserOnline   = "sys_user_online"
	TableSysNotice       = "sys_notice"
	TableSysUserRole     = "sys_user_role"
	TableSysRoleMenu     = "sys_role_menu"
	TableSysRoleDept     = "sys_role_department"
	TableSysUserPost     = "sys_user_position"
)

// 定义关系字段映射
type RelationFieldConfig struct {
	TableName   string
	FieldName   string
	TargetTable string
	Options     map[string]interface{}
}

// TableCreationRequest 表创建请求结构
type TableCreationRequest struct {
	TableName        string
	IsAuthCollection bool
	AuthOptions      *AuthCollectionOptions
	Fields           core.FieldsList
	Indexes          []string
	Relations        []RelationFieldConfig
}

// AuthCollectionOptions 认证集合选项
type AuthCollectionOptions struct {
	TokenKey string
}

// DataImportRequest 数据导入请求结构
type DataImportRequest struct {
	Table        string                   `json:"table"`
	Data         []map[string]interface{} `json:"data"`
	UniqueFields []string                 `json:"uniqueFields"`
}

// CreateTable 创建表
func CreateTable(txApp core.App, request TableCreationRequest) error {
	// 关闭外键约束检查以支持自引用表
	if _, err := txApp.DB().NewQuery("PRAGMA foreign_keys = OFF").Execute(); err != nil {
		return fmt.Errorf("failed to disable foreign key constraints: %w", err)
	}

	// 确保在函数结束时重新开启外键约束检查
	defer func() {
		if _, err := txApp.DB().NewQuery("PRAGMA foreign_keys = ON").Execute(); err != nil {
			// 记录错误但不影响主流程
			fmt.Printf("Warning: failed to re-enable foreign key constraints: %v\n", err)
		}
	}()

	collection := core.NewBaseCollection(request.TableName)

	if request.IsAuthCollection {
		authCollection := core.NewAuthCollection(request.TableName)
		if request.AuthOptions != nil {
			authCollection.AuthToken.Duration = 1209600 // 14 days
		}
		collection = authCollection
	}

	// 添加字段
	for _, field := range request.Fields {
		if field != nil {
			collection.Fields.Add(field)
		}
	}

	// 统一添加通用字段
	collection.Fields.Add(&core.TextField{Name: "created_by", Max: 64})
	collection.Fields.Add(&core.AutodateField{Name: "created", OnCreate: true})
	collection.Fields.Add(&core.TextField{Name: "updated_by", Max: 64})
	collection.Fields.Add(&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true})

	// 保存集合
	if err := txApp.Save(collection); err != nil {
		return fmt.Errorf("failed to save collection: %w", err)
	}

	// 创建索引
	for _, indexSQL := range request.Indexes {
		formattedSQL := fmt.Sprintf(indexSQL, collection.Name, collection.Name)
		if _, err := txApp.DB().NewQuery(formattedSQL).Execute(); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	return nil
}

// AddRelationFields 为已存在的表添加关联字段
func AddRelationFields(txApp core.App, tableName string, fields core.FieldsList) error {
	collection, err := txApp.FindCollectionByNameOrId(tableName)
	if err != nil {
		return fmt.Errorf("failed to find collection %s: %w", tableName, err)
	}

	for _, field := range fields {
		if field != nil {
			collection.Fields.Add(field)
		}
	}

	if err := txApp.Save(collection); err != nil {
		return fmt.Errorf("failed to save collection with relation fields: %w", err)
	}

	return nil
}

// AddSingleRelationField 为指定表添加单个关联字段
func AddSingleRelationField(app core.App, tableName, fieldName, targetTable string, options map[string]interface{}) error {
	collection, err := app.FindCollectionByNameOrId(tableName)
	if err != nil {
		return err
	}

	// 获取目标集合
	targetCollection, err := app.FindCollectionByNameOrId(targetTable)
	if err != nil {
		return err
	}

	// 创建关系字段
	relationField := CreateRelationFieldHelper(fieldName, targetCollection.Id, options)

	// 添加字段到集合
	collection.Fields.Add(relationField)

	return app.Save(collection)
}

// createField 创建字段
// CreateFieldHelper 创建字段的辅助函数，用于简化字段创建过程
func CreateFieldHelper(name, fieldType string, options map[string]interface{}) core.Field {
	switch fieldType {
	case core.FieldTypeText:
		field := &core.TextField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if max, ok := options["max"].(int); ok {
			field.Max = max
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeEmail:
		field := &core.EmailField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeNumber:
		field := &core.NumberField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if min, ok := options["min"].(*float64); ok {
			field.Min = min
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeSelect:
		field := &core.SelectField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if maxSelect, ok := options["maxSelect"].(int); ok && maxSelect > 0 {
			field.MaxSelect = maxSelect
		} else {
			field.MaxSelect = 1
		}
		if values, ok := options["values"].([]string); ok {
			field.Values = values
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeAutodate:
		field := &core.AutodateField{Name: name}
		if onCreate, ok := options["onCreate"].(bool); ok {
			field.OnCreate = onCreate
		}
		if onUpdate, ok := options["onUpdate"].(bool); ok {
			field.OnUpdate = onUpdate
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeDate:
		field := &core.DateField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeEditor:
		field := &core.EditorField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	case core.FieldTypeRelation:
		field := &core.RelationField{Name: name}
		if required, ok := options["required"].(bool); ok {
			field.Required = required
		}
		if maxSelect, ok := options["maxSelect"].(int); ok && maxSelect > 0 {
			field.MaxSelect = maxSelect
		} else {
			field.MaxSelect = 1
		}
		if system, ok := options["system"].(bool); ok {
			field.System = system
		}
		if hidden, ok := options["hidden"].(bool); ok {
			field.Hidden = hidden
		}
		return field

	default:
		return nil
	}
}

// CreateRelationFieldHelper 创建关系字段的辅助函数
func CreateRelationFieldHelper(name, collectionId string, options map[string]interface{}) core.Field {
	field := &core.RelationField{
		Name:         name,
		CollectionId: collectionId,
	}
	if required, ok := options["required"].(bool); ok {
		field.Required = required
	}
	if maxSelect, ok := options["maxSelect"].(int); ok && maxSelect > 0 {
		field.MaxSelect = maxSelect
	} else {
		field.MaxSelect = 1
	}
	if system, ok := options["system"].(bool); ok {
		field.System = system
	}
	if hidden, ok := options["hidden"].(bool); ok {
		field.Hidden = hidden
	}
	return field
}

// deleteUsersTable 删除users表
func deleteUsersTable(txApp core.App) error {
	usersCollection, err := txApp.FindCollectionByNameOrId("users")
	if err != nil {
		// 如果表不存在，不需要删除
		return nil
	}

	if err := txApp.Delete(usersCollection); err != nil {
		return fmt.Errorf("failed to delete users collection: %w", err)
	}

	return nil
}

// getValidFields 获取集合的有效字段列表（私有函数）
func getValidFields(collection *core.Collection) map[string]bool {
	validFields := make(map[string]bool)
	for _, field := range collection.Fields {
		validFields[field.GetName()] = true
	}
	// 添加系统字段
	validFields["id"] = true
	validFields["created"] = true
	validFields["updated"] = true
	return validFields
}

// findExistingRecord 根据唯一字段查找现有记录（私有函数）
func findExistingRecord(txApp core.App, tableName string, dataItem map[string]interface{}, uniqueFields []string) (*core.Record, error) {
	// 构建查询条件
	var conditions []string
	for _, uniqueField := range uniqueFields {
		if uniqueValue, exists := dataItem[uniqueField]; exists {
			conditions = append(conditions, fmt.Sprintf("%s = '%v'", uniqueField, uniqueValue))
		}
	}

	if len(conditions) == 0 {
		return nil, nil
	}

	filter := fmt.Sprintf("(%s)", conditions[0])
	for j := 1; j < len(conditions); j++ {
		filter += fmt.Sprintf(" && (%s)", conditions[j])
	}

	existingRecord, _ := txApp.FindFirstRecordByFilter(tableName, filter)
	return existingRecord, nil
}

// setRecordFields 设置记录字段值（私有函数）
func setRecordFields(record *core.Record, dataItem map[string]interface{}, validFields map[string]bool, isUpdate bool) error {
	// 设置字段值，只设置表中存在的字段
	for fieldName, fieldValue := range dataItem {
		if validFields[fieldName] {
			record.Set(fieldName, fieldValue)
		}
	}

	// 设置系统字段
	if !isUpdate {
		// 新记录：设置创建时间（如果没有手动设置）
		if _, exists := dataItem["created"]; !exists {
			record.Set("created", time.Now())
		}
	}
	// 更新时间总是设置为当前时间（除非手动指定）
	if _, exists := dataItem["updated"]; !exists {
		record.Set("updated", time.Now())
	}

	return nil
}

// ImportData 通用数据导入函数
// 支持 upsert 逻辑：通过 uniqueFields 指定的字段组合作为唯一标识
// 如果查找到已存在记录则更新，不存在则创建新记录
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体（包含 UniqueFields 配置）
//
// 返回:
//   - error: 错误信息
func ImportData(txApp core.App, request DataImportRequest) error {
	// 验证必要字段
	if request.Table == "" {
		return fmt.Errorf("表名不能为空")
	}
	if len(request.Data) == 0 {
		return fmt.Errorf("数据不能为空")
	}

	// 获取目标集合
	collection, err := txApp.FindCollectionByNameOrId(request.Table)
	if err != nil {
		return fmt.Errorf("找不到表 %s: %w", request.Table, err)
	}

	// 获取集合的字段信息
	validFields := getValidFields(collection)

	// 插入或更新数据
	for i, dataItem := range request.Data {
		var record *core.Record
		var isUpdate bool

		// 如果指定了唯一字段，检查是否已存在
		if len(request.UniqueFields) > 0 {
			existingRecord, err := findExistingRecord(txApp, request.Table, dataItem, request.UniqueFields)
			if err != nil {
				return fmt.Errorf("查找现有记录失败: %w", err)
			}
			if existingRecord != nil {
				// 使用已存在的记录进行更新
				record = existingRecord
				isUpdate = true
			}
		}

		// 如果没有找到已存在的记录，创建新记录
		if record == nil {
			record = core.NewRecord(collection)
			isUpdate = false
		}

		// 设置字段值和系统字段
		if err := setRecordFields(record, dataItem, validFields, isUpdate); err != nil {
			return fmt.Errorf("设置记录字段失败: %w", err)
		}

		// 保存记录
		if err := txApp.Save(record); err != nil {
			action := "保存"
			if isUpdate {
				action = "更新"
			}
			return fmt.Errorf("%s第 %d 条数据失败: %w", action, i+1, err)
		}
	}

	return nil
}

// RollbackData 根据唯一字段组合回滚数据
// 通过 DataImportRequest 中指定的数据和唯一字段组合来回滚（删除）对应的记录
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据回滚请求结构体（包含 Table、Data 和 UniqueFields 配置）
//
// 返回:
//   - error: 错误信息
func RollbackData(txApp core.App, request DataImportRequest) error {
	// 验证必要字段
	if request.Table == "" {
		return fmt.Errorf("表名不能为空")
	}
	if len(request.Data) == 0 {
		return fmt.Errorf("回滚数据不能为空")
	}
	if len(request.UniqueFields) == 0 {
		return fmt.Errorf("回滚操作必须指定唯一字段")
	}

	// 遍历每条数据，根据唯一字段组合查找并删除对应记录
	for i, dataItem := range request.Data {
		// 根据唯一字段查找现有记录
		existingRecord, err := findExistingRecord(txApp, request.Table, dataItem, request.UniqueFields)
		if err != nil {
			return fmt.Errorf("查找第 %d 条记录失败: %w", i+1, err)
		}

		// 如果找到记录，则删除
		if existingRecord != nil {
			if err := txApp.Delete(existingRecord); err != nil {
				return fmt.Errorf("删除第 %d 条记录失败: %w", i+1, err)
			}
		}
		// 如果没有找到记录，跳过（不报错）
	}

	return nil
}
