// Package handler 处理 HTTP 请求
// 只负责请求解析和响应格式化，不包含业务逻辑
package handler

import (
	"net/http"
	"strconv"

	"image-hosting/internal/model"
	"image-hosting/internal/service"

	"github.com/gin-gonic/gin"
)

// ImageHandler 图片相关 HTTP 处理器
type ImageHandler struct {
	imageService *service.ImageService
}

// NewImageHandler 创建图片处理器
func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// Upload 处理图片上传请求
// POST /api/v1/upload
// Content-Type: multipart/form-data
// 表单字段: file (图片文件)
func (h *ImageHandler) Upload(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"failed to get uploaded file: "+err.Error(),
		))
		return
	}
	defer file.Close()

	// 调用 service 处理上传
	result, err := h.imageService.Upload(c.Request.Context(), file, header.Size)
	if err != nil {
		// 根据错误类型返回不同的错误码
		code := model.CodeInternalError
		status := http.StatusInternalServerError

		errMsg := err.Error()
		if contains(errMsg, "invalid file type") {
			code = model.CodeInvalidFileType
			status = http.StatusBadRequest
		} else if contains(errMsg, "file too large") {
			code = model.CodeFileTooLarge
			status = http.StatusBadRequest
		} else if contains(errMsg, "failed to process") {
			code = model.CodeProcessingFailed
			status = http.StatusInternalServerError
		} else if contains(errMsg, "failed to save") {
			code = model.CodeStorageFailed
			status = http.StatusInternalServerError
		}

		c.JSON(status, model.NewErrorResponse(code, errMsg))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// List 获取图片列表
// GET /api/v1/images?page=1&page_size=20
func (h *ImageHandler) List(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// 调用 service 获取列表
	result, err := h.imageService.ListImages(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// Get 获取单张图片信息
// GET /api/v1/image/:id
func (h *ImageHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"image id is required",
		))
		return
	}

	// 调用 service 获取图片
	img, err := h.imageService.GetImage(c.Request.Context(), id)
	if err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(
				model.CodeNotFound,
				err.Error(),
			))
			return
		}

		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(img))
}

// Delete 删除图片
// DELETE /api/v1/image/:id
func (h *ImageHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(
			model.CodeBadRequest,
			"image id is required",
		))
		return
	}

	// 调用 service 删除图片
	err := h.imageService.DeleteImage(c.Request.Context(), id)
	if err != nil {
		if contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(
				model.CodeNotFound,
				err.Error(),
			))
			return
		}

		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(
			model.CodeInternalError,
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(nil))
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsImpl(s, substr))
}

func containsImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
