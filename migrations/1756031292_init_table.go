package migrations

import (
	"fmt"
	"slices"

	"github.com/pocketbase/pocketbase/core"
)

// TODO: 任务信息表 任务执行日志表
const (
	Table_Jobs = "_jobs"
)

func init() {
	// 按依赖关系排序的表创建配置
	// 创建顺序：基础表 -> 依赖表 -> 关联表
	var tableCreationRequests = []TableCreationRequest{
		{
			TableName: Table_Jobs,
			Fields: []core.Field{
				&core.TextField{Name: "name", Required: true, Max: 30},
				&core.TextField{Name: "cron", Required: true, Max: 30},
				&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"0", "1"}},
				&core.TextField{Name: "service", Required: false, Max: 128}, //系统注册的服务，
				&core.TextField{Name: "script", Required: false},            // 长度不作限制，脚本内容
				&core.TextField{Name: "remark", Max: 30},
			},
			System: true,
		},
	}

	// 注册迁移函数
	core.AppMigrations.Register(func(txApp core.App) error {
		// 创建所有表
		for _, request := range tableCreationRequests {
			if err := CreateTable(txApp, request); err != nil {
				return fmt.Errorf("failed to create table %s: %w", request.TableName, err)
			}
		}

		return nil
	}, func(txApp core.App) error {
		var allTables []string
		for _, request := range tableCreationRequests {
			allTables = append(allTables, request.TableName)
		}
		slices.Reverse(allTables)
		for _, tableName := range allTables {
			collection, err := txApp.FindCollectionByNameOrId(tableName)
			if err == nil {
				if err := txApp.Delete(collection); err != nil {
					return fmt.Errorf("failed to delete table %s: %w", tableName, err)
				}
			}
		}

		return nil
	})
}
