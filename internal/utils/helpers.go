package utils

import (
	"crypto/md5"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// URLHelper URL处理工具
type URLHelper struct{}

// NewURLHelper 创建URL处理工具
func NewURLHelper() *URLHelper {
	return &URLHelper{}
}

// IsValidURL 检查URL是否有效
func (u *URLHelper) IsValidURL(rawURL string) bool {
	_, err := url.Parse(rawURL)
	return err == nil
}

// NormalizeURL 规范化URL
func (u *URLHelper) NormalizeURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 移除fragment
	parsedURL.Fragment = ""

	// 规范化路径
	if parsedURL.Path == "" {
		parsedURL.Path = "/"
	}

	return parsedURL.String(), nil
}

// GetDomain 获取域名
func (u *URLHelper) GetDomain(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	return parsedURL.Host, nil
}

// JoinURL 合并URL
func (u *URLHelper) JoinURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	relative, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(relative).String(), nil
}

// StringHelper 字符串处理工具
type StringHelper struct{}

// NewStringHelper 创建字符串处理工具
func NewStringHelper() *StringHelper {
	return &StringHelper{}
}

// CleanText 清理文本
func (s *StringHelper) CleanText(text string) string {
	// 移除HTML标签
	text = s.RemoveHTMLTags(text)
	
	// 移除多余的空白字符
	text = s.RemoveExtraWhitespace(text)
	
	// 移除特殊字符
	text = s.RemoveSpecialChars(text)
	
	return strings.TrimSpace(text)
}

// RemoveHTMLTags 移除HTML标签
func (s *StringHelper) RemoveHTMLTags(text string) string {
	// 简单的HTML标签移除
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(text, "")
}

// RemoveExtraWhitespace 移除多余的空白字符
func (s *StringHelper) RemoveExtraWhitespace(text string) string {
	// 替换多个空白字符为单个空格
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(text, " ")
}

// RemoveSpecialChars 移除特殊字符
func (s *StringHelper) RemoveSpecialChars(text string) string {
	// 移除不可见字符
	re := regexp.MustCompile(`[\x00-\x1F\x7F-\x9F]`)
	return re.ReplaceAllString(text, "")
}

// TruncateText 截断文本
func (s *StringHelper) TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

// ExtractEmails 提取邮箱地址
func (s *StringHelper) ExtractEmails(text string) []string {
	re := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	return re.FindAllString(text, -1)
}

// ExtractPhones 提取电话号码
func (s *StringHelper) ExtractPhones(text string) []string {
	// 简单的电话号码正则
	re := regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b|\b\d{10}\b|\b\(\d{3}\)\s*\d{3}-\d{4}\b`)
	return re.FindAllString(text, -1)
}

// HashHelper 哈希处理工具
type HashHelper struct{}

// NewHashHelper 创建哈希处理工具
func NewHashHelper() *HashHelper {
	return &HashHelper{}
}

// MD5Hash 计算MD5哈希
func (h *HashHelper) MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return fmt.Sprintf("%x", hash)
}

// GenerateID 生成ID
func (h *HashHelper) GenerateID(url string) string {
	return h.MD5Hash(url)
}

// TimeHelper 时间处理工具
type TimeHelper struct{}

// NewTimeHelper 创建时间处理工具
func NewTimeHelper() *TimeHelper {
	return &TimeHelper{}
}

// FormatTime 格式化时间
func (t *TimeHelper) FormatTime(time time.Time, format string) string {
	if format == "" {
		format = "2006-01-02 15:04:05"
	}
	return time.Format(format)
}

// ParseTime 解析时间
func (t *TimeHelper) ParseTime(timeStr string) (time.Time, error) {
	// 常用时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析时间格式: %s", timeStr)
}

// GetCurrentTime 获取当前时间
func (t *TimeHelper) GetCurrentTime() time.Time {
	return time.Now()
}

// GetTimeAgo 获取相对时间
func (t *TimeHelper) GetTimeAgo(past time.Time) string {
	duration := time.Since(past)
	
	if duration < time.Minute {
		return "刚刚"
	} else if duration < time.Hour {
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	} else if duration < 30*24*time.Hour {
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	} else {
		return past.Format("2006-01-02")
	}
}

// ValidatorHelper 验证工具
type ValidatorHelper struct{}

// NewValidatorHelper 创建验证工具
func NewValidatorHelper() *ValidatorHelper {
	return &ValidatorHelper{}
}

// ValidateURL 验证URL
func (v *ValidatorHelper) ValidateURL(rawURL string) error {
	if rawURL == "" {
		return fmt.Errorf("URL不能为空")
	}
	
	_, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("无效的URL格式: %w", err)
	}
	
	return nil
}

// ValidateEmail 验证邮箱
func (v *ValidatorHelper) ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}
	
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !re.MatchString(email) {
		return fmt.Errorf("无效的邮箱格式")
	}
	
	return nil
}

// FileHelper 文件处理工具
type FileHelper struct{}

// NewFileHelper 创建文件处理工具
func NewFileHelper() *FileHelper {
	return &FileHelper{}
}

// GetFileExtension 获取文件扩展名
func (f *FileHelper) GetFileExtension(filename string) string {
	parts := strings.Split(filename, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

// IsImageFile 检查是否是图片文件
func (f *FileHelper) IsImageFile(filename string) bool {
	ext := strings.ToLower(f.GetFileExtension(filename))
	imageExts := []string{"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"}
	
	for _, imageExt := range imageExts {
		if ext == imageExt {
			return true
		}
	}
	return false
}

// GetSafeFilename 获取安全的文件名
func (f *FileHelper) GetSafeFilename(filename string) string {
	// 移除或替换不安全的字符
	re := regexp.MustCompile(`[<>:"/\\|?*]`)
	safe := re.ReplaceAllString(filename, "_")
	
	// 限制长度
	if len(safe) > 100 {
		safe = safe[:100]
	}
	
	return safe
}

// ArrayHelper 数组处理工具
type ArrayHelper struct{}

// NewArrayHelper 创建数组处理工具
func NewArrayHelper() *ArrayHelper {
	return &ArrayHelper{}
}

// UniqueStrings 字符串数组去重
func (a *ArrayHelper) UniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// ContainsString 检查字符串是否在数组中
func (a *ArrayHelper) ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FilterEmptyStrings 过滤空字符串
func (a *ArrayHelper) FilterEmptyStrings(slice []string) []string {
	result := []string{}
	for _, item := range slice {
		if strings.TrimSpace(item) != "" {
			result = append(result, item)
		}
	}
	return result
} 