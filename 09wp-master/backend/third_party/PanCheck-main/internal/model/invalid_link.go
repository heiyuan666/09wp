package model

import "time"

// InvalidLink 失效链接表
type InvalidLink struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Link          string    `gorm:"type:varchar(500);uniqueIndex;not null" json:"link"`      // 分享链接
	Platform      Platform  `gorm:"type:varchar(20);not null;index" json:"platform"`         // 网盘平台类型
	FailureReason string    `gorm:"type:text" json:"failure_reason"`                         // 失败原因
	CheckDuration *int64    `gorm:"type:bigint" json:"check_duration"`                       // 检测耗时（毫秒）
	IsRateLimited bool      `gorm:"type:boolean;default:false;index" json:"is_rate_limited"` // 是否被平台限制
	SubmissionID  *uint     `gorm:"index" json:"submission_id"`                              // 来源提交记录ID
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`                              // 创建时间
}

// TableName 指定表名
func (InvalidLink) TableName() string {
	return "invalid_links"
}
