package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func init() {
	core.AppMigrations.Register(func(txApp core.App) error {
		// 删除users表（如果存在）
		if err := deleteUsersTable(txApp); err != nil {
			return fmt.Errorf("users table deletion error: %w", err)
		}

		// 创建部门表
		if err := createSysDeptCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysDept, err)
		}

		// 创建用户表（在所有基础表创建之后，关联表之前）
		if err := createSysUserCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysUser, err)
		}

		// 创建岗位表
		if err := createSysPostCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysPost, err)
		}

		// 创建角色表
		if err := createSysRoleCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysRole, err)
		}

		// 创建菜单表
		if err := createSysMenuCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysMenu, err)
		}

		// 创建操作日志表
		if err := createSysOperLogCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysOperLog, err)
		}

		// 创建字典类型表
		if err := createSysDictTypeCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysDictType, err)
		}

		// 创建字典数据表
		if err := createSysDictDataCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysDictData, err)
		}

		// 创建参数配置表
		if err := createSysConfigCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysConfig, err)
		}

		// 创建系统访问记录表
		if err := createSysLogininforCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysLogininfor, err)
		}

		// 创建在线用户记录表
		if err := createSysUserOnlineCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysUserOnline, err)
		}

		// 创建通知公告表
		if err := createSysNoticeCollection(txApp); err != nil {
			return fmt.Errorf("%s creation error: %w", TableSysNotice, err)
		}

		// 添加关联字段（在所有表创建完成后）
		if err := addRelationFields(txApp); err != nil {
			return fmt.Errorf("relation fields creation error: %w", err)
		}

		return nil
	}, func(txApp core.App) error {
		// 回滚时删除所有创建的表
		tables := []string{
			TableSysNotice,
			TableSysUserOnline,
			TableSysLogininfor,
			TableSysConfig,
			TableSysDictData,
			TableSysDictType,
			TableSysOperLog,
			TableSysUserPost,
			TableSysRoleDept,
			TableSysRoleMenu,
			TableSysUserRole,
			TableSysUser,
			TableSysMenu,
			TableSysRole,
			TableSysPost,
			TableSysDept,
		}

		for _, name := range tables {
			collection, err := txApp.FindCollectionByNameOrId(name)
			if err == nil {
				if err := txApp.Delete(collection); err != nil {
					return fmt.Errorf("failed to delete collection %s: %w", name, err)
				}
			}
		}

		return nil
	})
}

// 创建部门表
func createSysDeptCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysDept)
	col.System = false

	// 添加字段（parent_id将在addRelationFields中添加）
	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      30,
	})
	col.Fields.Add(&core.TextField{
		Name:     "ancestors",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "leader",
		Required: false,
		Max:      20,
	})
	col.Fields.Add(&core.TextField{
		Name:     "phone",
		Required: false,
		Max:      11,
	})
	col.Fields.Add(&core.EmailField{
		Name:     "email",
		Required: false,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "delete_flag",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "2"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "order_num",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建字典类型表
func createSysDictTypeCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysDictType)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "type",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建字典数据表
func createSysDictDataCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysDictData)
	col.System = false

	col.Fields.Add(&core.NumberField{
		Name:     "code",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "label",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "value",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "type",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "css_class",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "list_class",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "is_default",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"Y", "N"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "order_num",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建参数配置表
func createSysConfigCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysConfig)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "key",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "value",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"Y", "N"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建系统访问记录表
func createSysLogininforCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysLogininfor)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "ipaddr",
		Required: false,
		Max:      128,
	})
	col.Fields.Add(&core.TextField{
		Name:     "location",
		Required: false,
		Max:      255,
	})
	col.Fields.Add(&core.TextField{
		Name:     "browser",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "os",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "msg",
		Required: false,
		Max:      255,
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建在线用户记录表
func createSysUserOnlineCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysUserOnline)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "username",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "dept_name",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "ipaddr",
		Required: false,
		Max:      128,
	})
	col.Fields.Add(&core.TextField{
		Name:     "location",
		Required: false,
		Max:      255,
	})
	col.Fields.Add(&core.TextField{
		Name:     "browser",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "os",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "status",
		Required: false,
		Max:      10,
	})
	col.Fields.Add(&core.DateField{
		Name:     "start_timestamp",
		Required: false,
	})
	col.Fields.Add(&core.DateField{
		Name:     "last_access_time",
		Required: false,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "expire_time",
		Required: false,
		Min:      types.Pointer(0.0),
	})

	return txApp.Save(col)
}

