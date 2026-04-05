package repository

import (
	"PanCheck/internal/model"
	"PanCheck/pkg/database"
)

// InvalidLinkRepository 失效链接仓库
type InvalidLinkRepository struct{}

// NewInvalidLinkRepository 创建失效链接仓库
func NewInvalidLinkRepository() *InvalidLinkRepository {
	return &InvalidLinkRepository{}
}

// FindByLinks 根据链接列表查找失效链接
func (r *InvalidLinkRepository) FindByLinks(links []string) ([]model.InvalidLink, error) {
	var invalidLinks []model.InvalidLink
	err := database.DB.Where("link IN ?", links).Find(&invalidLinks).Error
	return invalidLinks, err
}

// CreateOrUpdate 创建或更新失效链接
func (r *InvalidLinkRepository) CreateOrUpdate(invalidLink *model.InvalidLink) error {
	var existing model.InvalidLink
	err := database.DB.Where("link = ?", invalidLink.Link).First(&existing).Error

	if err != nil {
		// 不存在，创建新记录
		return database.DB.Create(invalidLink).Error
	}

	// 存在，更新记录
	existing.FailureReason = invalidLink.FailureReason
	existing.CheckDuration = invalidLink.CheckDuration
	existing.IsRateLimited = invalidLink.IsRateLimited
	// 如果新记录有 SubmissionID，则更新（允许追踪最新来源）
	if invalidLink.SubmissionID != nil {
		existing.SubmissionID = invalidLink.SubmissionID
	}

	return database.DB.Save(&existing).Error
}

// List 分页查询失效链接
func (r *InvalidLinkRepository) List(page, pageSize int, platform *model.Platform) ([]model.InvalidLink, int64, error) {
	var invalidLinks []model.InvalidLink
	var total int64

	query := database.DB.Model(&model.InvalidLink{})
	if platform != nil {
		query = query.Where("platform = ?", *platform)
	}

	offset := (page - 1) * pageSize
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&invalidLinks).Error
	return invalidLinks, total, err
}

// Exists 检查链接是否已存在于失效链接表（排除被限制的记录）
func (r *InvalidLinkRepository) Exists(link string) (bool, error) {
	var count int64
	err := database.DB.Model(&model.InvalidLink{}).
		Where("link = ? AND is_rate_limited = ?", link, false).
		Count(&count).Error
	return count > 0, err
}

// Count 统计失效链接总数
func (r *InvalidLinkRepository) Count(count *int64) error {
	return database.DB.Model(&model.InvalidLink{}).Where("is_rate_limited = ?", false).Count(count).Error
}

// CountByPlatform 按平台统计失效链接数
func (r *InvalidLinkRepository) CountByPlatform(platform model.Platform, count *int64) error {
	return database.DB.Model(&model.InvalidLink{}).
		Where("platform = ?", platform).
		Count(count).Error
}

// CountByRateLimited 统计被限制的失效链接数
func (r *InvalidLinkRepository) CountByRateLimited(count *int64) error {
	return database.DB.Model(&model.InvalidLink{}).
		Where("is_rate_limited = ?", true).
		Count(count).Error
}

// ListRateLimited 分页查询被限制的失效链接
func (r *InvalidLinkRepository) ListRateLimited(page, pageSize int, platform *model.Platform) ([]model.InvalidLink, int64, error) {
	var invalidLinks []model.InvalidLink
	var total int64

	query := database.DB.Model(&model.InvalidLink{}).Where("is_rate_limited = ?", true)
	if platform != nil {
		query = query.Where("platform = ?", *platform)
	}

	offset := (page - 1) * pageSize
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&invalidLinks).Error
	return invalidLinks, total, err
}

// DeleteRateLimited 删除所有被限制的失效链接
func (r *InvalidLinkRepository) DeleteRateLimited() error {
	return database.DB.Where("is_rate_limited = ?", true).Delete(&model.InvalidLink{}).Error
}
