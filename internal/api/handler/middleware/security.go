package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPSOnly 强制使用HTTPS中间件
func HTTPSOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在生产环境中，检查是否使用HTTPS
		if gin.Mode() == gin.ReleaseMode {
			if c.GetHeader("X-Forwarded-Proto") != "https" && c.Request.TLS == nil {
				// 重定向到HTTPS
				httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
				c.Redirect(http.StatusMovedPermanently, httpsURL)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// SecurityHeaders 添加安全头部中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 防止页面被嵌入到iframe中（防止点击劫持）
		c.Header("X-Frame-Options", "DENY")

		// 防止MIME类型嗅探
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS保护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 强制HTTPS（在支持的浏览器中）
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 内容安全策略（可根据需要调整）
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// NoCache 防止敏感页面被缓存
func NoCache() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
