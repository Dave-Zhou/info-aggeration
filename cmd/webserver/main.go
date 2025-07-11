package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/m/v2/internal/api"
	"example.com/m/v2/internal/config"
	"example.com/m/v2/internal/database"
	"example.com/m/v2/internal/utils"
	"github.com/gin-gonic/gin"
)

var (
	configPath = flag.String("config", "config/config.yaml", "配置文件路径")
	port       = flag.String("port", "8080", "服务器端口")
)

func main() {
	flag.Parse()

	// 初始化日志
	logger := utils.NewLogger()
	logger.Info("启动Web服务器...")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 连接数据库
	db, err := database.NewConnection(cfg.Storage.Database)
	if err != nil {
		log.Fatalf("连接数据库失败: %v", err)
	}
	defer db.Close()

	// 初始化数据库表
	if err := database.InitTables(db); err != nil {
		log.Fatalf("初始化数据库表失败: %v", err)
	}

	// 设置Gin模式，与日志级别关联
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := gin.New()

	// 添加中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(api.CORSMiddleware())
	router.Use(api.LoggerMiddleware(logger))

	// 静态文件服务
	router.Static("/static", "./web/build/static")
	router.StaticFile("/", "./web/build/index.html")
	router.StaticFile("/favicon.ico", "./web/build/favicon.ico")

	// API路由
	apiGroup := router.Group("/api/v1")
	api.SetupRoutes(apiGroup, db, logger, cfg)

	// 启动服务器
	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// 优雅关闭
	go func() {
		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		<-sigterm

		logger.Info("正在关闭Web服务器...")
		if err := server.Close(); err != nil {
			logger.Error("关闭服务器失败", "error", err)
		}
	}()

	logger.Info("Web服务器启动", "port", *port)
	logger.Info("访问地址", "url", "http://localhost:"+*port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("启动服务器失败: %v", err)
	}

	logger.Info("Web服务器已关闭")
}
