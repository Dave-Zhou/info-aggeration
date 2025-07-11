# Colly 爬虫 API 文档

## 概述

本文档描述了 Golang Colly 爬虫项目的 API 接口和使用方法。

## 项目结构

```
info-aggeration/
├── cmd/                    # 主程序入口
│   └── crawler/           # 爬虫主程序
│       └── main.go        # 程序入口点
├── internal/              # 内部包（不对外暴露）
│   ├── crawler/           # 爬虫核心逻辑
│   │   ├── spider.go      # 爬虫实现
│   │   └── parser.go      # 数据解析器
│   ├── config/            # 配置管理
│   │   └── config.go      # 配置文件读取
│   ├── storage/           # 数据存储
│   │   ├── storage.go     # 存储接口
│   │   ├── file.go        # 文件存储
│   │   ├── database.go    # 数据库存储
│   │   └── excel.go       # Excel 文件存储
│   └── utils/             # 工具函数
│       ├── logger.go      # 日志工具
│       └── helpers.go     # 辅助函数
├── pkg/                   # 公共包（可被其他项目使用）
│   ├── models/            # 数据模型
│   │   └── item.go        # 数据结构定义
│   └── constants/         # 常量定义
│       └── constants.go   # 项目常量
├── config/                # 配置文件
│   ├── config.yaml        # 主配置文件
│   └── sites.yaml         # 站点配置文件
├── data/                  # 数据输出目录
│   ├── output/            # 爬取结果
│   └── logs/              # 日志文件
├── scripts/               # 脚本文件
│   ├── build.sh           # 构建脚本
│   └── run.sh             # 运行脚本
└── docs/                  # 文档
    └── api.md             # API 文档
```

## 核心接口

### 1. 爬虫接口 (Spider)

#### 创建爬虫实例

```go
func NewSpider(cfg *config.Config, store storage.Storage, logger utils.Logger) *Spider
```

**参数说明:**
- `cfg`: 配置对象
- `store`: 存储接口实现
- `logger`: 日志记录器

**返回值:**
- 返回爬虫实例

#### 启动爬虫

```go
func (s *Spider) Start() error
```

**功能:** 启动爬虫开始抓取数据

**返回值:**
- `error`: 错误信息，成功时为 nil

#### 停止爬虫

```go
func (s *Spider) Stop() error
```

**功能:** 停止爬虫

**返回值:**
- `error`: 错误信息，成功时为 nil

### 2. 存储接口 (Storage)

#### 存储接口定义

```go
type Storage interface {
    Save(item *models.Item) error
    SaveBatch(items []*models.Item) error
    Close() error
}
```

#### 创建存储实例

```go
func NewStorage(cfg config.StorageConfig) (Storage, error)
```

**参数说明:**
- `cfg`: 存储配置

**返回值:**
- `Storage`: 存储接口实现
- `error`: 错误信息

### 3. 数据解析器 (Parser)

#### 解析器接口定义

```go
type Parser interface {
    Parse(e *colly.HTMLElement) (*models.Item, error)
    ParseList(e *colly.HTMLElement) ([]*models.Item, error)
}
```

#### 创建默认解析器

```go
func NewDefaultParser() *DefaultParser
```

**返回值:**
- 返回默认解析器实例

### 4. 配置管理 (Config)

#### 加载主配置

```go
func LoadConfig() (*Config, error)
```

**功能:** 加载主配置文件

**返回值:**
- `*Config`: 配置对象
- `error`: 错误信息

#### 加载站点配置

```go
func LoadSiteConfig() (*SiteConfig, error)
```

**功能:** 加载站点配置文件

**返回值:**
- `*SiteConfig`: 站点配置对象
- `error`: 错误信息

## 数据模型

### 数据项 (Item)

```go
type Item struct {
    // 基础信息
    ID          string    `json:"id"`
    URL         string    `json:"url"`
    Title       string    `json:"title"`
    Content     string    `json:"content"`
    Description string    `json:"description"`
    
    // 元数据
    Author      string    `json:"author"`
    Source      string    `json:"source"`
    PublishDate time.Time `json:"publish_date"`
    Timestamp   time.Time `json:"timestamp"`
    
    // 标签和分类
    Keywords []string `json:"keywords"`
    Tags     []string `json:"tags"`
    Category string   `json:"category"`
    
    // 链接和媒体
    Links  []string `json:"links"`
    Images []string `json:"images"`
    Videos []string `json:"videos"`
    
    // 额外信息
    Language string                 `json:"language"`
    Status   string                 `json:"status"`
    Metadata map[string]interface{} `json:"metadata"`
}
```

### 配置结构

#### 主配置 (Config)

```go
type Config struct {
    Spider  SpiderConfig  `yaml:"spider"`
    Storage StorageConfig `yaml:"storage"`
    Logging LoggingConfig `yaml:"logging"`
}
```

#### 爬虫配置 (SpiderConfig)

```go
type SpiderConfig struct {
    Concurrent int    `yaml:"concurrent"`
    Delay      int    `yaml:"delay"`
    UserAgent  string `yaml:"user_agent"`
    Timeout    int    `yaml:"timeout"`
    Retries    int    `yaml:"retries"`
    ProxyURL   string `yaml:"proxy_url"`
}
```

#### 存储配置 (StorageConfig)

```go
type StorageConfig struct {
    Type      string   `yaml:"type"`
    OutputDir string   `yaml:"output_dir"`
    Database  DBConfig `yaml:"database"`
}
```

## 使用示例

### 1. 基本使用

