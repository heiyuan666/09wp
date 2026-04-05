package handler

import (
	"PanCheck/internal/model"
	"PanCheck/internal/service"
	"PanCheck/pkg/utils"
	"PanCheck/pkg/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LinkHandler 链接处理器
type LinkHandler struct {
	linkService    *service.LinkService
	checkerService *service.CheckerService
}

// NewLinkHandler 创建链接处理器
func NewLinkHandler(linkService *service.LinkService, checkerService *service.CheckerService) *LinkHandler {
	return &LinkHandler{
		linkService:    linkService,
		checkerService: checkerService,
	}
}

// CheckLinks 检测链接
func (h *LinkHandler) CheckLinks(c *gin.Context) {
	var req service.CheckLinksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取客户端IP
	clientIP := c.ClientIP()

	// 从请求头获取设备信息
	userAgent := c.GetHeader("User-Agent")
	acceptLanguage := c.GetHeader("Accept-Language")

	deviceInfo := utils.ParseDeviceInfo(userAgent, acceptLanguage)

	// 调用服务
	resp, err := h.linkService.CheckLinks(&req, clientIP, deviceInfo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果选择了平台，对选中平台的链接进行即时检测
	// 如果选择了全部平台，则等同于即时检测所有链接
	if len(req.SelectedPlatforms) > 0 && len(resp.PendingLinks) > 0 {
		// 保存原始待检测链接列表
		originalPendingLinks := make([]string, len(resp.PendingLinks))
		copy(originalPendingLinks, resp.PendingLinks)

		// 构建选中平台的映射
		selectedPlatformMap := make(map[model.Platform]bool)
		for _, platform := range req.SelectedPlatforms {
			selectedPlatformMap[platform] = true
		}

		// 检查是否选择了全部平台（9个平台）
		allPlatforms := model.AllPlatforms()
		selectedAllPlatforms := len(req.SelectedPlatforms) >= len(allPlatforms)
		if selectedAllPlatforms {
			// 检查是否真的包含了所有平台
			selectedAllPlatforms = true
			for _, platform := range allPlatforms {
				if !selectedPlatformMap[platform] {
					selectedAllPlatforms = false
					break
				}
			}
		}

		var record *model.SubmissionRecord
		var err error

		if selectedAllPlatforms {
			// 选择了全部平台，等同于即时检测所有链接
			record, err = h.checkerService.CheckRealtime(resp.SubmissionID, resp.PendingLinks)
		} else {
			// 只检测选中平台的链接
			record, err = h.checkerService.CheckRealtimeWithPlatformFilter(resp.SubmissionID, resp.PendingLinks, req.SelectedPlatforms)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 更新响应：包含即时检测的结果
		validLinksList := []string(record.ValidLinks)
		remainingPendingLinks := []string(record.PendingLinks)
		resp.ValidLinks = validLinksList
		resp.PendingLinks = remainingPendingLinks // 如果选择了全部平台，这里应该是空的
		resp.TotalDuration = record.TotalDuration

		// 计算新检测到的失效链接
		// 失效链接 = 原始待检测链接中属于选中平台的链接 - 有效链接 - 已知失效链接
		knownInvalidMap := make(map[string]bool)
		for _, il := range resp.InvalidLinks {
			knownInvalidMap[il] = true
		}

		validLinksMap := make(map[string]bool)
		for _, vl := range validLinksList {
			validLinksMap[vl] = true
		}

		remainingPendingMap := make(map[string]bool)
		for _, rpl := range remainingPendingLinks {
			remainingPendingMap[rpl] = true
		}

		// 从原始待检测链接中找出失效的
		newInvalidLinks := make([]string, 0)
		for _, pendingLink := range originalPendingLinks {
			// 解析链接获取平台（ParseLink 会规范化链接）
			linkInfo := validator.ParseLink(pendingLink)
			normalizedLink := linkInfo.Link

			if selectedAllPlatforms {
				// 如果选择了全部平台，所有链接都应该被检测过
				// 如果不在有效链接中，也不在剩余待检测链接中，说明是新检测到的失效链接
				if !validLinksMap[normalizedLink] && !remainingPendingMap[normalizedLink] && !knownInvalidMap[normalizedLink] {
					newInvalidLinks = append(newInvalidLinks, normalizedLink)
				}
			} else {
				// 如果链接属于选中的平台，且已经检测过（不在 remainingPendingLinks 中）
				if selectedPlatformMap[linkInfo.Platform] && !remainingPendingMap[normalizedLink] {
					// 如果不在有效链接中，也不在已知失效链接中，说明是新检测到的失效链接
					if !validLinksMap[normalizedLink] && !knownInvalidMap[normalizedLink] {
						newInvalidLinks = append(newInvalidLinks, normalizedLink)
					}
				}
			}
		}
		resp.InvalidLinks = append(resp.InvalidLinks, newInvalidLinks...)
	}

	c.JSON(http.StatusOK, resp)
}

// GetSubmission 获取提交记录
func (h *LinkHandler) GetSubmission(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	record, err := h.linkService.GetSubmission(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	c.JSON(http.StatusOK, record)
}

// ListSubmissions 分页查询提交记录
func (h *LinkHandler) ListSubmissions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	records, total, err := h.linkService.ListSubmissions(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      records,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ListRateLimitedLinks 分页查询被限制的失效链接
func (h *LinkHandler) ListRateLimitedLinks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	platformStr := c.Query("platform")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var platform *model.Platform
	if platformStr != "" {
		p := model.Platform(platformStr)
		platform = &p
	}

	links, total, err := h.linkService.ListRateLimitedLinks(page, pageSize, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      links,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// ClearRateLimitedLinks 清空所有被限制的失效链接
func (h *LinkHandler) ClearRateLimitedLinks(c *gin.Context) {
	err := h.linkService.ClearRateLimitedLinks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "已清空所有受限链接"})
}
