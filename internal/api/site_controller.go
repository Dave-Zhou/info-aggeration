package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/m/v2/internal/utils"
	"example.com/m/v2/pkg/models"
)

// SiteController 站点控制器
type SiteController struct {
	db     *sql.DB
	logger utils.Logger
}

// NewSiteController 创建站点控制器
func NewSiteController(db *sql.DB, logger utils.Logger) *SiteController {
	return &SiteController{
		db:     db,
		logger: logger,
	}
}

// SiteRequest 站点请求结构
type SiteRequest struct {
	Name        string            `json:"name" binding:"required"`
	BaseURL     string            `json:"base_url" binding:"required"`
	Description string            `json:"description"`
	StartURLs   []string          `json:"start_urls" binding:"required"`
	Selectors   map[string]string `json:"selectors" binding:"required"`
	Rules       SiteRules         `json:"rules"`
	Enabled     bool              `json:"enabled"`
}

// SiteRules 站点规则
type SiteRules struct {
	MaxDepth     int      `json:"max_depth"`
	MaxPages     int      `json:"max_pages"`
	URLPatterns  []string `json:"url_patterns"`
	ContentTypes []string `json:"content_types"`
	Concurrent   int      `json:"concurrent"`
	Delay        int      `json:"delay"`
}

// SiteResponse 站点响应结构
type SiteResponse struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	Description string            `json:"description"`
	StartURLs   []string          `json:"start_urls"`
	Selectors   map[string]string `json:"selectors"`
	Rules       SiteRules         `json:"rules"`
	Enabled     bool              `json:"enabled"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Status      string            `json:"status"`
	LastRunAt   *time.Time        `json:"last_run_at"`
}

// ListSites 获取站点列表
func (sc *SiteController) ListSites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	enabled := c.Query("enabled")

	offset := (page - 1) * pageSize

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}
	
	if enabled != "" {
		where += " AND enabled = ?"
		args = append(args, enabled == "true")
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM sites WHERE " + where
	err := sc.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		sc.logger.Error("查询站点总数失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 查询站点列表
	query := `
		SELECT id, name, base_url, description, start_urls, selectors, rules, 
		       enabled, created_at, updated_at, status, last_run_at
		FROM sites 
		WHERE ` + where + `
		ORDER BY created_at DESC 
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := sc.db.Query(query, args...)
	if err != nil {
		sc.logger.Error("查询站点列表失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var sites []SiteResponse
	for rows.Next() {
		var site SiteResponse
		var startURLsJSON, selectorsJSON, rulesJSON string
		var lastRunAt sql.NullTime

		err := rows.Scan(
			&site.ID, &site.Name, &site.BaseURL, &site.Description,
			&startURLsJSON, &selectorsJSON, &rulesJSON,
			&site.Enabled, &site.CreatedAt, &site.UpdatedAt,
			&site.Status, &lastRunAt,
		)
		if err != nil {
			sc.logger.Error("扫描站点数据失败", "error", err)
			continue
		}

		// 解析JSON字段
		json.Unmarshal([]byte(startURLsJSON), &site.StartURLs)
		json.Unmarshal([]byte(selectorsJSON), &site.Selectors)
		json.Unmarshal([]byte(rulesJSON), &site.Rules)

		if lastRunAt.Valid {
			site.LastRunAt = &lastRunAt.Time
		}

		sites = append(sites, site)
	}

	c.JSON(200, gin.H{
		"data": sites,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

// CreateSite 创建站点
func (sc *SiteController) CreateSite(c *gin.Context) {
	var req SiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 检查站点名称是否已存在
	var count int
	err := sc.db.QueryRow("SELECT COUNT(*) FROM sites WHERE name = ?", req.Name).Scan(&count)
	if err != nil {
		sc.logger.Error("检查站点名称失败", "error", err)
		c.JSON(500, gin.H{"error": "服务器错误"})
		return
	}
	if count > 0 {
		c.JSON(400, gin.H{"error": "站点名称已存在"})
		return
	}

	// 序列化JSON字段
	startURLsJSON, _ := json.Marshal(req.StartURLs)
	selectorsJSON, _ := json.Marshal(req.Selectors)
	rulesJSON, _ := json.Marshal(req.Rules)

	// 插入站点
	query := `
		INSERT INTO sites (name, base_url, description, start_urls, selectors, rules, enabled, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'ready', NOW(), NOW())
	`
	result, err := sc.db.Exec(query, req.Name, req.BaseURL, req.Description, startURLsJSON, selectorsJSON, rulesJSON, req.Enabled)
	if err != nil {
		sc.logger.Error("创建站点失败", "error", err)
		c.JSON(500, gin.H{"error": "创建失败"})
		return
	}

	id, _ := result.LastInsertId()
	sc.logger.Info("创建站点成功", "id", id, "name", req.Name)

	c.JSON(201, gin.H{
		"message": "站点创建成功",
		"id":      id,
	})
}

// GetSite 获取站点详情
func (sc *SiteController) GetSite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的站点ID"})
		return
	}

	query := `
		SELECT id, name, base_url, description, start_urls, selectors, rules, 
		       enabled, created_at, updated_at, status, last_run_at
		FROM sites WHERE id = ?
	`
	
	var site SiteResponse
	var startURLsJSON, selectorsJSON, rulesJSON string
	var lastRunAt sql.NullTime

	err = sc.db.QueryRow(query, id).Scan(
		&site.ID, &site.Name, &site.BaseURL, &site.Description,
		&startURLsJSON, &selectorsJSON, &rulesJSON,
		&site.Enabled, &site.CreatedAt, &site.UpdatedAt,
		&site.Status, &lastRunAt,
	)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "站点不存在"})
		return
	}
	if err != nil {
		sc.logger.Error("查询站点详情失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 解析JSON字段
	json.Unmarshal([]byte(startURLsJSON), &site.StartURLs)
	json.Unmarshal([]byte(selectorsJSON), &site.Selectors)
	json.Unmarshal([]byte(rulesJSON), &site.Rules)

	if lastRunAt.Valid {
		site.LastRunAt = &lastRunAt.Time
	}

	c.JSON(200, gin.H{"data": site})
}

// UpdateSite 更新站点
func (sc *SiteController) UpdateSite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的站点ID"})
		return
	}

	var req SiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	// 检查站点是否存在
	var count int
	err = sc.db.QueryRow("SELECT COUNT(*) FROM sites WHERE id = ?", id).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(404, gin.H{"error": "站点不存在"})
		return
	}

	// 序列化JSON字段
	startURLsJSON, _ := json.Marshal(req.StartURLs)
	selectorsJSON, _ := json.Marshal(req.Selectors)
	rulesJSON, _ := json.Marshal(req.Rules)

	// 更新站点
	query := `
		UPDATE sites 
		SET name=?, base_url=?, description=?, start_urls=?, selectors=?, rules=?, enabled=?, updated_at=NOW()
		WHERE id=?
	`
	_, err = sc.db.Exec(query, req.Name, req.BaseURL, req.Description, startURLsJSON, selectorsJSON, rulesJSON, req.Enabled, id)
	if err != nil {
		sc.logger.Error("更新站点失败", "error", err)
		c.JSON(500, gin.H{"error": "更新失败"})
		return
	}

	sc.logger.Info("更新站点成功", "id", id, "name", req.Name)
	c.JSON(200, gin.H{"message": "站点更新成功"})
}

