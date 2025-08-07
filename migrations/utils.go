package migrations

import (
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
