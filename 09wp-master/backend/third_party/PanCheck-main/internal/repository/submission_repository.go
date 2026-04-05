package repository

import (
	"PanCheck/internal/model"
	"PanCheck/pkg/database"
	"time"
)

// SubmissionRepository 提交记录仓库
type SubmissionRepository struct{}

// NewSubmissionRepository 创建提交记录仓库
func NewSubmissionRepository() *SubmissionRepository {
	return &SubmissionRepository{}
}

// Create 创建提交记录
func (r *SubmissionRepository) Create(record *model.SubmissionRecord) error {
	return database.DB.Create(record).Error
}

// GetByID 根据ID获取提交记录
func (r *SubmissionRepository) GetByID(id uint) (*model.SubmissionRecord, error) {
	var record model.SubmissionRecord
	err := database.DB.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// Update 更新提交记录
func (r *SubmissionRepository) Update(record *model.SubmissionRecord) error {
	return database.DB.Save(record).Error
}

// GetPendingRecords 获取待检测的记录
func (r *SubmissionRepository) GetPendingRecords(limit int) ([]model.SubmissionRecord, error) {
	var records []model.SubmissionRecord
	query := database.DB.Where("status = ?", "pending")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Find(&records).Error
	return records, err
}

// UpdateStatusToChecking 原子性地将状态从 pending 更新为 checking（如果当前状态是 pending）
// 返回更新的行数，如果为 0 说明状态已经不是 pending（可能已被其他 goroutine 处理）
func (r *SubmissionRepository) UpdateStatusToChecking(id uint) (int64, error) {
	result := database.DB.Model(&model.SubmissionRecord{}).
		Where("id = ? AND status = ?", id, "pending").
		Update("status", "checking")
	return result.RowsAffected, result.Error
}

// List 分页查询提交记录
func (r *SubmissionRepository) List(page, pageSize int) ([]model.SubmissionRecord, int64, error) {
	var records []model.SubmissionRecord
	var total int64

	offset := (page - 1) * pageSize
	err := database.DB.Model(&model.SubmissionRecord{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = database.DB.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// Count 统计总记录数
func (r *SubmissionRepository) Count(count *int64) error {
	return database.DB.Model(&model.SubmissionRecord{}).Count(count).Error
}

// CountByStatus 按状态统计记录数
func (r *SubmissionRepository) CountByStatus(status string, count *int64) error {
	return database.DB.Model(&model.SubmissionRecord{}).
		Where("status = ?", status).
		Count(count).Error
}

// GetTimeSeries 获取时间序列数据
// granularity: "hour" 按小时分组, "day" 按天分组
func (r *SubmissionRepository) GetTimeSeries(startTime, endTime *time.Time, granularity string) ([]struct {
	Date  string
	Count int64
}, error) {
	var results []struct {
		Date  string
		Count int64
	}

	// 获取数据库类型
	dbType := database.DB.Dialector.Name()

	var selectClause, groupClause string
	if granularity == "hour" {
		// 按小时分组
		if dbType == "mysql" {
			// MySQL: DATE_FORMAT
			selectClause = "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as date"
			groupClause = "DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00')"
		} else {
			// PostgreSQL: DATE_TRUNC
			selectClause = "TO_CHAR(DATE_TRUNC('hour', created_at), 'YYYY-MM-DD HH24:00:00') as date"
			groupClause = "DATE_TRUNC('hour', created_at)"
		}
	} else {
		// 按天分组
		if dbType == "mysql" {
			selectClause = "DATE(created_at) as date"
			groupClause = "DATE(created_at)"
		} else {
			// PostgreSQL
			selectClause = "DATE(created_at) as date"
			groupClause = "DATE(created_at)"
		}
	}

	query := database.DB.Model(&model.SubmissionRecord{}).
		Select(selectClause + ", COUNT(*) as count").
		Group(groupClause)

	if startTime != nil {
		query = query.Where("created_at >= ?", *startTime)
	}
	if endTime != nil {
		query = query.Where("created_at <= ?", *endTime)
	}

	err := query.Order("date ASC").Scan(&results).Error
	return results, err
}