// DeleteSite 删除站点
func (sc *SiteController) DeleteSite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的站点ID"})
		return
	}

	// 检查站点是否存在
	var count int
	err = sc.db.QueryRow("SELECT COUNT(*) FROM sites WHERE id = ?", id).Scan(&count)
	if err != nil || count == 0 {
		c.JSON(404, gin.H{"error": "站点不存在"})
		return
	}

	// 删除站点
	_, err = sc.db.Exec("DELETE FROM sites WHERE id = ?", id)
	if err != nil {
		sc.logger.Error("删除站点失败", "error", err)
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	sc.logger.Info("删除站点成功", "id", id)
	c.JSON(200, gin.H{"message": "站点删除成功"})
}

// TestSite 测试站点配置
func (sc *SiteController) TestSite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的站点ID"})
		return
	}

	// TODO: 实现站点测试逻辑
	// 1. 获取站点配置
	// 2. 创建测试爬虫
	// 3. 抓取第一个URL
	// 4. 返回测试结果

	c.JSON(200, gin.H{
		"message": "测试功能开发中",
		"status":  "pending",
	})
}

// ToggleSite 切换站点状态
func (sc *SiteController) ToggleSite(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的站点ID"})
		return
	}

	// 获取当前状态
	var enabled bool
	err = sc.db.QueryRow("SELECT enabled FROM sites WHERE id = ?", id).Scan(&enabled)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "站点不存在"})
		return
	}
	if err != nil {
		sc.logger.Error("查询站点状态失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 切换状态
	newEnabled := !enabled
	_, err = sc.db.Exec("UPDATE sites SET enabled = ?, updated_at = NOW() WHERE id = ?", newEnabled, id)
	if err != nil {
		sc.logger.Error("切换站点状态失败", "error", err)
		c.JSON(500, gin.H{"error": "操作失败"})
		return
	}

	status := "禁用"
	if newEnabled {
		status = "启用"
	}

	sc.logger.Info("切换站点状态成功", "id", id, "enabled", newEnabled)
	c.JSON(200, gin.H{
		"message": "站点状态已" + status,
		"enabled": newEnabled,
	})
} 