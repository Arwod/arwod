package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

// 按依赖关系排序的表创建配置
// 创建顺序：基础表 -> 依赖表 -> 关联表
var tableCreationRequests = []TableCreationRequest{
	// 1. 部门表（基础表，自引用）
	{
		TableName: TableSysDept,
		Fields: []core.Field{
			&core.TextField{Name: "name", Required: true, Max: 30},
			&core.RelationField{
				Name:         "parent_id",
				Required:     false,
				MaxSelect:    1,
				CollectionId: TableSysDept, // 此处使用表名，后续会统一转换为表的id
			},
			&core.TextField{Name: "ancestors", Max: 50},
			&core.TextField{Name: "leader", Max: 20},
			&core.TextField{Name: "phone", Max: 11},
			&core.EmailField{Name: "email"},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.SelectField{Name: "delete_flag", MaxSelect: 1, Values: []string{"0", "2"}},
			&core.TextField{Name: "remark", Max: 500},
			&core.NumberField{Name: "order_num", Min: types.Pointer(0.0)},
		},
	},
	// 2. 岗位表（基础表）
	{
		TableName: TableSysPost,
		Fields: []core.Field{
			&core.TextField{Name: "code", Required: true, Max: 64},
			&core.TextField{Name: "name", Required: true, Max: 50},
			&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "remark", Max: 500},
			&core.NumberField{Name: "order_num", Min: types.Pointer(0.0)},
		},
	},
	// 3. 角色表（基础表）
	{
		TableName: TableSysRole,
		Fields: []core.Field{
			&core.TextField{Name: "name", Required: true, Max: 30},
			&core.TextField{Name: "key", Max: 100, Required: true},
			&core.SelectField{Name: "data_scope", MaxSelect: 1, Values: []string{"1", "2", "3", "4"}},
			&core.SelectField{Name: "status", Required: true, MaxSelect: 1, Values: []string{"0", "1"}},
			&core.SelectField{Name: "delete_flag", MaxSelect: 1, Values: []string{"0", "2"}},
			&core.TextField{Name: "remark", Max: 500},
			&core.NumberField{Name: "order_num", Min: types.Pointer(0.0)},
		},
		Indexes: []string{
			"CREATE UNIQUE INDEX `idx_name_%s` ON `%s` (`name`)",
		},
	},
	// 4. 菜单表（基础表，自引用）
	{
		TableName: TableSysMenu,
		Fields: []core.Field{
			&core.TextField{Name: "name", Required: true, Max: 50},
			&core.RelationField{
				Name:         "parent_id",
				Required:     false,
				MaxSelect:    1,
				CollectionId: TableSysMenu,
			},
			&core.TextField{Name: "url", Max: 200},
			&core.TextField{Name: "target", Max: 20},
			&core.SelectField{Name: "type", MaxSelect: 1, Values: []string{"M", "C", "F"}},
			&core.SelectField{Name: "visible", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.SelectField{Name: "is_refresh", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "perms", Max: 100},
			&core.TextField{Name: "icon", Max: 100},
			&core.TextField{Name: "remark", Max: 500},
			&core.NumberField{Name: "order_num", Min: types.Pointer(0.0)},
		},
	},
	// 5. 字典类型表（基础表）
	{
		TableName: TableSysDictType,
		Fields: []core.Field{
			&core.TextField{Name: "name", Max: 100},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "remark", Max: 500},
		},
	},
	// 6. 用户表（依赖部门表）
	{
		TableName: TableSysUser,
		TableType: core.CollectionTypeAuth,
		Fields: []core.Field{
			&core.TextField{Name: "username", Required: true, Max: 30},
			&core.RelationField{
				Name:         "dept_id",
				Required:     false,
				MaxSelect:    1,
				CollectionId: TableSysDept,
			},
			&core.SelectField{Name: "type", MaxSelect: 1, Values: []string{"00", "01"}},
			&core.TextField{Name: "phone", Max: 11},
			&core.SelectField{Name: "sex", MaxSelect: 1, Values: []string{"0", "1", "2"}},
			&core.TextField{Name: "salt", Max: 20},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.SelectField{Name: "delete_flag", MaxSelect: 1, Values: []string{"0", "2"}},
			&core.TextField{Name: "login_ip", Max: 128},
			&core.DateField{Name: "login_date"},
			&core.TextField{Name: "msg", Max: 255},
			&core.TextField{Name: "remark", Max: 500},
		},
	},
	// 7. 字典数据表（依赖字典类型表）
	{
		TableName: TableSysDictData,
		Fields: []core.Field{
			&core.NumberField{Name: "name", Min: types.Pointer(0.0)},
			&core.RelationField{
				Name:         "type_id",
				Required:     false,
				MaxSelect:    1,
				CollectionId: TableSysDictType,
			},
			&core.TextField{Name: "label", Max: 100},
			&core.TextField{Name: "value", Max: 100},
			&core.TextField{Name: "css_class", Max: 100},
			&core.TextField{Name: "list_class", Max: 100},
			&core.SelectField{Name: "is_default", MaxSelect: 1, Values: []string{"Y", "N"}},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "remark", Max: 500},
			&core.NumberField{Name: "order_num", Min: types.Pointer(0.0)},
		},
	},
	// 8. 系统配置表（独立表）
	{
		TableName: TableSysConfig,
		Fields: []core.Field{
			&core.TextField{Name: "name", Max: 100},
			&core.TextField{Name: "key", Max: 100},
			&core.TextField{Name: "value", Max: 500},
			&core.SelectField{Name: "type", Values: []string{"Y", "N"}, MaxSelect: 1},
			&core.TextField{Name: "remark", Max: 500},
		},
	},
	// 9. 操作日志表（独立表）
	{
		TableName: TableSysOperationLog,
		Fields: []core.Field{
			&core.TextField{Name: "title", Max: 50},
			&core.NumberField{Name: "business_type", Min: types.Pointer(0.0)},
			&core.TextField{Name: "method", Max: 100},
			&core.TextField{Name: "request_method", Max: 10},
			&core.NumberField{Name: "operator_type", Min: types.Pointer(0.0)},
			&core.TextField{Name: "oper_name", Max: 50},
			&core.TextField{Name: "dept_name", Max: 50},
			&core.TextField{Name: "oper_url", Max: 255},
			&core.TextField{Name: "oper_ip", Max: 128},
			&core.TextField{Name: "oper_location", Max: 255},
			&core.TextField{Name: "oper_param", Max: 2000},
			&core.TextField{Name: "json_result", Max: 2000},
			&core.NumberField{Name: "status", Min: types.Pointer(0.0)},
			&core.TextField{Name: "error_msg", Max: 2000},
			&core.DateField{Name: "oper_time"},
			&core.NumberField{Name: "cost_time", Min: types.Pointer(0.0)},
		},
	},
	// 10. 系统访问记录表（独立表）
	{
		TableName: TableSysLoginInfo,
		Fields: []core.Field{
			&core.TextField{Name: "name", Max: 50},
			&core.TextField{Name: "ipaddr", Max: 128},
			&core.TextField{Name: "location", Max: 255},
			&core.TextField{Name: "browser", Max: 50},
			&core.TextField{Name: "os", Max: 50},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "msg", Max: 255},
			&core.TextField{Name: "remark", Max: 500},
		},
	},
	// 11. 在线用户记录表（独立表）
	{
		TableName: TableSysUserOnline,
		Fields: []core.Field{
			&core.TextField{Name: "username", Max: 50},
			&core.TextField{Name: "dept_name", Max: 50},
			&core.TextField{Name: "ipaddr", Max: 128},
			&core.TextField{Name: "location", Max: 255},
			&core.TextField{Name: "browser", Max: 50},
			&core.TextField{Name: "os", Max: 50},
			&core.TextField{Name: "status", Max: 10},
			&core.DateField{Name: "start_timestamp"},
			&core.DateField{Name: "last_access_time"},
			&core.NumberField{Name: "expire_time", Min: types.Pointer(0.0)},
		},
	},
	// 12. 通知公告表（独立表）
	{
		TableName: TableSysNotice,
		Fields: []core.Field{
			&core.TextField{Name: "title", Required: true, Max: 50},
			&core.SelectField{Name: "type", Values: []string{"1", "2"}, Required: true, MaxSelect: 1},
			&core.EditorField{Name: "content"},
			&core.SelectField{Name: "status", MaxSelect: 1, Values: []string{"0", "1"}},
			&core.TextField{Name: "remark", Max: 500},
		},
	},
	// 13. 用户角色关联表
	{
		TableName: TableSysUserRole,
		Fields: []core.Field{
			&core.RelationField{
				Name:         "user_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysUser,
			},
			&core.RelationField{
				Name:         "role_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysRole,
			},
		},
	},
	// 14. 角色菜单关联表
	{
		TableName: TableSysRoleMenu,
		Fields: []core.Field{
			&core.RelationField{
				Name:         "role_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysRole,
			},
			&core.RelationField{
				Name:         "menu_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysMenu,
			},
		},
	},
	// 15. 角色部门关联表
	{
		TableName: TableSysRoleDept,
		Fields: []core.Field{
			&core.RelationField{
				Name:         "role_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysRole,
			},
			&core.RelationField{
				Name:         "dept_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysDept,
			},
		},
	},
	// 16. 用户岗位关联表
	{
		TableName: TableSysUserPost,
		Fields: []core.Field{
			&core.RelationField{
				Name:         "user_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysUser,
			},
			&core.RelationField{
				Name:         "position_id",
				Required:     true,
				MaxSelect:    1,
				CollectionId: TableSysPost,
			},
		},
	},
}

func init() {
	// 注册迁移函数
	core.AppMigrations.Register(func(txApp core.App) error {
		// 删除默认的users表
		if err := deleteUsersTable(txApp); err != nil {
			return fmt.Errorf("failed to delete users table: %w", err)
		}

		// 创建所有表
		for _, request := range tableCreationRequests {
			if err := CreateTable(txApp, request); err != nil {
				return fmt.Errorf("failed to create table %s: %w", request.TableName, err)
			}
		}

		return nil
	}, func(txApp core.App) error {
		// 回滚操作：删除所有创建的表
		allTables := append([]string{},
			TableSysUserPost, TableSysRoleDept, TableSysRoleMenu, TableSysUserRole, // 关联表先删除
			TableSysNotice, TableSysUserOnline, TableSysLoginInfo, TableSysConfig,
			TableSysOperationLog, TableSysDictData, TableSysDictType, TableSysMenu,
			TableSysUser, TableSysRole, TableSysPost, TableSysDept, // 基础表后删除
		)

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
