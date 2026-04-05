package model

import (
	"time"
)

// ScheduledTask 任务计划表
type ScheduledTask struct {
	ID              uint        `gorm:"primaryKey" json:"id"`
	Name            string      `gorm:"type:varchar(100);not null" json:"name"`                    // 任务名称
	Description     string      `gorm:"type:text" json:"description"`                              // 任务描述
	Tags            StringArray `gorm:"type:json" json:"tags"`                                     // 标签（JSON数组）
	CurlCommand     string      `gorm:"type:text;not null" json:"curl_command"`                    // curl命令
	TransformScript string      `gorm:"type:text" json:"transform_script"`                         // 数据转换脚本（JavaScript）
	CronExpression  string      `gorm:"type:varchar(100);not null" json:"cron_expression"`         // Cron表达式
	AutoDestroyAt   *time.Time  `json:"auto_destroy_at"`                                           // 自我销毁时间
	Status          string      `gorm:"type:varchar(20);not null;default:'stopped'" json:"status"` // active/stopped/expired
	LastRunAt       *time.Time  `json:"last_run_at"`                                               // 最后执行时间
	NextRunAt       *time.Time  `json:"next_run_at"`                                               // 下次执行时间
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// TableName 指定表名
func (ScheduledTask) TableName() string {
	return "scheduled_tasks"
}
