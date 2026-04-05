package model

import "time"

// UserPasswordReset 用户密码重置记录（用于“忘记密码/找回密码”流程）
type UserPasswordReset struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	// UserID 若用户存在则记录，便于重置时直接更新密码
	UserID *uint64 `gorm:"index" json:"user_id,omitempty"`

	Email string `gorm:"size:120;index" json:"email"`

	// TokenHash 为重置 token 的哈希（明文 token 只在生成/下发阶段出现）
	TokenHash string `gorm:"size:64;index" json:"token_hash"`

	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (UserPasswordReset) TableName() string {
	return "user_password_resets"
}

