package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件，验证管理员密码
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头读取密码
		password := c.GetHeader("X-Admin-Password")
		if password == "" {
			// 也支持从 Authorization 头读取（Bearer <password>）
			authHeader := c.GetHeader("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				password = authHeader[7:]
			}
		}

		// 如果密码为空，返回未授权错误
		if password == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权：缺少管理员密码",
			})
			c.Abort()
			return
		}

		// 从环境变量读取管理员密码
		adminPassword := os.Getenv("ADMIN_PASSWORD")
		if adminPassword == "" {
			// 如果环境变量未设置，使用默认密码
			adminPassword = "admin"
		}

		// 验证密码
		if password != adminPassword {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "未授权：密码错误",
			})
			c.Abort()
			return
		}

		// 验证通过，继续处理请求
		c.Next()
	}
}




