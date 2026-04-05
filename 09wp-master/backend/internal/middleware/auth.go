package middleware

import (
	"net/http"
	"strings"

	"dfan-netdisk-backend/pkg/jwtutil"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware 校验 JWT，adminOnly 为 true 时要求管理员身份
func AuthMiddleware(secret string, adminOnly bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "缺少或非法 token"})
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims, err := jwtutil.ParseToken(tokenStr, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "token 无效"})
			return
		}
		if adminOnly && !claims.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 403, "message": "无管理员权限"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("is_admin", claims.IsAdmin)
		c.Next()
	}
}

