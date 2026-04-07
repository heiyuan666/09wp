package handler

import (
	"dfan-netdisk-backend/internal/version"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

// PublicVersion 前台可读的程序版本（无需登录）
// @Summary      后端版本
// @Description  返回当前后端程序版本号（无需登录）
// @Tags         public
// @Produce      json
// @Success      200 {object} map[string]interface{} "与 pkg/response.OK 一致：含 code、message、data"
// @Router       /public/version [get]
func PublicVersion(c *gin.Context) {
	response.OK(c, gin.H{
		"version": version.Version,
	})
}
