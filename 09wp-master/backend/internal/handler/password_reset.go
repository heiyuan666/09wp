package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token           string `json:"token" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,min=6"`
}

func UserPasswordForgot(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		response.Error(c, 400, "邮箱不能为空")
		return
	}

	token := newResetToken()
	tokenHash := sha256Hex(token)

	var user model.User
	userErr := database.DB().Where("email = ?", email).First(&user).Error

	// 无用户时也返回成功，避免邮箱枚举（前端将拿不到 token，无法进入重置步骤）。
	if userErr != nil {
		response.OK(c, gin.H{
			"reset_token": "",
			"expires_at":  "",
		})
		return
	}

	exp := time.Now().Add(30 * time.Minute)
	rec := model.UserPasswordReset{
		UserID:    &user.ID,
		Email:     email,
		TokenHash: tokenHash,
		ExpiresAt: exp,
	}
	if err := database.DB().Create(&rec).Error; err != nil {
		response.Error(c, 500, "生成重置记录失败")
		return
	}

	// 真正功能：发邮件。若 SMTP 未配置，则回退为“演示模式”直接返回 token。
	resetURL := buildResetURL(c, token)
	html := fmt.Sprintf(
		`<div style="font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;line-height:1.7">
<h2>重置密码</h2>
<p>我们收到了你的重置密码请求。点击下方按钮继续：</p>
<p style="margin:18px 0">
  <a href="%s" style="display:inline-block;padding:10px 16px;border-radius:10px;background:#2563eb;color:#fff;text-decoration:none">重置密码</a>
</p>
<p style="color:#6b7280;font-size:12px">链接 30 分钟内有效。如非本人操作请忽略此邮件。</p>
</div>`,
		resetURL,
	)

	if err := service.SendHTMLEmail(email, "重置密码", html); err == nil {
		response.OK(c, gin.H{
			"reset_token": "",
			"expires_at":  exp.Format(time.RFC3339),
		})
		return
	}

	response.OK(c, gin.H{
		"reset_token": token,
		"expires_at":  exp.Format(time.RFC3339),
	})
}

func UserPasswordReset(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if strings.TrimSpace(req.NewPassword) != req.NewPassword || strings.TrimSpace(req.ConfirmPassword) != req.ConfirmPassword {
		// 明确拒绝首尾空格，避免“看起来相同但实际不同”
		response.Error(c, 400, "密码格式不正确")
		return
	}

	token := strings.TrimSpace(req.Token)
	tokenHash := sha256Hex(token)
	now := time.Now()

	var rec model.UserPasswordReset
	if err := database.DB().
		Where("token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, now).
		First(&rec).Error; err != nil {
		response.Error(c, 400, "重置链接已失效或不正确")
		return
	}

	if rec.UserID == nil {
		response.Error(c, 400, "重置记录无用户信息")
		return
	}

	var user model.User
	if err := database.DB().First(&user, *rec.UserID).Error; err != nil {
		response.Error(c, 404, "用户不存在")
		return
	}
	if user.Status != 1 {
		response.Error(c, 403, "账号已禁用")
		return
	}

	if req.NewPassword != req.ConfirmPassword {
		response.Error(c, 400, "两次输入密码不一致")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "密码加密失败")
		return
	}

	if err := database.DB().Model(&user).Update("password_hash", string(hash)).Error; err != nil {
		response.Error(c, 500, "更新密码失败")
		return
	}

	used := now
	_ = database.DB().Model(&model.UserPasswordReset{}).Where("id = ?", rec.ID).Update("used_at", &used).Error

	response.OK(c, nil)
}

func newResetToken() string {
	// 32 bytes => 43-44 chars (base64url)
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func buildResetURL(c *gin.Context, token string) string {
	origin := strings.TrimSpace(c.GetHeader("Origin"))
	if origin == "" {
		// 兼容某些环境没有 Origin（例如 curl）
		host := strings.TrimSpace(c.Request.Host)
		if host != "" {
			origin = "http://" + host
		}
	}
	if origin == "" {
		origin = "http://localhost:3007"
	}
	return fmt.Sprintf("%s/login?mode=forgot&token=%s", strings.TrimRight(origin, "/"), token)
}

