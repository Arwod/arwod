package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

// 定义初始化数据配置
var initDataRequests = []DataImportRequest{
	// 字典类型数据
	{
		Table:        TableSysDictType,
		UniqueFields: []string{"id"},
		Data: []map[string]interface{}{
			{"id": "upbepvx05k6j1pb", "name": "用户性别", "remark": "用户性别列表", "status": "0", "type": "sys_user_sex"},
			{"id": "9ew4nxq6202v35w", "name": "菜单状态", "remark": "菜单状态列表", "status": "0", "type": "sys_show_hide"},
			{"id": "htxvbu2rnoreyqw", "name": "系统开关", "remark": "系统开关列表", "status": "0", "type": "sys_normal_disable"},
			{"id": "pb3yap6wr35rxk2", "name": "任务状态", "remark": "任务状态列表", "status": "0", "type": "sys_job_status"},
			{"id": "u3kcjk8gjeirzuv", "name": "任务分组", "remark": "任务分组列表", "status": "0", "type": "sys_job_group"},
			{"id": "hz5ybddleja7g7g", "name": "系统是否", "remark": "系统是否列表", "status": "0", "type": "sys_yes_no"},
			{"id": "efusb55rhbyn8c1", "name": "通知类型", "remark": "通知类型列表", "status": "0", "type": "sys_notice_type"},
			{"id": "w91tyto4wvuh22g", "name": "通知状态", "remark": "通知状态列表", "status": "0", "type": "sys_notice_status"},
			{"id": "y6rxzctulm0vdju", "name": "操作类型", "remark": "操作类型列表", "status": "0", "type": "sys_oper_type"},
			{"id": "15xzpv4hl3sitzg", "name": "系统状态", "remark": "登录状态列表", "status": "0", "type": "sys_common_status"},
		},
	},
	// 字典数据
	{
		Table:        TableSysDictData,
		UniqueFields: []string{"type", "value"},
		Data: []map[string]interface{}{
			// 用户性别
			{"order_num": 1, "label": "男", "value": "0", "code": 1, "type": "sys_user_sex", "dict_type_id": "upbepvx05k6j1pb", "css_class": "", "list_class": "", "is_default": "Y", "status": "0", "create_by": "system", "remark": "性别男"},
			{"order_num": 2, "label": "女", "value": "1", "code": 2, "type": "sys_user_sex", "dict_type_id": "upbepvx05k6j1pb", "css_class": "", "list_class": "", "is_default": "N", "status": "0", "create_by": "system", "remark": "性别女"},
			{"order_num": 3, "label": "未知", "value": "2", "code": 3, "type": "sys_user_sex", "dict_type_id": "upbepvx05k6j1pb", "css_class": "", "list_class": "", "is_default": "N", "status": "0", "create_by": "system", "remark": "性别未知"},
			// 菜单状态
			{"order_num": 1, "label": "显示", "value": "0", "code": 1, "type": "sys_show_hide", "dict_type_id": "9ew4nxq6202v35w", "css_class": "", "list_class": "primary", "is_default": "Y", "status": "0", "create_by": "system", "remark": "显示菜单"},
			{"order_num": 2, "label": "隐藏", "value": "1", "code": 2, "type": "sys_show_hide", "dict_type_id": "9ew4nxq6202v35w", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "隐藏菜单"},
			// 系统开关
			{"order_num": 1, "label": "正常", "value": "0", "code": 1, "type": "sys_normal_disable", "dict_type_id": "htxvbu2rnoreyqw", "css_class": "", "list_class": "primary", "is_default": "Y", "status": "0", "create_by": "system", "remark": "正常状态"},
			{"order_num": 2, "label": "停用", "value": "1", "code": 2, "type": "sys_normal_disable", "dict_type_id": "htxvbu2rnoreyqw", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "停用状态"},
			// 任务状态
			{"order_num": 1, "label": "正常", "value": "0", "code": 1, "type": "sys_job_status", "dict_type_id": "pb3yap6wr35rxk2", "css_class": "", "list_class": "primary", "is_default": "Y", "status": "0", "create_by": "system", "remark": "正常状态"},
			{"order_num": 2, "label": "暂停", "value": "1", "code": 2, "type": "sys_job_status", "dict_type_id": "pb3yap6wr35rxk2", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "停用状态"},
			// 任务分组
			{"order_num": 1, "label": "默认", "value": "DEFAULT", "code": 1, "type": "sys_job_group", "dict_type_id": "u3kcjk8gjeirzuv", "css_class": "", "list_class": "", "is_default": "Y", "status": "0", "create_by": "system", "remark": "默认分组"},
			{"order_num": 2, "label": "系统", "value": "SYSTEM", "code": 2, "type": "sys_job_group", "dict_type_id": "u3kcjk8gjeirzuv", "css_class": "", "list_class": "", "is_default": "N", "status": "0", "create_by": "system", "remark": "系统分组"},
			// 系统是否
			{"order_num": 1, "label": "是", "value": "Y", "code": 1, "type": "sys_yes_no", "dict_type_id": "hz5ybddleja7g7g", "css_class": "", "list_class": "primary", "is_default": "Y", "status": "0", "create_by": "system", "remark": "系统默认是"},
			{"order_num": 2, "label": "否", "value": "N", "code": 2, "type": "sys_yes_no", "dict_type_id": "hz5ybddleja7g7g", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "系统默认否"},
			// 通知类型
			{"order_num": 1, "label": "通知", "value": "1", "code": 1, "type": "sys_notice_type", "dict_type_id": "efusb55rhbyn8c1", "css_class": "", "list_class": "warning", "is_default": "Y", "status": "0", "create_by": "system", "remark": "通知"},
			{"order_num": 2, "label": "公告", "value": "2", "code": 2, "type": "sys_notice_type", "dict_type_id": "efusb55rhbyn8c1", "css_class": "", "list_class": "success", "is_default": "N", "status": "0", "create_by": "system", "remark": "公告"},
			// 通知状态
			{"order_num": 1, "label": "正常", "value": "0", "code": 1, "type": "sys_notice_status", "dict_type_id": "w91tyto4wvuh22g", "css_class": "", "list_class": "primary", "is_default": "Y", "status": "0", "create_by": "system", "remark": "正常状态"},
			{"order_num": 2, "label": "关闭", "value": "1", "code": 2, "type": "sys_notice_status", "dict_type_id": "w91tyto4wvuh22g", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "关闭状态"},
			// 操作类型
			{"order_num": 99, "label": "其他", "value": "0", "code": 99, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "info", "is_default": "N", "status": "0", "create_by": "system", "remark": "其他操作"},
			{"order_num": 1, "label": "新增", "value": "1", "code": 1, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "info", "is_default": "N", "status": "0", "create_by": "system", "remark": "新增操作"},
			{"order_num": 2, "label": "修改", "value": "2", "code": 2, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "info", "is_default": "N", "status": "0", "create_by": "system", "remark": "修改操作"},
			{"order_num": 3, "label": "删除", "value": "3", "code": 3, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "删除操作"},
			{"order_num": 4, "label": "授权", "value": "4", "code": 4, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "primary", "is_default": "N", "status": "0", "create_by": "system", "remark": "授权操作"},
			{"order_num": 5, "label": "导出", "value": "5", "code": 5, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "warning", "is_default": "N", "status": "0", "create_by": "system", "remark": "导出操作"},
			{"order_num": 6, "label": "导入", "value": "6", "code": 6, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "warning", "is_default": "N", "status": "0", "create_by": "system", "remark": "导入操作"},
			{"order_num": 7, "label": "强退", "value": "7", "code": 7, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "强退操作"},
			{"order_num": 8, "label": "生成代码", "value": "8", "code": 8, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "warning", "is_default": "N", "status": "0", "create_by": "system", "remark": "生成操作"},
			{"order_num": 9, "label": "清空数据", "value": "9", "code": 9, "type": "sys_oper_type", "dict_type_id": "y6rxzctulm0vdju", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "清空操作"},
			// 系统状态
			{"order_num": 1, "label": "成功", "value": "0", "code": 1, "type": "sys_common_status", "dict_type_id": "15xzpv4hl3sitzg", "css_class": "", "list_class": "primary", "is_default": "N", "status": "0", "create_by": "system", "remark": "正常状态"},
			{"order_num": 2, "label": "失败", "value": "1", "code": 2, "type": "sys_common_status", "dict_type_id": "15xzpv4hl3sitzg", "css_class": "", "list_class": "danger", "is_default": "N", "status": "0", "create_by": "system", "remark": "停用状态"},
		},
	},
	// 角色信息数据
	{
		Table:        TableSysRole,
		UniqueFields: []string{"key"},
		Data: []map[string]interface{}{
			{"name": "超级管理员", "key": "admin", "data_scope": "1", "order_num": 1, "status": "0", "delete_flag": "0", "created_by": "admin", "updated_by": "", "remark": "超级管理员"},
			{"name": "普通角色", "key": "common", "data_scope": "2", "order_num": 2, "status": "0", "delete_flag": "0", "created_by": "admin", "updated_by": "", "remark": "普通角色"},
		},
	},
	// 系统用户数据
	{
		Table:        TableSysUser,
		UniqueFields: []string{"username"},
		Data: []map[string]interface{}{
			{"username": "admin", "email": "admin@example.com", "password": "123!@#qwe", "type": "00", "status": "0", "delete_flag": "0", "created_by": "system", "updated_by": "system", "remark": "系统默认管理员账号", "verified": true},
		},
	},
}

func init() {
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
