package handler

import (
	"strings"

	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type quarkTransferReq struct {
	Link     string `json:"link" binding:"required"`
	Passcode string `json:"passcode"`
}

// QuarkTransferByLink 管理端手动触发夸克转存
func QuarkTransferByLink(c *gin.Context) {
	var req quarkTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	req.Link = strings.TrimSpace(req.Link)
	if req.Link == "" {
		response.Error(c, 400, "缺少 link")
		return
	}
	res, err := service.QuarkSaveByShareLink(req.Link, req.Passcode)
	if err != nil {
		response.Error(c, 500, "转存失败: "+err.Error())
		return
	}
	response.OK(c, res)
}

