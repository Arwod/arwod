package migrations

import (
	"fmt"

	"github.com/pocketbase/pocketbase/core"
)

func init() {
	// 定义JavaScript脚本管理系统初始化数据配置
	var initDataRequests = []DataImportRequest{
		{
			Table:        Table_EventsDict,
			UniqueFields: []string{"name"},
			Data: []map[string]interface{}{
				{
					"name":        "App Hooks",
					"description": "",
					"childData": []map[string]interface{}{
						{
							"name":        "onBootstrap",
							"description": "onBootstrap hook is triggered when initializing the main application resources (db, app settings, etc).",
						},
						{
							"name":        "onSettingsReload",
							"description": "onSettingsReload hook is triggered every time when the $app.settings() is being replaced with a new state.",
						},
						{
							"name":        "onBackupCreate",
							"description": "onBackupCreate is triggered on each $app.createBackup call.",
						},
						{
							"name":        "onBackupRestore",
							"description": "onBackupRestore is triggered before app backup restore (aka. on $app.restoreBackup call).",
						},
						{
							"name":        "onTerminate",
							"description": "onTerminate hook is triggered when the app is in the process of being terminated (ex. on SIGTERM signal).",
						},
					},
				},
				{
					"name":        "Mailer Hooks",
					"description": "",
					"childData": []map[string]interface{}{
						{
							"name":        "onMailerSend",
							"description": "onMailerSend hook is triggered every time when a new email is being send using the $app.newMailClient() instance.",
						},
						//onMailerRecordAuthAlertSend
						{
							"name":        "onMailerRecordAuthAlertSend",
							"description": "onMailerRecordAuthAlertSend hook is triggered when sending a new device login auth alert email, allowing you to intercept and customize the email message that is being sent.",
						},
						//onMailerRecordPasswordResetSend
						{
							"name":        "onMailerRecordPasswordResetSend",
							"description": "onMailerRecordPasswordResetSend hook is triggered when sending a password reset email to an auth record, allowing you to intercept and customize the email message that is being sent.",
						},
						//onMailerRecordVerificationSend
						{
							"name":        "onMailerRecordVerificationSend",
							"description": "onMailerRecordVerificationSend hook is triggered when sending a verification email to an auth record, allowing you to intercept and customize the email message that is being sent.",
						},
						//onMailerRecordEmailChangeSend
						{
							"name":        "onMailerRecordEmailChangeSend",
							"description": "onMailerRecordEmailChangeSend hook is triggered when sending a confirmation new address email to an auth record, allowing you to intercept and customize the email message that is being sent.",
						},
						//onMailerRecordOTPSend
						{
							"name":        "onMailerRecordOTPSend",
							"description": "onMailerRecordOTPSend hook is triggered when sending an OTP email to an auth record, allowing you to intercept and customize the email message that is being sent.",
						},
					},
				},
			},
		},
		{
			Table:        "_superusers",
			UniqueFields: []string{"email"},
			Data: []map[string]interface{}{
				{
					"email":    "admin@example.com",
					"password": "1234!@#qwe",
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
