package middleware

import (
	"net/http"
	"strings"

	"auth-service/pkg/jwt"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
// 接收JWT密钥作为参数，从配置中传入
func JWTAuth(jwtSecret string) gin.HandlerFunc {
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
		claims, err := jwt.ParseToken(parts[1], jwtSecret) // 使用从配置传入的密钥
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的令牌: " + err.Error()})
			c.Abort()
			return
		}

		// 将用户信息存入上下文，供后续处理使用
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username) // 可选：也可以存储用户名
		c.Next()
	}
}
