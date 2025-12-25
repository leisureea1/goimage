package handler

import (
	"image-hosting/internal/config"
	"image-hosting/internal/middleware"
	"image-hosting/internal/service"
	"image-hosting/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置并返回 Gin 路由器
// 集中管理所有路由和中间件配置
func SetupRouter(cfg *config.Config, store storage.Storage, imageService *service.ImageService) *gin.Engine {
	// 生产环境使用 release 模式
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggerMiddleware())

	// CORS 配置 - 允许前端跨域访问
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // 生产环境应限制为具体域名
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 静态文件服务 - 提供图片访问
	// 将 /images 路径映射到存储目录
	if ls, ok := store.(*storage.LocalStorage); ok {
		r.Static("/images", ls.GetBasePath())
	}

	// 创建 Handler
	imageHandler := NewImageHandler(imageService)

	// API 路由组
	api := r.Group("/api/v1")
	{
		// 应用鉴权中间件
		api.Use(middleware.AuthMiddleware(&cfg.Auth))

		// 图片上传
		api.POST("/upload", imageHandler.Upload)

		// 图片列表
		api.GET("/images", imageHandler.List)

		// 单张图片信息
		api.GET("/image/:id", imageHandler.Get)

		// 删除图片
		api.DELETE("/image/:id", imageHandler.Delete)
	}

	// 健康检查接口 (不需要鉴权)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return r
}
