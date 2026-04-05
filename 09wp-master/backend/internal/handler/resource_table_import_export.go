package handler

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
	"dfan-netdisk-backend/internal/service"
	"dfan-netdisk-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminResourceImportTable(c *gin.Context) {
	// multipart/form-data: file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, 400, "缺少 file")
		return
	}
	defer func() { _ = file.Close() }()

	if header.Size > 20*1024*1024 {
		response.Error(c, 400, "文件过大（>20MB）")
		return
	}

	raw, err := io.ReadAll(file)
	if err != nil {
		response.Error(c, 500, "读取文件失败: "+err.Error())
		return
	}

	rows, err := service.ParseResourceSpreadsheet(header.Filename, raw)
	if err != nil {
		response.Error(c, 400, "解析失败: "+err.Error())
		return
	}
	if len(rows) == 0 {
		response.Error(c, 400, "表格无有效行")
		return
	}

	added := 0
	updated := 0
	skipped := 0
	errs := make([]gin.H, 0, 10)

	// 用于去重/唯一键：以外部字段 external_id 为准；没有则用 title+link+category_id 计算
	hashExternalID := func(r service.ResourceImportRow) string {
		if strings.TrimSpace(r.ExternalID) != "" {
			return strings.TrimSpace(r.ExternalID)
		}
		rawKey := strings.TrimSpace(r.Title) + "|" + strings.TrimSpace(r.Link) + "|" + strconv.FormatUint(r.CategoryID, 10)
		sum := sha1.Sum([]byte(rawKey))
		return hex.EncodeToString(sum[:])
	}

	// 逐行导入：每行做一次分类校验，避免整批都失败
	if err := database.DB().Transaction(func(tx *gorm.DB) error {
		for i, row := range rows {
			rowIdx := i + 2 // 第一行表头，行号从 2 开始
			row.Title = strings.TrimSpace(row.Title)
			row.Link = strings.TrimSpace(row.Link)
			row.ExtractCode = strings.TrimSpace(row.ExtractCode)
			row.Cover = strings.TrimSpace(row.Cover)
			row.Tags = strings.TrimSpace(row.Tags)
			row.Description = strings.TrimSpace(row.Description)

			if row.Title == "" || row.Link == "" || row.CategoryID == 0 {
				skipped++
				continue
			}

			// 分类校验（要求存在；不限制显示/隐藏，保证可原样回填）
			var cat model.Category
			if err := tx.Where("id = ?", row.CategoryID).First(&cat).Error; err != nil {
				skipped++
				errs = append(errs, gin.H{"row": rowIdx, "message": "分类不存在"})
				continue
			}

			externalID := hashExternalID(row)
			src := strings.TrimSpace(row.Source)
			if src == "" {
				src = "table_import"
			}

			var existing model.Resource
			err := tx.Where("external_id = ?", externalID).First(&existing).Error
			if err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					return err
				}
				// create
				newRes := model.Resource{
					ID: row.ID,

					Title:       row.Title,
					Link:        row.Link,
					CategoryID:  row.CategoryID,
					Source:      src,
					ExternalID:  externalID,
					Description: row.Description,
					ExtractCode: row.ExtractCode,
					Cover:       row.Cover,
					Tags:        row.Tags,

					LinkValid:     row.LinkValid,
					LinkCheckMsg:  row.LinkCheckMsg,
					LinkCheckedAt: row.LinkCheckedAt,

					TransferStatus:     row.TransferStatus,
					TransferMsg:        row.TransferMsg,
					TransferRetryCount: row.TransferRetryCount,
					TransferLastAt:     row.TransferLastAt,

					ViewCount: row.ViewCount,
					SortOrder: row.SortOrder,
					Status:    row.Status,
				}
				// 可选：导出文件带时间戳则原样回填
				if row.CreatedAt != nil {
					newRes.CreatedAt = *row.CreatedAt
				}
				if row.UpdatedAt != nil {
					newRes.UpdatedAt = *row.UpdatedAt
				}
				if err := tx.Create(&newRes).Error; err != nil {
					return err
				}
				added++
				continue
			}

			// update
			updates := map[string]any{
				"title":        row.Title,
				"link":         row.Link,
				"category_id":  row.CategoryID,
				"description":  row.Description,
				"extract_code": row.ExtractCode,
				"cover":        row.Cover,
				"tags":         row.Tags,

				"source":      src,
				"external_id": externalID,

				"link_valid":      row.LinkValid,
				"link_check_msg":  row.LinkCheckMsg,
				"link_checked_at": row.LinkCheckedAt,

				"transfer_status":      row.TransferStatus,
				"transfer_msg":         row.TransferMsg,
				"transfer_retry_count": row.TransferRetryCount,
				"transfer_last_at":     row.TransferLastAt,

				"view_count": row.ViewCount,
				"sort_order": row.SortOrder,
				"status":     row.Status,
			}
			if row.CreatedAt != nil {
				updates["created_at"] = *row.CreatedAt
			}
			if row.UpdatedAt != nil {
				updates["updated_at"] = *row.UpdatedAt
			}
			if err := tx.Model(&model.Resource{}).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
				return err
			}
			updated++
		}
		return nil
	}); err != nil {
		response.Error(c, 500, "导入失败: "+err.Error())
		return
	}

	response.OK(c, gin.H{
		"added":   added,
		"updated": updated,
		"skipped": skipped,
		"errors":  errs,
	})
}

