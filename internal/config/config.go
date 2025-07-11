package config

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 主配置结构
type Config struct {
	Spider  SpiderConfig  `yaml:"spider"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
	Web     WebConfig     `yaml:"web"`
}

// SpiderConfig 爬虫配置
type SpiderConfig struct {
	Concurrent int    `yaml:"concurrent"`     // 并发数
	Delay      int    `yaml:"delay"`          // 请求间隔（毫秒）
	UserAgent  string `yaml:"user_agent"`     // 用户代理
	Timeout    int    `yaml:"timeout"`        // 超时时间（秒）
	Retries    int    `yaml:"retries"`        // 重试次数
	ProxyURL   string `yaml:"proxy_url"`      // 代理地址
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type      string `yaml:"type"`       // 存储类型: file, database, excel
	OutputDir string `yaml:"output_dir"` // 输出目录
	Database  DBConfig `yaml:"database"`  // 数据库配置
}

// DBConfig 数据库配置
type DBConfig struct {
	Driver   string `yaml:"driver"`   // 数据库驱动
	Host     string `yaml:"host"`     // 主机地址
	Port     int    `yaml:"port"`     // 端口
	Username string `yaml:"username"` // 用户名
	Password string `yaml:"password"` // 密码
	Database string `yaml:"database"` // 数据库名
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level string `yaml:"level"` // 日志级别
	File  string `yaml:"file"`  // 日志文件路径
}

// WebConfig Web服务器配置
type WebConfig struct {
	Port       int       `yaml:"port"`        // 服务器端口
	StaticPath string    `yaml:"static_path"` // 静态文件路径
	APIPrefix  string    `yaml:"api_prefix"`  // API前缀
	CORS       CORSConfig `yaml:"cors"`       // CORS配置
	Auth       AuthConfig `yaml:"auth"`       // 认证配置
}

// CORSConfig CORS配置
type CORSConfig struct {
	Origins []string `yaml:"origins"` // 允许的来源
	Methods []string `yaml:"methods"` // 允许的方法
	Headers []string `yaml:"headers"` // 允许的头部
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enable      bool   `yaml:"enable"`       // 是否启用认证
	JWTSecret   string `yaml:"jwt_secret"`   // JWT密钥
	TokenExpire int    `yaml:"token_expire"` // Token过期时间（小时）
}

// SiteConfig 站点配置结构
type SiteConfig struct {
	Sites []Site `yaml:"sites"`
}

// Site 单个站点配置
type Site struct {
	Name        string            `yaml:"name"`         // 站点名称
	BaseURL     string            `yaml:"base_url"`     // 基础URL
	StartURLs   []string          `yaml:"start_urls"`   // 起始URL列表
	Selectors   map[string]string `yaml:"selectors"`    // CSS选择器
	Enabled     bool              `yaml:"enabled"`      // 是否启用
	Description string            `yaml:"description"`  // 描述
}

// LoadConfig 加载主配置文件
func LoadConfig() (*Config, error) {
	configPath := filepath.Join("config", "config.yaml")
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	setDefaults(&config)

	return &config, nil
}

// LoadSiteConfig 加载站点配置文件
func LoadSiteConfig() (*SiteConfig, error) {
	configPath := filepath.Join("config", "sites.yaml")
	
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取站点配置文件失败: %w", err)
	}

	var siteConfig SiteConfig
	if err := yaml.Unmarshal(data, &siteConfig); err != nil {
		return nil, fmt.Errorf("解析站点配置文件失败: %w", err)
	}

	return &siteConfig, nil
}

// setDefaults 设置默认配置值
func setDefaults(config *Config) {
	if config.Spider.Concurrent == 0 {
		config.Spider.Concurrent = 5
	}
	if config.Spider.Delay == 0 {
		config.Spider.Delay = 1000
	}
	if config.Spider.UserAgent == "" {
		config.Spider.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	}
	if config.Spider.Timeout == 0 {
		config.Spider.Timeout = 30
	}
	if config.Spider.Retries == 0 {
		config.Spider.Retries = 3
	}
	if config.Storage.Type == "" {
		config.Storage.Type = "file"
	}
	if config.Storage.OutputDir == "" {
		config.Storage.OutputDir = "./data/output"
	}
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}
	if config.Logging.File == "" {
		config.Logging.File = "./data/logs/crawler.log"
	}
	
	// Web服务器默认配置
	if config.Web.Port == 0 {
		config.Web.Port = 8080
	}
	if config.Web.StaticPath == "" {
		config.Web.StaticPath = "./web/build"
	}
	if config.Web.APIPrefix == "" {
		config.Web.APIPrefix = "/api/v1"
	}
} 