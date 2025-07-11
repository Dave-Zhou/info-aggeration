package storage

import (
	"fmt"
	"example.com/m/v2/internal/config"
	"example.com/m/v2/pkg/models"
)

// Storage 存储接口
type Storage interface {
	Save(item *models.Item) error
	SaveBatch(items []*models.Item) error
	Close() error
}

// NewStorage 创建存储实例
func NewStorage(cfg config.StorageConfig) (Storage, error) {
	switch cfg.Type {
	case "file":
		return NewFileStorage(cfg)
	case "database":
		return NewDatabaseStorage(cfg)
	case "excel":
		return NewExcelStorage(cfg)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", cfg.Type)
	}
} 