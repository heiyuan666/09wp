package handler

import (
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

func formatSubmissionAuthor(userID uint64, names ...string) string {
	for _, item := range names {
		name := strings.TrimSpace(item)
		if name != "" {
			return name
		}
	}
	return "用户投稿#" + strconv.FormatUint(userID, 10)
}

// UserSubmissionCreate handles frontend user submissions.
func UserSubmissionCreate(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	var req struct {
		Title       string `json:"title" binding:"required"`
		Link        string `json:"link" binding:"required"`
		CategoryID  uint64 `json:"category_id"`
		GameID      uint64 `json:"game_id"`
		Description string `json:"description"`
		ExtractCode string `json:"extract_code"`
		Tags        string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	title := strings.TrimSpace(req.Title)
	link := strings.TrimSpace(req.Link)
	if title == "" || link == "" || (req.CategoryID == 0 && req.GameID == 0) {
		response.Error(c, 400, "参数错误")
		return
	}

	if req.CategoryID > 0 {
		var cat model.Category
		if err := database.DB().Where("id = ? AND status = 1", req.CategoryID).First(&cat).Error; err != nil {
			response.Error(c, 400, "分类不存在")
			return
		}
	}

	var gameIDPtr *uint64
	if req.GameID > 0 {
		var game model.Game
		if err := database.DB().Select("id").Where("id = ?", req.GameID).First(&game).Error; err != nil {
			response.Error(c, 400, "游戏不存在")
			return
		}
		gameID := req.GameID
		gameIDPtr = &gameID
	}

	needReview := true
	autoTransfer := false
	var sysCfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&sysCfg).Error; err == nil {
		needReview = sysCfg.SubmissionNeedReview
		autoTransfer = sysCfg.SubmissionAutoTransfer
	}

	sub := model.UserResourceSubmission{
		UserID:      userID,
		GameID:      gameIDPtr,
		Title:       title,
		Link:        link,
		CategoryID:  req.CategoryID,
		Description: strings.TrimSpace(req.Description),
		ExtractCode: strings.TrimSpace(req.ExtractCode),
		Tags:        strings.TrimSpace(req.Tags),
		Status:      "pending",
		ReviewMsg:   "",
	}

	if !needReview {
		var resourceID uint64
		var gameResourceID uint64
		if err := database.DB().Transaction(func(tx *gorm.DB) error {
			sub.Status = "approved"
			sub.ReviewMsg = "系统已自动通过"
			if err := tx.Create(&sub).Error; err != nil {
				return err
			}

			if sub.CategoryID > 0 {
				res := model.Resource{
					Title:       sub.Title,
					Link:        sub.Link,
					CategoryID:  sub.CategoryID,
					Source:      "user",
					Description: sub.Description,
					ExtractCode: sub.ExtractCode,
					Tags:        sub.Tags,
					LinkValid:   true,
					SortOrder:   0,
					Status:      1,
				}
				if err := tx.Create(&res).Error; err != nil {
					return err
				}
				resourceID = res.ID
			}

			if sub.GameID != nil && *sub.GameID > 0 {
				gameRes := model.GameResource{
					GameID:       *sub.GameID,
					Title:        sub.Title,
					ResourceType: "submission",
					DownloadType: "用户投稿",
					PanType:      mergePanType("", []string{sub.Link}),
					DownloadURL:  sub.Link,
					ExtractCode:  sub.ExtractCode,
					Tested:       false,
					Author:       formatSubmissionAuthor(userID),
				}
				if err := tx.Create(&gameRes).Error; err != nil {
					return err
				}
				gameResourceID = gameRes.ID
			}
			return nil
		}); err != nil {
			response.Error(c, 500, "提交失败")
			return
		}

		response.OK(c, gin.H{
			"submission":       sub,
			"resource_id":      resourceID,
			"game_resource_id": gameResourceID,
			"auto_approved":    true,
		})

		if autoTransfer && resourceID > 0 {
			var res model.Resource
			if err := database.DB().First(&res, resourceID).Error; err == nil {
				if cred, err := service.LoadNetdiskCredentials(); err == nil && service.ShouldAutoTransferOnCreate(cred, res.Link) {
					service.MarkResourceTransferPending(res.ID, "用户投稿自动通过后等待转存")
					rid := res.ID
					go func() {
						defer func() { recover() }()
						_ = service.TransferResourceWithRetry(rid, 3)
					}()
				}
			}
		}
		return
	}

	if err := database.DB().Create(&sub).Error; err != nil {
		response.Error(c, 500, "提交失败")
		return
	}
	response.OK(c, sub)
}

// UserSubmissionMyList returns the current user's submissions.
func UserSubmissionMyList(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)

	var list []model.UserResourceSubmission
	if err := database.DB().
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(200).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}
	response.OK(c, list)
}

