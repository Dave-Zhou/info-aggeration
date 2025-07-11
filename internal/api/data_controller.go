package api

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"example.com/m/v2/internal/utils"
)

// DataController 数据控制器
type DataController struct {
	db     *sql.DB
	logger utils.Logger
}

// NewDataController 创建数据控制器
func NewDataController(db *sql.DB, logger utils.Logger) *DataController {
	return &DataController{
		db:     db,
		logger: logger,
	}
}

// ItemResponse 数据项响应结构
type ItemResponse struct {
	ID           int                    `json:"id"`
	TaskID       *int                   `json:"task_id"`
	SiteID       int                    `json:"site_id"`
	SiteName     string                 `json:"site_name"`
	URL          string                 `json:"url"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Source       string                 `json:"source"`
	Language     string                 `json:"language"`
	PublishDate  *time.Time             `json:"publish_date"`
	Keywords     []string               `json:"keywords"`
	Tags         []string               `json:"tags"`
	Links        []string               `json:"links"`
	Images       []string               `json:"images"`
	Videos       []string               `json:"videos"`
	Metadata     map[string]interface{} `json:"metadata"`
	ViewCount    int                    `json:"view_count"`
	CommentCount int                    `json:"comment_count"`
	LikeCount    int                    `json:"like_count"`
	ShareCount   int                    `json:"share_count"`
	Status       string                 `json:"status"`
	CrawlTime    time.Time              `json:"crawl_time"`
}

// ListItems 获取数据项列表
func (dc *DataController) ListItems(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	siteID := c.Query("site_id")
	status := c.Query("status")
	keyword := c.Query("keyword")

	offset := (page - 1) * pageSize

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}

	if siteID != "" {
		where += " AND cd.site_id = ?"
		args = append(args, siteID)
	}
	if status != "" {
		where += " AND cd.status = ?"
		args = append(args, status)
	}
	if keyword != "" {
		where += " AND (cd.title LIKE ? OR cd.content LIKE ?)"
		args = append(args, "%"+keyword+"%", "%"+keyword+"%")
	}

	// 查询总数
	var total int
	countQuery := `SELECT COUNT(*) FROM crawl_data cd WHERE ` + where
	err := dc.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		dc.logger.Error("查询数据总数失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	// 查询数据列表
	query := `
		SELECT cd.id, cd.task_id, cd.site_id, s.name as site_name, cd.url, cd.title, 
		       cd.description, cd.author, cd.source, cd.language, cd.publish_date,
		       cd.keywords, cd.tags, cd.view_count, cd.comment_count, cd.like_count, 
		       cd.share_count, cd.status, cd.crawl_time
		FROM crawl_data cd
		LEFT JOIN sites s ON cd.site_id = s.id
		WHERE ` + where + `
		ORDER BY cd.crawl_time DESC 
		LIMIT ? OFFSET ?
	`
	args = append(args, pageSize, offset)

	rows, err := dc.db.Query(query, args...)
	if err != nil {
		dc.logger.Error("查询数据列表失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var items []ItemResponse
	for rows.Next() {
		var item ItemResponse
		var taskID sql.NullInt64
		var publishDate sql.NullTime
		var keywordsJSON, tagsJSON string

		err := rows.Scan(
			&item.ID, &taskID, &item.SiteID, &item.SiteName, &item.URL, &item.Title,
			&item.Description, &item.Author, &item.Source, &item.Language, &publishDate,
			&keywordsJSON, &tagsJSON, &item.ViewCount, &item.CommentCount, &item.LikeCount,
			&item.ShareCount, &item.Status, &item.CrawlTime,
		)
		if err != nil {
			dc.logger.Error("扫描数据失败", "error", err)
			continue
		}

		if taskID.Valid {
			id := int(taskID.Int64)
			item.TaskID = &id
		}
		if publishDate.Valid {
			item.PublishDate = &publishDate.Time
		}

		// 解析JSON字段
		if keywordsJSON != "" && keywordsJSON != "null" {
			json.Unmarshal([]byte(keywordsJSON), &item.Keywords)
		}
		if tagsJSON != "" && tagsJSON != "null" {
			json.Unmarshal([]byte(tagsJSON), &item.Tags)
		}

		items = append(items, item)
	}

	c.JSON(200, gin.H{
		"data": items,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + pageSize - 1) / pageSize,
		},
	})
}

// GetItem 获取数据项详情
func (dc *DataController) GetItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的数据ID"})
		return
	}

	query := `
		SELECT cd.id, cd.task_id, cd.site_id, s.name as site_name, cd.url, cd.title, 
		       cd.content, cd.description, cd.author, cd.source, cd.language, cd.publish_date,
		       cd.keywords, cd.tags, cd.links, cd.images, cd.videos, cd.metadata,
		       cd.view_count, cd.comment_count, cd.like_count, cd.share_count, 
		       cd.status, cd.crawl_time
		FROM crawl_data cd
		LEFT JOIN sites s ON cd.site_id = s.id
		WHERE cd.id = ?
	`

	var item ItemResponse
	var taskID sql.NullInt64
	var publishDate sql.NullTime
	var keywordsJSON, tagsJSON, linksJSON, imagesJSON, videosJSON, metadataJSON string

	err = dc.db.QueryRow(query, id).Scan(
		&item.ID, &taskID, &item.SiteID, &item.SiteName, &item.URL, &item.Title,
		&item.Content, &item.Description, &item.Author, &item.Source, &item.Language, &publishDate,
		&keywordsJSON, &tagsJSON, &linksJSON, &imagesJSON, &videosJSON, &metadataJSON,
		&item.ViewCount, &item.CommentCount, &item.LikeCount, &item.ShareCount,
		&item.Status, &item.CrawlTime,
	)
	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "数据不存在"})
		return
	}
	if err != nil {
		dc.logger.Error("查询数据详情失败", "error", err)
		c.JSON(500, gin.H{"error": "查询失败"})
		return
	}

	if taskID.Valid {
		id := int(taskID.Int64)
		item.TaskID = &id
	}
	if publishDate.Valid {
		item.PublishDate = &publishDate.Time
	}

	// 解析JSON字段
	if keywordsJSON != "" && keywordsJSON != "null" {
		json.Unmarshal([]byte(keywordsJSON), &item.Keywords)
	}
	if tagsJSON != "" && tagsJSON != "null" {
		json.Unmarshal([]byte(tagsJSON), &item.Tags)
	}
	if linksJSON != "" && linksJSON != "null" {
		json.Unmarshal([]byte(linksJSON), &item.Links)
	}
	if imagesJSON != "" && imagesJSON != "null" {
		json.Unmarshal([]byte(imagesJSON), &item.Images)
	}
	if videosJSON != "" && videosJSON != "null" {
		json.Unmarshal([]byte(videosJSON), &item.Videos)
	}
	if metadataJSON != "" && metadataJSON != "null" {
		json.Unmarshal([]byte(metadataJSON), &item.Metadata)
	}

	c.JSON(200, gin.H{"data": item})
}

// DeleteItem 删除数据项
func (dc *DataController) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的数据ID"})
		return
	}

	// 删除数据
	result, err := dc.db.Exec("DELETE FROM crawl_data WHERE id = ?", id)
	if err != nil {
		dc.logger.Error("删除数据失败", "error", err)
		c.JSON(500, gin.H{"error": "删除失败"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "数据不存在"})
		return
	}

	dc.logger.Info("删除数据成功", "id", id)
	c.JSON(200, gin.H{"message": "数据删除成功"})
}

// SearchItems 搜索数据项
func (dc *DataController) SearchItems(c *gin.Context) {
	var req struct {
		Keyword   string `json:"keyword"`
		SiteID    int    `json:"site_id"`
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
		Page      int    `json:"page"`
		PageSize  int    `json:"page_size"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误", "details": err.Error()})
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}

	if req.Keyword != "" {
		where += " AND MATCH(cd.title, cd.content) AGAINST(? IN BOOLEAN MODE)"
		args = append(args, req.Keyword)
	}
	if req.SiteID > 0 {
		where += " AND cd.site_id = ?"
		args = append(args, req.SiteID)
	}
	if req.StartDate != "" {
		where += " AND cd.crawl_time >= ?"
		args = append(args, req.StartDate)
	}
	if req.EndDate != "" {
		where += " AND cd.crawl_time <= ?"
		args = append(args, req.EndDate)
	}

	// 查询总数
	var total int
	countQuery := `SELECT COUNT(*) FROM crawl_data cd WHERE ` + where
	err := dc.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		dc.logger.Error("搜索数据总数失败", "error", err)
		c.JSON(500, gin.H{"error": "搜索失败"})
		return
	}

	// 查询数据
	query := `
		SELECT cd.id, cd.site_id, s.name as site_name, cd.url, cd.title, 
		       cd.description, cd.author, cd.source, cd.crawl_time
		FROM crawl_data cd
		LEFT JOIN sites s ON cd.site_id = s.id
		WHERE ` + where + `
		ORDER BY cd.crawl_time DESC 
		LIMIT ? OFFSET ?
	`
	args = append(args, req.PageSize, offset)

	rows, err := dc.db.Query(query, args...)
	if err != nil {
		dc.logger.Error("搜索数据失败", "error", err)
		c.JSON(500, gin.H{"error": "搜索失败"})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, siteID int
		var siteName, url, title, description, author, source string
		var crawlTime time.Time

		err := rows.Scan(&id, &siteID, &siteName, &url, &title, &description, &author, &source, &crawlTime)
		if err != nil {
			continue
		}

		items = append(items, map[string]interface{}{
			"id":          id,
			"site_id":     siteID,
			"site_name":   siteName,
			"url":         url,
			"title":       title,
			"description": description,
			"author":      author,
			"source":      source,
			"crawl_time":  crawlTime,
		})
	}

	c.JSON(200, gin.H{
		"data": items,
		"pagination": gin.H{
			"page":       req.Page,
			"page_size":  req.PageSize,
			"total":      total,
			"total_page": (total + req.PageSize - 1) / req.PageSize,
		},
	})
}