// 删除users表
func deleteUsersTable(txApp core.App) error {
	return nil
	// 查找users表
	usersCollection, err := txApp.FindCollectionByNameOrId("users")
	if err != nil {
		// 如果表不存在，直接返回成功
		return nil
	}

	// 删除users表
	if err := txApp.Delete(usersCollection); err != nil {
		return fmt.Errorf("failed to delete users collection: %w", err)
	}

	return nil
}

// addRelationFields 在所有表创建完成后添加关联字段和创建关联表
func addRelationFields(txApp core.App) error {
	// 为sys_user表添加dept_id关联字段
	sysUserCol, err := txApp.FindCollectionByNameOrId(TableSysUser)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysUser, err)
	}

	// 获取sys_dept集合的实际ID
	sysDeptCol, err := txApp.FindCollectionByNameOrId(TableSysDept)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysDept, err)
	}

	// 获取菜单表集合
	sysMenuCol, err := txApp.FindCollectionByNameOrId(TableSysMenu)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysMenu, err)
	}

	// 添加部门ID关联字段
	sysUserCol.Fields.Add(&core.RelationField{
		Name:         "dept_id",
		Required:     false,
		CollectionId: sysDeptCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})

	if err := txApp.Save(sysUserCol); err != nil {
		return fmt.Errorf("failed to update %s collection: %w", TableSysUser, err)
	}

	// 为部门表添加parent_id自引用关联字段
	sysDeptCol.Fields.Add(&core.RelationField{
		Name:         "parent_id",
		Required:     false,
		CollectionId: sysDeptCol.Id, // 自引用
		MaxSelect:    1,
	})

	if err := txApp.Save(sysDeptCol); err != nil {
		return fmt.Errorf("failed to update %s collection: %w", TableSysDept, err)
	}

	// 为菜单表添加parent_id自引用关联字段
	sysMenuCol.Fields.Add(&core.RelationField{
		Name:         "parent_id",
		Required:     false,
		CollectionId: sysMenuCol.Id, // 自引用
		MaxSelect:    1,
	})

	if err := txApp.Save(sysMenuCol); err != nil {
		return fmt.Errorf("failed to update %s collection: %w", TableSysMenu, err)
	}

	// 为字典数据表添加字典类型关联字段
	sysDictDataCol, err := txApp.FindCollectionByNameOrId(TableSysDictData)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysDictData, err)
	}

	sysDictTypeCol, err := txApp.FindCollectionByNameOrId(TableSysDictType)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysDictType, err)
	}

	// 添加字典类型关联字段
	sysDictDataCol.Fields.Add(&core.RelationField{
		Name:         "dict_type_id",
		Required:     false,
		CollectionId: sysDictTypeCol.Id,
		MaxSelect:    1,
	})

	if err := txApp.Save(sysDictDataCol); err != nil {
		return fmt.Errorf("failed to update %s collection: %w", TableSysDictData, err)
	}

	// 创建关联表
	if err := createSysUserRoleCollection(txApp); err != nil {
		return fmt.Errorf("%s creation error: %w", TableSysUserRole, err)
	}

	if err := createSysRoleMenuCollection(txApp); err != nil {
		return fmt.Errorf("%s creation error: %w", TableSysRoleMenu, err)
	}

	if err := createSysRoleDeptCollection(txApp); err != nil {
		return fmt.Errorf("%s creation error: %w", TableSysRoleDept, err)
	}

	if err := createSysUserPostCollection(txApp); err != nil {
		return fmt.Errorf("%s creation error: %w", TableSysUserPost, err)
	}

	return nil
}

