package handler

import (
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type openNetdiskResourceItem struct {
	ID             uint64     `json:"id"`
	Title          string     `json:"title"`
	CategoryID     uint64     `json:"category_id"`
	CategoryName   string     `json:"category_name"`
	Platform       string     `json:"platform"`
	Link           string     `json:"link"`
	ExtraLinks     []string   `json:"extra_links"`
	ExtractCode    string     `json:"extract_code"`
	Description    string     `json:"description"`
	Cover          string     `json:"cover"`
	Tags           []string   `json:"tags"`
	Source         string     `json:"source"`
	LinkValid      bool       `json:"link_valid"`
	LinkCheckMsg   string     `json:"link_check_msg"`
	TransferStatus string     `json:"transfer_status"`
	TransferMsg    string     `json:"transfer_msg"`
	ViewCount      uint64     `json:"view_count"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	LinkCheckedAt  *time.Time `json:"link_checked_at,omitempty"`
	TransferLastAt *time.Time `json:"transfer_last_at,omitempty"`
}

type openNetdiskResourceDetail struct {
	openNetdiskResourceItem
	LatestTransfer *openNetdiskTransferInfo `json:"latest_transfer,omitempty"`
}

type openNetdiskTransferInfo struct {
	Platform    string    `json:"platform"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	OwnShareURL string    `json:"own_share_url"`
	CreatedAt   time.Time `json:"created_at"`
}

func detectOpenResourcePlatform(link string) string {
	u := strings.ToLower(strings.TrimSpace(link))
	switch {
	case strings.Contains(u, "pan.baidu.com"):
		return "baidu"
	case strings.Contains(u, "pan.quark.cn"):
		return "quark"
	case strings.Contains(u, "pan.xunlei.com"):
		return "xunlei"
	case strings.Contains(u, "aliyundrive.com"), strings.Contains(u, "alipan.com"):
		return "aliyun"
	case strings.Contains(u, "cloud.189.cn"), strings.Contains(u, "caiyun.189"), strings.Contains(u, "tianyi"):
		return "tianyi"
	case strings.Contains(u, "yun.139.com"), strings.Contains(u, "caiyun.139.com"):
		return "yidong"
	case strings.Contains(u, "115.com"), strings.Contains(u, "115cdn.com"):
		return "115"
	case strings.Contains(u, "123pan"), strings.Contains(u, "123684"), strings.Contains(u, "123685"),
		strings.Contains(u, "123912"), strings.Contains(u, "123592"), strings.Contains(u, "123865"), strings.Contains(u, "123.net"):
		return "123pan"
	case strings.Contains(u, "drive-h.uc.cn"), strings.Contains(u, "drive.uc.cn"):
		return "uc"
	default:
		return "other"
	}
}

func splitResourceTags(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		switch r {
		case ',', ';', '|', '/', '\n', '\r', '\t':
			return true
		default:
			return false
		}
	})
	result := make([]string, 0, len(fields))
	seen := make(map[string]struct{}, len(fields))
	for _, item := range fields {
		tag := strings.TrimSpace(item)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		result = append(result, tag)
	}
	return result
}

func loadCategoryNameMap() map[uint64]string {
	var categories []model.Category
	if err := database.DB().Model(&model.Category{}).Where("status = ?", 1).Find(&categories).Error; err != nil {
		return map[uint64]string{}
	}
	result := make(map[uint64]string, len(categories))
	for _, item := range categories {
		result[item.ID] = item.Name
	}
	return result
}

func toOpenNetdiskResourceItem(res model.Resource, categoryNameMap map[uint64]string) openNetdiskResourceItem {
	ex := []string(nil)
	if len(res.ExtraLinks) > 0 {
		ex = append(ex, []string(res.ExtraLinks)...)
	}
	return openNetdiskResourceItem{
		ID:             res.ID,
		Title:          res.Title,
		CategoryID:     res.CategoryID,
		CategoryName:   categoryNameMap[res.CategoryID],
		Platform:       detectOpenResourcePlatform(res.Link),
		Link:           res.Link,
		ExtraLinks:     ex,
		ExtractCode:    res.ExtractCode,
		Description:    res.Description,
		Cover:          res.Cover,
		Tags:           splitResourceTags(res.Tags),
		Source:         res.Source,
		LinkValid:      res.LinkValid,
		LinkCheckMsg:   res.LinkCheckMsg,
		TransferStatus: res.TransferStatus,
		TransferMsg:    res.TransferMsg,
		ViewCount:      res.ViewCount,
		CreatedAt:      res.CreatedAt,
		UpdatedAt:      res.UpdatedAt,
		LinkCheckedAt:  res.LinkCheckedAt,
		TransferLastAt: res.TransferLastAt,
	}
}

func openNetdiskResourceListQuery(c *gin.Context, keyword string) *gorm.DB {
	q := database.DB().Model(&model.Resource{}).Where("status = ?", 1)
	q = applyResourceFilters(q, c)
	if keyword != "" {
		q = q.Where("title LIKE ? OR description LIKE ? OR tags LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	return q
}

func OpenNetdiskResourceList(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("q"))
	var res []model.Resource

	var total int64
	if err := openNetdiskResourceListQuery(c, keyword).Count(&total).Error; err != nil {
		response.Error(c, 500, "query total failed")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	orderExpr := sortOrderExpr(c.DefaultQuery("sort", "latest"))
	if err := openNetdiskResourceListQuery(c, keyword).
		Order(orderExpr).
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&res).Error; err != nil {
		response.Error(c, 500, "query resource list failed")
		return
	}

	categoryNameMap := loadCategoryNameMap()
	list := make([]openNetdiskResourceItem, 0, len(res))
	for _, item := range res {
		list = append(list, toOpenNetdiskResourceItem(item, categoryNameMap))
	}

	response.OK(c, gin.H{
		"list":      list,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func OpenNetdiskResourceDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if id == 0 {
		response.Error(c, 400, "invalid id")
		return
	}

	var res model.Resource
	if err := database.DB().Where("id = ? AND status = ?", id, 1).First(&res).Error; err != nil {
		response.Error(c, 404, "resource not found")
		return
	}

	item := openNetdiskResourceDetail{
		openNetdiskResourceItem: toOpenNetdiskResourceItem(res, loadCategoryNameMap()),
	}

	var transfer model.ResourceTransferLog
	if err := database.DB().
		Model(&model.ResourceTransferLog{}).
		Where("resource_id = ?", res.ID).
		Order("id DESC").
		First(&transfer).Error; err == nil {
		item.LatestTransfer = &openNetdiskTransferInfo{
			Platform:    transfer.Platform,
			Status:      transfer.Status,
			Message:     transfer.Message,
			OwnShareURL: transfer.OwnShareURL,
			CreatedAt:   transfer.CreatedAt,
		}
	}

	response.OK(c, item)
}
