package handler

import (
	"PanCheck/internal/model"
	"PanCheck/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ScheduledTaskHandler 任务计划处理器
type ScheduledTaskHandler struct {
	taskService      *service.ScheduledTaskService
	schedulerService *service.TaskSchedulerService
	scheduler        interface {
		ReloadTask(taskID uint) error
	}
}

// NewScheduledTaskHandler 创建任务计划处理器
func NewScheduledTaskHandler(taskService *service.ScheduledTaskService, schedulerService *service.TaskSchedulerService, scheduler interface {
	ReloadTask(taskID uint) error
}) *ScheduledTaskHandler {
	return &ScheduledTaskHandler{
		taskService:      taskService,
		schedulerService: schedulerService,
		scheduler:        scheduler,
	}
}

// ListTasks 获取任务列表
func (h *ScheduledTaskHandler) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 解析标签筛选
	var tags []string
	if tagsParam := c.Query("tags"); tagsParam != "" {
		// 假设标签以逗号分隔
		tagList := c.QueryArray("tags")
		tags = tagList
	}

	status := c.Query("status")

	tasks, total, err := h.taskService.ListTasks(page, pageSize, tags, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      tasks,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetTask 获取任务详情
func (h *ScheduledTaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// parseDateTimeLocal 解析 datetime-local 格式的时间字符串
// 支持格式: "2006-01-02T15:04" 或 "2006-01-02T15:04:05" 或 RFC3339 格式
func parseDateTimeLocal(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	// 尝试多种格式
	formats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05",       // 带秒，无时区
		"2006-01-02T15:04",          // datetime-local 格式
		"2006-01-02 15:04:05",       // 空格分隔
		"2006-01-02 15:04",          // 空格分隔，无秒
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			// 如果没有时区信息，使用本地时区
			if format == "2006-01-02T15:04:05" || format == "2006-01-02T15:04" ||
				format == "2006-01-02 15:04:05" || format == "2006-01-02 15:04" {
				// 使用本地时区
				loc, _ := time.LoadLocation("Local")
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), loc)
			}
			return &t, nil
		}
	}

	return nil, &time.ParseError{
		Layout:     "2006-01-02T15:04",
		Value:      timeStr,
		LayoutElem: "",
		ValueElem:  "",
		Message:    "unable to parse time",
	}
}

