package models

import (
	"time"
)

// Item 抓取的数据项
type Item struct {
	// 基础信息
	ID          string    `json:"id"`          // 唯一标识
	URL         string    `json:"url"`         // 原始URL
	Title       string    `json:"title"`       // 标题
	Content     string    `json:"content"`     // 内容
	Description string    `json:"description"` // 描述
	
	// 元数据
	Author      string    `json:"author"`       // 作者
	Source      string    `json:"source"`       // 来源
	PublishDate time.Time `json:"publish_date"` // 发布时间
	Timestamp   time.Time `json:"timestamp"`    // 抓取时间
	
	// 标签和分类
	Keywords []string `json:"keywords"` // 关键词
	Tags     []string `json:"tags"`     // 标签
	Category string   `json:"category"` // 分类
	
	// 链接和媒体
	Links  []string `json:"links"`  // 链接
	Images []string `json:"images"` // 图片
	Videos []string `json:"videos"` // 视频
	
	// 额外信息
	Language string                 `json:"language"` // 语言
	Status   string                 `json:"status"`   // 状态
	Metadata map[string]interface{} `json:"metadata"` // 额外元数据
	
	// 统计信息
	ViewCount    int `json:"view_count"`    // 浏览量
	CommentCount int `json:"comment_count"` // 评论数
	LikeCount    int `json:"like_count"`    // 点赞数
	ShareCount   int `json:"share_count"`   // 分享数
}

// NewItem 创建新的数据项
func NewItem(url string) *Item {
	return &Item{
		URL:       url,
		Timestamp: time.Now(),
		Keywords:  []string{},
		Tags:      []string{},
		Links:     []string{},
		Images:    []string{},
		Videos:    []string{},
		Metadata:  make(map[string]interface{}),
		Status:    "new",
	}
}

// SetMetadata 设置元数据
func (i *Item) SetMetadata(key string, value interface{}) {
	if i.Metadata == nil {
		i.Metadata = make(map[string]interface{})
	}
	i.Metadata[key] = value
}

// GetMetadata 获取元数据
func (i *Item) GetMetadata(key string) (interface{}, bool) {
	if i.Metadata == nil {
		return nil, false
	}
	value, exists := i.Metadata[key]
	return value, exists
}

// AddKeyword 添加关键词
func (i *Item) AddKeyword(keyword string) {
	if !i.containsString(i.Keywords, keyword) {
		i.Keywords = append(i.Keywords, keyword)
	}
}

// AddTag 添加标签
func (i *Item) AddTag(tag string) {
	if !i.containsString(i.Tags, tag) {
		i.Tags = append(i.Tags, tag)
	}
}

// AddLink 添加链接
func (i *Item) AddLink(link string) {
	if !i.containsString(i.Links, link) {
		i.Links = append(i.Links, link)
	}
}

// AddImage 添加图片
func (i *Item) AddImage(image string) {
	if !i.containsString(i.Images, image) {
		i.Images = append(i.Images, image)
	}
}

// AddVideo 添加视频
func (i *Item) AddVideo(video string) {
	if !i.containsString(i.Videos, video) {
		i.Videos = append(i.Videos, video)
	}
}

// containsString 检查字符串是否在切片中
func (i *Item) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// IsValid 检查数据项是否有效
func (i *Item) IsValid() bool {
	return i.URL != "" && i.Title != ""
}

// GetHash 获取内容哈希
func (i *Item) GetHash() string {
	// 可以使用URL和标题生成哈希
	return i.URL + "|" + i.Title
}

