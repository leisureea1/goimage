// Package model 定义数据模型
// 这些模型用于 API 响应和内部数据传递
package model

import "time"

// Image 图片信息模型
// 包含图片的所有元数据，用于 API 响应
type Image struct {
	ID             string    `json:"id"`               // 图片唯一标识 (UUID)
	URL            string    `json:"url"`              // 图片访问 URL
	OriginalFormat string    `json:"original_format"`  // 原始格式 (jpeg/png/webp)
	OriginalSize   int64     `json:"original_size"`    // 原始文件大小 (bytes)
	ProcessedSize  int64     `json:"processed_size"`   // 处理后文件大小 (bytes)
	Width          int       `json:"width"`            // 图片宽度
	Height         int       `json:"height"`           // 图片高度
	CreatedAt      time.Time `json:"created_at"`       // 上传时间
	Filename       string    `json:"filename"`         // 存储文件名
	StoragePath    string    `json:"-"`                // 存储路径 (不暴露给前端)
}

// ImageListItem 图片列表项
// 用于列表展示，包含必要的展示信息
type ImageListItem struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	ThumbnailURL  string    `json:"thumbnail_url,omitempty"` // 缩略图 URL (预留)
	OriginalFormat string   `json:"original_format"`
	ProcessedSize int64     `json:"processed_size"`
	Width         int       `json:"width"`
	Height        int       `json:"height"`
	CreatedAt     time.Time `json:"created_at"`
}

// UploadResult 上传结果
// 上传成功后返回的完整信息
type UploadResult struct {
	ID             string    `json:"id"`
	URL            string    `json:"url"`
	OriginalFormat string    `json:"original_format"`
	OriginalSize   int64     `json:"original_size"`
	ProcessedSize  int64     `json:"processed_size"`
	Width          int       `json:"width"`
	Height         int       `json:"height"`
	CreatedAt      time.Time `json:"created_at"`
}
