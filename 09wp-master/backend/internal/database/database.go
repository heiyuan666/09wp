package database

import (
	"log"

	"dfan-netdisk-backend/internal/config"
	"dfan-netdisk-backend/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

// InitMySQL 初始化全局 GORM DB
func InitMySQL(dsn string) error {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

// DB 获取全局 DB 实例
func DB() *gorm.DB {
	if db == nil {
		log.Panic("database not initialized")
	}
	return db
}

// AutoMigrate 自动迁移核心表结构
func AutoMigrate() error {
	// 须先于 AutoMigrate：历史 uniqueIndex 会使多条 external_id='' 触发 Duplicate entry（MySQL 允许多 NULL 不允许多 ''）
	if err := migrateResourcesExternalIDDropUniqueIndex(); err != nil {
		return err
	}
	if err := DB().AutoMigrate(
		&model.Admin{},
		&model.User{},
		&model.Role{},
		&model.SystemConfig{},
		&model.TelegramAuthState{},
		&model.TelegramChannel{},
		&model.Category{},
		&model.HaokaCategory{},
		&model.HaokaProduct{},
		&model.HaokaSku{},
		&model.Resource{},
		&model.ResourceTransferLog{},
		&model.UserFavorite{},
		&model.KeywordBlock{},
		&model.Menu{},
		&model.SearchHotWord{},
		&model.NavigationMenu{},
		&model.TMDBSearchCache{},
		&model.NetdiskCredential{},
		&model.ResourceFeedback{},
		&model.RSSSubscription{},
		&model.GameCategory{},
		&model.Game{},
		&model.GameResource{},
		&model.UserPasswordReset{},
		&model.UserResourceSubmission{},
		&model.CaptchaChallenge{},
		&model.EmailVerificationCode{},
		&model.QRLoginSession{},
	); err != nil {
		return err
	}
	if err := ensureRequiredColumns(); err != nil {
		return err
	}
	return ensureUTF8MB4Collation()
}

func ensureRequiredColumns() error {
	m := DB().Migrator()

	// gorm 在部分 MySQL 场景下可能不会补齐新列，这里做一次兜底。
	if !m.HasColumn(&model.SystemConfig{}, "PanCheckBaseURL") {
		if err := m.AddColumn(&model.SystemConfig{}, "PanCheckBaseURL"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "LinkCheckEnabled") {
		if err := m.AddColumn(&model.SystemConfig{}, "LinkCheckEnabled"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "LinkCheckInterval") {
		if err := m.AddColumn(&model.SystemConfig{}, "LinkCheckInterval"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "DoubanHotNavEnabled") {
		if err := m.AddColumn(&model.SystemConfig{}, "DoubanHotNavEnabled"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "HotSearchEnabled") {
		if err := m.AddColumn(&model.SystemConfig{}, "HotSearchEnabled"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "ShowSiteTitle") {
		if err := m.AddColumn(&model.SystemConfig{}, "ShowSiteTitle"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "HomeRankBoardEnabled") {
		if err := m.AddColumn(&model.SystemConfig{}, "HomeRankBoardEnabled"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "DoubanCoverProxyURL") {
		if err := m.AddColumn(&model.SystemConfig{}, "DoubanCoverProxyURL"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "TgImageProxyURL") {
		if err := m.AddColumn(&model.SystemConfig{}, "TgImageProxyURL"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "TMDBBearerToken") {
		if err := m.AddColumn(&model.SystemConfig{}, "TMDBBearerToken"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "TMDBProxyURL") {
		if err := m.AddColumn(&model.SystemConfig{}, "TMDBProxyURL"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "FooterQuickLinks") {
		if err := m.AddColumn(&model.SystemConfig{}, "FooterQuickLinks"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "FooterHotPlatforms") {
		if err := m.AddColumn(&model.SystemConfig{}, "FooterHotPlatforms"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "FooterSocialLinks") {
		if err := m.AddColumn(&model.SystemConfig{}, "FooterSocialLinks"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "FooterWechat") {
		if err := m.AddColumn(&model.SystemConfig{}, "FooterWechat"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "AutoDeleteInvalidLinks") {
		if err := m.AddColumn(&model.SystemConfig{}, "AutoDeleteInvalidLinks"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "HideInvalidLinksInSearch") {
		if err := m.AddColumn(&model.SystemConfig{}, "HideInvalidLinksInSearch"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "ClarityProjectID") {
		if err := m.AddColumn(&model.SystemConfig{}, "ClarityProjectID"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "ClarityEnabled") {
		if err := m.AddColumn(&model.SystemConfig{}, "ClarityEnabled"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.SystemConfig{}, "ResourceDetailAutoTransfer") {
		if err := m.AddColumn(&model.SystemConfig{}, "ResourceDetailAutoTransfer"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "LinkValid") {
		if err := m.AddColumn(&model.Resource{}, "LinkValid"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "LinkCheckMsg") {
		if err := m.AddColumn(&model.Resource{}, "LinkCheckMsg"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "LinkCheckedAt") {
		if err := m.AddColumn(&model.Resource{}, "LinkCheckedAt"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "TransferStatus") {
		if err := m.AddColumn(&model.Resource{}, "TransferStatus"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "TransferMsg") {
		if err := m.AddColumn(&model.Resource{}, "TransferMsg"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "TransferRetryCount") {
		if err := m.AddColumn(&model.Resource{}, "TransferRetryCount"); err != nil {
			return err
		}
	}
	if !m.HasColumn(&model.Resource{}, "TransferLastAt") {
		if err := m.AddColumn(&model.Resource{}, "TransferLastAt"); err != nil {
			return err
		}
	}
	// SQL 级兜底：先查列再加列，避免重复列报错和日志污染。
	requiredCols := []struct {
		table  string
		column string
		sql    string
	}{
		{"system_configs", "pancheck_base_url", "ALTER TABLE system_configs ADD COLUMN pancheck_base_url varchar(255) NOT NULL DEFAULT ''"},
		{"system_configs", "link_check_enabled", "ALTER TABLE system_configs ADD COLUMN link_check_enabled tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "link_check_interval", "ALTER TABLE system_configs ADD COLUMN link_check_interval int NOT NULL DEFAULT 3600"},
		{"system_configs", "friend_links", "ALTER TABLE system_configs ADD COLUMN friend_links text NOT NULL DEFAULT '[]'"},
		{"system_configs", "submission_need_review", "ALTER TABLE system_configs ADD COLUMN submission_need_review tinyint(1) NOT NULL DEFAULT 1"},
		{"system_configs", "submission_auto_transfer", "ALTER TABLE system_configs ADD COLUMN submission_auto_transfer tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "resource_detail_auto_transfer", "ALTER TABLE system_configs ADD COLUMN resource_detail_auto_transfer tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "haoka_user_id", "ALTER TABLE system_configs ADD COLUMN haoka_user_id varchar(120) NOT NULL DEFAULT ''"},
		{"system_configs", "haoka_secret", "ALTER TABLE system_configs ADD COLUMN haoka_secret varchar(255) NOT NULL DEFAULT ''"},
		{"system_configs", "haoka_sync_enabled", "ALTER TABLE system_configs ADD COLUMN haoka_sync_enabled tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "haoka_sync_interval", "ALTER TABLE system_configs ADD COLUMN haoka_sync_interval int NOT NULL DEFAULT 3600"},
		{"system_configs", "haoka_order_url", "ALTER TABLE system_configs ADD COLUMN haoka_order_url varchar(500) NOT NULL DEFAULT ''"},
		{"system_configs", "haoka_agent_reg_url", "ALTER TABLE system_configs ADD COLUMN haoka_agent_reg_url varchar(500) NOT NULL DEFAULT ''"},
		{"system_configs", "hot_search_enabled", "ALTER TABLE system_configs ADD COLUMN hot_search_enabled tinyint(1) NOT NULL DEFAULT 1"},
		{"system_configs", "show_site_title", "ALTER TABLE system_configs ADD COLUMN show_site_title tinyint(1) NOT NULL DEFAULT 1"},
		{"system_configs", "home_rank_board_enabled", "ALTER TABLE system_configs ADD COLUMN home_rank_board_enabled tinyint(1) NOT NULL DEFAULT 1"},
		{"system_configs", "clarity_project_id", "ALTER TABLE system_configs ADD COLUMN clarity_project_id varchar(64) NOT NULL DEFAULT ''"},
		{"system_configs", "clarity_enabled", "ALTER TABLE system_configs ADD COLUMN clarity_enabled tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "tmdb_bearer_token", "ALTER TABLE system_configs ADD COLUMN tmdb_bearer_token varchar(600) NOT NULL DEFAULT ''"},
		{"system_configs", "tmdb_proxy_url", "ALTER TABLE system_configs ADD COLUMN tmdb_proxy_url varchar(500) NOT NULL DEFAULT ''"},
		{"system_configs", "footer_quick_links", "ALTER TABLE system_configs ADD COLUMN footer_quick_links text NOT NULL"},
		{"system_configs", "footer_hot_platforms", "ALTER TABLE system_configs ADD COLUMN footer_hot_platforms text NOT NULL"},
		{"system_configs", "footer_social_links", "ALTER TABLE system_configs ADD COLUMN footer_social_links text NOT NULL"},
		{"system_configs", "footer_wechat", "ALTER TABLE system_configs ADD COLUMN footer_wechat varchar(120) NOT NULL DEFAULT ''"},
		{"system_configs", "quark_cookie", "ALTER TABLE system_configs ADD COLUMN quark_cookie text NOT NULL"},
		{"system_configs", "quark_auto_save", "ALTER TABLE system_configs ADD COLUMN quark_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "quark_target_folder_id", "ALTER TABLE system_configs ADD COLUMN quark_target_folder_id varchar(64) NOT NULL DEFAULT '0'"},
		{"system_configs", "quark_ad_filter_enabled", "ALTER TABLE system_configs ADD COLUMN quark_ad_filter_enabled tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "quark_banned_keywords", "ALTER TABLE system_configs ADD COLUMN quark_banned_keywords text NOT NULL"},
		{"netdisk_credentials", "quark_ad_filter_enabled", "ALTER TABLE netdisk_credentials ADD COLUMN quark_ad_filter_enabled tinyint(1) NOT NULL DEFAULT 0"},
		{"netdisk_credentials", "quark_banned_keywords", "ALTER TABLE netdisk_credentials ADD COLUMN quark_banned_keywords text NOT NULL"},
		{"netdisk_credentials", "aliyun_renew_api_url", "ALTER TABLE netdisk_credentials ADD COLUMN aliyun_renew_api_url varchar(500) NOT NULL DEFAULT ''"},
		{"system_configs", "pan115_cookie", "ALTER TABLE system_configs ADD COLUMN pan115_cookie text NOT NULL"},
		{"system_configs", "pan115_auto_save", "ALTER TABLE system_configs ADD COLUMN pan115_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "pan115_target_folder_id", "ALTER TABLE system_configs ADD COLUMN pan115_target_folder_id varchar(64) NOT NULL DEFAULT ''"},
		{"system_configs", "tianyi_cookie", "ALTER TABLE system_configs ADD COLUMN tianyi_cookie text NOT NULL"},
		{"system_configs", "tianyi_auto_save", "ALTER TABLE system_configs ADD COLUMN tianyi_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "tianyi_target_folder_id", "ALTER TABLE system_configs ADD COLUMN tianyi_target_folder_id varchar(32) NOT NULL DEFAULT '-11'"},
		{"system_configs", "pan123_cookie", "ALTER TABLE system_configs ADD COLUMN pan123_cookie text NOT NULL"},
		{"system_configs", "pan123_auto_save", "ALTER TABLE system_configs ADD COLUMN pan123_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "pan123_target_folder_id", "ALTER TABLE system_configs ADD COLUMN pan123_target_folder_id varchar(64) NOT NULL DEFAULT '0'"},
		{"system_configs", "baidu_cookie", "ALTER TABLE system_configs ADD COLUMN baidu_cookie text NOT NULL"},
		{"system_configs", "baidu_auto_save", "ALTER TABLE system_configs ADD COLUMN baidu_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "baidu_target_path", "ALTER TABLE system_configs ADD COLUMN baidu_target_path varchar(500) NOT NULL DEFAULT '/'"},
		{"system_configs", "xunlei_cookie", "ALTER TABLE system_configs ADD COLUMN xunlei_cookie text NOT NULL"},
		{"system_configs", "xunlei_auto_save", "ALTER TABLE system_configs ADD COLUMN xunlei_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"system_configs", "xunlei_target_folder_id", "ALTER TABLE system_configs ADD COLUMN xunlei_target_folder_id varchar(64) NOT NULL DEFAULT '0'"},
		{"netdisk_credentials", "xunlei_cookie", "ALTER TABLE netdisk_credentials ADD COLUMN xunlei_cookie text NOT NULL"},
		{"netdisk_credentials", "xunlei_auto_save", "ALTER TABLE netdisk_credentials ADD COLUMN xunlei_auto_save tinyint(1) NOT NULL DEFAULT 0"},
		{"netdisk_credentials", "xunlei_target_folder_id", "ALTER TABLE netdisk_credentials ADD COLUMN xunlei_target_folder_id varchar(64) NOT NULL DEFAULT '0'"},
		{"system_configs", "replace_link_after_transfer", "ALTER TABLE system_configs ADD COLUMN replace_link_after_transfer tinyint(1) NOT NULL DEFAULT 0"},
		{"netdisk_credentials", "replace_link_after_transfer", "ALTER TABLE netdisk_credentials ADD COLUMN replace_link_after_transfer tinyint(1) NOT NULL DEFAULT 0"},
		{"resources", "link_valid", "ALTER TABLE resources ADD COLUMN link_valid tinyint(1) NOT NULL DEFAULT 1"},
		{"resources", "link_check_msg", "ALTER TABLE resources ADD COLUMN link_check_msg varchar(255) NOT NULL DEFAULT ''"},
		{"resources", "link_checked_at", "ALTER TABLE resources ADD COLUMN link_checked_at datetime NULL"},
		{"resources", "transfer_status", "ALTER TABLE resources ADD COLUMN transfer_status varchar(20) NOT NULL DEFAULT ''"},
		{"resources", "transfer_msg", "ALTER TABLE resources ADD COLUMN transfer_msg varchar(255) NOT NULL DEFAULT ''"},
		{"resources", "transfer_retry_count", "ALTER TABLE resources ADD COLUMN transfer_retry_count int NOT NULL DEFAULT 0"},
		{"resources", "transfer_last_at", "ALTER TABLE resources ADD COLUMN transfer_last_at datetime NULL"},
		{"game_resources", "resource_type", "ALTER TABLE game_resources ADD COLUMN resource_type varchar(30) NOT NULL DEFAULT 'game'"},
		{"user_resource_submissions", "game_id", "ALTER TABLE user_resource_submissions ADD COLUMN game_id bigint unsigned NULL"},
		{"resource_transfer_logs", "filter_log", "ALTER TABLE resource_transfer_logs ADD COLUMN filter_log text NOT NULL"},
	}
	for _, item := range requiredCols {
		exists, err := columnExists(item.table, item.column)
		if err != nil {
			return err
		}
		if exists {
			continue
		}
		if err := DB().Exec(item.sql).Error; err != nil {
			return err
		}
	}
	// 兼容历史列名 pan_check_base_url：若存在则把值迁移到 pancheck_base_url。
	oldExists, err := columnExists("system_configs", "pan_check_base_url")
	if err != nil {
		return err
	}
	if oldExists {
		_ = DB().Exec(
			"UPDATE system_configs SET pancheck_base_url = pan_check_base_url WHERE pancheck_base_url = '' AND pan_check_base_url <> ''",
		).Error
	}
	// 盘查服务地址为空时写入默认失效检测接口（新列或历史空值）
	_ = DB().Exec(
		"UPDATE system_configs SET pancheck_base_url = ? WHERE pancheck_base_url = '' OR pancheck_base_url IS NULL",
		config.DefaultPanCheckBaseURL,
	).Error
	return nil
}

// migrateResourcesExternalIDDropUniqueIndex 删除 resources.external_id 上旧的唯一索引，改由 AutoMigrate 按模型重建普通 index。
func migrateResourcesExternalIDDropUniqueIndex() error {
	ok, err := tableExists("resources")
	if err != nil || !ok {
		return err
	}
	var cnt int64
	if err := DB().Raw(
		`SELECT COUNT(1) FROM information_schema.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'resources' AND INDEX_NAME = 'idx_resources_external_id'`,
	).Scan(&cnt).Error; err != nil {
		return err
	}
	if cnt == 0 {
		return nil
	}
	return DB().Exec("ALTER TABLE `resources` DROP INDEX `idx_resources_external_id`").Error
}

func columnExists(tableName, columnName string) (bool, error) {
	var count int64
	err := DB().Raw(
		"SELECT COUNT(1) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		tableName, columnName,
	).Scan(&count).Error
	return count > 0, err
}

func tableExists(tableName string) (bool, error) {
	var count int64
	err := DB().Raw(
		"SELECT COUNT(1) FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?",
		tableName,
	).Scan(&count).Error
	return count > 0, err
}

// ensureUTF8MB4Collation 统一关键表排序规则，避免 utf8/utf8mb4 混用比较报错（如 emoji 场景）。
func ensureUTF8MB4Collation() error {
	// 先设置当前数据库默认字符集/排序规则，保障后续新表默认一致。
	if err := DB().Exec("ALTER DATABASE CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci").Error; err != nil {
		return err
	}

	// 历史表可能是 utf8_general_ci，这里按需转换。
	tables := []string{
		"resources",
		"resource_transfer_logs",
		"search_hot_words",
		"user_resource_submissions",
		"games",
		"game_resources",
		"rss_subscriptions",
	}
	for _, t := range tables {
		ok, err := tableExists(t)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if err := DB().Exec("ALTER TABLE `" + t + "` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci").Error; err != nil {
			return err
		}
	}
	return nil
}
