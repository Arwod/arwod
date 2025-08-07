package migrations

import (
	"fmt"
	"time"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	core.AppMigrations.Register(func(txApp core.App) error {
		// 为sys_user表添加默认的admin用户
		if err := addDefaultAdminUser(txApp); err != nil {
			return fmt.Errorf("添加默认admin用户失败: %w", err)
		}

		return nil
	}, func(txApp core.App) error {
		// 回滚操作：删除默认的admin用户
		if err := removeDefaultAdminUser(txApp); err != nil {
			return fmt.Errorf("删除默认admin用户失败: %w", err)
		}

		return nil
	})
}

// addDefaultAdminUser 添加默认的admin用户
func addDefaultAdminUser(txApp core.App) error {
	// 获取sys_user集合
	collection, err := txApp.FindCollectionByNameOrId("sys_user")
	if err != nil {
		return fmt.Errorf("找不到sys_user集合: %w", err)
	}

	// 检查是否已存在admin用户
	existingRecord, _ := txApp.FindFirstRecordByFilter("sys_user", "username = 'admin'")
	if existingRecord != nil {
		// 如果已存在admin用户，跳过创建
		return nil
	}

	// 创建新的用户记录
	record := core.NewRecord(collection)

	// 设置用户名
	record.Set("username", "admin")

	// 设置邮箱（PocketBase auth collection需要email字段）
	record.SetEmail("admin@example.com")

	// 设置密码（使用PocketBase的SetPassword方法）
	record.SetPassword("123!@#qwe")

	// 设置邮箱验证状态
	record.SetVerified(true)

	// 设置用户类型为管理员
	record.Set("type", "00")

	// 设置账号状态为正常
	record.Set("status", "0")

	// 设置删除标志为未删除
	record.Set("delete_flag", "0")

	// 设置创建时间
	now := time.Now()
	record.Set("created", now)
	record.Set("updated", now)

	// 设置创建者
	record.Set("created_by", "system")
	record.Set("updated_by", "system")

	// 设置备注
	record.Set("remark", "系统默认管理员账号")

	// 保存记录
	if err := txApp.Save(record); err != nil {
		return fmt.Errorf("保存admin用户失败: %w", err)
	}

	return nil
}

// removeDefaultAdminUser 删除默认的admin用户（回滚操作）
func removeDefaultAdminUser(txApp core.App) error {
	// 查找admin用户
	record, err := txApp.FindFirstRecordByFilter("sys_user", "username = 'admin'")
	if err != nil {
		// 如果找不到记录，可能已经被删除，不需要报错
		return nil
	}

	// 删除记录
	if err := txApp.Delete(record); err != nil {
		return fmt.Errorf("删除admin用户失败: %w", err)
	}

	return nil
}