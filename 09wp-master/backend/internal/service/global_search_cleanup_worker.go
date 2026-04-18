package service

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"
)

func writeCleanupTaskLog(resourceID uint64, platform, action, status, message string) {
	if resourceID == 0 {
		return
	}
	_ = database.DB().Create(&model.CleanupTaskLog{
		Task:       "global_search_cleanup",
		ResourceID: resourceID,
		Platform:   strings.TrimSpace(platform),
		Action:     strings.TrimSpace(action),
		Status:     strings.TrimSpace(status),
		Message:    trimTo500(strings.TrimSpace(message)),
	}).Error
}

// StartGlobalSearchCleanupWorker 定时清理全网搜来源的历史分享链接资源。
func StartGlobalSearchCleanupWorker() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			var cfg model.SystemConfig
			if err := database.DB().Order("id ASC").First(&cfg).Error; err != nil {
				continue
			}
			if !cfg.GlobalSearchCleanupEnabled {
				continue
			}
			minutes := cfg.GlobalSearchCleanupMinutes
			if minutes <= 0 {
				days := cfg.GlobalSearchCleanupDays
				if days < 1 {
					days = 1
				}
				minutes = days * 24 * 60
			}
			if minutes < 1 {
				minutes = 1
			}
			expireAt := time.Now().Add(-time.Duration(minutes) * time.Minute)

			var ids []uint64
			if err := database.DB().
				Model(&model.Resource{}).
				Where("source = ? AND created_at < ?", "global_search", expireAt).
				Pluck("id", &ids).Error; err != nil {
				continue
			}
			if len(ids) == 0 {
				continue
			}
			targetIDs := ids
			if cfg.GlobalSearchCleanupDeleteNetdiskFiles {
				targetIDs = filterResourcesDeletedInNetdisk(ids)
			}
			if len(targetIDs) == 0 {
				continue
			}
			_ = database.DB().Where("resource_id IN ?", targetIDs).Delete(&model.UserFavorite{}).Error
			if err := database.DB().Where("id IN ?", targetIDs).Delete(&model.Resource{}).Error; err == nil {
				for _, rid := range targetIDs {
					writeCleanupTaskLog(rid, "", "delete_resource", "success", "站内资源记录已删除")
				}
				log.Printf("global search cleanup done: deleted=%d older_than=%dmin", len(targetIDs), minutes)
			}
		}
	}()
}

func filterResourcesDeletedInNetdisk(resourceIDs []uint64) []uint64 {
	if len(resourceIDs) == 0 {
		return nil
	}
	allowed := make([]uint64, 0, len(resourceIDs))
	hasLog := map[uint64]bool{}
	var logs []model.ResourceTransferLog
	if err := database.DB().
		Where("resource_id IN ? AND status = ? AND saved_file_ids <> ''", resourceIDs, "success").
		Order("id DESC").
		Find(&logs).Error; err != nil {
		return resourceIDs
	}
	seen := make(map[uint64]bool)
	for _, lg := range logs {
		if seen[lg.ResourceID] {
			continue
		}
		seen[lg.ResourceID] = true
		hasLog[lg.ResourceID] = true
		var fids []string
		if err := json.Unmarshal([]byte(lg.SavedFileIDs), &fids); err != nil || len(fids) == 0 {
			continue
		}
		var err error
		switch strings.TrimSpace(strings.ToLower(lg.Platform)) {
		case "quark":
			err = DeleteQuarkFilesByFids(fids)
		case "baidu":
			err = DeleteBaiduByPaths(fids)
		case "uc":
			err = DeleteUcFilesByFids(fids)
		case "xunlei":
			err = DeleteXunleiFiles(fids)
		default:
			err = nil
		}
		if err != nil {
			log.Printf("global search cleanup delete netdisk files failed: resource_id=%d platform=%s err=%v", lg.ResourceID, lg.Platform, err)
			writeCleanupTaskLog(lg.ResourceID, lg.Platform, "delete_netdisk_file", "failed", err.Error())
		} else {
			log.Printf("global search cleanup delete netdisk files ok: resource_id=%d platform=%s items=%d", lg.ResourceID, lg.Platform, len(fids))
			writeCleanupTaskLog(lg.ResourceID, lg.Platform, "delete_netdisk_file", "success", "网盘文件删除成功")
			allowed = append(allowed, lg.ResourceID)
		}
	}
	for _, rid := range resourceIDs {
		if !hasLog[rid] {
			writeCleanupTaskLog(rid, "", "delete_netdisk_file", "skipped", "未命中转存日志，跳过网盘文件删除")
			allowed = append(allowed, rid)
		}
	}
	return allowed
}
