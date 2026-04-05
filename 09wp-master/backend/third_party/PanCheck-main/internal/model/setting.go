package model

import "time"

// Setting 设置表
type Setting struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"type:varchar(100);uniqueIndex;not null;column:key" json:"key"`        // 配置键
	Value       string    `gorm:"type:text;not null" json:"value"`                          // 配置值
	Description string    `gorm:"type:varchar(500)" json:"description"`                     // 配置描述
	Category    string    `gorm:"type:varchar(50);index" json:"category"`                   // 配置分类
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

