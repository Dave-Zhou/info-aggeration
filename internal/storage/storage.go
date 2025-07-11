package storage

import (
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
	// 移除了文件和Excel存储后，只返回数据库存储
	return NewDatabaseStorage(cfg)
}
