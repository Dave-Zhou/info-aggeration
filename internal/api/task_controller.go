package api

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/utils"
)

// TaskController 任务控制器
type TaskController struct {
	db     *sql.DB
	logger utils.Logger
	config *config.Config
}

// NewTaskController 创建任务控制器
func NewTaskController(db *sql.DB, logger utils.Logger, cfg *config.Config) *TaskController {
	return &TaskController{
		db:     db,
		logger: logger,
		config: cfg,
	}
}

// TaskRequest 任务请求结构
type TaskRequest struct {
	Name   string                 `json:"name" binding:"required"`
	SiteID int                    `json:"site_id" binding:"required"`
	Config map[string]interface{} `json:"config"`
}

// TaskResponse 任务响应结构
type TaskResponse struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	SiteID        int                    `json:"site_id"`
	SiteName      string                 `json:"site_name"`
	Status        string                 `json:"status"`
	Config        map[string]interface{} `json:"config"`
	StartTime     *time.Time             `json:"start_time"`
	EndTime       *time.Time             `json:"end_time"`
	TotalURLs     int                    `json:"total_urls"`
	ProcessedURLs int                    `json:"processed_urls"`
	SuccessURLs   int                    `json:"success_urls"`
	FailedURLs    int                    `json:"failed_urls"`
	ItemsCount    int                    `json:"items_count"`
	ErrorMessage  string                 `json:"error_message"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Duration      int64                  `json:"duration"` // 持续时间（秒）
	Progress      float64                `json:"progress"` // 进度百分比
}

// ListTasks 获取任务列表
func (tc *TaskController) ListTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")
	siteID := c.Query("site_id")

	offset := (page - 1) * pageSize

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}

	if status != "" {
		where += " AND t.status = ?"
		args = append(args, status)
	}
	if siteID != "" {
		where += " AND t.site_id = ?"
		args = append(args, siteID)
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM tasks t WHERE " + where
	err := tc.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		tc.logger.Error("查询任务总数失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 查询任务列表
	query := `
		SELECT t.id, t.name, t.site_id, s.name as site_name, t.status, t.config,
		       t.start_time, t.end_time, t.total_urls, t.processed_urls, 
		       t.success_urls, t.failed_urls, t.items_count, t.error_message,
		       t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN sites s ON t.site_id = s.id
		WHERE ` + where + `
		ORDER BY t.created_at DESC 
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := tc.db.Query(query, args...)
	if err != nil {
		tc.logger.Error("查询任务列表失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var tasks []TaskResponse
	for rows.Next() {
		var task TaskResponse
		var configJSON string
		var startTime, endTime sql.NullTime

		err := rows.Scan(
			&task.ID, &task.Name, &task.SiteID, &task.SiteName, &task.Status,
			&configJSON, &startTime, &endTime, &task.TotalURLs,
			&task.ProcessedURLs, &task.SuccessURLs, &task.FailedURLs,
			&task.ItemsCount, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
		)
		if err != nil {
			tc.logger.Error("扫描任务数据失败", "error", err)
			continue
		}

		// 解析JSON字段
		if configJSON != "" {
			json.Unmarshal([]byte(configJSON), &task.Config)
		}

		if startTime.Valid {
			task.StartTime = &startTime.Time
		}
		if endTime.Valid {
			task.EndTime = &endTime.Time
		}

		// 计算持续时间和进度
		if task.StartTime != nil {
			if task.EndTime != nil {
				task.Duration = int64(task.EndTime.Sub(*task.StartTime).Seconds())
			} else {
				task.Duration = int64(time.Since(*task.StartTime).Seconds())
			}
		}

		if task.TotalURLs > 0 {
			task.Progress = float64(task.ProcessedURLs) / float64(task.TotalURLs) * 100
		}

		tasks = append(tasks, task)
	}

	c.JSON(200, gin.H{
		"data": tasks,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

// CreateTask 创建任务
func (tc *TaskController) CreateTask(c *gin.Context) {
	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 检查站点是否存在
	var count int
	err := tc.db.QueryRow("SELECT COUNT(*) FROM sites WHERE id = ? AND enabled = TRUE", req.SiteID).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(400, gin.H{"error": "站点不存在或已禁用"})
		return
	}

	// 序列化配置
	configJSON, _ := json.Marshal(req.Config)

	// 插入任务
	query := `
		INSERT INTO tasks (name, site_id, config, status, created_at, updated_at)
		VALUES (?, ?, ?, 'pending', NOW(), NOW())
	`
	result, err := tc.db.Exec(query, req.Name, req.SiteID, configJSON)
	if err != nil {
		tc.logger.Error("创建任务失败", "error", err)
		c.JSON(500, gin.H{"error": "创建失败"})
		return
	}

	id, _ := result.LastInsertId()
	tc.logger.Info("创建任务成功", "id", id, "name", req.Name)

	c.JSON(201, gin.H{
		"message": "任务创建成功",
		"id":      id,
	})
}

// GetTask 获取任务详情
func (tc *TaskController) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	query := `
		SELECT t.id, t.name, t.site_id, s.name as site_name, t.status, t.config,
		       t.start_time, t.end_time, t.total_urls, t.processed_urls, 
		       t.success_urls, t.failed_urls, t.items_count, t.error_message,
		       t.created_at, t.updated_at
		FROM tasks t
		LEFT JOIN sites s ON t.site_id = s.id
		WHERE t.id = ?
	`

	var task TaskResponse
	var configJSON string
	var startTime, endTime sql.NullTime

	err = tc.db.QueryRow(query, id).Scan(
		&task.ID, &task.Name, &task.SiteID, &task.SiteName, &task.Status,
		&configJSON, &startTime, &endTime, &task.TotalURLs,
		&task.ProcessedURLs, &task.SuccessURLs, &task.FailedURLs,
		&task.ItemsCount, &task.ErrorMessage, &task.CreatedAt, &task.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if err != nil {
		tc.logger.Error("查询任务详情失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 解析JSON字段
	if configJSON != "" {
		json.Unmarshal([]byte(configJSON), &task.Config)
	}

	if startTime.Valid {
		task.StartTime = &startTime.Time
	}
	if endTime.Valid {
		task.EndTime = &endTime.Time
	}

	// 计算持续时间和进度
	if task.StartTime != nil {
		if task.EndTime != nil {
			task.Duration = int64(task.EndTime.Sub(*task.StartTime).Seconds())
		} else {
			task.Duration = int64(time.Since(*task.StartTime).Seconds())
		}
	}

	if task.TotalURLs > 0 {
		task.Progress = float64(task.ProcessedURLs) / float64(task.TotalURLs) * 100
	}

	c.JSON(200, gin.H{"data": task})
}

// UpdateTask 更新任务
func (tc *TaskController) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	var req TaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 检查任务是否存在且未运行
	var status string
	err = tc.db.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if status == "running" {
		c.JSON(400, gin.H{"error": "运行中的任务无法修改"})
		return
	}

	// 序列化配置
	configJSON, _ := json.Marshal(req.Config)

	// 更新任务
	query := `
		UPDATE tasks 
		SET name=?, site_id=?, config=?, updated_at=NOW()
		WHERE id=?
	`
	_, err = tc.db.Exec(query, req.Name, req.SiteID, configJSON, id)
	if err != nil {
		tc.logger.Error("更新任务失败", "error", err)
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	tc.logger.Info("更新任务成功", "id", id, "name", req.Name)
	c.JSON(200, gin.H{"message": "任务更新成功"})
}

// DeleteTask 删除任务
func (tc *TaskController) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	// 检查任务是否存在且未运行
	var status string
	err = tc.db.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if status == "running" {
		c.JSON(400, gin.H{"error": "运行中的任务无法删除"})
		return
	}

	// 删除任务
	_, err = tc.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		tc.logger.Error("删除任务失败", "error", err)
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	tc.logger.Info("删除任务成功", "id", id)
	c.JSON(200, gin.H{"message": "任务删除成功"})
}

// StartTask 启动任务
func (tc *TaskController) StartTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	// 检查任务状态
	var status string
	err = tc.db.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if status == "running" {
		c.JSON(400, gin.H{"error": "任务已在运行中"})
		return
	}

	// 更新任务状态
	_, err = tc.db.Exec("UPDATE tasks SET status='running', start_time=NOW(), updated_at=NOW() WHERE id=?", id)
	if err != nil {
		tc.logger.Error("启动任务失败", "error", err)
		c.JSON(500, gin.H{"error": "启动失败"})
		return
	}

	// TODO: 实际启动爬虫逻辑
	// 这里应该异步启动爬虫任务

	tc.logger.Info("启动任务成功", "id", id)
	c.JSON(200, gin.H{"message": "任务启动成功"})
}

// StopTask 停止任务
func (tc *TaskController) StopTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	// 检查任务状态
	var status string
	err = tc.db.QueryRow("SELECT status FROM tasks WHERE id = ?", id).Scan(&status)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if status != "running" {
		c.JSON(400, gin.H{"error": "任务未在运行中"})
		return
	}

	// 更新任务状态
	_, err = tc.db.Exec("UPDATE tasks SET status='stopped', end_time=NOW(), updated_at=NOW() WHERE id=?", id)
	if err != nil {
		tc.logger.Error("停止任务失败", "error", err)
		c.JSON(500, gin.H{"error": "停止失败"})
		return
	}

	// TODO: 实际停止爬虫逻辑

	tc.logger.Info("停止任务成功", "id", id)
	c.JSON(200, gin.H{"message": "任务停止成功"})
}

// GetTaskLogs 获取任务日志
func (tc *TaskController) GetTaskLogs(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	level := c.Query("level")

	offset := (page - 1) * pageSize

	// 构建查询条件
	where := "task_id = ?"
	args := []interface{}{id}

	if level != "" {
		where += " AND level = ?"
		args = append(args, level)
	}

	// 查询日志
	query := `
		SELECT id, level, message, details, created_at
		FROM task_logs 
		WHERE ` + where + `
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := tc.db.Query(query, args...)
	if err != nil {
		tc.logger.Error("查询任务日志失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var logID int
		var logLevel, message, details string
		var createdAt time.Time

		err := rows.Scan(&logID, &logLevel, &message, &details, &createdAt)
		if err != nil {
			continue
		}

		logItem := map[string]interface{}{
			"id":         logID,
			"level":      logLevel,
			"message":    message,
			"created_at": createdAt,
		}

		if details != "" {
			var detailsObj map[string]interface{}
			json.Unmarshal([]byte(details), &detailsObj)
			logItem["details"] = detailsObj
		}

		logs = append(logs, logItem)
	}

	c.JSON(200, gin.H{"data": logs})
}

// GetTaskStatus 获取任务状态
func (tc *TaskController) GetTaskStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的任务ID"})
		return
	}

	query := `
		SELECT status, total_urls, processed_urls, success_urls, failed_urls, items_count
		FROM tasks WHERE id = ?
	`

	var status string
	var totalURLs, processedURLs, successURLs, failedURLs, itemsCount int

	err = tc.db.QueryRow(query, id).Scan(&status, &totalURLs, &processedURLs, &successURLs, &failedURLs, &itemsCount)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "任务不存在"})
		return
	}
	if err != nil {
		tc.logger.Error("查询任务状态失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	progress := 0.0
	if totalURLs > 0 {
		progress = float64(processedURLs) / float64(totalURLs) * 100
	}

	c.JSON(200, gin.H{
		"status":         status,
		"total_urls":     totalURLs,
		"processed_urls": processedURLs,
		"success_urls":   successURLs,
		"failed_urls":    failedURLs,
		"items_count":    itemsCount,
		"progress":       progress,
	})
} 