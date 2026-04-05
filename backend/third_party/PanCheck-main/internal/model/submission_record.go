package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// SubmissionRecord 提交记录表
type SubmissionRecord struct {
	ID                uint          `gorm:"primaryKey" json:"id"`
	OriginalLinks     StringArray   `gorm:"type:json;not null" json:"original_links"`                  // 用户原始提交内容
	PendingLinks      StringArray   `gorm:"type:json" json:"pending_links"`                            // 待检测的链接
	ValidLinks        StringArray   `gorm:"type:json" json:"valid_links"`                              // 检测完成后的有效链接
	SelectedPlatforms PlatformArray `gorm:"type:json" json:"selected_platforms"`                       // 用户提交时选择的网盘平台类型
	Status            string        `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending/checked
	TotalDuration     *int64        `gorm:"type:bigint" json:"total_duration"`                         // 总耗时（毫秒）
	TotalLinks        int           `gorm:"default:0" json:"total_links"`                              // 总提交的链接数量
	ClientIP          string        `gorm:"type:varchar(45)" json:"client_ip"`                         // 客户端IP
	Browser           string        `gorm:"type:varchar(50)" json:"browser"`                           // 浏览器
	OS                string        `gorm:"type:varchar(50)" json:"os"`                                // 操作系统
	Device            string        `gorm:"type:varchar(20)" json:"device"`                            // 设备类型（desktop/mobile）
	Language          string        `gorm:"type:varchar(10)" json:"language"`                          // 语言
	Country           string        `gorm:"type:varchar(10)" json:"country"`                           // 国家
	Region            string        `gorm:"type:varchar(50)" json:"region"`                            // 地区（由IP反查得到）
	City              string        `gorm:"type:varchar(50)" json:"city"`                              // 城市（由IP反查得到）
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
	CheckedAt         *time.Time    `json:"checked_at"` // 检测完成时间
}

// TableName 指定表名
func (SubmissionRecord) TableName() string {
	return "submission_records"
}

// StringArray 字符串数组类型，用于JSON存储
type StringArray []string

// Value 实现 driver.Valuer 接口
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, a)
}

// PlatformArray 平台数组类型，用于JSON存储
type PlatformArray []Platform

// Value 实现 driver.Valuer 接口
func (a PlatformArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan 实现 sql.Scanner 接口
func (a *PlatformArray) Scan(value interface{}) error {
	if value == nil {
		*a = []Platform{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return nil
	}

	return json.Unmarshal(bytes, a)
}
