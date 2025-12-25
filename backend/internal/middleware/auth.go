// Package middleware 提供 HTTP 中间件
// 包括鉴权、日志、CORS 等通用功能
package middleware

import (
	"net/http"
	"strings"

	"image-hosting/internal/config"
	"image-hosting/internal/model"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 创建鉴权中间件
// 使用 Bearer Token 方式进行 API 鉴权
// 设计为可配置开关，便于开发调试
func AuthMiddleware(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果鉴权未启用，直接放行
		if !cfg.Enabled {
			c.Next()
			return
		}

		// 从 Header 获取 Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				model.CodeUnauthorized,
				"missing authorization header",
			))
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				model.CodeUnauthorized,
				"invalid authorization format, expected: Bearer <token>",
			))
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 Token
		if !validateToken(token, cfg.Tokens) {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(
				model.CodeUnauthorized,
				"invalid token",
			))
			c.Abort()
			return
		}

		// Token 有效，继续处理
		c.Next()
	}
}

// validateToken 验证 Token 是否在允许列表中
func validateToken(token string, allowedTokens []string) bool {
	for _, t := range allowedTokens {
		if t == token {
			return true
		}
	}
	return false
}

// OptionalAuthMiddleware 可选鉴权中间件
// 用于某些接口需要区分已认证和未认证用户的场景
func OptionalAuthMiddleware(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 未提供 Token，标记为未认证用户
			c.Set("authenticated", false)
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.Set("authenticated", false)
			c.Next()
			return
		}

		token := parts[1]
		if validateToken(token, cfg.Tokens) {
			c.Set("authenticated", true)
		} else {
			c.Set("authenticated", false)
		}

		c.Next()
	}
}
