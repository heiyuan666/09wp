package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func applyResourceFilters(db *gorm.DB, c *gin.Context) *gorm.DB {
	if cid := c.Query("category_id"); cid != "" {
		db = db.Where("category_id = ?", cid)
	}

	switch c.Query("platform") {
	case "baidu":
		db = db.Where("link LIKE ?", "%pan.baidu.com%")
	case "aliyun":
		db = db.Where("link LIKE ? OR link LIKE ?", "%aliyundrive.com%", "%alipan.com%")
	case "quark":
		db = db.Where("link LIKE ?", "%pan.quark.cn%")
	case "xunlei":
		db = db.Where("link LIKE ?", "%pan.xunlei.com%")
	case "uc":
		db = db.Where("link LIKE ? OR link LIKE ?", "%drive-h.uc.cn%", "%drive.uc.cn%")
	case "tianyi":
		db = db.Where("link LIKE ? OR link LIKE ? OR link LIKE ?", "%cloud.189.cn%", "%caiyun.189%", "%tianyi%")
	case "yidong":
		db = db.Where("link LIKE ? OR link LIKE ?", "%yun.139.com%", "%caiyun.139.com%")
	case "pan115":
		db = db.Where("link LIKE ? OR link LIKE ?", "%115.com%", "%115cdn.com%")
	case "pan123":
		db = db.Where(
			"link LIKE ? OR link LIKE ? OR link LIKE ? OR link LIKE ? OR link LIKE ? OR link LIKE ? OR link LIKE ?",
			"%123pan%", "%123684%", "%123685%", "%123912%", "%123592%", "%123865%", "%123.net%",
		)
	case "other":
		db = db.Where(`
			link NOT LIKE '%pan.baidu.com%' AND
			link NOT LIKE '%aliyundrive.com%' AND link NOT LIKE '%alipan.com%' AND
			link NOT LIKE '%pan.quark.cn%' AND
			link NOT LIKE '%pan.xunlei.com%' AND
			link NOT LIKE '%drive-h.uc.cn%' AND link NOT LIKE '%drive.uc.cn%' AND
			link NOT LIKE '%cloud.189.cn%' AND link NOT LIKE '%caiyun.189%' AND link NOT LIKE '%tianyi%' AND
			link NOT LIKE '%yun.139.com%' AND link NOT LIKE '%caiyun.139.com%' AND
			link NOT LIKE '%115.com%' AND link NOT LIKE '%115cdn.com%' AND
			link NOT LIKE '%123pan%' AND link NOT LIKE '%123684%' AND link NOT LIKE '%123685%' AND
			link NOT LIKE '%123912%' AND link NOT LIKE '%123592%' AND link NOT LIKE '%123865%' AND link NOT LIKE '%123.net%'
		`)
	}

	if linkValid := strings.TrimSpace(c.Query("link_valid")); linkValid != "" {
		switch strings.ToLower(linkValid) {
		case "1", "true":
			db = db.Where("link_valid = ?", true)
		case "0", "false":
			db = db.Where("link_valid = ?", false)
		}
	}

	return db
}

func sortOrderExpr(sort string) string {
	if sort == "hot" {
		return "view_count DESC, id DESC"
	}
	return "created_at DESC, id DESC"
}

func buildSearchRelevanceOrder(keywordNorm string, tokens []string) (string, []any) {
	exact := keywordNorm
	prefixLike := keywordNorm + "%"
	containsLike := "%" + keywordNorm + "%"

	parts := []string{
		"CASE WHEN title = ? THEN 140 ELSE 0 END",
		"CASE WHEN title LIKE ? THEN 90 ELSE 0 END",
		"CASE WHEN title LIKE ? THEN 45 ELSE 0 END",
		"CASE WHEN tags LIKE ? THEN 35 ELSE 0 END",
		"CASE WHEN description LIKE ? THEN 25 ELSE 0 END",
	}
	args := []any{exact, prefixLike, containsLike, containsLike, containsLike}

	if len(tokens) > 6 {
		tokens = tokens[:6]
	}
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		like := "%" + t + "%"
		parts = append(parts,
			"CASE WHEN title LIKE ? THEN 8 ELSE 0 END",
			"CASE WHEN tags LIKE ? THEN 5 ELSE 0 END",
			"CASE WHEN description LIKE ? THEN 4 ELSE 0 END",
		)
		args = append(args, like, like, like)
	}

	return "(" + strings.Join(parts, " + ") + ") DESC, created_at DESC, id DESC", args
}

