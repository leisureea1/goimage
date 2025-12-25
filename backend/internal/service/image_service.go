package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"image-hosting/internal/config"
	"image-hosting/internal/model"
	"image-hosting/internal/storage"

	"github.com/google/uuid"
)

// ImageService 图片服务
// 处理所有图片相关的业务逻辑
type ImageService struct {
	storage   storage.Storage
	processor *ImageProcessor
	config    *config.Config
	metadata  *MetadataStore
}

// MetadataStore 图片元数据存储
// 使用 JSON 文件存储元数据，便于简单部署
// 生产环境建议替换为数据库
type MetadataStore struct {
	mu       sync.Mutex  // 改用互斥锁，确保读写串行
	images   map[string]*model.Image
	filePath string
}

// NewMetadataStore 创建元数据存储
func NewMetadataStore(basePath string) (*MetadataStore, error) {
	filePath := filepath.Join(basePath, "metadata.json")
	store := &MetadataStore{
		images:   make(map[string]*model.Image),
		filePath: filePath,
	}

	// 尝试加载已有数据
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// load 从文件加载元数据（内部方法，调用前需持有锁）
func (s *MetadataStore) loadLocked() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.images)
}

// load 从文件加载元数据
func (s *MetadataStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// saveLocked 保存元数据到文件（内部方法，调用前需持有锁）
func (s *MetadataStore) saveLocked() error {
	data, err := json.MarshalIndent(s.images, "", "  ")
	if err != nil {
		return err
	}

	// 先写入临时文件，再原子替换，防止写入中断导致数据损坏
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	// 原子替换
	if err := os.Rename(tmpFile, s.filePath); err != nil {
		os.Remove(tmpFile) // 清理临时文件
		return err
	}

	return nil
}

// Add 添加图片元数据
func (s *MetadataStore) Add(img *model.Image) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.images[img.ID] = img
	return s.saveLocked()
}

// Delete 删除图片元数据
func (s *MetadataStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.images, id)
	return s.saveLocked()
}

// List 列出所有图片
func (s *MetadataStore) List() []*model.Image {
	s.mu.Lock()
	defer s.mu.Unlock()

	images := make([]*model.Image, 0, len(s.images))
	for _, img := range s.images {
		// 复制一份，避免外部修改
		imgCopy := *img
		images = append(images, &imgCopy)
	}

	// 按创建时间倒序排列
	sort.Slice(images, func(i, j int) bool {
		return images[i].CreatedAt.After(images[j].CreatedAt)
	})

	return images
}

// Get 获取图片元数据
func (s *MetadataStore) Get(id string) (*model.Image, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	img, ok := s.images[id]
	if !ok {
		return nil, false
	}
	// 返回副本
	imgCopy := *img
	return &imgCopy, true
}

// Count 获取图片总数
func (s *MetadataStore) Count() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return int64(len(s.images))
}

// Reload 重新从文件加载数据
func (s *MetadataStore) Reload() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loadLocked()
}

// NewImageService 创建图片服务
func NewImageService(cfg *config.Config, store storage.Storage) (*ImageService, error) {
	// 获取存储基础路径用于元数据存储
	basePath := cfg.Storage.BasePath
	if ls, ok := store.(*storage.LocalStorage); ok {
		basePath = ls.GetBasePath()
	}

	metadata, err := NewMetadataStore(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create metadata store: %w", err)
	}

	return &ImageService{
		storage:   store,
		processor: NewImageProcessor(cfg.Image.Quality),
		config:    cfg,
		metadata:  metadata,
	}, nil
}

