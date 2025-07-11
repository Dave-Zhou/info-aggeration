package api

import (
	"database/sql"
	"encoding/json"
	"runtime"
	"strconv"
	"strings"
	"time"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/utils"
	"github.com/gin-gonic/gin"
)

// SystemController 系统控制器
type SystemController struct {
	db     *sql.DB
	logger utils.Logger
	config *config.Config
}

// NewSystemController 创建系统控制器
func NewSystemController(db *sql.DB, logger utils.Logger, cfg *config.Config) *SystemController {
	return &SystemController{
		db:     db,
		logger: logger,
		config: cfg,
	}
}

// GetSystemStatus 获取系统状态
func (sc *SystemController) GetSystemStatus(c *gin.Context) {
	// 获取系统信息
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 数据库连接状态
	dbStatus := "connected"
	if err := sc.db.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	// 运行中的任务数
	var runningTasks int
	sc.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'running'").Scan(&runningTasks)

	// 活跃站点数
	var activeSites int
	sc.db.QueryRow("SELECT COUNT(*) FROM sites WHERE enabled = TRUE").Scan(&activeSites)

	// 系统负载（简化版）
	cpuUsage := 0.0 // TODO: 实现CPU使用率计算
	memUsage := float64(memStats.Alloc) / float64(memStats.Sys) * 100

	status := gin.H{
		"status":        "running",
		"uptime":        time.Now().Unix(), // TODO: 实现实际运行时间
		"version":       "1.0.0",
		"go_version":    runtime.Version(),
		"goroutines":    runtime.NumGoroutine(),
		"cpu_usage":     cpuUsage,
		"memory_usage":  memUsage,
		"memory_alloc":  memStats.Alloc,
		"memory_sys":    memStats.Sys,
		"database":      dbStatus,
		"running_tasks": runningTasks,
		"active_sites":  activeSites,
		"timestamp":     time.Now(),
	}

	c.JSON(200, status)
}

// GetConfig 获取系统配置
func (sc *SystemController) GetConfig(c *gin.Context) {
	// 返回配置信息（隐藏敏感信息）
	cfg := map[string]interface{}{
		"spider": map[string]interface{}{
			"user_agent": sc.config.Spider.UserAgent,
			"concurrent": sc.config.Spider.Concurrent,
			"delay":      sc.config.Spider.Delay,
			"timeout":    sc.config.Spider.Timeout,
			"retries":    sc.config.Spider.Retries,
			"proxy_url":  sc.config.Spider.ProxyURL,
		},
		"storage": map[string]interface{}{
			"type":       sc.config.Storage.Type,
			"output_dir": sc.config.Storage.OutputDir,
			"database": map[string]interface{}{
				"driver": sc.config.Storage.Database.Driver,
				"host":   sc.config.Storage.Database.Host,
				"port":   sc.config.Storage.Database.Port,
				"name":   sc.config.Storage.Database.Database,
				// 不返回用户名和密码
			},
		},
		"logging": map[string]interface{}{
			"level": sc.config.Logging.Level,
			"file":  sc.config.Logging.File,
		},
		"web": map[string]interface{}{
			"port":        sc.config.Web.Port,
			"static_path": sc.config.Web.StaticPath,
			"api_prefix":  sc.config.Web.APIPrefix,
		},
	}

	c.JSON(200, gin.H{"config": cfg})
}

// UpdateConfig 更新系统配置
func (sc *SystemController) UpdateConfig(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 将配置保存到数据库
	configJSON, err := json.Marshal(req)
	if err != nil {
		c.JSON(400, gin.H{"error": "配置格式错误"})
		return
	}

	// 更新或插入配置
	_, err = sc.db.Exec(`
		INSERT INTO system_config (config_key, config_value, description, updated_at)
		VALUES ('system_config', ?, '系统配置', NOW())
		ON DUPLICATE KEY UPDATE config_value = ?, updated_at = NOW()
	`, configJSON, configJSON)

	if err != nil {
		sc.logger.Error("更新系统配置失败", "error", err)
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	sc.logger.Info("更新系统配置成功")
	c.JSON(200, gin.H{"message": "配置更新成功"})
}

// GetLogs 获取系统日志
func (sc *SystemController) GetLogs(c *gin.Context) {
	level := c.DefaultQuery("level", "")
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 50
	}
	// 为了安全和性能，限制每页最大数量
	if pageSize > 200 {
		pageSize = 200
	}

	offset := (page - 1) * pageSize

	// 使用 strings.Builder 安全地构建查询
	var queryBuilder strings.Builder
	args := []interface{}{}

	queryBuilder.WriteString("SELECT level, message, details, created_at FROM task_logs WHERE 1=1")

	if level != "" {
		queryBuilder.WriteString(" AND level = ?")
		args = append(args, level)
	}

	queryBuilder.WriteString(" ORDER BY created_at DESC LIMIT ? OFFSET ?")
	args = append(args, pageSize, offset)

	rows, err := sc.db.Query(queryBuilder.String(), args...)
	if err != nil {
		sc.logger.Error("查询系统日志失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var logLevel, message string
		var createdAt time.Time
		var details sql.NullString // 处理 details 可能为 NULL 的情况

		if err := rows.Scan(&logLevel, &message, &details, &createdAt); err != nil {
			sc.logger.Error("扫描日志行失败", "error", err)
			continue
		}

		logItem := map[string]interface{}{
			"level":      logLevel,
			"message":    message,
			"created_at": createdAt,
		}

		if details.Valid && details.String != "" {
			var detailsObj map[string]interface{}
			if json.Unmarshal([]byte(details.String), &detailsObj) == nil {
				logItem["details"] = detailsObj
			} else {
				logItem["details"] = details.String // 如果不是json，则作为原始字符串
			}
		} else {
			logItem["details"] = nil
		}

		logs = append(logs, logItem)
	}

	c.JSON(200, gin.H{"data": logs})
}

// CreateBackup 创建备份
func (sc *SystemController) CreateBackup(c *gin.Context) {
	backupName := "backup_" + time.Now().Format("20060102_150405")

	// TODO: 实现数据库备份逻辑
	// 1. 导出数据库
	// 2. 压缩文件
	// 3. 保存到指定位置

	sc.logger.Info("创建备份", "name", backupName)
	c.JSON(200, gin.H{
		"message": "备份创建成功",
		"name":    backupName,
	})
}

// ListBackups 获取备份列表
func (sc *SystemController) ListBackups(c *gin.Context) {
	// TODO: 实现备份文件列表获取
	backups := []map[string]interface{}{
		{
			"name":       "backup_20231201_120000",
			"size":       "10.5MB",
			"created_at": "2023-12-01 12:00:00",
		},
		{
			"name":       "backup_20231130_120000",
			"size":       "9.8MB",
			"created_at": "2023-11-30 12:00:00",
		},
	}

	c.JSON(200, gin.H{"data": backups})
}

// RestoreBackup 恢复备份
func (sc *SystemController) RestoreBackup(c *gin.Context) {
	var req struct {
		BackupName string `json:"backup_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// TODO: 实现数据库恢复逻辑
	// 1. 验证备份文件
	// 2. 停止相关服务
	// 3. 恢复数据库
	// 4. 重启服务

	sc.logger.Info("恢复备份", "name", req.BackupName)
	c.JSON(200, gin.H{
		"message": "备份恢复成功",
		"name":    req.BackupName,
	})
}
