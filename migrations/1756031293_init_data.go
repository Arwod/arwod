package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	// 定义初始化数据配置
	var initDataRequests = []DataImportRequest{
		{
			Table:        "_superusers",
			UniqueFields: []string{"email"},
			Data: []map[string]interface{}{
				{"email": "admin@admin.com", "password": "1234!@#qwe"}, //json格式的初始化数据
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
