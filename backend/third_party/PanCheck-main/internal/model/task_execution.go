package model

import "time"

// TaskExecution 任务执行记录表
type TaskExecution struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	TaskID            uint       `gorm:"not null;index" json:"task_id"`                           // 关联任务ID
	Status            string      `gorm:"type:varchar(20);not null" json:"status"`                 // running/success/failed
	LinksCount        int         `gorm:"default:0" json:"links_count"`                           // 获取到的链接数量
	CheckedCount      int         `gorm:"default:0" json:"checked_count"`                         // 已检测链接数量
	ValidCount        int         `gorm:"default:0" json:"valid_count"`                           // 有效链接数量
	InvalidCount      int         `gorm:"default:0" json:"invalid_count"`                          // 失效链接数量
	ErrorMessage      string      `gorm:"type:text" json:"error_message"`                        // 错误信息
	ExecutionDuration *int64      `gorm:"type:bigint" json:"execution_duration"`                  // 执行耗时（毫秒）
	StartedAt         time.Time   `gorm:"not null" json:"started_at"`                              // 开始时间
	FinishedAt        *time.Time  `json:"finished_at"`                                            // 结束时间
	CreatedAt         time.Time   `json:"created_at"`
}

// TableName 指定表名
func (TaskExecution) TableName() string {
	return "task_executions"
}

