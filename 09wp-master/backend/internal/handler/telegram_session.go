package handler

import (
	"context"
	"errors"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type tgSendCodeReq struct {
	Phone string `json:"phone"`
}

type tgSignInReq struct {
	Code string `json:"code"`
}

type tgPasswordReq struct {
	Password string `json:"password"`
}

func TelegramSessionSendCode(c *gin.Context) {
	var req tgSendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	cfg, err := service.GetMTProtoConfigFromDB()
	if err != nil {
		response.Error(c, 500, "系统配置不存在")
		return
	}
	if err := service.MTProtoSendCode(cfg.APIID, cfg.APIHash, cfg.ProxyURL, req.Phone); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			response.Error(c, 500, "发送验证码超时：请检查服务器到 Telegram 的网络连通性（必要时配置代理）")
			return
		}
		response.Error(c, 500, "发送验证码失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "验证码已发送"})
}

func TelegramSessionSignIn(c *gin.Context) {
	var req tgSignInReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	needPassword, err := service.MTProtoSignIn(req.Code)
	if err != nil {
		response.Error(c, 500, "登录失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{
		"need_password": needPassword,
		"message":       map[bool]string{true: "需要输入2FA密码", false: "登录成功，session已保存"}[needPassword],
	})
}

func TelegramSessionCheckPassword(c *gin.Context) {
	var req tgPasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := service.MTProtoCheckPassword(req.Password); err != nil {
		response.Error(c, 500, "2FA登录失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{"message": "2FA登录成功，session已保存"})
}

func TelegramSessionStatus(c *gin.Context) {
	data, err := service.MTProtoSessionStatus()
	if err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, data)
}

