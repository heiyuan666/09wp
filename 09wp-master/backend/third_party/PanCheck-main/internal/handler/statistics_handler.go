package handler

import (
	"PanCheck/internal/service"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatisticsHandler 统计处理器
type StatisticsHandler struct {
	statisticsService *service.StatisticsService
}

// NewStatisticsHandler 创建统计处理器
func NewStatisticsHandler() *StatisticsHandler {
	return &StatisticsHandler{
		statisticsService: service.NewStatisticsService(),
	}
}

// GetOverview 获取统计概览
func (h *StatisticsHandler) GetOverview(c *gin.Context) {
	overview, err := h.statisticsService.GetOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": overview})
}

// GetPlatformInvalidCounts 获取各大网盘失效记录数
func (h *StatisticsHandler) GetPlatformInvalidCounts(c *gin.Context) {
	counts, err := h.statisticsService.GetPlatformInvalidCounts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": counts})
}

// GetSubmissionTimeSeries 获取各个时间段提交记录数
func (h *StatisticsHandler) GetSubmissionTimeSeries(c *gin.Context) {
	var startTime, endTime *time.Time
	granularity := c.DefaultQuery("granularity", "day") // 默认按天，可选：hour, day

	// 解析开始时间
	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			startTime = &t
		}
	}

	// 解析结束时间
	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			// 设置为当天的23:59:59
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTime = &t
		}
	}

	data, err := h.statisticsService.GetSubmissionTimeSeries(startTime, endTime, granularity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": data})
}

