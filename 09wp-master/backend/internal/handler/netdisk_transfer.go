package handler

import (
	"strings"

	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type netdiskTransferReq struct {
	Link     string `json:"link" binding:"required"`
	Passcode string `json:"passcode"`
}

type netdiskBatchTransferReq struct {
	Items []netdiskTransferReq `json:"items" binding:"required"`
}

// NetdiskTransferByLink 管理端按链接识别网盘并一键转存
func NetdiskTransferByLink(c *gin.Context) {
	var req netdiskTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	req.Link = strings.TrimSpace(req.Link)
	if req.Link == "" {
		response.Error(c, 400, "缺少 link")
		return
	}
	switch service.DetectTransferPlatform(req.Link) {
	case service.PlatformBaidu:
		res, err := service.BaiduSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformQuark:
		res, err := service.QuarkSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformUC:
		res, err := service.UcSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformPan115:
		res, err := service.Pan115SaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformTianyi:
		res, err := service.TianyiSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformPan123:
		res, err := service.Pan123SaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformAliyun:
		res, err := service.AliyunSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	case service.PlatformXunlei:
		res, err := service.XunleiSaveByShareLink(req.Link, req.Passcode)
		if err != nil {
			response.Error(c, 500, "转存失败: "+err.Error())
			return
		}
		response.OK(c, res)
	default:
		response.Error(c, 400, "无法识别或暂不支持的网盘链接")
	}
}

// NetdiskTransferBatchByLinks 管理端批量按链接识别网盘并一键转存
func NetdiskTransferBatchByLinks(c *gin.Context) {
	var req netdiskBatchTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if len(req.Items) == 0 {
		response.Error(c, 400, "items 不能为空")
		return
	}

	results := make([]gin.H, 0, len(req.Items))
	for i, item := range req.Items {
		link := strings.TrimSpace(item.Link)
		passcode := strings.TrimSpace(item.Passcode)
		if link == "" {
			results = append(results, gin.H{
				"index":   i,
				"link":    item.Link,
				"success": false,
				"message": "缺少 link",
			})
			continue
		}

		platform := service.DetectTransferPlatform(link)
		var (
			data any
			err  error
		)
		switch platform {
		case service.PlatformBaidu:
			data, err = service.BaiduSaveByShareLink(link, passcode)
		case service.PlatformQuark:
			data, err = service.QuarkSaveByShareLink(link, passcode)
		case service.PlatformUC:
			data, err = service.UcSaveByShareLink(link, passcode)
		case service.PlatformPan115:
			data, err = service.Pan115SaveByShareLink(link, passcode)
		case service.PlatformTianyi:
			data, err = service.TianyiSaveByShareLink(link, passcode)
		case service.PlatformPan123:
			data, err = service.Pan123SaveByShareLink(link, passcode)
		case service.PlatformAliyun:
			data, err = service.AliyunSaveByShareLink(link, passcode)
		case service.PlatformXunlei:
			data, err = service.XunleiSaveByShareLink(link, passcode)
		default:
			err = errUnsupportedPlatform
		}

		if err != nil {
			results = append(results, gin.H{
				"index":    i,
				"link":     link,
				"platform": platform.String(),
				"success":  false,
				"message":  err.Error(),
			})
			continue
		}
		results = append(results, gin.H{
			"index":    i,
			"link":     link,
			"platform": platform.String(),
			"success":  true,
			"data":     data,
		})
	}

	okCount := 0
	for _, r := range results {
		if v, ok := r["success"].(bool); ok && v {
			okCount++
		}
	}
	response.OK(c, gin.H{
		"total":   len(results),
		"success": okCount,
		"failed":  len(results) - okCount,
		"results": results,
	})
}

var errUnsupportedPlatform = &unsupportedPlatformError{}

type unsupportedPlatformError struct{}

func (*unsupportedPlatformError) Error() string {
	return "无法识别或暂不支持的网盘链接"
}
