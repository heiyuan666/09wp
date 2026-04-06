package handler

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/jwtutil"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func randomPublicID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type qrLoginCreateRequest struct {
	ForAdmin bool `json:"for_admin"`
}

// QRLoginCreate 创建扫码登录会话（公开，供 Web/桌面生成二维码）
func QRLoginCreate(c *gin.Context) {
	var req qrLoginCreateRequest
	_ = c.ShouldBindJSON(&req)

	sid, err := randomPublicID()
	if err != nil {
		response.Error(c, 500, "生成会话失败")
		return
	}
	now := time.Now()
	sess := model.QRLoginSession{
		PublicID:  sid,
		ForAdmin:  req.ForAdmin,
		Status:    "pending",
		ExpiresAt: now.Add(10 * time.Minute),
		CreatedAt: now,
	}
	if err := database.DB().Create(&sess).Error; err != nil {
		response.Error(c, 500, "创建会话失败")
		return
	}

	var payload string
	var altType string
	if req.ForAdmin {
		payload = "dfannetdisk://qr-admin-login?sid=" + sid
		altType = "dfan_qr_admin_login"
	} else {
		payload = "dfannetdisk://qr-login?sid=" + sid
		altType = "dfan_qr_login"
	}

	response.OK(c, gin.H{
		"sid":        sid,
		"expires_at": sess.ExpiresAt.Format(time.RFC3339),
		"qr_payload": payload,
		"qr_payload_alt": gin.H{
			"type": altType,
			"sid":  sid,
		},
		"for_admin": req.ForAdmin,
	})
}

// QRLoginStatus Web 端轮询会话状态；确认后返回与普通登录一致的 token
func QRLoginStatus(c *gin.Context) {
	sid := strings.TrimSpace(c.Param("sid"))
	if sid == "" {
		response.Error(c, 400, "缺少 sid")
		return
	}
	var sess model.QRLoginSession
	if err := database.DB().Where("public_id = ?", sid).First(&sess).Error; err != nil {
		response.Error(c, 404, "会话不存在")
		return
	}
	if time.Now().After(sess.ExpiresAt) {
		if sess.Status == "pending" {
			_ = database.DB().Model(&sess).Update("status", "expired").Error
		}
		response.OK(c, gin.H{"status": "expired"})
		return
	}
	if sess.Status != "confirmed" || sess.Token == "" {
		response.OK(c, gin.H{"status": "pending"})
		return
	}

	if sess.ForAdmin {
		var admin model.Admin
		if err := database.DB().First(&admin, sess.UserID).Error; err != nil {
			response.OK(c, gin.H{"status": "pending"})
			return
		}
		response.OK(c, gin.H{
			"status": "confirmed",
			"token":  sess.Token,
			"user": gin.H{
				"id":         strconv.FormatUint(admin.ID, 10),
				"username":   admin.Username,
				"password":   "",
				"status":     "active",
				"createTime": admin.CreatedAt.Format(time.RFC3339),
				"updateTime": admin.UpdatedAt.Format(time.RFC3339),
			},
			"for_admin": true,
		})
		return
	}

	var user model.User
	if err := database.DB().First(&user, sess.UserID).Error; err != nil {
		response.OK(c, gin.H{"status": "pending"})
		return
	}
	response.OK(c, gin.H{
		"status": "confirmed",
		"token":  sess.Token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"for_admin": false,
	})
}

type qrLoginConfirmRequest struct {
	Sid      string `json:"sid" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// QRLoginConfirm 手机端扫码后提交账号密码，写入会话并返回 JWT
func QRLoginConfirm(c *gin.Context) {
	var req qrLoginConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	sid := strings.TrimSpace(req.Sid)
	identifier := strings.TrimSpace(req.Username)
	if sid == "" || identifier == "" {
		response.Error(c, 400, "参数错误")
		return
	}

	var sess model.QRLoginSession
	if err := database.DB().Where("public_id = ?", sid).First(&sess).Error; err != nil {
		response.Error(c, 404, "二维码已失效")
		return
	}
	if time.Now().After(sess.ExpiresAt) {
		_ = database.DB().Model(&sess).Update("status", "expired").Error
		response.Error(c, 400, "二维码已过期")
		return
	}
	if sess.Status != "pending" {
		response.Error(c, 400, "该二维码已使用或已失效")
		return
	}

	secret := c.MustGet("jwt_secret").(string)

	if sess.ForAdmin {
		confirmAdminQRLogin(c, &sess, identifier, req.Password, secret)
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

	token, err := jwtutil.GenerateToken(user.ID, false, secret, 24*time.Hour)
	if err != nil {
		response.Error(c, 500, "生成 token 失败")
		return
	}

	tx := database.DB().Begin()
	res := tx.Model(&model.QRLoginSession{}).
		Where("id = ? AND status = ?", sess.ID, "pending").
		Updates(map[string]interface{}{
			"status":     "confirmed",
			"token":      token,
			"user_id":    user.ID,
			"expires_at": time.Now().Add(30 * time.Minute),
		})
	if res.Error != nil {
		tx.Rollback()
		response.Error(c, 500, "确认失败")
		return
	}
	if res.RowsAffected != 1 {
		tx.Rollback()
		response.Error(c, 400, "该二维码已使用或已失效")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Error(c, 500, "确认失败")
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

func confirmAdminQRLogin(c *gin.Context, sess *model.QRLoginSession, username, password, secret string) {
	var admin model.Admin
	if err := database.DB().Where("username = ?", username).First(&admin).Error; err != nil {
		response.Error(c, 401, "账号或密码错误")
		return
	}
	if admin.Status != 1 {
		response.Error(c, 403, "账号已禁用")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		response.Error(c, 401, "账号或密码错误")
		return
	}

	token, err := jwtutil.GenerateToken(admin.ID, true, secret, 24*time.Hour)
	if err != nil {
		response.Error(c, 500, "生成 token 失败")
		return
	}

	tx := database.DB().Begin()
	res := tx.Model(&model.QRLoginSession{}).
		Where("id = ? AND status = ?", sess.ID, "pending").
		Updates(map[string]interface{}{
			"status":     "confirmed",
			"token":      token,
			"user_id":    admin.ID,
			"expires_at": time.Now().Add(30 * time.Minute),
		})
	if res.Error != nil {
		tx.Rollback()
		response.Error(c, 500, "确认失败")
		return
	}
	if res.RowsAffected != 1 {
		tx.Rollback()
		response.Error(c, 400, "该二维码已使用或已失效")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Error(c, 500, "确认失败")
		return
	}

	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":         strconv.FormatUint(admin.ID, 10),
			"username":   admin.Username,
			"password":     "",
			"status":       "active",
			"createTime":   admin.CreatedAt.Format(time.RFC3339),
			"updateTime":   admin.UpdatedAt.Format(time.RFC3339),
		},
	})
}
