package handler

import (
	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"strings"
)

type panCheckReq struct {
	Links             []string `json:"links"`
	SelectedPlatforms []string `json:"selectedPlatforms"`
}

func PanCheckLinks(c *gin.Context) {
	var req panCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	baseURL := ""
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		baseURL = cfg.PanCheckBaseURL
	}
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		response.Error(c, 400, "请先在系统配置填写“失效检测地址”")
		return
	}
	data, err := service.PanCheckLinks(service.PanCheckRequest{
		Links:             req.Links,
		SelectedPlatforms: req.SelectedPlatforms,
	}, baseURL)
	if err != nil {
		response.Error(c, 500, "检测失败: "+err.Error())
		return
	}
	response.OK(c, data)
}