func AdminResourceExportTable(c *gin.Context) {
	// 导出资源管理记录：默认导出全部（不写入 resources 记录）
	exportAll := true
	if v := strings.TrimSpace(c.Query("export_all")); v != "" {
		exportAll = v == "1" || strings.EqualFold(v, "true")
	}

	// 可选筛选（仅当 export_all=false 时生效）
	title := strings.TrimSpace(c.Query("title"))
	cid := strings.TrimSpace(c.Query("category_id"))
	status := strings.TrimSpace(c.Query("status"))

	format := strings.ToLower(strings.TrimSpace(c.DefaultQuery("format", "xlsx")))
	if format != "csv" && format != "xlsx" {
		format = "xlsx"
	}

	limit := 5000
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 50000 {
		limit = 50000
	}

	db := database.DB().Model(&model.Resource{})
	if !exportAll {
		if title != "" {
			db = db.Where("title LIKE ?", "%"+title+"%")
		}
		if cid != "" {
			if x, err := strconv.ParseUint(cid, 10, 64); err == nil && x > 0 {
				db = db.Where("category_id = ?", x)
			}
		}
		if status != "" {
			if st, err := strconv.ParseInt(status, 10, 8); err == nil {
				db = db.Where("status = ?", st)
			}
		}
	}

	var list []model.Resource
	if err := db.Order("sort_order DESC, id DESC").Limit(limit).Find(&list).Error; err != nil {
		response.Error(c, 500, "导出查询失败: "+err.Error())
		return
	}

	// 保存到磁盘
	exportDir := "./storage/exports"
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		response.Error(c, 500, "创建目录失败: "+err.Error())
		return
	}

	ts := time.Now().Format("20060102_150405")
	fileExt := format
	fileName := fmt.Sprintf("resources_export_%s_%d.%s", ts, time.Now().UnixMilli(), fileExt)
	savePath := filepath.Join(exportDir, fileName)

	var genErr error
	if format == "csv" {
		genErr = service.ExportResourcesToCSV(list, savePath)
	} else {
		genErr = service.ExportResourcesToXLSX(list, savePath)
	}
	if genErr != nil {
		response.Error(c, 500, "生成文件失败: "+genErr.Error())
		return
	}

	response.OK(c, gin.H{
		"link":     "/public/exports/" + fileName,
		"filename": fileName,
		"count":    len(list),
	})
}
