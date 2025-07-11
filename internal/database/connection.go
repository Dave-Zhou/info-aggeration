package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"example.com/m/v2/internal/config"
)

// NewConnection 创建数据库连接
func NewConnection(cfg config.DBConfig) (*sql.DB, error) {
	var dsn string
	
	switch cfg.Driver {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.Username,
			cfg.Password,
			cfg.Host,
			cfg.Port,
			cfg.Database,
		)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", cfg.Driver)
	}

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("打开数据库连接失败: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return db, nil
}

// InitTables 初始化数据库表
func InitTables(db *sql.DB) error {
	tables := []string{
		createSitesTable,
		createTasksTable,
		createCrawlDataTable,
		createTaskLogsTable,
		createSystemConfigTable,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("创建表失败: %w", err)
		}
	}

	return nil
}

// 站点表
const createSitesTable = `
CREATE TABLE IF NOT EXISTS sites (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE COMMENT '站点名称',
    base_url VARCHAR(1000) NOT NULL COMMENT '基础URL',
    description TEXT COMMENT '站点描述',
    start_urls JSON NOT NULL COMMENT '起始URL列表',
    selectors JSON NOT NULL COMMENT 'CSS选择器配置',
    rules JSON COMMENT '爬取规则',
    enabled BOOLEAN DEFAULT TRUE COMMENT '是否启用',
    status ENUM('ready', 'running', 'stopped', 'error') DEFAULT 'ready' COMMENT '状态',
    last_run_at DATETIME NULL COMMENT '最后运行时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_name (name),
    INDEX idx_status (status),
    INDEX idx_enabled (enabled),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='爬虫站点配置表';
`

// 任务表
const createTasksTable = `
CREATE TABLE IF NOT EXISTS tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL COMMENT '任务名称',
    site_id INT NOT NULL COMMENT '站点ID',
    status ENUM('pending', 'running', 'completed', 'failed', 'stopped') DEFAULT 'pending' COMMENT '任务状态',
    config JSON COMMENT '任务配置',
    start_time DATETIME NULL COMMENT '开始时间',
    end_time DATETIME NULL COMMENT '结束时间',
    total_urls INT DEFAULT 0 COMMENT '总URL数',
    processed_urls INT DEFAULT 0 COMMENT '已处理URL数',
    success_urls INT DEFAULT 0 COMMENT '成功URL数',
    failed_urls INT DEFAULT 0 COMMENT '失败URL数',
    items_count INT DEFAULT 0 COMMENT '抓取项目数',
    error_message TEXT COMMENT '错误信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    INDEX idx_site_id (site_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='爬虫任务表';
`

// 爬取数据表
const createCrawlDataTable = `
CREATE TABLE IF NOT EXISTS crawl_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id INT NULL COMMENT '任务ID',
    site_id INT NOT NULL COMMENT '站点ID',
    url VARCHAR(2000) NOT NULL COMMENT '原始URL',
    title TEXT COMMENT '标题',
    content LONGTEXT COMMENT '内容',
    description TEXT COMMENT '描述',
    author VARCHAR(255) COMMENT '作者',
    source VARCHAR(255) COMMENT '来源',
    language VARCHAR(10) COMMENT '语言',
    publish_date DATETIME NULL COMMENT '发布时间',
    keywords JSON COMMENT '关键词',
    tags JSON COMMENT '标签',
    links JSON COMMENT '链接',
    images JSON COMMENT '图片',
    videos JSON COMMENT '视频',
    metadata JSON COMMENT '额外元数据',
    view_count INT DEFAULT 0 COMMENT '浏览量',
    comment_count INT DEFAULT 0 COMMENT '评论数',
    like_count INT DEFAULT 0 COMMENT '点赞数',
    share_count INT DEFAULT 0 COMMENT '分享数',
    status ENUM('new', 'processed', 'failed', 'skipped') DEFAULT 'new' COMMENT '状态',
    crawl_time DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '抓取时间',
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
    UNIQUE KEY unique_url_site (url(500), site_id),
    INDEX idx_task_id (task_id),
    INDEX idx_site_id (site_id),
    INDEX idx_publish_date (publish_date),
    INDEX idx_crawl_time (crawl_time),
    INDEX idx_status (status),
    FULLTEXT INDEX ft_title_content (title, content)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='爬取数据表';
`

// 任务日志表
const createTaskLogsTable = `
CREATE TABLE IF NOT EXISTS task_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    task_id INT NOT NULL COMMENT '任务ID',
    level ENUM('debug', 'info', 'warn', 'error') DEFAULT 'info' COMMENT '日志级别',
    message TEXT NOT NULL COMMENT '日志消息',
    details JSON COMMENT '详细信息',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    INDEX idx_task_id (task_id),
    INDEX idx_level (level),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='任务日志表';
`

// 系统配置表
const createSystemConfigTable = `
CREATE TABLE IF NOT EXISTS system_config (
    id INT AUTO_INCREMENT PRIMARY KEY,
    config_key VARCHAR(255) NOT NULL UNIQUE COMMENT '配置键',
    config_value JSON COMMENT '配置值',
    description TEXT COMMENT '配置描述',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';
` 