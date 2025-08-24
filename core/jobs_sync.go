package core

import (
	"fmt"
	"log/slog"

	"github.com/pocketbase/pocketbase/tools/cron"
	"github.com/pocketbase/pocketbase/tools/hook"
)

// syncCronJobsToCollection 将系统注册的cron jobs同步到_jobs集合
func syncCronJobsToCollection(app App) error {
	// 获取_jobs集合
	collection, err := app.FindCollectionByNameOrId("_jobs")
	if err != nil {
		return fmt.Errorf("failed to find _jobs collection: %w", err)
	}

	// 获取所有注册的cron jobs
	cronJobs := app.Cron().Jobs()
	if len(cronJobs) == 0 {
		app.Logger().Debug("No cron jobs found to sync")
		return nil
	}

	app.Logger().Debug("Syncing cron jobs to _jobs collection", slog.Int("count", len(cronJobs)))

	// 遍历每个cron job并同步到数据库
	for _, job := range cronJobs {
		if err := syncSingleCronJob(app, collection, job); err != nil {
			app.Logger().Warn("Failed to sync cron job",
				slog.String("jobId", job.Id()),
				slog.String("error", err.Error()))
			continue
		}
	}

	app.Logger().Debug("Cron jobs sync completed")
	return nil
}

// syncSingleCronJob 同步单个cron job到_jobs集合
func syncSingleCronJob(app App, collection *Collection, job *cron.Job) error {
	// 检查是否已存在相同name的记录
	existingRecord, _ := app.FindFirstRecordByFilter(
		collection.Id,
		"name = {:name}",
		map[string]any{"name": job.Id()},
	)

	// 准备记录数据
	recordData := map[string]any{
		"name":    job.Id(),
		"cron":    job.Expression(),
		"status":  "1",      // 默认启用状态
		"service": "system", // 标记为系统服务
		"script":  "",       // 系统cron job没有脚本内容
		"remark":  fmt.Sprintf("System cron job: %s", job.Id()),
	}

	if existingRecord != nil {
		// 更新现有记录
		for key, value := range recordData {
			existingRecord.Set(key, value)
		}
		existingRecord.Set("updated_by", "system")

		if err := app.Save(existingRecord); err != nil {
			return fmt.Errorf("failed to update existing job record: %w", err)
		}
	} else {
		// 创建新记录
		newRecord := NewRecord(collection)
		for key, value := range recordData {
			newRecord.Set(key, value)
		}
		newRecord.Set("created_by", "system")
		newRecord.Set("updated_by", "system")

		if err := app.Save(newRecord); err != nil {
			return fmt.Errorf("failed to create new job record: %w", err)
		}
	}

	return nil
}

// registerJobsSyncHooks 注册jobs同步相关的hooks
func (app *BaseApp) registerJobsSyncHooks() {
	// 在Serve时同步cron jobs，确保所有系统cron jobs都已注册
	app.OnServe().Bind(&hook.Handler[*ServeEvent]{
		Id: "__pbJobsSync__",
		Func: func(e *ServeEvent) error {
			err := e.Next()
			if err != nil {
				return err
			}

			// 同步cron jobs到_jobs集合
			if syncErr := syncCronJobsToCollection(e.App); syncErr != nil {
				e.App.Logger().Error("Failed to sync cron jobs to _jobs collection",
					slog.String("error", syncErr.Error()))
			}

			return nil
		},
		Priority: -999, // 设置较低优先级，确保在cron启动之后执行
	})
}
