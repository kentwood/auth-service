package router

import (
	"auth-service/internal/api/handler"
	"auth-service/internal/api/handler/middleware"

	"github.com/gin-gonic/gin"
)

// Setup 注册所有路由
func Setup(r *gin.Engine,
	authHandler *handler.AuthHandler,
	oauth2Handler *handler.OAuth2Handler) {
	// 应用全局安全中间件
	r.Use(middleware.SecurityHeaders()) // 安全头部中间件
	r.Use(middleware.HTTPSOnly())       // 强制HTTPS中间件

	// 公开路由（无需登录）
	public := r.Group("/auth/v1")
	public.Use(middleware.NoCache()) // 认证相关接口不缓存
	{
		public.POST("/login", authHandler.Login)       // 登录
		public.POST("/register", authHandler.Register) // 注册

		// OAuth2 认证路由
		public.GET("/oauth2/github/login", oauth2Handler.GitHubLogin)
		public.GET("/oauth2/github/callback", oauth2Handler.GitHubCallback)
	}

	// 需认证的路由（JWT 验证）
	protected := r.Group("/auth")
	protected.Use(middleware.JWTAuth()) // 应用 JWT 中间件
	{
		protected.GET("/user/me", authHandler.GetCurrentUser) // 获取当前用户信息
	}
}
