package repository

import (
	"PanCheck/internal/model"
	"PanCheck/pkg/database"
)

// SettingsRepository 设置仓库
type SettingsRepository struct{}

// NewSettingsRepository 创建设置仓库
func NewSettingsRepository() *SettingsRepository {
	return &SettingsRepository{}
}

// GetByKey 根据key获取设置
func (r *SettingsRepository) GetByKey(key string) (*model.Setting, error) {
	var setting model.Setting
	err := database.DB.Where("`key` = ?", key).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// GetOrCreate 获取或创建设置
func (r *SettingsRepository) GetOrCreate(key, value, description, category string) (*model.Setting, error) {
	var setting model.Setting
	err := database.DB.Where("`key` = ?", key).First(&setting).Error

	if err != nil {
		// 不存在，创建新记录
		setting = model.Setting{
			Key:         key,
			Value:       value,
			Description: description,
			Category:    category,
		}
		err = database.DB.Create(&setting).Error
		if err != nil {
			return nil, err
		}
		return &setting, nil
	}

	return &setting, nil
}

// Update 更新设置（如果不存在则创建）
func (r *SettingsRepository) Update(setting *model.Setting) error {
	var existing model.Setting
	err := database.DB.Where("`key` = ?", setting.Key).First(&existing).Error

	if err != nil {
		// 不存在，创建新记录
		return database.DB.Create(setting).Error
	}

	// 存在，更新记录
	existing.Value = setting.Value
	if setting.Description != "" {
		existing.Description = setting.Description
	}
	if setting.Category != "" {
		existing.Category = setting.Category
	}
	return database.DB.Save(&existing).Error
}

// GetByCategory 根据分类获取设置列表
func (r *SettingsRepository) GetByCategory(category string) ([]model.Setting, error) {
	var settings []model.Setting
	err := database.DB.Where("`category` = ?", category).Find(&settings).Error
	return settings, err
}

// GetAll 获取所有设置
func (r *SettingsRepository) GetAll() ([]model.Setting, error) {
	var settings []model.Setting
	err := database.DB.Find(&settings).Error
	return settings, err
}
