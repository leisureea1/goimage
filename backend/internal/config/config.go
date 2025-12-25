// Package config 提供应用配置管理
// 支持从 YAML 文件加载配置，便于不同环境部署
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 应用全局配置结构
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Storage StorageConfig `yaml:"storage"`
	Auth    AuthConfig    `yaml:"auth"`
	Image   ImageConfig   `yaml:"image"`
}

// ServerConfig HTTP 服务器配置
type ServerConfig struct {
	Port string `yaml:"port"` // 监听端口
	Host string `yaml:"host"` // 监听地址
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type     string `yaml:"type"`      // 存储类型: local, s3, oss 等
	BasePath string `yaml:"base_path"` // 本地存储基础路径
	BaseURL  string `yaml:"base_url"`  // 图片访问基础 URL
}

// AuthConfig 鉴权配置
type AuthConfig struct {
	Enabled bool     `yaml:"enabled"` // 是否启用鉴权
	Tokens  []string `yaml:"tokens"`  // 允许的 API Token 列表
}

// ImageConfig 图片处理配置
type ImageConfig struct {
	Quality      int      `yaml:"quality"`       // WebP 压缩质量 (1-100)
	MaxSize      int64    `yaml:"max_size"`      // 最大上传文件大小 (bytes)
	AllowedTypes []string `yaml:"allowed_types"` // 允许的 MIME 类型
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Storage: StorageConfig{
			Type:     "local",
			BasePath: "./storage/images",
			BaseURL:  "/images",
		},
		Auth: AuthConfig{
			Enabled: false,
			Tokens:  []string{},
		},
		Image: ImageConfig{
			Quality:      75,
			MaxSize:      10 * 1024 * 1024, // 10MB
			AllowedTypes: []string{"image/jpeg", "image/png", "image/webp"},
		},
	}
}

// Load 从 YAML 文件加载配置
// 如果文件不存在，返回默认配置
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 检查配置文件是否存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// 确保存储目录存在
	if cfg.Storage.Type == "local" {
		if err := os.MkdirAll(cfg.Storage.BasePath, 0755); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// GetAbsStoragePath 获取存储路径的绝对路径
func (c *Config) GetAbsStoragePath() (string, error) {
	return filepath.Abs(c.Storage.BasePath)
}
