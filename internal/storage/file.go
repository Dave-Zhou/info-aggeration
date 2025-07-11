package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/pkg/models"
)

// FileStorage 文件存储实现
type FileStorage struct {
	config    config.StorageConfig
	outputDir string
	mu        sync.Mutex
	file      *os.File
	encoder   *json.Encoder
}

// NewFileStorage 创建文件存储实例
func NewFileStorage(cfg config.StorageConfig) (*FileStorage, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建输出文件
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("crawl_data_%s.json", timestamp)
	filePath := filepath.Join(cfg.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %w", err)
	}

	// 写入JSON数组开始标记
	if _, err := file.WriteString("[\n"); err != nil {
		file.Close()
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	storage := &FileStorage{
		config:    cfg,
		outputDir: cfg.OutputDir,
		file:      file,
		encoder:   json.NewEncoder(file),
	}

	// 设置JSON编码器格式
	storage.encoder.SetIndent("", "  ")

	return storage, nil
}

// Save 保存单个数据项
func (fs *FileStorage) Save(item *models.Item) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// 检查文件是否已关闭
	if fs.file == nil {
		return fmt.Errorf("文件存储已关闭")
	}

	// 写入逗号分隔符（除了第一个项目）
	if fs.needComma() {
		if _, err := fs.file.WriteString(",\n"); err != nil {
			return fmt.Errorf("写入分隔符失败: %w", err)
		}
	}

	// 编码并写入JSON数据
	if err := fs.encoder.Encode(item); err != nil {
		return fmt.Errorf("编码JSON失败: %w", err)
	}

	return nil
}

// SaveBatch 批量保存数据项
func (fs *FileStorage) SaveBatch(items []*models.Item) error {
	for _, item := range items {
		if err := fs.Save(item); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭文件存储
func (fs *FileStorage) Close() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.file == nil {
		return nil
	}

	// 写入JSON数组结束标记
	if _, err := fs.file.WriteString("\n]"); err != nil {
		return fmt.Errorf("写入结束标记失败: %w", err)
	}

	// 关闭文件
	if err := fs.file.Close(); err != nil {
		return fmt.Errorf("关闭文件失败: %w", err)
	}

	fs.file = nil
	return nil
}

// needComma 检查是否需要逗号分隔符
func (fs *FileStorage) needComma() bool {
	// 获取文件当前位置
	pos, err := fs.file.Seek(0, 1) // 相对于当前位置偏移0
	if err != nil {
		return false
	}
	
	// 如果文件位置大于2（"[\n"的长度），说明已经有数据
	return pos > 2
}

// GetOutputPath 获取输出文件路径
func (fs *FileStorage) GetOutputPath() string {
	if fs.file == nil {
		return ""
	}
	return fs.file.Name()
}

// CSVFileStorage CSV文件存储实现
type CSVFileStorage struct {
	config    config.StorageConfig
	outputDir string
	mu        sync.Mutex
	file      *os.File
	isFirstRow bool
}

// NewCSVFileStorage 创建CSV文件存储实例
func NewCSVFileStorage(cfg config.StorageConfig) (*CSVFileStorage, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 创建输出文件
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("crawl_data_%s.csv", timestamp)
	filePath := filepath.Join(cfg.OutputDir, filename)

	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("创建输出文件失败: %w", err)
	}

	storage := &CSVFileStorage{
		config:     cfg,
		outputDir:  cfg.OutputDir,
		file:       file,
		isFirstRow: true,
	}

	// 写入CSV标题行
	header := "URL,Title,Content,Description,Author,Source,PublishDate,Timestamp,Keywords,Tags,Links,Images\n"
	if _, err := file.WriteString(header); err != nil {
		file.Close()
		return nil, fmt.Errorf("写入CSV标题失败: %w", err)
	}

	return storage, nil
}

// Save 保存单个数据项到CSV
func (cs *CSVFileStorage) Save(item *models.Item) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.file == nil {
		return fmt.Errorf("CSV文件存储已关闭")
	}

	// 转换为CSV格式
	csvLine := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
		csvEscape(item.URL),
		csvEscape(item.Title),
		csvEscape(item.Content),
		csvEscape(item.Description),
		csvEscape(item.Author),
		csvEscape(item.Source),
		item.PublishDate.Format("2006-01-02 15:04:05"),
		item.Timestamp.Format("2006-01-02 15:04:05"),
		csvEscape(joinStrings(item.Keywords, ";")),
		csvEscape(joinStrings(item.Tags, ";")),
		csvEscape(joinStrings(item.Links, ";")),
		csvEscape(joinStrings(item.Images, ";")),
	)

	// 写入CSV行
	if _, err := cs.file.WriteString(csvLine); err != nil {
		return fmt.Errorf("写入CSV行失败: %w", err)
	}

	return nil
}

// SaveBatch 批量保存数据项到CSV
func (cs *CSVFileStorage) SaveBatch(items []*models.Item) error {
	for _, item := range items {
		if err := cs.Save(item); err != nil {
			return err
		}
	}
	return nil
}

// Close 关闭CSV文件存储
func (cs *CSVFileStorage) Close() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.file == nil {
		return nil
	}

	if err := cs.file.Close(); err != nil {
		return fmt.Errorf("关闭CSV文件失败: %w", err)
	}

	cs.file = nil
	return nil
}

// 辅助函数
func csvEscape(value string) string {
	// 如果包含逗号、引号或换行符，需要用双引号包围
	if containsSpecialChars(value) {
		// 转义内部的双引号
		escaped := ""
		for _, char := range value {
			if char == '"' {
				escaped += "\"\""
			} else {
				escaped += string(char)
			}
		}
		return "\"" + escaped + "\""
	}
	return value
}

func containsSpecialChars(value string) bool {
	for _, char := range value {
		if char == ',' || char == '"' || char == '\n' || char == '\r' {
			return true
		}
	}
	return false
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
} 