// Comment 评论模型
type Comment struct {
	ID        string    `json:"id"`
	ItemID    string    `json:"item_id"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	LikeCount int       `json:"like_count"`
	ReplyTo   string    `json:"reply_to"` // 回复的评论ID
}

// NewComment 创建新评论
func NewComment(itemID, author, content string) *Comment {
	return &Comment{
		ItemID:    itemID,
		Author:    author,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// CrawlTask 爬虫任务模型
type CrawlTask struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	URLs        []string               `json:"urls"`
	Status      string                 `json:"status"`      // pending, running, completed, failed
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	ItemsCount  int                    `json:"items_count"`
	ErrorsCount int                    `json:"errors_count"`
	Config      map[string]interface{} `json:"config"`
	Results     []string               `json:"results"` // 结果文件路径
}

// NewCrawlTask 创建新的爬虫任务
func NewCrawlTask(name string, urls []string) *CrawlTask {
	return &CrawlTask{
		Name:      name,
		URLs:      urls,
		Status:    "pending",
		StartTime: time.Now(),
		Config:    make(map[string]interface{}),
		Results:   []string{},
	}
}

// Start 开始任务
func (t *CrawlTask) Start() {
	t.Status = "running"
	t.StartTime = time.Now()
}

// Complete 完成任务
func (t *CrawlTask) Complete() {
	t.Status = "completed"
	t.EndTime = time.Now()
}

// Fail 任务失败
func (t *CrawlTask) Fail() {
	t.Status = "failed"
	t.EndTime = time.Now()
}

// GetDuration 获取任务持续时间
func (t *CrawlTask) GetDuration() time.Duration {
	if t.EndTime.IsZero() {
		return time.Since(t.StartTime)
	}
	return t.EndTime.Sub(t.StartTime)
}

// CrawlStatistics 爬虫统计信息
type CrawlStatistics struct {
	TotalItems      int           `json:"total_items"`
	SuccessfulItems int           `json:"successful_items"`
	FailedItems     int           `json:"failed_items"`
	TotalTime       time.Duration `json:"total_time"`
	AverageTime     time.Duration `json:"average_time"`
	ItemsPerSecond  float64       `json:"items_per_second"`
	Sources         map[string]int `json:"sources"`         // 来源统计
	Categories      map[string]int `json:"categories"`      // 分类统计
	Languages       map[string]int `json:"languages"`       // 语言统计
	Errors          []string       `json:"errors"`          // 错误列表
}

// NewCrawlStatistics 创建新的统计信息
func NewCrawlStatistics() *CrawlStatistics {
	return &CrawlStatistics{
		Sources:    make(map[string]int),
		Categories: make(map[string]int),
		Languages:  make(map[string]int),
		Errors:     []string{},
	}
}

// AddItem 添加统计项
func (s *CrawlStatistics) AddItem(item *Item, success bool) {
	s.TotalItems++
	if success {
		s.SuccessfulItems++
	} else {
		s.FailedItems++
	}
	
	// 统计来源
	if item.Source != "" {
		s.Sources[item.Source]++
	}
	
	// 统计分类
	if item.Category != "" {
		s.Categories[item.Category]++
	}
	
	// 统计语言
	if item.Language != "" {
		s.Languages[item.Language]++
	}
}

// AddError 添加错误
func (s *CrawlStatistics) AddError(error string) {
	s.Errors = append(s.Errors, error)
}

// Calculate 计算统计数据
func (s *CrawlStatistics) Calculate() {
	if s.TotalItems > 0 && s.TotalTime > 0 {
		s.AverageTime = s.TotalTime / time.Duration(s.TotalItems)
		s.ItemsPerSecond = float64(s.TotalItems) / s.TotalTime.Seconds()
	}
}

// SiteConfig 站点配置模型
type SiteConfig struct {
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	StartURLs   []string          `json:"start_urls"`
	Selectors   map[string]string `json:"selectors"`
	Enabled     bool              `json:"enabled"`
	Description string            `json:"description"`
	Rules       CrawlRules        `json:"rules"`
}

// CrawlRules 爬虫规则
type CrawlRules struct {
	MaxDepth        int      `json:"max_depth"`         // 最大深度
	MaxPages        int      `json:"max_pages"`         // 最大页面数
	RespectRobots   bool     `json:"respect_robots"`    // 是否遵守robots.txt
	AllowedDomains  []string `json:"allowed_domains"`   // 允许的域名
	ForbiddenDomains []string `json:"forbidden_domains"` // 禁止的域名
	URLPatterns     []string `json:"url_patterns"`      // URL模式
	ContentTypes    []string `json:"content_types"`     // 内容类型
} 