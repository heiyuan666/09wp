package model

import "time"

// GameResourceFeedback 游戏资源（详情页下载资源）反馈：用于“标记失效”等场景
// 与 ResourceFeedback 区分：这里的资源来源是 game_resources（不一定存在 resource_id）
type GameResourceFeedback struct {
	ID           uint64 `gorm:"primaryKey" json:"id"`
	GameID       uint64 `gorm:"index;not null" json:"game_id"`
	GameResource uint64 `gorm:"index;not null" json:"game_resource_id"`
	DownloadURL  string `gorm:"size:1000;default:''" json:"download_url"`
	ExtractCode  string `gorm:"size:50;default:''" json:"extract_code"`
	Type         string `gorm:"size:32;default:'';index" json:"type"`   // link_invalid / other
	Content      string `gorm:"type:text;not null" json:"content"`
	Contact      string `gorm:"size:255;default:''" json:"contact"`
	Status       string `gorm:"size:20;default:'pending';index" json:"status"` // pending / processed
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

