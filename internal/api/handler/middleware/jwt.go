package middleware

import (
	"net/http"
	"strings"

	"auth-service/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取令牌
		authHeader := c.Request.Header.Get("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未提供令牌"})
			c.Abort()
			return
		}

		// 检查令牌格式（Bearer <token>）
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "令牌格式错误"})
			c.Abort()
			return
		}

		// 验证令牌并解析用户ID
		claims, err := jwt.ParseToken(parts[1], "your_jwt_secret") // 实际应从配置获取密钥
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌: " + err.Error()})
			c.Abort()
			return
		}

		// 将用户ID存入上下文，供后续处理使用
		c.Set("userID", claims.UserID)
		c.Next()
	}
}
