package repository

import (
	"PanCheck/internal/model"
	"PanCheck/pkg/database"
	"time"
)

// ScheduledTaskRepository 任务计划仓库
type ScheduledTaskRepository struct{}

// NewScheduledTaskRepository 创建任务计划仓库
func NewScheduledTaskRepository() *ScheduledTaskRepository {
	return &ScheduledTaskRepository{}
}

// Create 创建任务计划
func (r *ScheduledTaskRepository) Create(task *model.ScheduledTask) error {
	return database.DB.Create(task).Error
}

// GetByID 根据ID获取任务计划
func (r *ScheduledTaskRepository) GetByID(id uint) (*model.ScheduledTask, error) {
	var task model.ScheduledTask
	err := database.DB.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update 更新任务计划
func (r *ScheduledTaskRepository) Update(task *model.ScheduledTask) error {
	// 使用 Omit 排除 CreatedAt 和 ID，避免零值时间字段导致 MySQL 报错
	// Updates 方法只会更新非零值字段
	return database.DB.Model(task).
		Omit("created_at", "id").
		Where("id = ?", task.ID).
		Updates(task).Error
}

// Delete 删除任务计划
func (r *ScheduledTaskRepository) Delete(id uint) error {
	return database.DB.Delete(&model.ScheduledTask{}, id).Error
}

// List 分页查询任务计划
// tags: 标签筛选（如果提供，只返回包含这些标签的任务）
// status: 状态筛选（如果提供）
func (r *ScheduledTaskRepository) List(page, pageSize int, tags []string, status string) ([]model.ScheduledTask, int64, error) {
	var tasks []model.ScheduledTask
	var total int64

	query := database.DB.Model(&model.ScheduledTask{})

	// 标签筛选
	if len(tags) > 0 {
		// 使用JSON_CONTAINS或JSONB操作符进行标签筛选
		dbType := database.DB.Dialector.Name()
		if dbType == "mysql" {
			// MySQL: 使用JSON_CONTAINS
			for _, tag := range tags {
				query = query.Where("JSON_CONTAINS(tags, ?)", `"`+tag+`"`)
			}
		} else {
			// PostgreSQL: 使用JSONB操作符
			for _, tag := range tags {
				query = query.Where("tags @> ?", `["`+tag+`"]`)
			}
		}
	}

	// 状态筛选
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

// GetAllActive 获取所有活跃的任务（用于调度器）
func (r *ScheduledTaskRepository) GetAllActive() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	err := database.DB.Where("status = ?", "active").Find(&tasks).Error
	return tasks, err
}

// GetExpiredTasks 获取已过期的任务
func (r *ScheduledTaskRepository) GetExpiredTasks() ([]model.ScheduledTask, error) {
	var tasks []model.ScheduledTask
	now := time.Now()
	err := database.DB.Where("status = ? AND auto_destroy_at IS NOT NULL AND auto_destroy_at <= ?", "active", now).Find(&tasks).Error
	return tasks, err
}

// GetAllTags 获取所有标签列表
func (r *ScheduledTaskRepository) GetAllTags() ([]string, error) {
	var tasks []model.ScheduledTask
	err := database.DB.Select("tags").Find(&tasks).Error
	if err != nil {
		return nil, err
	}

	// 收集所有标签并去重
	tagMap := make(map[string]bool)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			if tag != "" {
				tagMap[tag] = true
			}
		}
	}

	tags := make([]string, 0, len(tagMap))
	for tag := range tagMap {
		tags = append(tags, tag)
	}

	return tags, nil
}

// ExistsByName 检查指定名称的任务是否存在
// excludeID: 排除的任务ID（用于更新时检查，排除当前任务）
func (r *ScheduledTaskRepository) ExistsByName(name string, excludeID *uint) (bool, error) {
	var count int64
	query := database.DB.Model(&model.ScheduledTask{}).Where("name = ?", name)

	// 如果提供了排除ID，则排除该任务
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Count 统计定时任务总数
func (r *ScheduledTaskRepository) Count(count *int64) error {
	return database.DB.Model(&model.ScheduledTask{}).Count(count).Error
}