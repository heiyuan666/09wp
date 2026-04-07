package model

import "time"

// QRLoginSession 扫码登录会话：Web/桌面创建二维码，手机 App 扫码后提交账号密码确认，双方拿到同一 JWT。
// ForAdmin=true 时表示管理后台 /login 扫码，确认时使用管理员账号并签发 IsAdmin JWT。
type QRLoginSession struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	PublicID  string    `gorm:"size:64;uniqueIndex;not null;column:public_id" json:"-"`
	ForAdmin  bool      `gorm:"default:false;index" json:"-"`
	Status    string    `gorm:"size:20;default:pending;index" json:"status"` // pending, confirmed, expired
	Token     string    `gorm:"size:768;default:''" json:"-"`
	UserID    uint64    `gorm:"default:0;index" json:"-"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (QRLoginSession) TableName() string {
	return "qr_login_sessions"
}
