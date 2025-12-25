// 图床系统后端入口
// 提供图片上传、管理、访问的 RESTful API
package main

import (
	"flag"
	"fmt"
	"log"

	"image-hosting/internal/config"
	"image-hosting/internal/handler"
	"image-hosting/internal/service"
	"image-hosting/internal/storage"
)

func main() {
	// 命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting image hosting server...")
	log.Printf("Storage type: %s", cfg.Storage.Type)
	log.Printf("Storage path: %s", cfg.Storage.BasePath)
	log.Printf("Auth enabled: %v", cfg.Auth.Enabled)

	// 初始化存储
	var store storage.Storage
	switch cfg.Storage.Type {
	case "local":
		store, err = storage.NewLocalStorage(cfg.Storage.BasePath, cfg.Storage.BaseURL)
		if err != nil {
			log.Fatalf("Failed to create local storage: %v", err)
		}
	default:
		log.Fatalf("Unsupported storage type: %s", cfg.Storage.Type)
	}

	// 初始化服务
	imageService, err := service.NewImageService(cfg, store)
	if err != nil {
		log.Fatalf("Failed to create image service: %v", err)
	}

	// 设置路由
	router := handler.SetupRouter(cfg, store, imageService)

	// 启动服务器
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server listening on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