// AdminResourceCreate 后台添加资源
func AdminResourceCreate(c *gin.Context) {
	var req struct {
		Title       string   `json:"title" binding:"required"`
		Link        string   `json:"link" binding:"required"`
		ExtraLinks  []string `json:"extra_links"`
		CategoryID  uint64   `json:"category_id" binding:"required"`
		Description string   `json:"description"`
		ExtractCode string   `json:"extract_code"`
		Cover       string   `json:"cover"`
		Tags        string   `json:"tags"`
		SortOrder   int      `json:"sort_order"`
		Status      int8     `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	res := model.Resource{
		Title:       req.Title,
		Link:        req.Link,
		ExtraLinks:  model.NormalizeExtraShareLinks(req.ExtraLinks),
		CategoryID:  req.CategoryID,
		Description: req.Description,
		ExtractCode: req.ExtractCode,
		Cover:       req.Cover,
		Tags:        req.Tags,
		LinkValid:   true,
		SortOrder:   req.SortOrder,
		Status:      req.Status,
	}
	if res.Status == 0 {
		res.Status = 1
	}

	if err := database.DB().Create(&res).Error; err != nil {
		log.Printf("AdminResourceCreate: %v", err)
		if strings.Contains(err.Error(), "Duplicate") {
			response.Error(c, 409, "创建失败：与已有记录冲突（如 external_id / 唯一索引），请检查是否重复导入")
			return
		}
		response.Error(c, 500, "创建失败")
		return
	}

	cred, cerr := service.LoadNetdiskCredentials()
	if cerr == nil && service.ShouldAutoTransferOnCreateMulti(cred, req.Link, res.ExtraLinks) {
		service.MarkResourceTransferPending(res.ID, "\u7b49\u5f85\u81ea\u52a8\u8f6c\u5b58")
		go func(rid uint64) {
			_ = service.TransferResourceWithRetry(rid, 3)
		}(res.ID)
	}

	response.OK(c, res)
}

func AdminResourceRetryTransfer(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	service.MarkResourceTransferPending(id, "\u624b\u52a8\u91cd\u8bd5\u4e2d")
	if err := service.TransferResourceWithRetryForce(id, 3); err != nil {
		response.Error(c, 500, "转存失败: "+err.Error())
		return
	}
	response.OK(c, nil)
}

// AdminResourceTransferLogs 获取资源转存尝试日志
func AdminResourceTransferLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 50
	}

	db := database.DB().Model(&model.ResourceTransferLog{}).Where("resource_id = ?", id)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

	var logs []model.ResourceTransferLog
	if err := db.Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&logs).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OKPage(c, logs, total)
}

// AdminResourceUpdate 后台更新资源
func AdminResourceUpdate(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Title       string    `json:"title" binding:"required"`
		Link        string    `json:"link" binding:"required"`
		ExtraLinks  *[]string `json:"extra_links"`
		CategoryID  uint64    `json:"category_id" binding:"required"`
		Description string    `json:"description"`
		ExtractCode string    `json:"extract_code"`
		Cover       string    `json:"cover"`
		Tags        string    `json:"tags"`
		SortOrder   int       `json:"sort_order"`
		Status      int8      `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	updates := map[string]interface{}{
		"title":        req.Title,
		"link":         req.Link,
		"category_id":  req.CategoryID,
		"description":  req.Description,
		"extract_code": req.ExtractCode,
		"cover":        req.Cover,
		"tags":         req.Tags,
		"sort_order":   req.SortOrder,
		"status":       req.Status,
	}
	if req.ExtraLinks != nil {
		updates["extra_links"] = model.NormalizeExtraShareLinks(*req.ExtraLinks)
	}
	if err := database.DB().Model(&model.Resource{}).Where("id = ?", id).
		Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// AdminResourceDelete 删除资源
func AdminResourceDelete(c *gin.Context) {
	id := c.Param("id")
	if err := database.DB().Delete(&model.Resource{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

// AdminResourceBatchDelete 批量删除资源
func AdminResourceBatchDelete(c *gin.Context) {
	var req struct {
		IDs []uint64 `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Where("id IN ?", req.IDs).Delete(&model.Resource{}).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func AdminResourceBatchStatus(c *gin.Context) {
	var req struct {
		IDs    []uint64 `json:"ids" binding:"required"`
		Status int8     `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.IDs) == 0 {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.Resource{}).
		Where("id IN ?", req.IDs).
		Update("status", req.Status).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

// AdminResourceList 后台资源列表
func AdminResourceList(c *gin.Context) {
	var (
		db  = database.DB().Model(&model.Resource{})
		res []model.Resource
	)

	title := c.Query("title")
	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}
	if cid := c.Query("category_id"); cid != "" {
		db = db.Where("category_id = ?", cid)
	}
	if status := c.Query("status"); status != "" {
		db = db.Where("status = ?", status)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
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

	if err := db.Order("sort_order DESC, id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&res).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OKPage(c, res, total)
}

func ResourceList(c *gin.Context) {
	var (
		db  = database.DB().Model(&model.Resource{}).Where("status = 1")
		res []model.Resource
	)

	db = applyResourceFilters(db, c)

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
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

	if err := db.Order(orderExpr).
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&res).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OKPage(c, res, total)
}

// ResourceDetail 获取资源详情并增加浏览量
func ResourceDetail(c *gin.Context) {
	id := c.Param("id")
	var res model.Resource
	if err := database.DB().Where("id = ? AND status = 1", id).First(&res).Error; err != nil {
		response.Error(c, 404, "\u8d44\u6e90\u4e0d\u5b58\u5728")
		return
	}
	// 浏览量加 1，不阻塞详情返回
	_ = database.DB().Model(&model.Resource{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error

	response.OK(c, res)
}

// ResourceAccessLink 获取资源访问链接，必要时触发后台转存
func ResourceAccessLink(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}

	var res model.Resource
	if err := database.DB().Where("id = ? AND status = 1", id).First(&res).Error; err != nil {
		response.Error(c, 404, "\u8d44\u6e90\u4e0d\u5b58\u5728")
		return
	}

	currentLink := strings.TrimSpace(res.Link)
	switch strings.TrimSpace(res.TransferStatus) {
	case "success":
		if currentLink != "" {
			var fresh model.Resource
			_ = database.DB().Where("id = ?", id).First(&fresh).Error
			links := fresh.AllShareLinks()
			ex := []string(nil)
			if len(fresh.ExtraLinks) > 0 {
				ex = append(ex, []string(fresh.ExtraLinks)...)
			}
			response.OK(c, gin.H{
				"status":      "success",
				"link":        fresh.Link,
				"extra_links": ex,
				"links":       links,
				"message":     "\u5df2\u4e3a\u4f60\u5207\u6362\u5230\u672c\u7ad9\u8f6c\u5b58\u94fe\u63a5",
			})
			return
		}
	case "pending":
		response.OK(c, gin.H{
			"status":  "pending",
			"message": "\u6b63\u5728\u4e3a\u4f60\u51c6\u5907\u53ef\u7528\u94fe\u63a5...",
		})
		return
	}

	cred, err := service.LoadNetdiskCredentials()
	var cfg model.SystemConfig
	resourceDetailAutoTransfer := false
	if dbErr := database.DB().Order("id ASC").First(&cfg).Error; dbErr == nil {
		resourceDetailAutoTransfer = cfg.ResourceDetailAutoTransfer
	}
	canAutoTransfer := resourceDetailAutoTransfer && err == nil && service.ShouldAutoTransferOnCreateMulti(cred, currentLink, res.ExtraLinks)
	if !canAutoTransfer {
		links := res.AllShareLinks()
		ex := []string(nil)
		if len(res.ExtraLinks) > 0 {
			ex = append(ex, []string(res.ExtraLinks)...)
		}
		response.OK(c, gin.H{
			"status":      "direct",
			"link":        currentLink,
			"extra_links": ex,
			"links":       links,
			"message":     "\u5f53\u524d\u672a\u5f00\u542f\u8be6\u60c5\u9875\u81ea\u52a8\u8f6c\u5b58\uff0c\u5df2\u8fd4\u56de\u53ef\u7528\u94fe\u63a5",
		})
		return
	}

	tx := database.DB().Model(&model.Resource{}).
		Where("id = ? AND status = 1 AND (transfer_status = '' OR transfer_status IS NULL OR transfer_status = ?)", id, "failed").
		Updates(map[string]any{
			"transfer_status": "pending",
			"transfer_msg":    "\u7528\u6237\u8bbf\u95ee\u8d44\u6e90\uff0c\u5f00\u59cb\u540e\u53f0\u8f6c\u5b58",
		})
	if tx.Error != nil {
		response.Error(c, 500, "\u521b\u5efa\u8f6c\u5b58\u4efb\u52a1\u5931\u8d25")
		return
	}

	if tx.RowsAffected > 0 {
		go func(resourceID uint64) {
			defer func() { recover() }()
			_ = service.TransferResourceWithRetry(resourceID, 3)
		}(id)
	}

	response.OK(c, gin.H{
		"status":  "pending",
		"message": "\u6b63\u5728\u4e3a\u4f60\u51c6\u5907\u53ef\u7528\u94fe\u63a5...",
	})
}

// ResourceLatestTransferLog 前台获取该资源最近一次成功转存日志
func ResourceLatestTransferLog(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var log model.ResourceTransferLog
	if err := database.DB().
		Model(&model.ResourceTransferLog{}).
		Where("resource_id = ? AND status = ?", id, "success").
		Order("id DESC").
		First(&log).Error; err != nil {
		response.OK(c, gin.H{
			"exists": false,
		})
		return
	}

	// filter_log 是 JSON 字符串，这里尽量解析成对象返回给前端
	var filter any
	if strings.TrimSpace(log.FilterLog) != "" {
		_ = json.Unmarshal([]byte(log.FilterLog), &filter)
	}
	response.OK(c, gin.H{
		"exists":        true,
		"platform":      log.Platform,
		"message":       log.Message,
		"own_share_url": log.OwnShareURL,
		"created_at":    log.CreatedAt,
		"filter_log":    filter,
	})
}

// ResourceSearch
func ResourceSearch(c *gin.Context) {
	keywordRaw := strings.TrimSpace(c.Query("q"))
	if keywordRaw == "" {
		response.Error(c, 400, "缺少搜索关键词")
		return
	}
	keywordNorm := service.NormalizeSearchQuery(keywordRaw)
	if keywordNorm == "" {
		response.OKPage(c, []model.Resource{}, 0)
		return
	}
	if service.IsKeywordBlockedText(keywordRaw) {
		response.OKPage(c, []model.Resource{}, 0)
		return
	}

	blocks, err := service.ListEnabledKeywordBlocks()
	if err != nil {
		blocks = nil
	}
	if len(blocks) > 200 {
		blocks = blocks[:200]
	}

	var (
		db  = database.DB().Model(&model.Resource{}).Where("status = 1")
		res []model.Resource
	)
	db = applyResourceFilters(db, c)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	sortParam := strings.TrimSpace(c.DefaultQuery("sort", "relevance"))

	hideInvalidInSearch := false
	if c.Query("link_valid") == "" {
		var cfg model.SystemConfig
		if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
			hideInvalidInSearch = cfg.HideInvalidLinksInSearch
			if hideInvalidInSearch {
				db = db.Where("link_valid = ?", true)
			}
		}
	}

	tokens := service.TokenizeSearchQuery(keywordNorm)
	phraseLike := "%" + keywordNorm + "%"
	db = db.Where(
		"(title LIKE ? OR description LIKE ? OR tags LIKE ?)",
		phraseLike, phraseLike, phraseLike,
	)
	for _, part := range strings.Fields(keywordNorm) {
		part = strings.TrimSpace(part)
		if part == "" || part == keywordNorm {
			continue
		}
		partLike := "%" + part + "%"
		db = db.Where(
			"(title LIKE ? OR description LIKE ? OR tags LIKE ?)",
			partLike, partLike, partLike,
		)
	}

	for _, w := range blocks {
		if w == "" {
			continue
		}
		notLike := "%" + service.EscapeLikePattern(w) + "%"
		db = db.Where(
			"title NOT LIKE ? ESCAPE '\\\\' AND description NOT LIKE ? ESCAPE '\\\\' AND tags NOT LIKE ? ESCAPE '\\\\'",
			notLike, notLike, notLike,
		)
	}

	type searchCachePage struct {
		List  []model.Resource `json:"list"`
		Total int64            `json:"total"`
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(keywordNorm))
	kwKey := fmt.Sprintf("%x", h.Sum64())
	cacheKey := fmt.Sprintf(
		"search:v4:%s:%s:%s:%s:%s:%d:%d:%t:%d",
		kwKey,
		strings.TrimSpace(c.Query("category_id")),
		strings.TrimSpace(c.Query("platform")),
		strings.TrimSpace(c.Query("link_valid")),
		sortParam,
		page,
		pageSize,
		hideInvalidInSearch,
		len(blocks),
	)
	if b, ok := service.GetSearchCache(context.Background(), cacheKey); ok {
		var cached searchCachePage
		if err := json.Unmarshal(b, &cached); err == nil {
			response.OKPage(c, cached.List, cached.Total)
			return
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "统计失败")
		return
	}

		query := db.Limit(pageSize).Offset((page - 1) * pageSize)
	switch sortParam {
	case "latest":
		query = query.Order(sortOrderExpr("latest"))
	case "hot":
		query = query.Order(sortOrderExpr("hot"))
	default:
		relevanceOrderSQL, relevanceArgs := buildSearchRelevanceOrder(keywordNorm, tokens)
		query = query.Clauses(clause.OrderBy{
			Expression: clause.Expr{
				SQL:  relevanceOrderSQL,
				Vars: relevanceArgs,
			},
		})
	}
	if err := query.Find(&res).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	if raw, err := json.Marshal(searchCachePage{List: res, Total: total}); err == nil {
		service.SetSearchCache(context.Background(), cacheKey, raw)
	}

	if page <= 1 {
		kw := keywordRaw
		go func() {
			defer func() { recover() }()
			service.RecordSearchKeyword(kw)
		}()
	}

	response.OKPage(c, res, total)
}

// AdminResourceSyncTelegram
func AdminResourceSyncTelegram(c *gin.Context) {
	synced, added, skipped, err := service.SyncAllEnabledTelegramChannels()
	if err != nil {
		response.Error(c, 500, "同步失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{
		"synced":  synced,
		"added":   added,
		"skipped": skipped,
	})
}

func AdminResourceCheckLinks(c *gin.Context) {
	var req struct {
		IDs               []uint64 `json:"ids"`
		SelectedPlatforms []string `json:"selectedPlatforms"`
		OneByOne          bool     `json:"one_by_one"`
	}
	_ = c.ShouldBindJSON(&req)

	stats, err := service.CheckResourceLinks(req.IDs, req.SelectedPlatforms, req.OneByOne)
	if err != nil {
		response.Error(c, 500, "检测失败: "+err.Error())
		return
	}
	response.OK(c, gin.H{
		"submission_id": stats.SubmissionID,
		"valid":         stats.Valid,
		"invalid":       stats.Invalid,
		"unknown":       stats.Unknown,
		"checked":       stats.Checked,
		"details":       stats.Details,
	})
}

