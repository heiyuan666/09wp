package model

import "time"

// EmailVerificationCode 邮箱验证码（注册/找回密码等）
type EmailVerificationCode struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	Email string `gorm:"size:120;index" json:"email"`

	// Purpose: register / reset_password 等
	Purpose string `gorm:"size:32;index" json:"purpose"`

	CodeHash  string    `gorm:"size:64;index" json:"code_hash"`
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`

	// 限流：同一邮箱短时间发送次数
	SentAt time.Time `gorm:"index" json:"sent_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (EmailVerificationCode) TableName() string {
	return "email_verification_codes"
}

