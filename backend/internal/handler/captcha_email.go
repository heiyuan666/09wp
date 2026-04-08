package handler

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// GetCaptcha 返回数字验证码（SVG）
func GetCaptcha(c *gin.Context) {
	code := randomDigits(4)
	captchaID := randomID(24)

	exp := time.Now().Add(5 * time.Minute)
	rec := model.CaptchaChallenge{
		CaptchaID:  captchaID,
		AnswerHash: sha256Hex(code),
		ExpiresAt:  exp,
	}
	_ = database.DB().Create(&rec).Error

	svg := buildCaptchaSVG(code)

	response.OK(c, gin.H{
		"captcha_id": captchaID,
		"svg":        svg,
		"expires_at": exp.Format(time.RFC3339),
	})
}

type sendRegisterCodeReq struct {
	Email       string `json:"email" binding:"required,email"`
	CaptchaID   string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// SendRegisterEmailCode 发送“注册邮箱验证码”（需先通过图形验证码）
func SendRegisterEmailCode(c *gin.Context) {
	var req sendRegisterCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		response.Error(c, 400, "邮箱不能为空")
		return
	}

	if !verifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		response.Error(c, 400, "验证码错误或已过期")
		return
	}

	// 发送频率限制：同邮箱 60 秒内只允许一次
	var recent model.EmailVerificationCode
	if err := database.DB().
		Where("email = ? AND purpose = ? AND sent_at > ?", email, "register", time.Now().Add(-60*time.Second)).
		Order("id DESC").
		First(&recent).Error; err == nil {
		response.Error(c, 429, "发送过于频繁，请稍后再试")
		return
	}

	code := randomDigits(6)
	exp := time.Now().Add(10 * time.Minute)

	rec := model.EmailVerificationCode{
		Email:     email,
		Purpose:   "register",
		CodeHash:  sha256Hex(code),
		ExpiresAt: exp,
		SentAt:    time.Now(),
	}
	if err := database.DB().Create(&rec).Error; err != nil {
		response.Error(c, 500, "生成验证码失败")
		return
	}

	html := fmt.Sprintf(
		`<div style="font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;line-height:1.7">
<h2>邮箱验证码</h2>
<p>你的注册验证码是：</p>
<div style="font-size:28px;font-weight:800;letter-spacing:4px;margin:12px 0">%s</div>
<p style="color:#6b7280;font-size:12px">验证码 10 分钟内有效。如非本人操作请忽略此邮件。</p>
</div>`,
		code,
	)

	if err := service.SendHTMLEmail(email, "注册验证码", html); err != nil {
		response.Error(c, 500, "邮件发送失败（请检查 SMTP 配置）")
		return
	}

	response.OK(c, gin.H{
		"expires_at": exp.Format(time.RFC3339),
	})
}

func verifyCaptcha(captchaID string, captchaCode string) bool {
	captchaID = strings.TrimSpace(captchaID)
	captchaCode = strings.TrimSpace(captchaCode)
	if captchaID == "" || captchaCode == "" {
		return false
	}

	var rec model.CaptchaChallenge
	if err := database.DB().
		Where("captcha_id = ? AND used_at IS NULL AND expires_at > ?", captchaID, time.Now()).
		First(&rec).Error; err != nil {
		return false
	}

	if sha256Hex(captchaCode) != rec.AnswerHash {
		return false
	}

	now := time.Now()
	_ = database.DB().Model(&model.CaptchaChallenge{}).Where("id = ?", rec.ID).Update("used_at", &now).Error
	return true
}

func buildCaptchaSVG(code string) string {
	// 极简 SVG：数字 + 轻微干扰线（不追求强对抗，主要防机器批量）
	return fmt.Sprintf(
		`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40">
  <rect x="0" y="0" width="120" height="40" rx="10" fill="#F3F4F6"/>
  <path d="M8 28 C 28 10, 48 34, 68 16 S 98 34, 112 12" stroke="#CBD5E1" stroke-width="2" fill="none" opacity="0.9"/>
  <path d="M10 12 C 34 30, 52 8, 78 26 S 102 10, 116 28" stroke="#E2E8F0" stroke-width="2" fill="none" opacity="0.9"/>
  <text x="60" y="27" text-anchor="middle" font-size="22" font-weight="800"
        font-family="ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto" fill="#111827"
        textLength="92" lengthAdjust="spacingAndGlyphs">%s</text>
</svg>`,
		code,
	)
}

func randomDigits(n int) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		v, _ := rand.Int(rand.Reader, big.NewInt(10))
		out = append(out, byte('0'+v.Int64()))
	}
	return string(out)
}

func randomID(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		v, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		out = append(out, alphabet[v.Int64()])
	}
	return string(out)
}