// 创建用户表
func createSysUserCollection(txApp core.App) error {
	col := core.NewAuthCollection(TableSysUser, "_pb_sys_users_auth_")
	col.Type = core.CollectionTypeAuth
	col.Name = TableSysUser
	col.System = false

	// 登录账号
	col.Fields.Add(&core.TextField{
		Name:     "username",
		Required: true,
		Max:      30,
	})

	// 用户类型
	col.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"00", "01"},
	})

	// 手机号码
	col.Fields.Add(&core.TextField{
		Name:     "phone",
		Required: false,
		Max:      11,
	})

	// 用户性别
	col.Fields.Add(&core.SelectField{
		Name:      "sex",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1", "2"},
	})

	// 盐加密
	col.Fields.Add(&core.TextField{
		Name:     "salt",
		Required: false,
		Max:      20,
	})

	// 账号状态
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})

	// 删除标志
	col.Fields.Add(&core.SelectField{
		Name:      "delete_flag",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "2"},
	})

	// 最后登录IP
	col.Fields.Add(&core.TextField{
		Name:     "login_ip",
		Required: false,
		Max:      128,
	})

	// 最后登录时间
	col.Fields.Add(&core.DateField{
		Name:     "login_date",
		Required: false,
	})

	// 密码最后更新时间
	col.Fields.Add(&core.DateField{
		Name:     "pwd_update_date",
		Required: false,
	})

	// 创建者
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})

	// 创建时间
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	// 更新者
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})

	// 更新时间
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	// 备注
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})

	return txApp.Save(col)
}

// 创建岗位表
func createSysPostCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysPost)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "code",
		Required: true,
		Max:      64,
	})
	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      50,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  true,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "order_num",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}

// 创建角色表
func createSysRoleCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysRole)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      30,
	})
	col.Fields.Add(&core.TextField{
		Name:     "key",
		Required: true,
		Max:      100,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "data_scope",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"1", "2", "3", "4"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  true,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "delete_flag",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "2"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "order_num",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})

	return txApp.Save(col)
}

// 创建菜单表
func createSysMenuCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysMenu)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "name",
		Required: true,
		Max:      50,
	})
	// parent_id字段将在addRelationFields中添加
	col.Fields.Add(&core.TextField{
		Name:     "url",
		Required: false,
		Max:      200,
	})
	col.Fields.Add(&core.TextField{
		Name:     "target",
		Required: false,
		Max:      20,
	})
	col.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"M", "C", "F"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "visible",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.SelectField{
		Name:      "is_refresh",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})
	col.Fields.Add(&core.TextField{
		Name:     "perms",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "icon",
		Required: false,
		Max:      100,
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "order_num",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})

	return txApp.Save(col)
}

// 创建用户角色关联表
func createSysUserRoleCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysUserRole)
	col.System = false

	// 获取关联集合的实际ID
	sysUserCol, err := txApp.FindCollectionByNameOrId(TableSysUser)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysUser, err)
	}
	sysRoleCol, err := txApp.FindCollectionByNameOrId(TableSysRole)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysRole, err)
	}

	col.Fields.Add(&core.RelationField{
		Name:         "user_id",
		Required:     true,
		CollectionId: sysUserCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})
	col.Fields.Add(&core.RelationField{
		Name:         "role_id",
		Required:     true,
		CollectionId: sysRoleCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})

	return txApp.Save(col)
}

// 创建角色菜单关联表
func createSysRoleMenuCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysRoleMenu)
	col.System = false

	// 获取关联集合的实际ID
	sysRoleCol, err := txApp.FindCollectionByNameOrId(TableSysRole)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysRole, err)
	}
	sysMenuCol, err := txApp.FindCollectionByNameOrId(TableSysMenu)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysMenu, err)
	}

	col.Fields.Add(&core.RelationField{
		Name:         "role_id",
		Required:     true,
		CollectionId: sysRoleCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})
	col.Fields.Add(&core.RelationField{
		Name:         "menu_id",
		Required:     true,
		CollectionId: sysMenuCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})

	return txApp.Save(col)
}

