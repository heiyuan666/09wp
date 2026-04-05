package model

import "time"

// UserResourceSubmission 用户提交的资源（待审核）
type UserResourceSubmission struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	UserID uint64  `gorm:"index;not null" json:"user_id"`
	GameID *uint64 `gorm:"index" json:"game_id,omitempty"`

	Title       string `gorm:"size:200;not null" json:"title"`
	Link        string `gorm:"size:500;not null" json:"link"`
	CategoryID  uint64 `gorm:"index;default:0" json:"category_id"`
	Description string `gorm:"type:text" json:"description"`
	ExtractCode string `gorm:"size:50" json:"extract_code"`
	Tags        string `gorm:"size:255" json:"tags"`

	// pending / approved / rejected
	Status    string `gorm:"size:20;default:'pending';index" json:"status"`
	ReviewMsg string `gorm:"size:255;default:''" json:"review_msg"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserResourceSubmission) TableName() string {
	return "user_resource_submissions"
}
