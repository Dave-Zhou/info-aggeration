package crawler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/debug"
	"github.com/gocolly/colly/v2/extensions"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/storage"
	"example.com/m/v2/internal/utils"
	"example.com/m/v2/pkg/models"
)

// Spider 爬虫结构
type Spider struct {
	config  *config.Config
	storage storage.Storage
	logger  utils.Logger
	
	collector *colly.Collector
	running   bool
	mu        sync.RWMutex
}

// NewSpider 创建新的爬虫实例
func NewSpider(cfg *config.Config, store storage.Storage, logger utils.Logger) *Spider {
	return &Spider{
		config:  cfg,
		storage: store,
		logger:  logger,
		running: false,
	}
}

// Start 启动爬虫
func (s *Spider) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("爬虫已在运行中")
	}

	s.logger.Info("初始化爬虫...")

	// 创建 Collector
	c := colly.NewCollector(
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
	)

	// 配置 Collector
	s.setupCollector(c)

	// 设置请求处理
	s.setupHandlers(c)

	// 限制并发和延迟
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.config.Spider.Concurrent,
		Delay:       time.Duration(s.config.Spider.Delay) * time.Millisecond,
	})

	s.collector = c
	s.running = true

	// 加载站点配置
	siteConfig, err := config.LoadSiteConfig()
	if err != nil {
		return fmt.Errorf("加载站点配置失败: %w", err)
	}

	// 开始爬取
	s.logger.Info("开始爬取数据...")
	for _, site := range siteConfig.Sites {
		if !site.Enabled {
			s.logger.Info("跳过禁用的站点", "site", site.Name)
			continue
		}

		s.logger.Info("开始爬取站点", "site", site.Name)
		for _, url := range site.StartURLs {
			s.logger.Info("访问URL", "url", url)
			c.Visit(url)
		}
	}

	// 等待所有请求完成
	c.Wait()

	s.logger.Info("爬取完成")
	return nil
}

// Stop 停止爬虫
func (s *Spider) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	s.logger.Info("停止爬虫...")
	s.running = false

	if s.collector != nil {
		s.collector.OnError(nil)
	}

	return nil
}

// setupCollector 配置 Collector
func (s *Spider) setupCollector(c *colly.Collector) {
	// 设置用户代理
	c.UserAgent = s.config.Spider.UserAgent

	// 设置超时
	c.SetRequestTimeout(time.Duration(s.config.Spider.Timeout) * time.Second)

	// 添加扩展
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// 设置代理（如果配置了）
	if s.config.Spider.ProxyURL != "" {
		c.SetProxyFunc(colly.ProxyURL(s.config.Spider.ProxyURL))
	}

	// 允许重复访问
	c.AllowURLRevisit = false

	// 设置允许的域名
	c.AllowedDomains = []string{} // 根据需要设置
}

// setupHandlers 设置请求处理器
func (s *Spider) setupHandlers(c *colly.Collector) {
	// 请求前处理
	c.OnRequest(func(r *colly.Request) {
		s.logger.Info("访问页面", "url", r.URL.String())
	})

	// 响应处理
	c.OnResponse(func(r *colly.Response) {
		s.logger.Info("收到响应", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	// HTML 处理
	c.OnHTML("html", func(e *colly.HTMLElement) {
		s.processPage(e)
	})

	// 错误处理
	c.OnError(func(r *colly.Response, err error) {
		s.logger.Error("请求失败", "url", r.Request.URL.String(), "error", err.Error())
		
		// 重试逻辑
		retryCount := r.Request.Headers.Get("Retry-Count")
		if retryCount == "" {
			retryCount = "0"
		}
		
		if retryCount < fmt.Sprintf("%d", s.config.Spider.Retries) {
			newCount := fmt.Sprintf("%d", getRetryCount(retryCount)+1)
			r.Request.Headers.Set("Retry-Count", newCount)
			
			s.logger.Info("重试请求", "url", r.Request.URL.String(), "retry", newCount)
			time.Sleep(time.Second * 2) // 等待2秒后重试
			r.Request.Retry()
		}
	})

	// 抓取完成
	c.OnScraped(func(r *colly.Response) {
		s.logger.Info("页面抓取完成", "url", r.Request.URL.String())
	})
}

// processPage 处理页面数据
func (s *Spider) processPage(e *colly.HTMLElement) {
	url := e.Request.URL.String()
	
	// 创建数据项
	item := &models.Item{
		URL:       url,
		Title:     strings.TrimSpace(e.ChildText("title")),
		Content:   strings.TrimSpace(e.ChildText("body")),
		Timestamp: time.Now(),
		Source:    getDomainFromURL(url),
	}

	// 提取其他数据
	s.extractData(e, item)

	// 保存数据
	if err := s.storage.Save(item); err != nil {
		s.logger.Error("保存数据失败", "url", url, "error", err)
	} else {
		s.logger.Info("数据保存成功", "url", url)
	}
}

// extractData 提取页面数据
func (s *Spider) extractData(e *colly.HTMLElement, item *models.Item) {
	// 这里可以根据具体需求添加数据提取逻辑
	// 例如：
	// item.Description = e.ChildText("meta[name='description']")
	// item.Keywords = e.ChildText("meta[name='keywords']")
	// item.Author = e.ChildText(".author")
	// item.PublishDate = e.ChildText(".publish-date")
	
	// 提取链接
	e.ForEach("a[href]", func(i int, el *colly.HTMLElement) {
		link := el.Attr("href")
		if link != "" {
			item.Links = append(item.Links, link)
		}
	})

	// 提取图片
	e.ForEach("img[src]", func(i int, el *colly.HTMLElement) {
		src := el.Attr("src")
		if src != "" {
			item.Images = append(item.Images, src)
		}
	})
}

// 辅助函数
func getRetryCount(retryStr string) int {
	if retryStr == "" {
		return 0
	}
	// 简单解析，实际项目中可能需要更严格的错误处理
	if retryStr == "1" {
		return 1
	} else if retryStr == "2" {
		return 2
	} else if retryStr == "3" {
		return 3
	}
	return 0
}

func getDomainFromURL(url string) string {
	// 简单的域名提取，实际项目中可能需要更复杂的逻辑
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
} 