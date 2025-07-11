package crawler

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"example.com/m/v2/pkg/models"
)

// Parser 数据解析器接口
type Parser interface {
	Parse(e *colly.HTMLElement) (*models.Item, error)
	ParseList(e *colly.HTMLElement) ([]*models.Item, error)
}

// DefaultParser 默认解析器
type DefaultParser struct{}

// NewDefaultParser 创建默认解析器
func NewDefaultParser() *DefaultParser {
	return &DefaultParser{}
}

// Parse 解析单个页面
func (p *DefaultParser) Parse(e *colly.HTMLElement) (*models.Item, error) {
	item := &models.Item{
		URL:       e.Request.URL.String(),
		Title:     p.extractTitle(e),
		Content:   p.extractContent(e),
		Timestamp: time.Now(),
		Source:    p.extractSource(e),
	}

	// 提取其他字段
	item.Description = p.extractDescription(e)
	item.Keywords = p.extractKeywords(e)
	item.Author = p.extractAuthor(e)
	item.PublishDate = p.extractPublishDate(e)
	item.Links = p.extractLinks(e)
	item.Images = p.extractImages(e)
	item.Tags = p.extractTags(e)

	return item, nil
}

// ParseList 解析列表页面
func (p *DefaultParser) ParseList(e *colly.HTMLElement) ([]*models.Item, error) {
	var items []*models.Item

	// 通用的列表项选择器
	listSelectors := []string{
		"article", ".article", ".post", ".item", ".entry",
		".news-item", ".content-item", "li", ".list-item",
	}

	for _, selector := range listSelectors {
		e.ForEach(selector, func(i int, el *colly.HTMLElement) {
			item := &models.Item{
				Title:     p.extractTitleFromElement(el),
				Content:   p.extractContentFromElement(el),
				URL:       p.extractURLFromElement(el, e.Request.URL.String()),
				Timestamp: time.Now(),
				Source:    p.extractSource(e),
			}

			// 只有当提取到标题时才添加
			if item.Title != "" {
				items = append(items, item)
			}
		})

		// 如果找到了项目，就不再尝试其他选择器
		if len(items) > 0 {
			break
		}
	}

	return items, nil
}

// extractTitle 提取标题
func (p *DefaultParser) extractTitle(e *colly.HTMLElement) string {
	// 尝试多种标题选择器
	selectors := []string{
		"title",
		"h1",
		".title",
		".headline",
		"h1.title",
		"h1.headline",
		"meta[property='og:title']",
		"meta[name='twitter:title']",
	}

	for _, selector := range selectors {
		if selector == "meta[property='og:title']" || selector == "meta[name='twitter:title']" {
			if title := e.ChildAttr(selector, "content"); title != "" {
				return strings.TrimSpace(title)
			}
		} else {
			if title := e.ChildText(selector); title != "" {
				return strings.TrimSpace(title)
			}
		}
	}

	return ""
}

// extractContent 提取内容
func (p *DefaultParser) extractContent(e *colly.HTMLElement) string {
	// 尝试多种内容选择器
	selectors := []string{
		".content", ".article-content", ".post-content",
		".entry-content", ".main-content", "article",
		".article-body", ".post-body", "main",
	}

	for _, selector := range selectors {
		if content := e.ChildText(selector); content != "" {
			return p.cleanText(content)
		}
	}

	// 如果没有找到特定的内容区域，返回body内容
	if content := e.ChildText("body"); content != "" {
		return p.cleanText(content)
	}

	return ""
}

// extractDescription 提取描述
func (p *DefaultParser) extractDescription(e *colly.HTMLElement) string {
	// 尝试多种描述选择器
	selectors := []string{
		"meta[name='description']",
		"meta[property='og:description']",
		"meta[name='twitter:description']",
		".description",
		".summary",
		".excerpt",
	}

	for _, selector := range selectors {
		if strings.Contains(selector, "meta[") {
			if desc := e.ChildAttr(selector, "content"); desc != "" {
				return strings.TrimSpace(desc)
			}
		} else {
			if desc := e.ChildText(selector); desc != "" {
				return strings.TrimSpace(desc)
			}
		}
	}

	return ""
}

// extractKeywords 提取关键词
func (p *DefaultParser) extractKeywords(e *colly.HTMLElement) []string {
	// 从meta标签提取关键词
	keywords := e.ChildAttr("meta[name='keywords']", "content")
	if keywords != "" {
		return strings.Split(keywords, ",")
	}

	// 从标签中提取
	var tags []string
	e.ForEach(".tag, .keyword, .label", func(i int, el *colly.HTMLElement) {
		if tag := strings.TrimSpace(el.Text); tag != "" {
			tags = append(tags, tag)
		}
	})

	return tags
}