```go
package main

import (
    "log"
    "example.com/m/v2/internal/config"
    "example.com/m/v2/internal/crawler"
    "example.com/m/v2/internal/storage"
    "example.com/m/v2/internal/utils"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建日志记录器
    logger := utils.NewLogger()
    
    // 创建存储
    store, err := storage.NewStorage(cfg.Storage)
    if err != nil {
        log.Fatal(err)
    }
    defer store.Close()
    
    // 创建爬虫
    spider := crawler.NewSpider(cfg, store, logger)
    
    // 启动爬虫
    if err := spider.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### 2. 自定义解析器

```go
type CustomParser struct {
    *crawler.DefaultParser
}

func (p *CustomParser) Parse(e *colly.HTMLElement) (*models.Item, error) {
    // 自定义解析逻辑
    item := &models.Item{
        URL:   e.Request.URL.String(),
        Title: e.ChildText("h1.custom-title"),
        // ... 其他字段
    }
    return item, nil
}
```

### 3. 自定义存储

```go
type CustomStorage struct {
    // 自定义存储字段
}

func (s *CustomStorage) Save(item *models.Item) error {
    // 自定义存储逻辑
    return nil
}

func (s *CustomStorage) SaveBatch(items []*models.Item) error {
    // 批量存储逻辑
    return nil
}

func (s *CustomStorage) Close() error {
    // 清理资源
    return nil
}
```

## 工具函数

### 日志工具

```go
// 创建日志记录器
logger := utils.NewLogger()

// 记录不同级别的日志
logger.Info("信息日志", "key", "value")
logger.Error("错误日志", "error", err)
logger.Warn("警告日志")
logger.Debug("调试日志")
```

### 辅助函数

```go
// URL 处理
urlHelper := utils.NewURLHelper()
isValid := urlHelper.IsValidURL("https://example.com")
domain, _ := urlHelper.GetDomain("https://example.com/path")

// 字符串处理
strHelper := utils.NewStringHelper()
cleaned := strHelper.CleanText("原始文本")
emails := strHelper.ExtractEmails("文本中的邮箱")

// 时间处理
timeHelper := utils.NewTimeHelper()
formatted := timeHelper.FormatTime(time.Now(), "2006-01-02")
```

## 配置说明

### 主配置文件 (config.yaml)

```yaml
spider:
  concurrent: 5
  delay: 1000
  timeout: 30
  retries: 3
  user_agent: "Mozilla/5.0 ..."

storage:
  type: "file"
  output_dir: "./data/output"

logging:
  level: "info"
  file: "./data/logs/crawler.log"
```

### 站点配置文件 (sites.yaml)

```yaml
sites:
  - name: "示例网站"
    base_url: "https://example.com"
    enabled: true
    start_urls:
      - "https://example.com/page1"
      - "https://example.com/page2"
    selectors:
      title: "h1.title"
      content: ".content"
      author: ".author"
```

## 错误处理

### 常见错误类型

- `ErrInvalidURL`: 无效的 URL
- `ErrConnectionFailed`: 连接失败
- `ErrTimeout`: 请求超时
- `ErrParsingFailed`: 解析失败
- `ErrStorageFailed`: 存储失败

### 错误处理示例

```go
if err := spider.Start(); err != nil {
    switch err.Error() {
    case constants.ErrConnectionFailed:
        log.Println("连接失败，请检查网络")
    case constants.ErrInvalidConfig:
        log.Println("配置文件错误，请检查配置")
    default:
        log.Printf("未知错误: %v", err)
    }
}
```

## 性能优化

### 并发控制

```yaml
spider:
  concurrent: 10  # 并发数
  delay: 500      # 请求间隔
```

### 内存管理

```yaml
performance:
  max_memory: 1024      # 最大内存使用（MB）
  queue_size: 10000     # 队列大小
  batch_size: 100       # 批处理大小
```

### 存储优化

```yaml
storage:
  type: "database"      # 使用数据库存储
  database:
    driver: "sqlite3"
    sqlite_file: "./data/crawler.db"
```

## 扩展开发

### 1. 添加新的存储类型

1. 实现 `Storage` 接口
2. 在 `storage.go` 中注册新类型
3. 更新配置文件

### 2. 添加新的解析器

1. 实现 `Parser` 接口
2. 在爬虫中使用自定义解析器

### 3. 添加新的工具函数

1. 在 `utils` 包中添加新的工具类
2. 遵循现有的命名规范

## 部署指南

### 1. 构建项目

```bash
# 使用构建脚本
./scripts/build.sh

# 或者直接使用 go build
go build -o bin/crawler ./cmd/crawler
```

### 2. 运行项目

```bash
# 使用运行脚本
./scripts/run.sh start

# 或者直接运行
./bin/crawler
```

### 3. 后台运行

```bash
# 后台模式
./scripts/run.sh start -d

# 查看状态
./scripts/run.sh status

# 查看日志
./scripts/run.sh logs
```

## 常见问题 (FAQ)

### Q: 如何配置代理？

A: 在配置文件中设置 `proxy_url`：

```yaml
spider:
  proxy_url: "http://proxy.example.com:8080"
```

### Q: 如何处理动态内容？

A: Colly 主要处理静态内容，对于动态内容建议使用 Selenium 等工具。

### Q: 如何避免被反爬？

A: 
- 设置合理的请求间隔
- 使用不同的 User-Agent
- 配置代理池
- 遵守 robots.txt

### Q: 如何处理大量数据？

A: 
- 使用数据库存储
- 启用批处理模式
- 适当增加并发数
- 定期清理日志

## 许可证

本项目基于 MIT 许可证开源。 