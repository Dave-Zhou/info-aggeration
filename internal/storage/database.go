package storage

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/pkg/models"
)

// DatabaseStorage 数据库存储实现
type DatabaseStorage struct {
	config config.StorageConfig
	db     *sql.DB
	driver string
}

// NewDatabaseStorage 创建数据库存储实例
func NewDatabaseStorage(cfg config.StorageConfig) (*DatabaseStorage, error) {
	storage := &DatabaseStorage{
		config: cfg,
		driver: cfg.Database.Driver,
	}

	// 连接数据库
	if err := storage.connect(); err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	// 创建表
	if err := storage.createTable(); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return storage, nil
}

// connect 连接数据库
func (ds *DatabaseStorage) connect() error {
	var dsn string
	
	switch ds.driver {
	case "sqlite3":
		dsn = "./data/crawler.db"
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			ds.config.Database.Username,
			ds.config.Database.Password,
			ds.config.Database.Host,
			ds.config.Database.Port,
			ds.config.Database.Database,
		)
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", ds.driver)
	}

	db, err := sql.Open(ds.driver, dsn)
	if err != nil {
		return err
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return err
	}

	ds.db = db
	return nil
}

// createTable 创建表
func (ds *DatabaseStorage) createTable() error {
	var createSQL string
	
	switch ds.driver {
	case "sqlite3":
		createSQL = `
		CREATE TABLE IF NOT EXISTS crawl_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			url TEXT NOT NULL,
			title TEXT,
			content TEXT,
			description TEXT,
			author TEXT,
			source TEXT,
			publish_date DATETIME,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			keywords TEXT,
			tags TEXT,
			links TEXT,
			images TEXT,
			UNIQUE(url)
		)`
	case "mysql":
		createSQL = `
		CREATE TABLE IF NOT EXISTS crawl_data (
			id INT AUTO_INCREMENT PRIMARY KEY,
			url VARCHAR(1000) NOT NULL,
			title TEXT,
			content LONGTEXT,
			description TEXT,
			author VARCHAR(255),
			source VARCHAR(255),
			publish_date DATETIME,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			keywords TEXT,
			tags TEXT,
			links TEXT,
			images TEXT,
			UNIQUE KEY unique_url (url)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", ds.driver)
	}

	_, err := ds.db.Exec(createSQL)
	return err
}

// Save 保存单个数据项
func (ds *DatabaseStorage) Save(item *models.Item) error {
	insertSQL := `
	INSERT OR REPLACE INTO crawl_data 
	(url, title, content, description, author, source, publish_date, timestamp, keywords, tags, links, images)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	if ds.driver == "mysql" {
		insertSQL = `
		INSERT INTO crawl_data 
		(url, title, content, description, author, source, publish_date, timestamp, keywords, tags, links, images)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		title = VALUES(title),
		content = VALUES(content),
		description = VALUES(description),
		author = VALUES(author),
		source = VALUES(source),
		publish_date = VALUES(publish_date),
		timestamp = VALUES(timestamp),
		keywords = VALUES(keywords),
		tags = VALUES(tags),
		links = VALUES(links),
		images = VALUES(images)`
	}

	// 处理时间字段
	var publishDate interface{}
	if item.PublishDate.IsZero() {
		publishDate = nil
	} else {
		publishDate = item.PublishDate
	}

	// 将数组转换为字符串
	keywords := strings.Join(item.Keywords, ";")
	tags := strings.Join(item.Tags, ";")
	links := strings.Join(item.Links, ";")
	images := strings.Join(item.Images, ";")

	_, err := ds.db.Exec(insertSQL,
		item.URL,
		item.Title,
		item.Content,
		item.Description,
		item.Author,
		item.Source,
		publishDate,
		item.Timestamp,
		keywords,
		tags,
		links,
		images,
	)

	return err
}

// SaveBatch 批量保存数据项
func (ds *DatabaseStorage) SaveBatch(items []*models.Item) error {
	// 开始事务
	tx, err := ds.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 准备插入语句
	insertSQL := `
	INSERT OR REPLACE INTO crawl_data 
	(url, title, content, description, author, source, publish_date, timestamp, keywords, tags, links, images)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	if ds.driver == "mysql" {
		insertSQL = `
		INSERT INTO crawl_data 
		(url, title, content, description, author, source, publish_date, timestamp, keywords, tags, links, images)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
		title = VALUES(title),
		content = VALUES(content),
		description = VALUES(description),
		author = VALUES(author),
		source = VALUES(source),
		publish_date = VALUES(publish_date),
		timestamp = VALUES(timestamp),
		keywords = VALUES(keywords),
		tags = VALUES(tags),
		links = VALUES(links),
		images = VALUES(images)`
	}

	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return fmt.Errorf("准备语句失败: %w", err)
	}
	defer stmt.Close()

	// 批量插入
	for _, item := range items {
		// 处理时间字段
		var publishDate interface{}
		if item.PublishDate.IsZero() {
			publishDate = nil
		} else {
			publishDate = item.PublishDate
		}

		// 将数组转换为字符串
		keywords := strings.Join(item.Keywords, ";")
		tags := strings.Join(item.Tags, ";")
		links := strings.Join(item.Links, ";")
		images := strings.Join(item.Images, ";")

		_, err := stmt.Exec(
			item.URL,
			item.Title,
			item.Content,
			item.Description,
			item.Author,
			item.Source,
			publishDate,
			item.Timestamp,
			keywords,
			tags,
			links,
			images,
		)
		if err != nil {
			return fmt.Errorf("执行插入失败: %w", err)
		}
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (ds *DatabaseStorage) Close() error {
	if ds.db != nil {
		return ds.db.Close()
	}
	return nil
}

// Query 查询数据
func (ds *DatabaseStorage) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return ds.db.Query(query, args...)
}

// Count 统计数据条数
func (ds *DatabaseStorage) Count() (int, error) {
	var count int
	err := ds.db.QueryRow("SELECT COUNT(*) FROM crawl_data").Scan(&count)
	return count, err
}

// GetLatest 获取最新数据
func (ds *DatabaseStorage) GetLatest(limit int) ([]*models.Item, error) {
	query := "SELECT url, title, content, description, author, source, publish_date, timestamp, keywords, tags, links, images FROM crawl_data ORDER BY timestamp DESC LIMIT ?"
	
	rows, err := ds.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.Item
	for rows.Next() {
		item := &models.Item{}
		var publishDate sql.NullTime
		var keywords, tags, links, images string

		err := rows.Scan(
			&item.URL,
			&item.Title,
			&item.Content,
			&item.Description,
			&item.Author,
			&item.Source,
			&publishDate,
			&item.Timestamp,
			&keywords,
			&tags,
			&links,
			&images,
		)
		if err != nil {
			return nil, err
		}

		// 处理时间字段
		if publishDate.Valid {
			item.PublishDate = publishDate.Time
		}

		// 将字符串转换为数组
		if keywords != "" {
			item.Keywords = strings.Split(keywords, ";")
		}
		if tags != "" {
			item.Tags = strings.Split(tags, ";")
		}
		if links != "" {
			item.Links = strings.Split(links, ";")
		}
		if images != "" {
			item.Images = strings.Split(images, ";")
		}

		items = append(items, item)
	}

	return items, nil
} 