// 创建角色部门关联表
func createSysRoleDeptCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysRoleDept)
	col.System = false

	// 获取关联集合的实际ID
	sysRoleCol, err := txApp.FindCollectionByNameOrId(TableSysRole)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysRole, err)
	}
	sysDeptCol, err := txApp.FindCollectionByNameOrId(TableSysDept)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysDept, err)
	}

	col.Fields.Add(&core.RelationField{
		Name:         "role_id",
		Required:     true,
		CollectionId: sysRoleCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})
	col.Fields.Add(&core.RelationField{
		Name:         "dept_id",
		Required:     true,
		CollectionId: sysDeptCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})

	return txApp.Save(col)
}

// 创建用户岗位关联表
func createSysUserPostCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysUserPost)
	col.System = false

	// 获取关联集合的实际ID
	sysUserCol, err := txApp.FindCollectionByNameOrId(TableSysUser)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysUser, err)
	}
	sysPostCol, err := txApp.FindCollectionByNameOrId(TableSysPost)
	if err != nil {
		return fmt.Errorf("failed to find %s collection: %w", TableSysPost, err)
	}

	col.Fields.Add(&core.RelationField{
		Name:         "user_id",
		Required:     true,
		CollectionId: sysUserCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})
	col.Fields.Add(&core.RelationField{
		Name:         "post_id",
		Required:     true,
		CollectionId: sysPostCol.Id, // 使用实际的集合ID
		MaxSelect:    1,
	})

	return txApp.Save(col)
}

// 创建操作日志表
func createSysOperLogCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysOperLog)
	col.System = false

	col.Fields.Add(&core.TextField{
		Name:     "title",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "business_type",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "method",
		Required: false,
		Max:      200,
	})
	col.Fields.Add(&core.TextField{
		Name:     "request_method",
		Required: false,
		Max:      10,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "type",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "username",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "dept_name",
		Required: false,
		Max:      50,
	})
	col.Fields.Add(&core.TextField{
		Name:     "url",
		Required: false,
		Max:      200,
	})
	col.Fields.Add(&core.TextField{
		Name:     "ip",
		Required: false,
		Max:      128,
	})
	col.Fields.Add(&core.TextField{
		Name:     "location",
		Required: false,
		Max:      255,
	})
	col.Fields.Add(&core.TextField{
		Name:     "param",
		Required: false,
		Max:      2000,
	})
	col.Fields.Add(&core.TextField{
		Name:     "json_result",
		Required: false,
		Max:      2000,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "status",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "error_msg",
		Required: false,
		Max:      2000,
	})
	col.Fields.Add(&core.DateField{
		Name:     "time",
		Required: false,
	})
	col.Fields.Add(&core.NumberField{
		Name:     "cost_time",
		Required: false,
		Min:      types.Pointer(0.0),
	})
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      500,
	})
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	return txApp.Save(col)
}

// 创建通知公告表
func createSysNoticeCollection(txApp core.App) error {
	col := core.NewBaseCollection(TableSysNotice)
	col.System = false

	// 公告标题
	col.Fields.Add(&core.TextField{
		Name:     "title",
		Required: true,
		Max:      50,
	})

	// 公告类型（1通知 2公告）
	col.Fields.Add(&core.SelectField{
		Name:      "type",
		Required:  true,
		MaxSelect: 1,
		Values:    []string{"1", "2"},
	})

	// 公告内容
	col.Fields.Add(&core.EditorField{
		Name:     "content",
		Required: false,
	})

	// 公告状态（0正常 1关闭）
	col.Fields.Add(&core.SelectField{
		Name:      "status",
		Required:  false,
		MaxSelect: 1,
		Values:    []string{"0", "1"},
	})

	// 备注
	col.Fields.Add(&core.TextField{
		Name:     "remark",
		Required: false,
		Max:      255,
	})

	// 创建者
	col.Fields.Add(&core.TextField{
		Name:     "created_by",
		Required: false,
		Max:      64,
	})

	// 创建时间
	col.Fields.Add(&core.AutodateField{
		Name:     "created",
		OnCreate: true,
	})

	// 更新者
	col.Fields.Add(&core.TextField{
		Name:     "updated_by",
		Required: false,
		Max:      64,
	})

	// 更新时间
	col.Fields.Add(&core.AutodateField{
		Name:     "updated",
		OnCreate: true,
		OnUpdate: true,
	})

	return txApp.Save(col)
}