// extractAuthor 提取作者
func (p *DefaultParser) extractAuthor(e *colly.HTMLElement) string {
	selectors := []string{
		".author", ".by-author", ".post-author",
		".article-author", "meta[name='author']",
		"meta[property='article:author']",
		".byline", ".writer",
	}

	for _, selector := range selectors {
		if strings.Contains(selector, "meta[") {
			if author := e.ChildAttr(selector, "content"); author != "" {
				return strings.TrimSpace(author)
			}
		} else {
			if author := e.ChildText(selector); author != "" {
				return strings.TrimSpace(author)
			}
		}
	}

	return ""
}

// extractPublishDate 提取发布日期
func (p *DefaultParser) extractPublishDate(e *colly.HTMLElement) time.Time {
	selectors := []string{
		"meta[property='article:published_time']",
		"meta[name='publish_date']",
		".publish-date", ".date", ".post-date",
		".article-date", "time",
	}

	for _, selector := range selectors {
		var dateStr string
		if strings.Contains(selector, "meta[") {
			dateStr = e.ChildAttr(selector, "content")
		} else if selector == "time" {
			dateStr = e.ChildAttr(selector, "datetime")
			if dateStr == "" {
				dateStr = e.ChildText(selector)
			}
		} else {
			dateStr = e.ChildText(selector)
		}

		if dateStr != "" {
			if date := p.parseDate(dateStr); !date.IsZero() {
				return date
			}
		}
	}

	return time.Time{}
}

// extractLinks 提取链接
func (p *DefaultParser) extractLinks(e *colly.HTMLElement) []string {
	var links []string
	e.ForEach("a[href]", func(i int, el *colly.HTMLElement) {
		if href := el.Attr("href"); href != "" {
			links = append(links, href)
		}
	})
	return links
}

// extractImages 提取图片
func (p *DefaultParser) extractImages(e *colly.HTMLElement) []string {
	var images []string
	e.ForEach("img[src]", func(i int, el *colly.HTMLElement) {
		if src := el.Attr("src"); src != "" {
			images = append(images, src)
		}
	})
	return images
}

// extractTags 提取标签
func (p *DefaultParser) extractTags(e *colly.HTMLElement) []string {
	var tags []string
	e.ForEach(".tag, .tags a, .category", func(i int, el *colly.HTMLElement) {
		if tag := strings.TrimSpace(el.Text); tag != "" {
			tags = append(tags, tag)
		}
	})
	return tags
}

// extractSource 提取来源
func (p *DefaultParser) extractSource(e *colly.HTMLElement) string {
	host := e.Request.URL.Host
	if host != "" {
		return host
	}
	return "unknown"
}

// 辅助方法用于列表解析
func (p *DefaultParser) extractTitleFromElement(el *colly.HTMLElement) string {
	selectors := []string{"h1", "h2", "h3", ".title", "a"}
	for _, selector := range selectors {
		if title := el.ChildText(selector); title != "" {
			return strings.TrimSpace(title)
		}
	}
	return ""
}

func (p *DefaultParser) extractContentFromElement(el *colly.HTMLElement) string {
	selectors := []string{".content", ".summary", ".excerpt", "p"}
	for _, selector := range selectors {
		if content := el.ChildText(selector); content != "" {
			return p.cleanText(content)
		}
	}
	return ""
}

func (p *DefaultParser) extractURLFromElement(el *colly.HTMLElement, baseURL string) string {
	if url := el.ChildAttr("a", "href"); url != "" {
		return url
	}
	return baseURL
}

// cleanText 清理文本
func (p *DefaultParser) cleanText(text string) string {
	// 移除多余的空白字符
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// parseDate 解析日期
func (p *DefaultParser) parseDate(dateStr string) time.Time {
	// 常见的日期格式
	formats := []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02-01-2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
	}

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date
		}
	}

	// 尝试解析相对时间（如 "2小时前"）
	if relativeTime := p.parseRelativeTime(dateStr); !relativeTime.IsZero() {
		return relativeTime
	}

	return time.Time{}
}

// parseRelativeTime 解析相对时间
func (p *DefaultParser) parseRelativeTime(timeStr string) time.Time {
	now := time.Now()
	
	// 匹配"X小时前"、"X天前"等格式
	patterns := map[string]time.Duration{
		`(\d+)分钟前`:  time.Minute,
		`(\d+)小时前`:  time.Hour,
		`(\d+)天前`:   24 * time.Hour,
		`(\d+)周前`:   7 * 24 * time.Hour,
		`(\d+)月前`:   30 * 24 * time.Hour,
		`(\d+)年前`:   365 * 24 * time.Hour,
	}

	for pattern, duration := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(timeStr)
		if len(matches) > 1 {
			if num, err := strconv.Atoi(matches[1]); err == nil {
				return now.Add(-time.Duration(num) * duration)
			}
		}
	}

	return time.Time{}
} 