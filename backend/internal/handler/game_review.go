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

func parseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
	return v
}

func parsePage(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	if v <= 0 {
		return 1
	}
	if v > 100000 {
		return 100000
	}
	return v
}

func parsePageSize(s string) int {
	v, _ := strconv.Atoi(strings.TrimSpace(s))
	if v <= 0 {
		return 10
	}
	if v > 50 {
		return 50
	}
	return v
}

type gameReviewCreateReq struct {
	GameID  uint64 `json:"game_id" binding:"required"`
	Rating  int    `json:"rating"` // 1~5 or 0
	Content string `json:"content" binding:"required"`
}

type gameReviewUserOut struct {
	ID       uint64  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

type gameReviewOut struct {
	ID        uint64           `json:"id"`
	GameID    uint64           `json:"game_id"`
	Rating    int              `json:"rating"`
	Content   string           `json:"content"`
	Helpful   uint64           `json:"helpful"`
	Unhelpful uint64           `json:"unhelpful"`
	CreatedAt time.Time        `json:"created_at"`
	User      gameReviewUserOut `json:"user"`
}

// GameReviewCreate 登录后发布评论
func GameReviewCreate(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)
	if userID == 0 {
		response.Error(c, 401, "未登录")
		return
	}

	var req gameReviewCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}

	content := strings.TrimSpace(req.Content)
	if len([]rune(content)) < 3 {
		response.Error(c, 400, "评论内容至少 3 个字")
		return
	}
	if len([]rune(content)) > 1000 {
		response.Error(c, 400, "评论内容过长（最多 1000 字）")
		return
	}
	if req.Rating != 0 && (req.Rating < 1 || req.Rating > 5) {
		response.Error(c, 400, "评分不合法（1~5）")
		return
	}

	// 确认游戏存在
	var game model.Game
	if err := database.DB().Select("id").First(&game, req.GameID).Error; err != nil {
		response.Error(c, 404, "游戏不存在")
		return
	}

	row := model.GameReview{
		GameID:  req.GameID,
		UserID:  userID,
		Rating:  req.Rating,
		Content: content,
		Status:  1,
	}
	if err := database.DB().Create(&row).Error; err != nil {
		response.Error(c, 500, "发布失败")
		return
	}
	response.OK(c, gin.H{"id": row.ID})
}

type gameReviewListResp struct {
	List         []gameReviewOut `json:"list"`
	Total        int64           `json:"total"`
	Average      float64         `json:"average"`
	Distribution []gin.H         `json:"distribution"` // [{stars:5,count:xx,percentage:xx}]
}

// GameReviewList 公开评论列表（按时间倒序）
func GameReviewList(c *gin.Context) {
	gameID := parseUint64(c.DefaultQuery("game_id", "0"))
	if gameID == 0 {
		response.Error(c, 400, "game_id 必填")
		return
	}
	page := parsePage(c.DefaultQuery("page", "1"))
	pageSize := parsePageSize(c.DefaultQuery("page_size", "10"))
	sortBy := strings.ToLower(strings.TrimSpace(c.DefaultQuery("sort", "recent")))

	db := database.DB().Model(&model.GameReview{}).Where("game_id = ? AND status = 1", gameID)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	var rows []model.GameReview
	order := "id DESC"
	if sortBy == "helpful" {
		// 最有帮助：先看 helpful_count，再看 (helpful - unhelpful)，最后按时间
		order = "helpful_count DESC, (helpful_count - unhelpful_count) DESC, id DESC"
	}
	if err := db.Order(order).Limit(pageSize).Offset((page-1)*pageSize).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询失败")
		return
	}

	// 取用户信息
	userIDs := make([]uint64, 0, len(rows))
	seen := map[uint64]struct{}{}
	for _, r := range rows {
		if _, ok := seen[r.UserID]; ok {
			continue
		}
		seen[r.UserID] = struct{}{}
		userIDs = append(userIDs, r.UserID)
	}
	users := make([]model.User, 0, len(userIDs))
	userByID := map[uint64]model.User{}
	if len(userIDs) > 0 {
		_ = database.DB().Select("id, username, avatar").Where("id IN ?", userIDs).Find(&users).Error
		for _, u := range users {
			userByID[u.ID] = u
		}
	}

	out := make([]gameReviewOut, 0, len(rows))
	for _, r := range rows {
		u := userByID[r.UserID]
		out = append(out, gameReviewOut{
			ID:        r.ID,
			GameID:    r.GameID,
			Rating:    r.Rating,
			Content:   r.Content,
			Helpful:   r.HelpfulCount,
			Unhelpful: r.UnhelpfulCount,
			CreatedAt: r.CreatedAt,
			User: gameReviewUserOut{
				ID:       u.ID,
				Username: u.Username,
				Avatar:   u.Avatar,
			},
		})
	}

	// 统计（评分分布与均分）
	type aggRow struct {
		Rating int
		Cnt    int64
	}
	var aggs []aggRow
	_ = database.DB().
		Model(&model.GameReview{}).
		Select("rating, COUNT(1) as cnt").
		Where("game_id = ? AND status = 1 AND rating BETWEEN 1 AND 5", gameID).
		Group("rating").
		Scan(&aggs).Error

	countByStar := map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0}
	var sum int64
	var cnt int64
	for _, a := range aggs {
		if a.Rating < 1 || a.Rating > 5 {
			continue
		}
		countByStar[a.Rating] = a.Cnt
		sum += int64(a.Rating) * a.Cnt
		cnt += a.Cnt
	}
	avg := 0.0
	if cnt > 0 {
		avg = float64(sum) / float64(cnt)
	}
	dist := make([]gin.H, 0, 5)
	for star := 5; star >= 1; star-- {
		cn := countByStar[star]
		percent := 0
		if cnt > 0 {
			percent = int(float64(cn) * 100.0 / float64(cnt))
		}
		dist = append(dist, gin.H{"stars": star, "count": cn, "percentage": percent})
	}

	response.OK(c, gameReviewListResp{
		List:         out,
		Total:        total,
		Average:      avg,
		Distribution: dist,
	})
}