// CreateTask 创建任务
func (h *ScheduledTaskHandler) CreateTask(c *gin.Context) {
	// 使用 map 接收 JSON，以便手动处理时间字段
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建任务对象
	var task model.ScheduledTask

	// 处理字符串字段
	if name, ok := req["name"].(string); ok {
		task.Name = name
	}
	if desc, ok := req["description"].(string); ok {
		task.Description = desc
	}
	if curlCmd, ok := req["curl_command"].(string); ok {
		task.CurlCommand = curlCmd
	}
	if transformScript, ok := req["transform_script"].(string); ok {
		task.TransformScript = transformScript
	}
	if cronExpr, ok := req["cron_expression"].(string); ok {
		task.CronExpression = cronExpr
	}
	if status, ok := req["status"].(string); ok {
		task.Status = status
	}

	// 处理标签
	if tags, ok := req["tags"].([]interface{}); ok {
		task.Tags = make(model.StringArray, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	// 处理 auto_destroy_at 时间字段
	if autoDestroyAt, ok := req["auto_destroy_at"].(string); ok && autoDestroyAt != "" {
		parsedTime, err := parseDateTimeLocal(autoDestroyAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auto_destroy_at format: " + err.Error()})
			return
		}
		task.AutoDestroyAt = parsedTime
	}

	// 验证必填字段
	if task.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if task.CurlCommand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "curl_command is required"})
		return
	}
	if task.CronExpression == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cron_expression is required"})
		return
	}

	if err := h.taskService.CreateTask(&task); err != nil {
		// 如果是名称重复错误，返回 409 Conflict
		if strings.Contains(err.Error(), "已存在") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果任务创建时就是active状态，需要重新加载调度器
	if task.Status == "active" {
		if h.scheduler != nil {
			_ = h.scheduler.ReloadTask(task.ID)
		}
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask 更新任务
func (h *ScheduledTaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	// 使用 map 接收 JSON，以便手动处理时间字段
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 构建任务对象
	var task model.ScheduledTask
	task.ID = uint(id)

	// 处理字符串字段
	if name, ok := req["name"].(string); ok {
		task.Name = name
	}
	if desc, ok := req["description"].(string); ok {
		task.Description = desc
	}
	if curlCmd, ok := req["curl_command"].(string); ok {
		task.CurlCommand = curlCmd
	}
	if transformScript, ok := req["transform_script"].(string); ok {
		task.TransformScript = transformScript
	}
	if cronExpr, ok := req["cron_expression"].(string); ok {
		task.CronExpression = cronExpr
	}
	if status, ok := req["status"].(string); ok {
		task.Status = status
	}

	// 处理标签
	if tags, ok := req["tags"].([]interface{}); ok {
		task.Tags = make(model.StringArray, 0, len(tags))
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				task.Tags = append(task.Tags, tagStr)
			}
		}
	}

	// 处理 auto_destroy_at 时间字段
	if autoDestroyAt, ok := req["auto_destroy_at"].(string); ok {
		if autoDestroyAt == "" {
			task.AutoDestroyAt = nil
		} else {
			parsedTime, err := parseDateTimeLocal(autoDestroyAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid auto_destroy_at format: " + err.Error()})
				return
			}
			task.AutoDestroyAt = parsedTime
		}
	}
	if err := h.taskService.UpdateTask(&task); err != nil {
		// 如果是名称重复错误，返回 409 Conflict
		if strings.Contains(err.Error(), "已存在") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载任务到调度器
	if h.scheduler != nil {
		_ = h.scheduler.ReloadTask(task.ID)
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTask 删除任务
func (h *ScheduledTaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

// TestTaskConfig 测试任务配置
func (h *ScheduledTaskHandler) TestTaskConfig(c *gin.Context) {
	// 支持两种方式：通过任务ID或直接提供配置
	var curlCommand, transformScript string

	if idParam := c.Param("id"); idParam != "" && idParam != "test" {
		// 通过任务ID测试
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
			return
		}

		task, err := h.taskService.GetTask(uint(id))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		curlCommand = task.CurlCommand
		transformScript = task.TransformScript
	} else {
		// 直接提供配置测试
		var req struct {
			CurlCommand     string `json:"curl_command" binding:"required"`
			TransformScript string `json:"transform_script"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		curlCommand = req.CurlCommand
		transformScript = req.TransformScript
	}

	// 执行测试
	links, err := h.taskService.TestTaskConfig(curlCommand, transformScript)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"links": links,
		"count": len(links),
	})
}

// RunTask 手动触发执行任务
func (h *ScheduledTaskHandler) RunTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	task, err := h.taskService.GetTask(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	// 异步执行任务
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Task %d execution panic recovered: %v", task.ID, r)
			}
		}()
		log.Printf("Starting manual execution of task %d: %s", task.ID, task.Name)
		if err := h.schedulerService.ExecuteTask(task); err != nil {
			log.Printf("Task %d execution failed: %v", task.ID, err)
		} else {
			log.Printf("Task %d execution completed successfully", task.ID)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "task execution started"})
}

// EnableTask 启用任务
func (h *ScheduledTaskHandler) EnableTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.EnableTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载任务到调度器
	if h.scheduler != nil {
		_ = h.scheduler.ReloadTask(uint(id))
	}

	c.JSON(http.StatusOK, gin.H{"message": "task enabled"})
}

// DisableTask 禁用任务
func (h *ScheduledTaskHandler) DisableTask(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	if err := h.taskService.DisableTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 重新加载任务到调度器
	if h.scheduler != nil {
		_ = h.scheduler.ReloadTask(uint(id))
	}

	c.JSON(http.StatusOK, gin.H{"message": "task disabled"})
}

// GetTaskExecutions 获取任务执行历史
func (h *ScheduledTaskHandler) GetTaskExecutions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	executions, total, err := h.taskService.GetTaskExecutions(uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      executions,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetAllTags 获取所有标签列表
func (h *ScheduledTaskHandler) GetAllTags(c *gin.Context) {
	tags, err := h.taskService.GetAllTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
