package service

import (
	"fmt"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

type LinkCheckStats struct {
	SubmissionID int64 `json:"submission_id"`
	Valid        int   `json:"valid"`
	Invalid      int   `json:"invalid"`
	Unknown      int   `json:"unknown"`
	Checked      int   `json:"checked"`
	Details      []LinkCheckDetail `json:"details,omitempty"`
}

type LinkCheckDetail struct {
	Link   string `json:"link"`
	Status string `json:"status"` // valid/invalid/unknown
	Msg    string `json:"msg"`
}

func CheckResourceLinks(ids []uint64, selectedPlatforms []string, oneByOne bool) (LinkCheckStats, error) {
	selectedPlatforms = normalizeSelectedPlatforms(selectedPlatforms)
	db := database.DB().Model(&model.Resource{}).Where("status = 1")
	if len(ids) > 0 {
		db = db.Where("id IN ?", ids)
	}
	var resources []model.Resource
	if err := db.Select("id", "link").Find(&resources).Error; err != nil {
		return LinkCheckStats{}, fmt.Errorf("查询资源失败")
	}
	if len(resources) == 0 {
		return LinkCheckStats{}, fmt.Errorf("没有可检测资源")
	}

	links := make([]string, 0, len(resources))
	linkToIDs := make(map[string][]uint64, len(resources))
	for _, r := range resources {
		link := strings.TrimSpace(r.Link)
		if link == "" {
			continue
		}
		if _, ok := linkToIDs[link]; !ok {
			links = append(links, link)
		}
		linkToIDs[link] = append(linkToIDs[link], r.ID)
	}
	if len(links) == 0 {
		return LinkCheckStats{}, fmt.Errorf("没有可检测链接")
	}

	baseURL := ""
	var cfg model.SystemConfig
	if err := database.DB().Order("id ASC").First(&cfg).Error; err == nil {
		baseURL = strings.TrimSpace(cfg.PanCheckBaseURL)
	}
	if baseURL == "" {
		return LinkCheckStats{}, fmt.Errorf("请先在系统配置填写“失效检测地址”")
	}

	autoDeleteInvalid := cfg.AutoDeleteInvalidLinks

	stats := LinkCheckStats{Checked: len(resources)}
	now := time.Now()
	validSet := map[string]struct{}{}
	invalidSet := map[string]struct{}{}
	if oneByOne {
		details := make([]LinkCheckDetail, 0, len(links))
		for _, link := range links {
			result, err := PanCheckLinksWithPolling(PanCheckRequest{
				Links:             []string{link},
				SelectedPlatforms: selectedPlatforms,
			}, baseURL, 2, 2*time.Second)
			if err != nil {
				details = append(details, LinkCheckDetail{
					Link:   link,
					Status: "unknown",
					Msg:    "检测请求失败",
				})
				continue
			}
			stats.SubmissionID = result.SubmissionID
			if len(result.ValidLinks) > 0 {
				validSet[link] = struct{}{}
				details = append(details, LinkCheckDetail{Link: link, Status: "valid", Msg: "有效"})
			} else if len(result.InvalidLinks) > 0 {
				invalidSet[link] = struct{}{}
				details = append(details, LinkCheckDetail{Link: link, Status: "invalid", Msg: "失效"})
			} else {
				details = append(details, LinkCheckDetail{Link: link, Status: "unknown", Msg: "待检测/未知"})
			}
		}
		stats.Details = details
	} else {
		result, err := PanCheckLinksWithPolling(PanCheckRequest{
			Links:             links,
			SelectedPlatforms: selectedPlatforms,
		}, baseURL, 2, 3*time.Second)
		if err != nil {
			return LinkCheckStats{}, err
		}
		stats.SubmissionID = result.SubmissionID
		for _, v := range result.ValidLinks {
			validSet[strings.TrimSpace(v)] = struct{}{}
		}
		for _, v := range result.InvalidLinks {
			invalidSet[strings.TrimSpace(v)] = struct{}{}
		}
	}

	for link, rids := range linkToIDs {
		updates := map[string]interface{}{"link_checked_at": &now}
		deletedOk := false
		if _, ok := invalidSet[link]; ok {
			updates["link_valid"] = false
			updates["link_check_msg"] = "失效"
			if autoDeleteInvalid {
				// 物理删除：直接删除资源，避免前台继续出现
				_ = database.DB().Where("resource_id IN ?", rids).Delete(&model.UserFavorite{}).Error
				delErr := database.DB().Where("id IN ?", rids).Delete(&model.Resource{}).Error
				// 删除失败兜底：至少下架，避免泄露失效链接
				if delErr != nil {
					updates["status"] = 0
					_ = database.DB().Model(&model.Resource{}).Where("id IN ?", rids).Updates(updates).Error
				} else {
					deletedOk = true
				}
			} else {
				// 不删除：默认下架（前台 status=1 不再展示）
				updates["status"] = 0
			}
			stats.Invalid += len(rids)
		} else if _, ok := validSet[link]; ok {
			updates["link_valid"] = true
			updates["link_check_msg"] = "有效"
			stats.Valid += len(rids)
		} else {
			updates["link_check_msg"] = "待检测/未知"
			stats.Unknown += len(rids)
		}
		// 仅在未物理删除成功时更新
		if !deletedOk {
			_ = database.DB().Model(&model.Resource{}).Where("id IN ?", rids).Updates(updates).Error
		}
	}
	return stats, nil
}

func normalizeSelectedPlatforms(selected []string) []string {
	if len(selected) > 0 {
		return selected
	}
	// 为空时默认全平台，避免 PanCheck 将结果全部放入 pending_links。
	return []string{"quark", "uc", "baidu", "tianyi", "pan123", "pan115", "aliyun", "xunlei", "cmcc"}
}