// ExportItems 导出数据
func (dc *DataController) ExportItems(c *gin.Context) {
	format := c.DefaultQuery("format", "json")
	siteID := c.Query("site_id")

	// 构建查询条件
	where := "1=1"
	args := []interface{}{}

	if siteID != "" {
		where += " AND site_id = ?"
		args = append(args, siteID)
	}

	// 查询数据
	query := `
		SELECT url, title, content, description, author, source, publish_date, crawl_time
		FROM crawl_data 
		WHERE ` + where + `
		ORDER BY crawl_time DESC
		LIMIT 1000
	`

	rows, err := dc.db.Query(query, args...)
	if err != nil {
		dc.logger.Error("导出数据查询失败", "error", err)
		c.JSON(500, gin.H{"error": "导出失败"})
		return
	}
	defer rows.Close()

	var data []map[string]interface{}
	for rows.Next() {
		var url, title, content, description, author, source string
		var publishDate sql.NullTime
		var crawlTime time.Time

		err := rows.Scan(&url, &title, &content, &description, &author, &source, &publishDate, &crawlTime)
		if err != nil {
			continue
		}

		item := map[string]interface{}{
			"url":         url,
			"title":       title,
			"content":     content,
			"description": description,
			"author":      author,
			"source":      source,
			"crawl_time":  crawlTime,
		}

		if publishDate.Valid {
			item["publish_date"] = publishDate.Time
		}

		data = append(data, item)
	}

	// 根据格式返回数据
	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment; filename=crawl_data.csv")
		// TODO: 实现CSV格式导出
		c.JSON(200, gin.H{"message": "CSV导出功能开发中"})
	case "excel":
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", "attachment; filename=crawl_data.xlsx")
		// TODO: 实现Excel格式导出
		c.JSON(200, gin.H{"message": "Excel导出功能开发中"})
	default:
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", "attachment; filename=crawl_data.json")
		c.JSON(200, gin.H{"data": data})
	}
}