// Upload 上传并处理图片
// 完整流程: 验证 -> 处理 -> 存储 -> 记录元数据
func (s *ImageService) Upload(ctx context.Context, file io.Reader, originalSize int64) (*model.UploadResult, error) {
	// 1. 读取文件内容
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// 2. 检测并验证 MIME 类型
	mimeType := detectMimeFromHeader(data)
	if !ValidateMimeType(mimeType, s.config.Image.AllowedTypes) {
		return nil, fmt.Errorf("invalid file type: %s", mimeType)
	}

	// 3. 检查文件大小
	if int64(len(data)) > s.config.Image.MaxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", len(data), s.config.Image.MaxSize)
	}

	// 4. 处理图片 (EXIF 修正 + WebP 转换 + 压缩)
	result, err := s.processor.Process(data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to process image: %w", err)
	}

	// 5. 生成存储路径 (年/月/uuid.webp)
	now := time.Now()
	id := uuid.New().String()
	filename := fmt.Sprintf("%s.webp", id) // WebP 格式输出
	storagePath := fmt.Sprintf("%d/%02d/%s", now.Year(), now.Month(), filename)

	// 6. 保存文件
	url, err := s.storage.Save(ctx, storagePath, bytes.NewReader(result.Data))
	if err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// 7. 提取原始格式
	originalFormat := mimeTypeToFormat(mimeType)

	// 8. 创建图片记录
	img := &model.Image{
		ID:             id,
		URL:            url,
		OriginalFormat: originalFormat,
		OriginalSize:   originalSize,
		ProcessedSize:  int64(len(result.Data)),
		Width:          result.Width,
		Height:         result.Height,
		CreatedAt:      now,
		Filename:       filename,
		StoragePath:    storagePath,
	}

	// 9. 保存元数据
	if err := s.metadata.Add(img); err != nil {
		// 元数据保存失败，删除已上传的文件
		s.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return &model.UploadResult{
		ID:             img.ID,
		URL:            img.URL,
		OriginalFormat: img.OriginalFormat,
		OriginalSize:   img.OriginalSize,
		ProcessedSize:  img.ProcessedSize,
		Width:          img.Width,
		Height:         img.Height,
		CreatedAt:      img.CreatedAt,
	}, nil
}

// GetImage 获取单张图片信息
func (s *ImageService) GetImage(ctx context.Context, id string) (*model.Image, error) {
	img, ok := s.metadata.Get(id)
	if !ok {
		return nil, fmt.Errorf("image not found: %s", id)
	}
	return img, nil
}

// ListImages 获取图片列表
func (s *ImageService) ListImages(ctx context.Context, page, pageSize int) (*model.PaginatedList, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	allImages := s.metadata.List()
	total := int64(len(allImages))

	// 计算分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= len(allImages) {
		start = len(allImages)
	}
	if end > len(allImages) {
		end = len(allImages)
	}

	// 转换为列表项
	items := make([]model.ImageListItem, 0, end-start)
	for _, img := range allImages[start:end] {
		items = append(items, model.ImageListItem{
			ID:             img.ID,
			URL:            img.URL,
			OriginalFormat: img.OriginalFormat,
			ProcessedSize:  img.ProcessedSize,
			Width:          img.Width,
			Height:         img.Height,
			CreatedAt:      img.CreatedAt,
		})
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &model.PaginatedList{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteImage 删除图片
func (s *ImageService) DeleteImage(ctx context.Context, id string) error {
	img, ok := s.metadata.Get(id)
	if !ok {
		return fmt.Errorf("image not found: %s", id)
	}

	// 验证存储路径是否有效
	// 有效路径格式: 年/月/文件名.webp 或 年/月/文件名.jpg
	storagePath := img.StoragePath
	if storagePath == "" {
		// 如果 StoragePath 为空，尝试从 URL 推断
		// URL 格式: /images/年/月/文件名.webp
		if img.URL != "" {
			// 移除 /images/ 前缀
			const prefix = "/images/"
			if len(img.URL) > len(prefix) && img.URL[:len(prefix)] == prefix {
				storagePath = img.URL[len(prefix):]
			}
		}
	}

	// 再次验证路径
	if storagePath == "" || storagePath == "/" || !isValidStoragePath(storagePath) {
		// 路径无效，只删除元数据，跳过文件删除
		if err := s.metadata.Delete(id); err != nil {
			return fmt.Errorf("failed to delete metadata: %w", err)
		}
		return nil
	}

	// 删除文件
	if err := s.storage.Delete(ctx, storagePath); err != nil {
		// 如果文件不存在，继续删除元数据
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to delete file: %w", err)
		}
	}

	// 删除元数据
	if err := s.metadata.Delete(id); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	return nil
}

// isValidStoragePath 验证存储路径是否有效
// 有效路径应该包含文件名，格式如: 2024/12/uuid.webp
func isValidStoragePath(path string) bool {
	if path == "" {
		return false
	}
	// 路径应该包含至少一个 / 和文件扩展名
	parts := filepath.SplitList(path)
	if len(parts) == 0 {
		return false
	}
	// 检查是否有文件扩展名
	ext := filepath.Ext(path)
	return ext == ".webp" || ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

// mimeTypeToFormat 将 MIME 类型转换为格式名称
func mimeTypeToFormat(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return "jpeg"
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "unknown"
	}
}
