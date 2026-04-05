package service

import (
	"PanCheck/internal/model"
	"PanCheck/internal/repository"
	"time"
)

// StatisticsService 统计服务
type StatisticsService struct {
	submissionRepo   *repository.SubmissionRepository
	invalidLinkRepo  *repository.InvalidLinkRepository
	scheduledTaskRepo *repository.ScheduledTaskRepository
}

// NewStatisticsService 创建统计服务
func NewStatisticsService() *StatisticsService {
	return &StatisticsService{
		submissionRepo:    repository.NewSubmissionRepository(),
		invalidLinkRepo:   repository.NewInvalidLinkRepository(),
		scheduledTaskRepo: repository.NewScheduledTaskRepository(),
	}
}

// StatisticsOverview 统计概览
type StatisticsOverview struct {
	TotalInvalidLinks      int64 `json:"total_invalid_links"`        // 总失效链接数
	TotalSubmissions       int64 `json:"total_submissions"`          // 总提交记录数
	CompletedSubmissions   int64 `json:"completed_submissions"`      // 已完成检测的记录数
	PendingSubmissions     int64 `json:"pending_submissions"`        // 待检测的记录数
	RateLimitedLinks       int64 `json:"rate_limited_links"`         // 可能被限制导致检测无效的链接数
	TotalScheduledTasks    int64 `json:"total_scheduled_tasks"`      // 总定时任务数
}

// PlatformInvalidCount 平台失效链接统计
type PlatformInvalidCount struct {
	Platform string `json:"platform"`
	Count    int64  `json:"count"`
}

// TimeSeriesData 时间序列数据
type TimeSeriesData struct {
	Date  string `json:"date"`  // 日期，格式：YYYY-MM-DD
	Count int64  `json:"count"` // 该日期的提交记录数
}

// GetOverview 获取统计概览
func (s *StatisticsService) GetOverview() (*StatisticsOverview, error) {
	var totalInvalidLinks int64
	var totalSubmissions int64
	var completedSubmissions int64
	var pendingSubmissions int64
	var rateLimitedLinks int64
	var totalScheduledTasks int64

	// 统计总失效链接数
	err := s.invalidLinkRepo.Count(&totalInvalidLinks)
	if err != nil {
		return nil, err
	}

	// 统计总提交记录数
	err = s.submissionRepo.Count(&totalSubmissions)
	if err != nil {
		return nil, err
	}

	// 统计已完成检测的记录数
	err = s.submissionRepo.CountByStatus("checked", &completedSubmissions)
	if err != nil {
		return nil, err
	}

	// 统计待检测的记录数
	err = s.submissionRepo.CountByStatus("pending", &pendingSubmissions)
	if err != nil {
		return nil, err
	}

	// 统计可能被限制导致检测无效的链接数
	err = s.invalidLinkRepo.CountByRateLimited(&rateLimitedLinks)
	if err != nil {
		return nil, err
	}

	// 统计总定时任务数
	err = s.scheduledTaskRepo.Count(&totalScheduledTasks)
	if err != nil {
		return nil, err
	}

	return &StatisticsOverview{
		TotalInvalidLinks:    totalInvalidLinks,
		TotalSubmissions:     totalSubmissions,
		CompletedSubmissions: completedSubmissions,
		PendingSubmissions:   pendingSubmissions,
		RateLimitedLinks:     rateLimitedLinks,
		TotalScheduledTasks:  totalScheduledTasks,
	}, nil
}

// GetPlatformInvalidCounts 获取各大网盘失效记录数
func (s *StatisticsService) GetPlatformInvalidCounts() ([]PlatformInvalidCount, error) {
	platforms := model.AllPlatforms()
	results := make([]PlatformInvalidCount, 0, len(platforms))

	for _, platform := range platforms {
		var count int64
		err := s.invalidLinkRepo.CountByPlatform(platform, &count)
		if err != nil {
			return nil, err
		}
		results = append(results, PlatformInvalidCount{
			Platform: string(platform),
			Count:    count,
		})
	}

	return results, nil
}

// GetSubmissionTimeSeries 获取各个时间段提交记录数
// granularity: "hour" 或 "day"
func (s *StatisticsService) GetSubmissionTimeSeries(startTime, endTime *time.Time, granularity string) ([]TimeSeriesData, error) {
	results, err := s.submissionRepo.GetTimeSeries(startTime, endTime, granularity)
	if err != nil {
		return nil, err
	}

	timeSeriesData := make([]TimeSeriesData, 0, len(results))
	for _, result := range results {
		timeSeriesData = append(timeSeriesData, TimeSeriesData{
			Date:  result.Date,
			Count: result.Count,
		})
	}

	return timeSeriesData, nil
}

