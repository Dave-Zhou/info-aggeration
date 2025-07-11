package constants

// 应用常量
const (
	AppName    = "Colly-Crawler"
	AppVersion = "1.0.0"
	AppDesc    = "基于Colly的网络爬虫框架"
)

// 存储类型常量
const (
	StorageTypeFile     = "file"
	StorageTypeDatabase = "database"
	StorageTypeExcel    = "excel"
	StorageTypeCSV      = "csv"
	StorageTypeJSON     = "json"
)

// 数据库驱动常量
const (
	DatabaseDriverSQLite = "sqlite3"
	DatabaseDriverMySQL  = "mysql"
	DatabaseDriverPostgres = "postgres"
)

// 日志级别常量
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// 爬虫状态常量
const (
	SpiderStatusPending   = "pending"
	SpiderStatusRunning   = "running"
	SpiderStatusCompleted = "completed"
	SpiderStatusFailed    = "failed"
	SpiderStatusStopped   = "stopped"
)

// 数据项状态常量
const (
	ItemStatusNew       = "new"
	ItemStatusProcessed = "processed"
	ItemStatusFailed    = "failed"
	ItemStatusSkipped   = "skipped"
)

// HTTP 常量
const (
	UserAgentChrome  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	UserAgentFirefox = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0"
	UserAgentSafari  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"
	UserAgentDefault = UserAgentChrome
)

// HTTP 状态码
const (
	HTTPStatusOK                  = 200
	HTTPStatusNotFound            = 404
	HTTPStatusInternalServerError = 500
	HTTPStatusTooManyRequests     = 429
	HTTPStatusForbidden           = 403
	HTTPStatusUnauthorized        = 401
)

// 文件扩展名
const (
	ExtensionJSON = ".json"
	ExtensionCSV  = ".csv"
	ExtensionXLSX = ".xlsx"
	ExtensionTXT  = ".txt"
	ExtensionSQL  = ".sql"
	ExtensionDB   = ".db"
)

// 默认配置值
const (
	DefaultConcurrency     = 5
	DefaultDelay           = 1000  // 毫秒
	DefaultTimeout         = 30    // 秒
	DefaultRetries         = 3
	DefaultMaxDepth        = 10
	DefaultMaxPages        = 1000
	DefaultOutputDir       = "./data/output"
	DefaultLogDir          = "./data/logs"
	DefaultConfigDir       = "./config"
	DefaultScriptsDir      = "./scripts"
	DefaultDocsDir         = "./docs"
	DefaultDatabaseFile    = "./data/crawler.db"
	DefaultLogFile         = "./data/logs/crawler.log"
	DefaultConfigFile      = "./config/config.yaml"
	DefaultSitesConfigFile = "./config/sites.yaml"
)

// 正则表达式模式
const (
	RegexEmail = `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	RegexPhone = `\b\d{3}-\d{3}-\d{4}\b|\b\d{10}\b|\b\(\d{3}\)\s*\d{3}-\d{4}\b`
	RegexURL   = `https?://[^\s]+`
	RegexIP    = `\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`
	RegexDate  = `\d{4}-\d{2}-\d{2}`
	RegexTime  = `\d{2}:\d{2}:\d{2}`
)

// CSS 选择器模式
const (
	SelectorTitle       = "title, h1, .title, .headline"
	SelectorContent     = ".content, .article-content, .post-content, .entry-content, .main-content, article, .article-body, .post-body, main"
	SelectorDescription = "meta[name='description'], meta[property='og:description'], .description, .summary, .excerpt"
	SelectorAuthor      = ".author, .by-author, .post-author, .article-author, meta[name='author'], .byline, .writer"
	SelectorDate        = ".publish-date, .date, .post-date, .article-date, time, meta[property='article:published_time']"
	SelectorTags        = ".tag, .tags a, .category, .label"
	SelectorLinks       = "a[href]"
	SelectorImages      = "img[src]"
	SelectorVideos      = "video[src], iframe[src*='youtube'], iframe[src*='youtu.be']"
)

// 文件大小限制
const (
	MaxFileSize     = 50 * 1024 * 1024 // 50MB
	MaxContentSize  = 10 * 1024 * 1024 // 10MB
	MaxLogFileSize  = 100 * 1024 * 1024 // 100MB
	MaxImageSize    = 5 * 1024 * 1024  // 5MB
)

// 队列配置
const (
	DefaultQueueSize     = 10000
	DefaultWorkerCount   = 10
	DefaultBatchSize     = 100
	DefaultFlushInterval = 5000 // 毫秒
)

// 错误消息
const (
	ErrInvalidURL        = "invalid URL"
	ErrInvalidConfig     = "invalid configuration"
	ErrConnectionFailed  = "connection failed"
	ErrTimeout           = "request timeout"
	ErrRateLimited       = "rate limited"
	ErrAccessDenied      = "access denied"
	ErrPageNotFound      = "page not found"
	ErrServerError       = "server error"
	ErrParsingFailed     = "parsing failed"
	ErrStorageFailed     = "storage failed"
	ErrDatabaseError     = "database error"
	ErrFileError         = "file error"
	ErrNetworkError      = "network error"
	ErrUnknownError      = "unknown error"
)

// 支持的内容类型
var SupportedContentTypes = []string{
	"text/html",
	"text/plain",
	"application/json",
	"application/xml",
	"text/xml",
	"application/rss+xml",
	"application/atom+xml",
}

// 支持的图片格式
var SupportedImageFormats = []string{
	"jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico",
}

// 支持的视频格式
var SupportedVideoFormats = []string{
	"mp4", "avi", "mov", "wmv", "flv", "webm", "mkv", "m4v",
}

// 支持的文档格式
var SupportedDocumentFormats = []string{
	"pdf", "doc", "docx", "xls", "xlsx", "ppt", "pptx", "txt", "rtf",
}

// 默认用户代理列表
var DefaultUserAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:89.0) Gecko/20100101 Firefox/89.0",
}

// 常见的停用词
var StopWords = []string{
	"的", "了", "在", "是", "我", "有", "和", "就", "不", "人", "都", "一", "一个", "上", "也", "很", "到", "说", "要", "去", "你", "会", "着", "没有", "看", "好", "自己", "这",
	"the", "be", "to", "of", "and", "a", "in", "that", "have", "i", "it", "for", "not", "on", "with", "he", "as", "you", "do", "at", "this", "but", "his", "by", "from",
}

// 语言检测模式
var LanguagePatterns = map[string]string{
	"zh": `[\u4e00-\u9fff]`,     // 中文
	"en": `[a-zA-Z]`,           // 英文
	"ja": `[\u3040-\u309f\u30a0-\u30ff\u4e00-\u9fff]`, // 日文
	"ko": `[\uac00-\ud7af]`,    // 韩文
	"ar": `[\u0600-\u06ff]`,    // 阿拉伯文
	"ru": `[\u0400-\u04ff]`,    // 俄文
} 