// AdminSubmissionList returns paginated user submissions for review.
func AdminSubmissionList(c *gin.Context) {
	db := database.DB().Model(&model.UserResourceSubmission{})
	if st := strings.TrimSpace(c.Query("status")); st != "" {
		db = db.Where("status = ?", st)
	}
	if q := strings.TrimSpace(c.Query("q")); q != "" {
		like := "%" + q + "%"
		db = db.Where("title LIKE ? OR link LIKE ? OR tags LIKE ?", like, like, like)
	}
	if uid := strings.TrimSpace(c.Query("user_id")); uid != "" {
		if u, err := strconv.ParseUint(uid, 10, 64); err == nil && u > 0 {
			db = db.Where("user_id = ?", u)
		}
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

	var list []model.UserResourceSubmission
	if err := db.Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&list).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	response.OK(c, gin.H{
		"list":  list,
		"total": total,
	})
}

// AdminSubmissionApprove approves a submission and writes it into resource tables.
func AdminSubmissionApprove(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}

	var sub model.UserResourceSubmission
	if err := database.DB().First(&sub, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	if sub.Status != "pending" {
		response.Error(c, 400, "当前状态不可审核")
		return
	}

	var user model.User
	_ = database.DB().Select("id", "username", "name").Where("id = ?", sub.UserID).First(&user).Error

	var resourceID uint64
	var gameResourceID uint64
	if err := database.DB().Transaction(func(tx *gorm.DB) error {
		if sub.CategoryID > 0 {
			res := model.Resource{
				Title:       sub.Title,
				Link:        sub.Link,
				CategoryID:  sub.CategoryID,
				Source:      "user",
				ExternalID:  "",
				Description: sub.Description,
				ExtractCode: sub.ExtractCode,
				Cover:       "",
				Tags:        sub.Tags,
				LinkValid:   true,
				SortOrder:   0,
				Status:      1,
			}
			if err := tx.Create(&res).Error; err != nil {
				return err
			}
			resourceID = res.ID
		}

		if sub.GameID != nil && *sub.GameID > 0 {
			gameRes := model.GameResource{
				GameID:       *sub.GameID,
				Title:        sub.Title,
				ResourceType: "submission",
				DownloadType: "用户投稿",
				PanType:      mergePanType("", []string{sub.Link}),
				DownloadURL:  sub.Link,
				Tested:       false,
				Author:       formatSubmissionAuthor(sub.UserID, user.Name, user.Username),
			}
			if err := tx.Create(&gameRes).Error; err != nil {
				return err
			}
			gameResourceID = gameRes.ID
		}

		return tx.Model(&model.UserResourceSubmission{}).Where("id = ?", id).Updates(map[string]any{
			"status":     "approved",
			"review_msg": "",
			"updated_at": time.Now(),
		}).Error
	}); err != nil {
		response.Error(c, 500, "审核失败")
		return
	}

	var sysCfg model.SystemConfig
	if resourceID > 0 && database.DB().Order("id ASC").First(&sysCfg).Error == nil && sysCfg.SubmissionAutoTransfer {
		var res model.Resource
		if err := database.DB().First(&res, resourceID).Error; err == nil {
			if cred, err := service.LoadNetdiskCredentials(); err == nil && service.ShouldAutoTransferOnCreate(cred, res.Link) {
				service.MarkResourceTransferPending(resourceID, "审核通过后等待自动转存")
				rid := resourceID
				go func() {
					defer func() { recover() }()
					_ = service.TransferResourceWithRetry(rid, 3)
				}()
			}
		}
	}

	response.OK(c, gin.H{
		"resource_id":      resourceID,
		"game_resource_id": gameResourceID,
	})
}

// AdminSubmissionReject rejects a submission with an optional reason.
func AdminSubmissionReject(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if id == 0 {
		response.Error(c, 400, "ID 错误")
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	var sub model.UserResourceSubmission
	if err := database.DB().First(&sub, id).Error; err != nil {
		response.Error(c, 404, "记录不存在")
		return
	}
	if sub.Status != "pending" {
		response.Error(c, 400, "当前状态不可审核")
		return
	}

	reason := strings.TrimSpace(req.Reason)
	if len(reason) > 255 {
		reason = reason[:255]
	}

	if err := database.DB().Model(&model.UserResourceSubmission{}).Where("id = ?", id).Updates(map[string]any{
		"status":     "rejected",
		"review_msg": reason,
		"updated_at": time.Now(),
	}).Error; err != nil {
		response.Error(c, 500, "驳回失败")
		return
	}
	response.OK(c, nil)
}
