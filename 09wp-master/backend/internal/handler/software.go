package handler

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func normalizeSoftwareURLs(in []string) []string {
	out := make([]string, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, raw := range in {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func parseSoftwareDate(v string) (*time.Time, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil, nil
	}
	layouts := []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, v, time.Local); err == nil {
			return &t, nil
		}
	}
	return nil, errors.New("invalid date")
}

func createSoftwareThumb(srcPath, thumbPath string, maxWidth int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return err
	}
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 {
		return errors.New("invalid image size")
	}
	if w <= maxWidth {
		maxWidth = w
	}
	newW := maxWidth
	newH := h * newW / w
	if newH <= 0 {
		newH = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		sy := b.Min.Y + y*h/newH
		for x := 0; x < newW; x++ {
			sx := b.Min.X + x*w/newW
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	if err := os.MkdirAll(filepath.Dir(thumbPath), os.ModePerm); err != nil {
		return err
	}
	out, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer out.Close()
	return jpeg.Encode(out, dst, &jpeg.Options{Quality: 85})
}

func SoftwareUploadCover(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "file is required")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	}
	if !allowed[ext] {
		response.Error(c, 400, "only jpg/jpeg/png/webp are supported")
		return
	}

	dir := filepath.Join("storage", "covers", "software")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		response.Error(c, 500, "create upload dir failed")
		return
	}
	base := fmt.Sprintf("%d_%d", time.Now().UnixNano(), os.Getpid())
	fileName := base + ext
	savePath := filepath.Join(dir, fileName)
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		response.Error(c, 500, "save file failed")
		return
	}

	thumbName := base + "_thumb.jpg"
	thumbPath := filepath.Join(dir, thumbName)
	if err := createSoftwareThumb(savePath, thumbPath, 320); err != nil {
		response.Error(c, 500, "generate thumbnail failed")
		return
	}

	response.OK(c, gin.H{
		"url":        "/public/covers/software/" + fileName,
		"thumb_url":  "/public/covers/software/" + thumbName,
		"preview_url": "/public/covers/software/" + fileName,
	})
}

