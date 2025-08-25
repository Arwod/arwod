package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	// 定义JavaScript脚本管理系统初始化数据配置
	var initDataRequests = []DataImportRequest{
		{
			Table:        Table_JsScripts,
			UniqueFields: []string{"name"},
			Data: []map[string]interface{}{
				{
					"name":           "示例Hook脚本",
					"description":    "演示如何在记录创建时执行自定义逻辑",
					"content":        "// 示例Hook脚本\nconsole.log('记录创建Hook被触发:', JSON.stringify(context));\n\n// 获取创建的记录\nconst record = context.record;\nconsole.log('新创建的记录ID:', record.id);\n\n// 可以在这里添加自定义业务逻辑\n// 例如：发送通知、更新相关数据等\n\nreturn { success: true, message: '脚本执行成功' };",
					"trigger_type":   "hook",
					"trigger_config": `{\"hookType\": \"onRecordCreate\", \"collections\": [\"users\"]}`,
					"status":         "active",
					"timeout":        30,
					"tags":           "示例,Hook,记录创建",
					"version":        1,
					"is_system":      false,
					"created_by":     "system",
					"updated_by":     "system",
				},
				{
					"name":           "定时清理脚本",
					"description":    "定时清理过期的执行日志",
					"content":        "// 定时清理脚本\nconsole.log('开始清理过期的执行日志...');\n\n// 计算30天前的时间\nconst thirtyDaysAgo = new Date();\nthirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);\n\ntry {\n    // 查询过期的执行日志\n    const expiredLogs = $app.findRecordsByFilter(\n        'js_execution_logs',\n        `start_time < '${thirtyDaysAgo.toISOString()}'`\n    );\n    \n    console.log(`找到 ${expiredLogs.length} 条过期日志`);\n    \n    // 删除过期日志\n    let deletedCount = 0;\n    for (const log of expiredLogs) {\n        $app.delete(log);\n        deletedCount++;\n    }\n    \n    console.log(`成功删除 ${deletedCount} 条过期日志`);\n    return { success: true, deletedCount: deletedCount };\n    \n} catch (error) {\n    console.error('清理过程中发生错误:', error);\n    return { success: false, error: error.message };\n}",
					"trigger_type":   "cron",
					"trigger_config": `{\"cron\": \"0 2 * * *\", \"timezone\": \"Asia/Shanghai\"}`,
					"status":         "active",
					"timeout":        60,
					"tags":           "清理,定时任务,日志管理",
					"version":        1,
					"is_system":      true,
					"created_by":     "system",
					"updated_by":     "system",
				},
				{
					"name":           "数据验证脚本",
					"description":    "手动执行的数据验证脚本",
					"content":        "// 数据验证脚本\nconsole.log('开始执行数据验证...');\n\nlet validationResults = {\n    totalRecords: 0,\n    validRecords: 0,\n    invalidRecords: 0,\n    errors: []\n};\n\ntry {\n    // 获取所有用户记录\n    const users = $app.findRecordsByFilter('users', '');\n    validationResults.totalRecords = users.length;\n    \n    for (const user of users) {\n        try {\n            // 验证邮箱格式\n            const emailRegex = /^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$/;\n            if (!emailRegex.test(user.email)) {\n                validationResults.errors.push({\n                    recordId: user.id,\n                    field: 'email',\n                    value: user.email,\n                    error: '邮箱格式不正确'\n                });\n                validationResults.invalidRecords++;\n            } else {\n                validationResults.validRecords++;\n            }\n        } catch (error) {\n            validationResults.errors.push({\n                recordId: user.id,\n                error: error.message\n            });\n            validationResults.invalidRecords++;\n        }\n    }\n    \n    console.log('验证完成:', JSON.stringify(validationResults, null, 2));\n    return validationResults;\n    \n} catch (error) {\n    console.error('验证过程中发生错误:', error);\n    return { success: false, error: error.message };\n}",
					"trigger_type":   "manual",
					"trigger_config": `{}`,
					"status":         "active",
					"timeout":        120,
					"tags":           "验证,手动执行,数据质量",
					"version":        1,
					"is_system":      false,
					"created_by":     "system",
					"updated_by":     "system",
				},
			},
		},
	}

	core.AppMigrations.Register(func(txApp core.App) error {
		// 循环执行数据初始化
		for i, request := range initDataRequests {
			if err := ImportData(txApp, request); err != nil {
				return fmt.Errorf("初始化第 %d 组数据失败 (表: %s): %w", i+1, request.Table, err)
			}
		}
		return nil
	}, func(txApp core.App) error {
		// 循环执行数据回滚（按相反顺序使用 RollbackData）
		for i := len(initDataRequests) - 1; i >= 0; i-- {
			request := initDataRequests[i]
			if err := RollbackData(txApp, request); err != nil {
				return fmt.Errorf("回滚第 %d 组数据失败 (表: %s): %w", i+1, request.Table, err)
			}
		}
		return nil
	})
}