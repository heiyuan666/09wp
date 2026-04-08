package service

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"dfan-netdisk-backend/internal/database"
	"dfan-netdisk-backend/internal/model"

	"github.com/meilisearch/meilisearch-go"
)

type MeiliReindexResult struct {
	Index     string `json:"index"`
	BatchSize int    `json:"batch_size"`
	Total     int64  `json:"total"`
	Indexed   int64  `json:"indexed"`
	Message   string `json:"message"`
}

func ensureMeiliIndexExists(primaryKey string) error {
	if !MeiliEnabled() {
		return errors.New("meili disabled")
	}
	uid := strings.TrimSpace(meili.cfg.Index)
	if uid == "" {
		uid = "resources"
	}
	pk := strings.TrimSpace(primaryKey)
	if pk == "" {
		pk = "id"
	}

	// 尝试读取索引；存在则确保 primaryKey 已设置
	if info, err := meili.client.GetIndex(uid); err == nil {
		// primaryKey 为空时，尝试设置为 id（若索引已写入且不允许修改，Meili 会返回错误）
		if strings.TrimSpace(info.PrimaryKey) == "" {
			_, err := meili.index.UpdateIndex(&meilisearch.UpdateIndexRequestParams{PrimaryKey: pk})
			return err
		}
		return nil
	}

	// 不存在则创建并指定 primaryKey
	_, err := meili.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        uid,
		PrimaryKey: pk,
	})
	return err
}

func ensureMeiliGameIndexExists(primaryKey string) error {
	if !MeiliGameEnabled() {
		return errors.New("meili disabled")
	}
	uid := strings.TrimSpace(meili.cfg.Index)
	if uid == "" {
		uid = "resources"
	}
	uid = uid + "_games"
	pk := strings.TrimSpace(primaryKey)
	if pk == "" {
		pk = "id"
	}

	if info, err := meili.client.GetIndex(uid); err == nil {
		if strings.TrimSpace(info.PrimaryKey) == "" {
			_, err := meili.gameIdx.UpdateIndex(&meilisearch.UpdateIndexRequestParams{PrimaryKey: pk})
			return err
		}
		return nil
	}

	_, err := meili.client.CreateIndex(&meilisearch.IndexConfig{
		Uid:        uid,
		PrimaryKey: pk,
	})
	return err
}

// MeiliReindexAll 全量重建 resources 索引（从 MySQL 扫描，分批 upsert 到 Meili）
func MeiliReindexAll(ctx context.Context, batchSize int) (MeiliReindexResult, error) {
	if !MeiliEnabled() {
		return MeiliReindexResult{}, errors.New("meili disabled")
	}
	if batchSize <= 0 || batchSize > 2000 {
		batchSize = 500
	}
	if err := ensureMeiliIndexExists(meili.cfg.PrimaryKey); err != nil {
		return MeiliReindexResult{}, err
	}
	_ = ensureMeiliIndexSettings()

	var total int64
	if err := database.DB().Model(&model.Resource{}).Count(&total).Error; err != nil {
		return MeiliReindexResult{}, err
	}

	var indexed int64
	var lastID uint64
	for {
		var rows []model.Resource
		tx := database.DB().Model(&model.Resource{}).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&rows)
		if tx.Error != nil {
			return MeiliReindexResult{}, tx.Error
		}
		if len(rows) == 0 {
			break
		}

		docs := make([]meiliResourceDoc, 0, len(rows))
		for _, r := range rows {
			docs = append(docs, toMeiliResourceDoc(r))
			lastID = r.ID
		}

		pk := "id"
		if _, err := meili.index.AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &pk}); err != nil {
			return MeiliReindexResult{}, err
		}
		indexed += int64(len(docs))
	}

	return MeiliReindexResult{
		Index:     strings.TrimSpace(meili.cfg.Index),
		BatchSize: batchSize,
		Total:     total,
		Indexed:   indexed,
		Message:   "reindex submitted (async tasks in meilisearch)",
	}, nil
}

// MeiliReindexGames 全量重建 games 索引（从 MySQL 扫描，分批 upsert 到 Meili）
func MeiliReindexGames(ctx context.Context, batchSize int) (MeiliReindexResult, error) {
	if !MeiliGameEnabled() {
		return MeiliReindexResult{}, errors.New("meili disabled")
	}
	if batchSize <= 0 || batchSize > 2000 {
		batchSize = 500
	}
	if err := ensureMeiliGameIndexExists(meili.cfg.PrimaryKey); err != nil {
		return MeiliReindexResult{}, err
	}
	_ = ensureMeiliGameIndexSettings()

	var total int64
	if err := database.DB().Model(&model.Game{}).Count(&total).Error; err != nil {
		return MeiliReindexResult{}, err
	}

	var indexed int64
	var lastID uint64
	for {
		var rows []model.Game
		tx := database.DB().Model(&model.Game{}).
			Where("id > ?", lastID).
			Order("id ASC").
			Limit(batchSize).
			Find(&rows)
		if tx.Error != nil {
			return MeiliReindexResult{}, tx.Error
		}
		if len(rows) == 0 {
			break
		}
		docs := make([]meiliGameDoc, 0, len(rows))
		for _, g := range rows {
			docs = append(docs, toMeiliGameDoc(g))
			lastID = g.ID
		}
		pk := "id"
		if _, err := meili.gameIdx.AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &pk}); err != nil {
			return MeiliReindexResult{}, err
		}
		indexed += int64(len(docs))
	}

	return MeiliReindexResult{
		Index:     strings.TrimSpace(meili.cfg.Index) + "_games",
		BatchSize: batchSize,
		Total:     total,
		Indexed:   indexed,
		Message:   "reindex submitted (async tasks in meilisearch)",
	}, nil
}

func ParseBatchSize(v string) int {
	v = strings.TrimSpace(v)
	if v == "" {
		return 0
	}
	n, _ := strconv.Atoi(v)
	return n
}
