package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/crawler"
	"example.com/m/v2/internal/storage"
	"example.com/m/v2/internal/utils"
)

func main() {
	// 初始化日志
	logger := utils.NewLogger()
	logger.Info("启动爬虫程序...")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化存储
	store, err := storage.NewStorage(cfg.Storage)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	// 创建爬虫实例
	spider := crawler.NewSpider(cfg, store, logger)

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动爬虫
	go func() {
		if err := spider.Start(); err != nil {
			logger.Error("爬虫启动失败", "error", err)
			os.Exit(1)
		}
	}()

	// 等待信号
	<-sigChan
	logger.Info("收到停止信号，正在关闭爬虫...")

	// 优雅停止
	if err := spider.Stop(); err != nil {
		logger.Error("爬虫停止失败", "error", err)
	}

	fmt.Println("爬虫程序已停止")
} 