package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct{}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Login 验证管理员密码
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "密码不能为空",
		})
		return
	}

	// 从环境变量读取管理员密码
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		// 如果环境变量未设置，使用默认密码
		adminPassword = "admin"
	}

	// 验证密码
	if req.Password == adminPassword {
		c.JSON(http.StatusOK, LoginResponse{
			Success: true,
			Message: "登录成功",
		})
	} else {
		c.JSON(http.StatusUnauthorized, LoginResponse{
			Success: false,
			Message: "密码错误",
		})
	}
}