func SoftwareCategoryCreate(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Slug      string `json:"slug" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Status    *int8  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	item := model.SoftwareCategory{
		Name:      strings.TrimSpace(req.Name),
		Slug:      strings.TrimSpace(req.Slug),
		SortOrder: req.SortOrder,
		Status:    1,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if item.Name == "" || item.Slug == "" {
		response.Error(c, 400, "name/slug 不能为空")
		return
	}
	if err := database.DB().Create(&item).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, item)
}

func SoftwareCategoryUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req struct {
		Name      string `json:"name"`
		Slug      string `json:"slug"`
		SortOrder *int   `json:"sort_order"`
		Status    *int8  `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if strings.TrimSpace(req.Slug) != "" {
		updates["slug"] = strings.TrimSpace(req.Slug)
	}
	if req.SortOrder != nil {
		updates["sort_order"] = *req.SortOrder
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		response.Error(c, 400, "无更新字段")
		return
	}
	if err := database.DB().Model(&model.SoftwareCategory{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareCategoryDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Delete(&model.SoftwareCategory{}, id).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareCategoryList(c *gin.Context) {
	var list []model.SoftwareCategory
	if err := database.DB().Order("sort_order DESC, id ASC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func PublicSoftwareCategoryList(c *gin.Context) {
	var list []model.SoftwareCategory
	if err := database.DB().Where("status = 1").Order("sort_order DESC, id ASC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func SoftwareCategorySort(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req struct {
		SortOrder int `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if err := database.DB().Model(&model.SoftwareCategory{}).Where("id = ?", id).Update("sort_order", req.SortOrder).Error; err != nil {
		response.Error(c, 500, "更新排序失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareCreate(c *gin.Context) {
	var req struct {
		Name              string   `json:"name" binding:"required"`
		Summary           string   `json:"summary"`
		CategoryID        uint64   `json:"category_id" binding:"required"`
		Version           string   `json:"version"`
		Cover             string   `json:"cover"`
		CoverThumb        string   `json:"cover_thumb"`
		Screenshots       []string `json:"screenshots"`
		Size              string   `json:"size"`
		Platforms         []string `json:"platforms"`
		Website           string   `json:"website"`
		DownloadDirect    []string `json:"download_direct"`
		DownloadPan       []string `json:"download_pan"`
		DownloadExtract   string   `json:"download_extract"`
		PublishedAt       string   `json:"published_at"`
		UpdatedAtOverride string   `json:"updated_at_override"`
		Status            *int8    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	publishedAt, err := parseSoftwareDate(req.PublishedAt)
	if err != nil {
		response.Error(c, 400, "published_at 格式错误")
		return
	}
	updatedAtOverride, err := parseSoftwareDate(req.UpdatedAtOverride)
	if err != nil {
		response.Error(c, 400, "updated_at_override 格式错误")
		return
	}

	item := model.Software{
		Name:              strings.TrimSpace(req.Name),
		Summary:           strings.TrimSpace(req.Summary),
		CategoryID:        req.CategoryID,
		Version:           strings.TrimSpace(req.Version),
		Cover:             strings.TrimSpace(req.Cover),
		CoverThumb:        strings.TrimSpace(req.CoverThumb),
		Screenshots:       model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.Screenshots)),
		Size:              strings.TrimSpace(req.Size),
		Platforms:         strings.Join(normalizeSoftwareURLs(req.Platforms), ","),
		Website:           strings.TrimSpace(req.Website),
		DownloadDirect:    model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadDirect)),
		DownloadPan:       model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadPan)),
		DownloadExtract:   strings.TrimSpace(req.DownloadExtract),
		PublishedAt:       publishedAt,
		UpdatedAtOverride: updatedAtOverride,
		Status:            1,
	}
	if req.Status != nil {
		item.Status = *req.Status
	}
	if item.Name == "" || item.CategoryID == 0 {
		response.Error(c, 400, "name/category_id 必填")
		return
	}
	if err := database.DB().Create(&item).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, item)
}

func SoftwareUpdate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var req struct {
		Name              string   `json:"name"`
		Summary           string   `json:"summary"`
		CategoryID        *uint64  `json:"category_id"`
		Version           string   `json:"version"`
		Cover             string   `json:"cover"`
		CoverThumb        string   `json:"cover_thumb"`
		Screenshots       []string `json:"screenshots"`
		Size              string   `json:"size"`
		Platforms         []string `json:"platforms"`
		Website           string   `json:"website"`
		DownloadDirect    []string `json:"download_direct"`
		DownloadPan       []string `json:"download_pan"`
		DownloadExtract   string   `json:"download_extract"`
		PublishedAt       string   `json:"published_at"`
		UpdatedAtOverride string   `json:"updated_at_override"`
		Status            *int8    `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Summary != "" {
		updates["summary"] = strings.TrimSpace(req.Summary)
	}
	if req.CategoryID != nil && *req.CategoryID > 0 {
		updates["category_id"] = *req.CategoryID
	}
	if req.Version != "" {
		updates["version"] = strings.TrimSpace(req.Version)
	}
	if req.Cover != "" {
		updates["cover"] = strings.TrimSpace(req.Cover)
	}
	if req.CoverThumb != "" {
		updates["cover_thumb"] = strings.TrimSpace(req.CoverThumb)
	}
	if req.Screenshots != nil {
		updates["screenshots"] = model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.Screenshots))
	}
	if req.Size != "" {
		updates["size"] = strings.TrimSpace(req.Size)
	}
	if req.Platforms != nil {
		updates["platforms"] = strings.Join(normalizeSoftwareURLs(req.Platforms), ",")
	}
	if req.Website != "" {
		updates["website"] = strings.TrimSpace(req.Website)
	}
	if req.DownloadDirect != nil {
		updates["download_direct"] = model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadDirect))
	}
	if req.DownloadPan != nil {
		updates["download_pan"] = model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadPan))
	}
	if req.DownloadExtract != "" {
		updates["download_extract"] = strings.TrimSpace(req.DownloadExtract)
	}
	if req.PublishedAt != "" {
		v, err := parseSoftwareDate(req.PublishedAt)
		if err != nil {
			response.Error(c, 400, "published_at 格式错误")
			return
		}
		updates["published_at"] = v
	}
	if req.UpdatedAtOverride != "" {
		v, err := parseSoftwareDate(req.UpdatedAtOverride)
		if err != nil {
			response.Error(c, 400, "updated_at_override 格式错误")
			return
		}
		updates["updated_at_override"] = v
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if len(updates) == 0 {
		response.Error(c, 400, "无更新字段")
		return
	}
	if err := database.DB().Model(&model.Software{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareDelete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	tx := database.DB().Begin()
	if tx.Error != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	if err := tx.Where("software_id = ?", id).Delete(&model.SoftwareVersion{}).Error; err != nil {
		tx.Rollback()
		response.Error(c, 500, "删除版本失败")
		return
	}
	if err := tx.Delete(&model.Software{}, id).Error; err != nil {
		tx.Rollback()
		response.Error(c, 500, "删除失败")
		return
	}
	if err := tx.Commit().Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	categoryID := strings.TrimSpace(c.Query("category_id"))
	version := strings.TrimSpace(c.Query("version"))

	db := database.DB().Model(&model.Software{})
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if categoryID != "" {
		db = db.Where("category_id = ?", categoryID)
	}
	if version != "" {
		db = db.Where("version LIKE ?", "%"+version+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	var list []model.Software
	if err := db.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, list, total)
}

func SoftwareDetail(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	var item model.Software
	if err := database.DB().First(&item, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	var versions []model.SoftwareVersion
	_ = database.DB().Where("software_id = ?", id).Order("published_at DESC, id DESC").Find(&versions).Error
	response.OK(c, gin.H{
		"software": item,
		"versions": versions,
	})
}

func SoftwareVersionCreate(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "software id 错误")
		return
	}
	var exists model.Software
	if err := database.DB().First(&exists, id).Error; err != nil {
		response.Error(c, 404, "软件不存在")
		return
	}
	var req struct {
		Version         string   `json:"version" binding:"required"`
		ReleaseNotes    string   `json:"release_notes"`
		PublishedAt     string   `json:"published_at"`
		DownloadDirect  []string `json:"download_direct"`
		DownloadPan     []string `json:"download_pan"`
		DownloadExtract string   `json:"download_extract"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	publishedAt, err := parseSoftwareDate(req.PublishedAt)
	if err != nil {
		response.Error(c, 400, "published_at 格式错误")
		return
	}
	item := model.SoftwareVersion{
		SoftwareID:      id,
		Version:         strings.TrimSpace(req.Version),
		ReleaseNotes:    strings.TrimSpace(req.ReleaseNotes),
		PublishedAt:     publishedAt,
		DownloadDirect:  model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadDirect)),
		DownloadPan:     model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadPan)),
		DownloadExtract: strings.TrimSpace(req.DownloadExtract),
	}
	if item.Version == "" {
		response.Error(c, 400, "version 不能为空")
		return
	}
	if err := database.DB().Create(&item).Error; err != nil {
		response.Error(c, 500, "创建失败")
		return
	}
	response.OK(c, item)
}

func SoftwareVersionUpdate(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)
	if versionID == 0 {
		response.Error(c, 400, "version id 错误")
		return
	}
	var req struct {
		Version         string   `json:"version"`
		ReleaseNotes    string   `json:"release_notes"`
		PublishedAt     string   `json:"published_at"`
		DownloadDirect  []string `json:"download_direct"`
		DownloadPan     []string `json:"download_pan"`
		DownloadExtract string   `json:"download_extract"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	updates := map[string]interface{}{}
	if req.Version != "" {
		updates["version"] = strings.TrimSpace(req.Version)
	}
	if req.ReleaseNotes != "" {
		updates["release_notes"] = strings.TrimSpace(req.ReleaseNotes)
	}
	if req.PublishedAt != "" {
		v, err := parseSoftwareDate(req.PublishedAt)
		if err != nil {
			response.Error(c, 400, "published_at 格式错误")
			return
		}
		updates["published_at"] = v
	}
	if req.DownloadDirect != nil {
		updates["download_direct"] = model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadDirect))
	}
	if req.DownloadPan != nil {
		updates["download_pan"] = model.NormalizeExtraShareLinks(normalizeSoftwareURLs(req.DownloadPan))
	}
	if req.DownloadExtract != "" {
		updates["download_extract"] = strings.TrimSpace(req.DownloadExtract)
	}
	if len(updates) == 0 {
		response.Error(c, 400, "无更新字段")
		return
	}
	if err := database.DB().Model(&model.SoftwareVersion{}).Where("id = ?", versionID).Updates(updates).Error; err != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareVersionDelete(c *gin.Context) {
	versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)
	if versionID == 0 {
		response.Error(c, 400, "version id 错误")
		return
	}
	if err := database.DB().Delete(&model.SoftwareVersion{}, versionID).Error; err != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	response.OK(c, nil)
}

func SoftwareVersionList(c *gin.Context) {
	softwareID, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if softwareID == 0 {
		response.Error(c, 400, "software id 错误")
		return
	}
	var list []model.SoftwareVersion
	if err := database.DB().Where("software_id = ?", softwareID).Order("published_at DESC, id DESC").Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list})
}

func PublicSoftwareList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	keyword := strings.TrimSpace(c.Query("keyword"))
	categoryID := strings.TrimSpace(c.Query("category_id"))
	version := strings.TrimSpace(c.Query("version"))

	db := database.DB().Model(&model.Software{}).Where("status = 1")
	if keyword != "" {
		db = db.Where("name LIKE ?", "%"+keyword+"%")
	}
	if categoryID != "" {
		db = db.Where("category_id = ?", categoryID)
	}
	if version != "" {
		db = db.Where("version LIKE ?", "%"+version+"%")
	}
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	var list []model.Software
	if err := db.Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OKPage(c, list, total)
}

func PublicSoftwareDetail(c *gin.Context) {
	var item model.Software
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}
	if err := database.DB().Where("id = ? AND status = 1", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, "记录不存在")
			return
		}
		response.Error(c, 500, "查询失败")
		return
	}
	var versions []model.SoftwareVersion
	_ = database.DB().Where("software_id = ?", id).Order("published_at DESC, id DESC").Find(&versions).Error
	response.OK(c, gin.H{
		"software": item,
		"versions": versions,
	})
}
