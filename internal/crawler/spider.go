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

// StartWithTask 启动针对单个任务的爬虫
func (s *Spider) StartWithTask(task *models.CrawlTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("爬虫已在运行中")
	}

	s.logger.Info("初始化爬虫任务...", "site", task.Name)

	// 创建 Collector
	c := colly.NewCollector(
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
	)

	// 配置 Collector
	s.setupCollector(c, task)

	// 设置请求处理
	s.setupHandlers(c, task)

	// 限制并发和延迟
	concurrent := s.config.Spider.Concurrent
	delay := s.config.Spider.Delay
	if task.Rules.Concurrent > 0 {
		concurrent = task.Rules.Concurrent
	}
	if task.Rules.Delay > 0 {
		delay = task.Rules.Delay
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: concurrent,
		Delay:       time.Duration(delay) * time.Millisecond,
	})

	s.collector = c
	s.running = true

	// 开始爬取
	s.logger.Info("开始爬取站点", "site", task.Name)
	for _, url := range task.StartURLs {
		s.logger.Info("访问URL", "url", url)
		c.Visit(url)
	}

	// 等待所有请求完成
	c.Wait()

	s.running = false
	s.logger.Info("站点爬取完成", "site", task.Name)
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
func (s *Spider) setupCollector(c *colly.Collector, task *models.CrawlTask) {
	// 设置用户代理
	c.UserAgent = s.config.Spider.UserAgent

	// 设置超时
	c.SetRequestTimeout(time.Duration(s.config.Spider.Timeout) * time.Second)

	// 添加扩展
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	// 设置代理（如果配置了）
	// if s.config.Spider.ProxyURL != "" {
	// 	c.SetProxyFunc(colly.ProxyURL(s.config.Spider.ProxyURL))
	// }

	// 允许重复访问
	c.AllowURLRevisit = false

	// 设置允许的域名
	domain := getDomainFromURL(task.BaseURL)
	c.AllowedDomains = []string{domain}
}

// setupHandlers 设置请求处理器
func (s *Spider) setupHandlers(c *colly.Collector, task *models.CrawlTask) {
	// 请求前处理
	c.OnRequest(func(r *colly.Request) {
		s.logger.Info("访问页面", "url", r.URL.String())
	})

	// 响应处理
	c.OnResponse(func(r *colly.Response) {
		s.logger.Info("收到响应", "url", r.Request.URL.String(), "status", r.StatusCode)
	})

	// HTML 处理 - 根据站点配置的item选择器来处理
	itemSelector := "html"
	if sel, ok := task.Selectors["item"]; ok && sel != "" {
		itemSelector = sel
	}

	c.OnHTML(itemSelector, func(e *colly.HTMLElement) {
		s.processPage(e, task)
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
func (s *Spider) processPage(e *colly.HTMLElement, task *models.CrawlTask) {
	url := e.Request.URL.String()

	// 创建数据项
	item := &models.Item{
		URL:       url,
		Timestamp: time.Now(),
		Source:    getDomainFromURL(url),
	}

	// 根据站点配置的选择器提取数据
	s.extractData(e, item, task.Selectors)

	// 保存数据
	if err := s.storage.Save(item); err != nil {
		s.logger.Error("保存数据失败", "url", url, "error", err)
	} else {
		s.logger.Info("数据保存成功", "url", url, "title", item.Title)
	}
}

// extractData 提取页面数据
func (s *Spider) extractData(e *colly.HTMLElement, item *models.Item, selectors map[string]string) {
	// 动态根据selectors提取数据
	if sel, ok := selectors["title"]; ok && sel != "" {
		item.Title = strings.TrimSpace(e.ChildText(sel))
	} else {
		item.Title = strings.TrimSpace(e.ChildText("title"))
	}

	if sel, ok := selectors["content"]; ok && sel != "" {
		item.Content = strings.TrimSpace(e.ChildText(sel))
	} else {
		item.Content = strings.TrimSpace(e.ChildText("body"))
	}

	if sel, ok := selectors["description"]; ok && sel != "" {
		item.Description = strings.TrimSpace(e.ChildAttr(sel, "content"))
	}

	if sel, ok := selectors["keywords"]; ok && sel != "" {
		item.Keywords = strings.Split(strings.TrimSpace(e.ChildAttr(sel, "content")), ",")
	}

	if sel, ok := selectors["author"]; ok && sel != "" {
		item.Author = strings.TrimSpace(e.ChildText(sel))
	}

	// 提取链接
	if sel, ok := selectors["links"]; ok && sel != "" {
		e.ForEach(sel, func(_ int, el *colly.HTMLElement) {
			link := el.Request.AbsoluteURL(el.Attr("href"))
			if link != "" {
				item.Links = append(item.Links, link)
			}
		})
	}

	// 提取图片
	if sel, ok := selectors["images"]; ok && sel != "" {
		e.ForEach(sel, func(_ int, el *colly.HTMLElement) {
			src := el.Request.AbsoluteURL(el.Attr("src"))
			if src != "" {
				item.Images = append(item.Images, src)
			}
		})
	}
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
