package handler

import (
	"dfan-netdisk-backend/internal/version"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// PublicVersion 前台可读的程序版本（无需登录）
func PublicVersion(c *gin.Context) {
	response.OK(c, gin.H{
		"version": version.Version,
	})
}
