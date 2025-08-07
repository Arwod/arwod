package migrations

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

// DataImportRequest 数据导入请求结构
type DataImportRequest struct {
	Table        string                   `json:"table"`
	Data         []map[string]interface{} `json:"data"`
	UniqueFields []string                 `json:"uniqueFields"`
}

// ImportDataFromJSON 通用数据填充函数
// 接受JSON格式的输入，根据table确定要填充的表，根据data确定要填充的数据
// 参数:
//   - txApp: 数据库应用实例
//   - jsonData: JSON格式的数据，格式为 {"table":"表名","data":[{"字段1":"值1","字段2":"值2"}]}
//   - uniqueField: 用于检查重复的唯一字段名（为空则直接追加新记录，不为空则根据该字段检查重复并更新）
//
// 返回:
//   - error: 错误信息
func ImportDataFromJSON(txApp core.App, jsonData string, uniqueField string) error {
	// 解析JSON数据
	var request DataImportRequest
	if err := json.Unmarshal([]byte(jsonData), &request); err != nil {
		return fmt.Errorf("解析JSON数据失败: %w", err)
	}

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
	validFields := make(map[string]bool)
	for _, field := range collection.Fields {
		validFields[field.GetName()] = true
	}
	// 添加系统字段
	validFields["id"] = true
	validFields["created"] = true
	validFields["updated"] = true

	// 插入或更新数据
	for i, dataItem := range request.Data {
		var record *core.Record
		var isUpdate bool

		// 如果指定了唯一字段，检查是否已存在
		if uniqueField != "" {
			if uniqueValue, exists := dataItem[uniqueField]; exists {
				existingRecord, _ := txApp.FindFirstRecordByFilter(request.Table, fmt.Sprintf("%s = '%v'", uniqueField, uniqueValue))
				if existingRecord != nil {
					// 使用已存在的记录进行更新
					record = existingRecord
					isUpdate = true
				}
			}
		}

		// 如果没有找到已存在的记录，创建新记录
		if record == nil {
			record = core.NewRecord(collection)
			isUpdate = false
		}

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

// ImportDataFromStruct 通用数据填充函数（结构体版本）
// 接受结构体格式的输入，根据table确定要填充的表，根据data确定要填充的数据
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体
//   - uniqueField: 用于检查重复的唯一字段名（为空则直接追加新记录，不为空则根据该字段检查重复并更新）
//
// 返回:
//   - error: 错误信息
func ImportDataFromStruct(txApp core.App, request DataImportRequest, uniqueField string) error {
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
	validFields := make(map[string]bool)
	for _, field := range collection.Fields {
		validFields[field.GetName()] = true
	}
	// 添加系统字段
	validFields["id"] = true
	validFields["created"] = true
	validFields["updated"] = true

	// 插入或更新数据
	for i, dataItem := range request.Data {
		var record *core.Record
		var isUpdate bool

		// 如果指定了唯一字段，检查是否已存在
		if uniqueField != "" {
			if uniqueValue, exists := dataItem[uniqueField]; exists {
				existingRecord, _ := txApp.FindFirstRecordByFilter(request.Table, fmt.Sprintf("%s = '%v'", uniqueField, uniqueValue))
				if existingRecord != nil {
					// 使用已存在的记录进行更新
					record = existingRecord
					isUpdate = true
				}
			}
		}

		// 如果没有找到已存在的记录，创建新记录
		if record == nil {
			record = core.NewRecord(collection)
			isUpdate = false
		}

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

// AppendDataFromJSON 追加数据函数（JSON版本）
// 直接追加新记录，不检查重复
// 参数:
//   - txApp: 数据库应用实例
//   - jsonData: JSON格式的数据
//
// 返回:
//   - error: 错误信息
func AppendDataFromJSON(txApp core.App, jsonData string) error {
	return ImportDataFromJSON(txApp, jsonData, "")
}

// AppendDataFromStruct 追加数据函数（结构体版本）
// 直接追加新记录，不检查重复
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体
//
// 返回:
//   - error: 错误信息
func AppendDataFromStruct(txApp core.App, request DataImportRequest) error {
	return ImportDataFromStruct(txApp, request, "")
}

// ImportDataFromStructWithUniqueFields 通用数据填充函数（使用结构体中的UniqueFields）
// 从结构体的UniqueFields字段读取唯一字段配置
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体（包含UniqueFields配置）
//
// 返回:
//   - error: 错误信息
func ImportDataFromStructWithUniqueFields(txApp core.App, request DataImportRequest) error {
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
	validFields := make(map[string]bool)
	for _, field := range collection.Fields {
		validFields[field.GetName()] = true
	}
	// 添加系统字段
	validFields["id"] = true
	validFields["created"] = true
	validFields["updated"] = true

	// 插入或更新数据
	for i, dataItem := range request.Data {
		var record *core.Record
		var isUpdate bool

		// 如果指定了唯一字段，检查是否已存在
		if len(request.UniqueFields) > 0 {
			// 构建查询条件
			var conditions []string
			for _, uniqueField := range request.UniqueFields {
				if uniqueValue, exists := dataItem[uniqueField]; exists {
					conditions = append(conditions, fmt.Sprintf("%s = '%v'", uniqueField, uniqueValue))
				}
			}

			if len(conditions) > 0 {
				filter := fmt.Sprintf("(%s)", conditions[0])
				for j := 1; j < len(conditions); j++ {
					filter += fmt.Sprintf(" && (%s)", conditions[j])
				}

				existingRecord, _ := txApp.FindFirstRecordByFilter(request.Table, filter)
				if existingRecord != nil {
					// 使用已存在的记录进行更新
					record = existingRecord
					isUpdate = true
				}
			}
		}

		// 如果没有找到已存在的记录，创建新记录
		if record == nil {
			record = core.NewRecord(collection)
			isUpdate = false
		}

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

// UpsertDataFromJSON 通用数据插入或更新函数（Upsert操作）
// 接受JSON格式的输入，根据table确定要操作的表，根据data确定要操作的数据
// 与ImportDataFromJSON的区别：此函数专门用于Upsert操作，必须指定uniqueField
// 参数:
//   - txApp: 数据库应用实例
//   - jsonData: JSON格式的数据，格式为 {"table":"表名","data":[{"字段1":"值1","字段2":"值2"}]}
//   - uniqueField: 用于检查重复的唯一字段名（必须指定）
//
// 返回:
//   - error: 错误信息
func UpsertDataFromJSON(txApp core.App, jsonData string, uniqueField string) error {
	if uniqueField == "" {
		return fmt.Errorf("Upsert操作必须指定uniqueField")
	}
	return ImportDataFromJSON(txApp, jsonData, uniqueField)
}

// UpsertDataFromStruct 通用数据插入或更新函数（结构体版本）
// 接受结构体格式的输入，根据table确定要操作的表，根据data确定要操作的数据
// 与ImportDataFromStruct的区别：此函数专门用于Upsert操作，必须指定uniqueField
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体
//   - uniqueField: 用于检查重复的唯一字段名（必须指定）
//
// 返回:
//   - error: 错误信息
func UpsertDataFromStruct(txApp core.App, request DataImportRequest, uniqueField string) error {
	if uniqueField == "" {
		return fmt.Errorf("Upsert操作必须指定uniqueField")
	}
	return ImportDataFromStruct(txApp, request, uniqueField)
}

// UpdateDataFromJSON 通用数据更新函数（仅更新已存在的记录）
// 接受JSON格式的输入，根据table确定要更新的表，根据data确定要更新的数据
// 只更新已存在的记录，不会创建新记录
// 参数:
//   - txApp: 数据库应用实例
//   - jsonData: JSON格式的数据，格式为 {"table":"表名","data":[{"字段1":"值1","字段2":"值2"}]}
//   - uniqueField: 用于查找记录的唯一字段名（必须指定）
//
// 返回:
//   - error: 错误信息
func UpdateDataFromJSON(txApp core.App, jsonData string, uniqueField string) error {
	if uniqueField == "" {
		return fmt.Errorf("更新操作必须指定uniqueField")
	}

	// 解析JSON数据
	var request DataImportRequest
	if err := json.Unmarshal([]byte(jsonData), &request); err != nil {
		return fmt.Errorf("解析JSON数据失败: %w", err)
	}

	return UpdateDataFromStruct(txApp, request, uniqueField)
}

// UpdateDataFromStruct 通用数据更新函数（结构体版本，仅更新已存在的记录）
// 接受结构体格式的输入，根据table确定要更新的表，根据data确定要更新的数据
// 只更新已存在的记录，不会创建新记录
// 参数:
//   - txApp: 数据库应用实例
//   - request: 数据导入请求结构体
//   - uniqueField: 用于查找记录的唯一字段名（必须指定）
//
// 返回:
//   - error: 错误信息
func UpdateDataFromStruct(txApp core.App, request DataImportRequest, uniqueField string) error {
	if uniqueField == "" {
		return fmt.Errorf("更新操作必须指定uniqueField")
	}

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
	validFields := make(map[string]bool)
	for _, field := range collection.Fields {
		validFields[field.GetName()] = true
	}
	// 添加系统字段
	validFields["id"] = true
	validFields["created"] = true
	validFields["updated"] = true

	// 更新数据
	for i, dataItem := range request.Data {
		// 检查uniqueField是否存在于数据中
		uniqueValue, exists := dataItem[uniqueField]
		if !exists {
			return fmt.Errorf("第 %d 条数据缺少唯一字段 %s", i+1, uniqueField)
		}

		// 查找已存在的记录
		existingRecord, err := txApp.FindFirstRecordByFilter(request.Table, fmt.Sprintf("%s = '%v'", uniqueField, uniqueValue))
		if err != nil {
			return fmt.Errorf("查找第 %d 条记录失败: %w", i+1, err)
		}
		if existingRecord == nil {
			return fmt.Errorf("第 %d 条数据对应的记录不存在，无法更新 (%s = %v)", i+1, uniqueField, uniqueValue)
		}

		// 设置字段值，只设置表中存在的字段
		for fieldName, fieldValue := range dataItem {
			if validFields[fieldName] {
				existingRecord.Set(fieldName, fieldValue)
			}
		}

		// 更新时间总是设置为当前时间（除非手动指定）
		if _, exists := dataItem["updated"]; !exists {
			existingRecord.Set("updated", time.Now())
		}

		// 保存记录
		if err := txApp.Save(existingRecord); err != nil {
			return fmt.Errorf("更新第 %d 条数据失败: %w", i+1, err)
		}
	}

	return nil
}

// DeleteDataByFilter 通用数据删除函数
// 根据过滤条件删除指定表中的数据，常用于迁移回滚操作
// 参数:
//   - txApp: 数据库应用实例
//   - tableName: 表名
//   - filter: 过滤条件，例如 "created_by = 'system'"
//
// 返回:
//   - error: 错误信息
func DeleteDataByFilter(txApp core.App, tableName string, filter string) error {
	if tableName == "" {
		return fmt.Errorf("表名不能为空")
	}
	if filter == "" {
		return fmt.Errorf("过滤条件不能为空")
	}

	// 查找符合条件的记录
	records, err := txApp.FindRecordsByFilter(tableName, filter, "", 0, 0)
	if err != nil {
		return fmt.Errorf("查找记录失败: %w", err)
	}

	// 删除找到的记录
	for _, record := range records {
		if err := txApp.Delete(record); err != nil {
			return fmt.Errorf("删除记录失败: %w", err)
		}
	}

	return nil
}