// GetStatistics 获取统计信息
func (dc *DataController) GetStatistics(c *gin.Context) {
	// 总数统计
	var totalItems, totalSites, totalTasks int
	dc.db.QueryRow("SELECT COUNT(*) FROM crawl_data").Scan(&totalItems)
	dc.db.QueryRow("SELECT COUNT(*) FROM sites").Scan(&totalSites)
	dc.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&totalTasks)

	// 今日数据
	var todayItems int
	dc.db.QueryRow("SELECT COUNT(*) FROM crawl_data WHERE DATE(crawl_time) = CURDATE()").Scan(&todayItems)

	// 按站点统计
	siteStatsQuery := `
		SELECT s.name, COUNT(cd.id) as count
		FROM sites s
		LEFT JOIN crawl_data cd ON s.id = cd.site_id
		GROUP BY s.id, s.name
		ORDER BY count DESC
		LIMIT 10
	`
	siteRows, _ := dc.db.Query(siteStatsQuery)
	defer siteRows.Close()

	var siteStats []map[string]interface{}
	for siteRows.Next() {
		var siteName string
		var count int
		siteRows.Scan(&siteName, &count)
		siteStats = append(siteStats, map[string]interface{}{
			"name":  siteName,
			"count": count,
		})
	}

	// 最近7天统计
	recentStatsQuery := `
		SELECT DATE(crawl_time) as date, COUNT(*) as count
		FROM crawl_data 
		WHERE crawl_time >= DATE_SUB(CURDATE(), INTERVAL 7 DAY)
		GROUP BY DATE(crawl_time)
		ORDER BY date DESC
	`
	recentRows, _ := dc.db.Query(recentStatsQuery)
	defer recentRows.Close()

	var recentStats []map[string]interface{}
	for recentRows.Next() {
		var date string
		var count int
		recentRows.Scan(&date, &count)
		recentStats = append(recentStats, map[string]interface{}{
			"date":  date,
			"count": count,
		})
	}

	c.JSON(200, gin.H{
		"total_items": totalItems,
		"total_sites": totalSites,
		"total_tasks": totalTasks,
		"today_items": todayItems,
		"site_stats":  siteStats,
		"recent_stats": recentStats,
	})
} 