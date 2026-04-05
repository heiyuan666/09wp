package repository

import (
	"PanCheck/internal/model"
	"PanCheck/pkg/database"
)

// TaskExecutionRepository 任务执行记录仓库
type TaskExecutionRepository struct{}

// NewTaskExecutionRepository 创建任务执行记录仓库
func NewTaskExecutionRepository() *TaskExecutionRepository {
	return &TaskExecutionRepository{}
}

// Create 创建任务执行记录
func (r *TaskExecutionRepository) Create(execution *model.TaskExecution) error {
	return database.DB.Create(execution).Error
}

// GetByID 根据ID获取任务执行记录
func (r *TaskExecutionRepository) GetByID(id uint) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := database.DB.First(&execution, id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// Update 更新任务执行记录
func (r *TaskExecutionRepository) Update(execution *model.TaskExecution) error {
	return database.DB.Save(execution).Error
}

// ListByTaskID 根据任务ID分页查询执行记录
func (r *TaskExecutionRepository) ListByTaskID(taskID uint, page, pageSize int) ([]model.TaskExecution, int64, error) {
	var executions []model.TaskExecution
	var total int64

	query := database.DB.Model(&model.TaskExecution{}).Where("task_id = ?", taskID)

	// 统计总数
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&executions).Error
	return executions, total, err
}

// GetLatestByTaskID 获取任务的最新执行记录
func (r *TaskExecutionRepository) GetLatestByTaskID(taskID uint) (*model.TaskExecution, error) {
	var execution model.TaskExecution
	err := database.DB.Where("task_id = ?", taskID).Order("created_at DESC").First(&execution).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

