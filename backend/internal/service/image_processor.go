// Package service 提供核心业务逻辑
// 所有图片处理、业务规则都在此层实现
package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
	xwebp "golang.org/x/image/webp"
)

// ImageProcessor 图片处理器
// 负责图片格式转换、压缩、EXIF 处理等
type ImageProcessor struct {
	quality int // WebP 压缩质量 (1-100)
}

// NewImageProcessor 创建图片处理器
func NewImageProcessor(quality int) *ImageProcessor {
	if quality < 1 || quality > 100 {
		quality = 75 // 默认质量
	}
	return &ImageProcessor{quality: quality}
}

// ProcessResult 图片处理结果
type ProcessResult struct {
	Data   []byte // 处理后的图片数据
	Width  int    // 图片宽度
	Height int    // 图片高度
	Format string // 输出格式 (webp)
}

// Process 处理图片
// 1. 解码图片
// 2. 修正 EXIF 方向
// 3. 转换为 WebP 格式
// 4. 压缩
func (p *ImageProcessor) Process(data []byte, mimeType string) (*ProcessResult, error) {
	// 解码图片
	img, format, err := p.decodeImage(data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// 修正 EXIF 方向 (仅 JPEG)
	if format == "jpeg" {
		img = p.fixOrientation(data, img)
	}

	// 编码为 WebP
	webpData, err := p.encodeWebP(img)
	if err != nil {
		return nil, fmt.Errorf("failed to encode webp: %w", err)
	}

	bounds := img.Bounds()
	return &ProcessResult{
		Data:   webpData,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
		Format: "webp",
	}, nil
}

// decodeImage 解码图片
// 支持 JPEG、PNG、WebP 格式
func (p *ImageProcessor) decodeImage(data []byte, mimeType string) (image.Image, string, error) {
	reader := bytes.NewReader(data)

	switch mimeType {
	case "image/jpeg":
		img, err := jpeg.Decode(reader)
		return img, "jpeg", err
	case "image/png":
		img, err := png.Decode(reader)
		return img, "png", err
	case "image/webp":
		img, err := xwebp.Decode(reader)
		return img, "webp", err
	default:
		return nil, "", fmt.Errorf("unsupported image type: %s", mimeType)
	}
}

// fixOrientation 根据 EXIF 信息修正图片方向
// JPEG 图片可能包含方向信息，需要根据此信息旋转图片
func (p *ImageProcessor) fixOrientation(data []byte, img image.Image) image.Image {
	reader := bytes.NewReader(data)
	x, err := exif.Decode(reader)
	if err != nil {
		return img // 无 EXIF 信息，返回原图
	}

	orientTag, err := x.Get(exif.Orientation)
	if err != nil {
		return img // 无方向信息，返回原图
	}

	orient, err := orientTag.Int(0)
	if err != nil {
		return img
	}

	// 根据 EXIF 方向值旋转图片
	// 方向值定义参考: https://exiftool.org/TagNames/EXIF.html
	switch orient {
	case 2:
		return imaging.FlipH(img)
	case 3:
		return imaging.Rotate180(img)
	case 4:
		return imaging.FlipV(img)
	case 5:
		return imaging.Transpose(img)
	case 6:
		return imaging.Rotate270(img)
	case 7:
		return imaging.Transverse(img)
	case 8:
		return imaging.Rotate90(img)
	default:
		return img
	}
}

// encodeWebP 将图片编码为 WebP 格式
// 使用 chai2010/webp 库进行真正的 WebP 编码
func (p *ImageProcessor) encodeWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer

	// 使用 webp 库编码，支持质量参数
	err := webp.Encode(&buf, img, &webp.Options{
		Lossless: false,
		Quality:  float32(p.quality),
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ValidateMimeType 验证 MIME 类型是否允许
func ValidateMimeType(mimeType string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if t == mimeType {
			return true
		}
	}
	return false
}

// DetectMimeType 检测文件的真实 MIME 类型
// 通过读取文件头部字节来判断，而非依赖文件扩展名
func DetectMimeType(reader io.Reader) (string, []byte, error) {
	// 读取文件头部用于检测
	header := make([]byte, 512)
	n, err := reader.Read(header)
	if err != nil && err != io.EOF {
		return "", nil, err
	}
	header = header[:n]

	// 检测 MIME 类型
	mimeType := detectMimeFromHeader(header)

	return mimeType, header, nil
}

// detectMimeFromHeader 从文件头检测 MIME 类型
func detectMimeFromHeader(header []byte) string {
	if len(header) < 4 {
		return "application/octet-stream"
	}

	// JPEG: FF D8 FF
	if header[0] == 0xFF && header[1] == 0xD8 && header[2] == 0xFF {
		return "image/jpeg"
	}

	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(header) >= 8 &&
		header[0] == 0x89 && header[1] == 0x50 && header[2] == 0x4E && header[3] == 0x47 &&
		header[4] == 0x0D && header[5] == 0x0A && header[6] == 0x1A && header[7] == 0x0A {
		return "image/png"
	}

	// WebP: RIFF....WEBP
	if len(header) >= 12 &&
		header[0] == 0x52 && header[1] == 0x49 && header[2] == 0x46 && header[3] == 0x46 &&
		header[8] == 0x57 && header[9] == 0x45 && header[10] == 0x42 && header[11] == 0x50 {
		return "image/webp"
	}

	return "application/octet-stream"
}