type gameReviewVoteReq struct {
	Vote int `json:"vote" binding:"required"` // 1=helpful -1=unhelpful 0=cancel
}

// GameReviewVote 登录后：对评论投票（有帮助/无帮助/取消）
func GameReviewVote(c *gin.Context) {
	userIDVal, _ := c.Get("user_id")
	userID, _ := userIDVal.(uint64)
	if userID == 0 {
		response.Error(c, 401, "未登录")
		return
	}
	reviewID := parseUint64(c.Param("id"))
	if reviewID == 0 {
		response.Error(c, 400, "评论 ID 错误")
		return
	}
	var req gameReviewVoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "参数错误")
		return
	}
	if req.Vote != -1 && req.Vote != 0 && req.Vote != 1 {
		response.Error(c, 400, "vote 不合法")
		return
	}

	// 确认评论存在
	var rev model.GameReview
	if err := database.DB().First(&rev, reviewID).Error; err != nil {
		response.Error(c, 404, "评论不存在")
		return
	}

	var existing model.GameReviewVote
	err := database.DB().Where("review_id = ? AND user_id = ?", reviewID, userID).First(&existing).Error
	prev := int8(0)
	if err == nil {
		prev = existing.Vote
	}
	next := int8(req.Vote)
	if prev == next {
		response.OK(c, gin.H{"ok": true})
		return
	}

	// 事务：更新 vote + 计数
	tx := database.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err == nil {
		if err2 := tx.Model(&model.GameReviewVote{}).Where("id = ?", existing.ID).Update("vote", next).Error; err2 != nil {
			tx.Rollback()
			response.Error(c, 500, "投票失败")
			return
		}
	} else {
		row := model.GameReviewVote{ReviewID: reviewID, UserID: userID, Vote: next}
		if err2 := tx.Create(&row).Error; err2 != nil {
			tx.Rollback()
			response.Error(c, 500, "投票失败")
			return
		}
	}

	// delta counters
	decHelpful := prev == 1
	decUnhelpful := prev == -1
	incHelpful := next == 1
	incUnhelpful := next == -1

	if decHelpful {
		_ = tx.Model(&model.GameReview{}).Where("id = ? AND helpful_count > 0", reviewID).Update("helpful_count", gorm.Expr("helpful_count - 1")).Error
	}
	if decUnhelpful {
		_ = tx.Model(&model.GameReview{}).Where("id = ? AND unhelpful_count > 0", reviewID).Update("unhelpful_count", gorm.Expr("unhelpful_count - 1")).Error
	}
	if incHelpful {
		_ = tx.Model(&model.GameReview{}).Where("id = ?", reviewID).Update("helpful_count", gorm.Expr("helpful_count + 1")).Error
	}
	if incUnhelpful {
		_ = tx.Model(&model.GameReview{}).Where("id = ?", reviewID).Update("unhelpful_count", gorm.Expr("unhelpful_count + 1")).Error
	}

	if err2 := tx.Commit().Error; err2 != nil {
		response.Error(c, 500, "投票失败")
		return
	}
	response.OK(c, gin.H{"ok": true})
}

