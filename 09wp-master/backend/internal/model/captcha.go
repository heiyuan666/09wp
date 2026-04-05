package model

import "time"

// CaptchaChallenge 数字验证码挑战（用于注册/发邮箱验证码等）
type CaptchaChallenge struct {
	ID uint64 `gorm:"primaryKey" json:"id"`

	// CaptchaID 供前端持有的标识（随机字符串）
	CaptchaID string `gorm:"size:64;uniqueIndex" json:"captcha_id"`

	// AnswerHash 为正确答案的 hash（不存明文）
	AnswerHash string `gorm:"size:64;index" json:"answer_hash"`

	ExpiresAt time.Time  `gorm:"index" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CaptchaChallenge) TableName() string {
	return "captcha_challenges"
}

