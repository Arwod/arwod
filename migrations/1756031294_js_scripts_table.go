package migrations

import (
	"fmt"
	"slices"

	"github.com/pocketbase/pocketbase/core"
)

// 辅助函数：创建float64指针
func float64Ptr(v float64) *float64 {
	return &v
}

// JavaScript脚本管理系统表名常量
const (
	Table_EventsDict      = "_events_dict"
	Table_JsScripts       = "js_scripts"
	Table_JsExecutionLogs = "js_execution_logs"
)

func init() {
	// 按依赖关系排序的表创建配置
	// 创建顺序：js_scripts -> js_execution_logs
	var tableCreationRequests = []TableCreationRequest{
		// 创建一个事件字典表，系统支持的可以通过js拓展的事件类型通过字典维护
		{
			TableName: Table_EventsDict,
			Fields: []core.Field{
				&core.TextField{Name: "name", Required: true, Max: 100},
				&core.RelationField{Name: "parent_id", Required: false, MaxSelect: 1, CollectionId: Table_EventsDict},
				&core.TextField{Name: "description", Required: false, Max: 500},
			},
			System: true,
		},
		{
			TableName: Table_JsScripts,
			Fields: []core.Field{
				&core.TextField{Name: "name", Required: true, Max: 100},
				&core.TextField{Name: "description", Required: false, Max: 500},
				&core.TextField{Name: "content", Required: true}, // 脚本内容，不限制长度
				&core.RelationField{Name: "trigger_type", Required: true, MaxSelect: 1, CollectionId: Table_EventsDict},
				&core.TextField{Name: "trigger_config", Required: false}, // JSON格式的触发器配置
				&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"active", "inactive"}},
				&core.NumberField{Name: "timeout", Required: false, Min: float64Ptr(1.0), Max: float64Ptr(300.0)}, // 超时时间（秒）
				&core.TextField{Name: "tags", Required: false, Max: 200},                                          // 标签，逗号分隔
				&core.NumberField{Name: "version", Required: true, Min: float64Ptr(1.0)},                          // 版本号
				&core.BoolField{Name: "is_system", Required: false},                                               // 是否系统脚本
			},
			Indexes: []string{
				"CREATE UNIQUE INDEX `idx_js_scripts_name_%s` ON `%s` (`name`)",
				"CREATE INDEX `idx_js_scripts_status_%s` ON `%s` (`status`)",
				"CREATE INDEX `idx_js_scripts_trigger_type_%s` ON `%s` (`trigger_type`)",
				"CREATE INDEX `idx_js_scripts_created_%s` ON `%s` (`created`)",
			},
			System: false,
		},
		{
			TableName: Table_JsExecutionLogs,
			Fields: []core.Field{
				&core.RelationField{Name: "script_id", Required: true, CollectionId: Table_JsScripts, MaxSelect: 1},
				&core.TextField{Name: "script_name", Required: true, Max: 100}, // 冗余字段，便于查询
				&core.SelectField{Name: "trigger_type", Required: true, MaxSelect: 1, Values: []string{"manual", "hook", "cron"}},
				&core.TextField{Name: "trigger_context", Required: false}, // JSON格式的触发上下文
				&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"running", "success", "error", "timeout"}},
				&core.DateField{Name: "start_time", Required: true},
				&core.DateField{Name: "end_time", Required: false},
				&core.NumberField{Name: "execution_time", Required: false, Min: float64Ptr(0.0)}, // 执行时间（毫秒）
				&core.TextField{Name: "output", Required: false},                                 // 脚本输出
				&core.TextField{Name: "error_message", Required: false},                          // 错误信息
				&core.TextField{Name: "stack_trace", Required: false},                            // 错误堆栈
				&core.NumberField{Name: "memory_usage", Required: false, Min: float64Ptr(0.0)},   // 内存使用量（KB）
				&core.NumberField{Name: "cpu_usage", Required: false, Min: float64Ptr(0.0)},      // CPU使用率（%）
			},
			Indexes: []string{
				"CREATE INDEX `idx_js_execution_logs_script_id_%s` ON `%s` (`script_id`)",
				"CREATE INDEX `idx_js_execution_logs_status_%s` ON `%s` (`status`)",
				"CREATE INDEX `idx_js_execution_logs_start_time_%s` ON `%s` (`start_time`)",
				"CREATE INDEX `idx_js_execution_logs_trigger_type_%s` ON `%s` (`trigger_type`)",
				"CREATE INDEX `idx_js_execution_logs_script_name_%s` ON `%s` (`script_name`)",
			},
			System: false,
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
		// 回滚：删除所有表（按相反顺序）
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
