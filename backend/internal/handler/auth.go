package handler

import (
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/jwtutil"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLogin 管理员登录，返回 JWT
func AdminLogin(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, 400, "参数错误")
			return
		}

		var admin model.Admin
		if err := database.DB().Where("username = ?", req.Username).First(&admin).Error; err != nil {
			response.Error(c, 401, "账号或密码错误")
			return
		}
		if admin.Status != 1 {
			response.Error(c, 403, "账号已禁用")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
			response.Error(c, 401, "账号或密码错误")
			return
		}

		token, err := jwtutil.GenerateToken(admin.ID, true, secret, 24*time.Hour)
		if err != nil {
			response.Error(c, 500, "生成 token 失败")
			return
		}

		response.OK(c, gin.H{
			"token": token,
			// DFAN Admin 前端期望字段名为 user
			"user": gin.H{
				"id":       strconv.FormatUint(admin.ID, 10),
				"username": admin.Username,
				"password": "",
				"status":   "active",
				"createTime": time.Now().Format(time.RFC3339),
				"updateTime": time.Now().Format(time.RFC3339),
			},
		})
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	EmailCode string `json:"email_code" binding:"required"`
	CaptchaID string `json:"captcha_id" binding:"required"`
	CaptchaCode string `json:"captcha_code" binding:"required"`
}

// UserRegister 普通用户注册
func UserRegister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if username == "" || email == "" {
		response.Error(c, 400, "用户名或邮箱不能为空")
		return
	}

	// 校验图形验证码
	if !verifyCaptcha(req.CaptchaID, req.CaptchaCode) {
		response.Error(c, 400, "验证码错误或已过期")
		return
	}

	// 校验邮箱验证码（10分钟有效、一次性使用）
	var ev model.EmailVerificationCode
	if err := database.DB().
		Where("email = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?", email, "register", time.Now()).
		Order("id DESC").
		First(&ev).Error; err != nil {
		response.Error(c, 400, "邮箱验证码无效或已过期")
		return
	}
	if sha256Hex(strings.TrimSpace(req.EmailCode)) != ev.CodeHash {
		response.Error(c, 400, "邮箱验证码错误")
		return
	}
	nowUsed := time.Now()
	_ = database.DB().Model(&model.EmailVerificationCode{}).Where("id = ?", ev.ID).Update("used_at", &nowUsed).Error

	// 后端也要强制校验：避免绕过前端直接注册
	var sysCfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&sysCfg).Error; err != nil {
		response.Error(c, 500, "系统配置读取失败")
		return
	}
	if !sysCfg.AllowRegister {
		response.Error(c, 403, "当前系统已关闭注册")
		return
	}

	// 检查是否存在
	var cnt int64
	database.DB().Model(&model.User{}).Where("username = ? OR email = ?", username, email).Count(&cnt)
	if cnt > 0 {
		response.Error(c, 409, "用户名或邮箱已存在")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		response.Error(c, 500, "密码加密失败")
		return
	}

	user := model.User{
		Username:     username,
		Email:        email,
		Name:         username, // 给前台用户中心留出“姓名”兜底
		PasswordHash: string(hash),
		Status:       1,
	}
	if err := database.DB().Create(&user).Error; err != nil {
		response.Error(c, 500, "注册失败")
		return
	}

	// 邮件通知（可选：若 SMTP 未配置则忽略）
	_ = func() error {
		html := `<div style="font-family:system-ui,-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;line-height:1.7">
<h2>注册成功</h2>
<p>欢迎加入！你的账号已经创建成功。</p>
<p style="color:#6b7280;font-size:12px">如非本人操作请忽略此邮件。</p>
</div>`
		return service.SendHTMLEmail(user.Email, "注册成功", html)
	}()

	response.OK(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
	})
}

// UserLogin 普通用户登录
func UserLogin(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	identifier := strings.TrimSpace(req.Username)
	if identifier == "" {
		response.Error(c, 400, "账号不能为空")
		return
	}

	var user model.User
	if err := database.DB().Where("username = ? OR email = ?", identifier, identifier).First(&user).Error; err != nil {
		response.Error(c, 401, "账号或密码错误")
		return
	}
	if user.Status != 1 {
		response.Error(c, 403, "账号已禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		response.Error(c, 401, "账号或密码错误")
		return
	}

	token, err := jwtutil.GenerateToken(user.ID, false, c.MustGet("jwt_secret").(string), 24*time.Hour)
	if err != nil {
		response.Error(c, 500, "生成 token 失败")
		return
	}

	now := time.Now()
	_ = database.DB().Model(&user).Update("last_login_at", &now).Error

	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

