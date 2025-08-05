package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"auth-service/internal/domain/user"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
	"auth-service/pkg/oauth2"
)

// OAuth2Handler OAuth2 认证处理器
type OAuth2Handler struct {
	userService  *user.Service
	jwtSecret    string
	logger       *logger.ZapLogger
	githubOAuth2 *oauth2.GitHubOAuth2Service
}

// NewOAuth2Handler 创建 OAuth2 处理器实例
func NewOAuth2Handler(userService *user.Service, jwtSecret string, logger *logger.ZapLogger, githubOAuth2 *oauth2.GitHubOAuth2Service) *OAuth2Handler {
	return &OAuth2Handler{
		userService:  userService,
		jwtSecret:    jwtSecret,
		logger:       logger,
		githubOAuth2: githubOAuth2,
	}
}

// GitHubLogin 发起 GitHub OAuth2 登录
// @Summary GitHub OAuth2 登录
// @Description 重定向到 GitHub 进行 OAuth2 认证
// @Tags oauth2
// @Accept json
// @Produce json
// @Success 302 {string} string "重定向到 GitHub"
// @Router /auth/oauth2/github/login [get]
func (h *OAuth2Handler) GitHubLogin(c *gin.Context) {
	// 生成随机状态码防止 CSRF 攻击
	state, err := h.generateRandomState()
	if err != nil {
		h.logger.Error("生成状态码失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	// 将状态码存储在会话中
	c.SetCookie("oauth_state", state, 600, "/", "", false, true) // 10分钟有效期

	// 获取授权 URL 并重定向
	authURL := h.githubOAuth2.GetAuthURL(state)
	h.logger.Info("发起 GitHub OAuth2 登录",
		zap.String("client_ip", c.ClientIP()),
		zap.String("user_agent", c.GetHeader("User-Agent")),
	)

	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GitHubCallback 处理 GitHub OAuth2 回调
// @Summary GitHub OAuth2 回调
// @Description 处理 GitHub OAuth2 认证回调
// @Tags oauth2
// @Accept json
// @Produce json
// @Param code query string true "授权码"
// @Param state query string true "状态码"
// @Success 200 {object} gin.H{token:string, user_id:uint, username:string, auth_type:string}
// @Failure 400 {object} gin.H{error:string}
// @Failure 401 {object} gin.H{error:string}
// @Failure 500 {object} gin.H{error:string}
// @Router /auth/oauth2/github/callback [get]
func (h *OAuth2Handler) GitHubCallback(c *gin.Context) {
	// 1. 验证状态码
	code := c.Query("code")
	state := c.Query("state")
	storedState, err := c.Cookie("oauth_state")

	if err != nil || state != storedState {
		h.logger.Warn("OAuth2 状态码验证失败",
			zap.String("received_state", state),
			zap.String("stored_state", storedState),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的状态码"})
		return
	}

	// 清除状态码 cookie
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	if code == "" {
		h.logger.Warn("OAuth2 回调缺少授权码", zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少授权码"})
		return
	}

	// 2. 交换授权码获取用户信息
	githubUser, err := h.githubOAuth2.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("GitHub OAuth2 授权失败",
			zap.String("code", code),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub 授权失败"})
		return
	}

	// 3. 登录或注册用户
	u, err := h.userService.LoginWithGitHub(githubUser)
	if err != nil {
		h.logger.Error("GitHub 用户登录失败",
			zap.Int64("github_id", githubUser.ID),
			zap.String("github_login", githubUser.Login),
			zap.String("github_email", githubUser.Email),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败"})
		return
	}

	// 4. 生成 JWT 令牌
	token, err := jwt.GenerateToken(u.ID, u.Username, h.jwtSecret, 24*time.Hour)
	if err != nil {
		h.logger.Error("生成 JWT 令牌失败",
			zap.Uint("user_id", u.ID),
			zap.String("username", u.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}

	// 5. 记录登录成功日志
	h.logger.Info("GitHub OAuth2 登录成功",
		zap.Uint("user_id", u.ID),
		zap.String("username", u.Username),
		zap.String("auth_type", u.AuthType),
		zap.String("github_login", githubUser.Login),
		zap.String("client_ip", c.ClientIP()),
	)

	// 6. 返回登录结果
	c.JSON(http.StatusOK, gin.H{
		"message":    "登录成功",
		"token":      token,
		"user_id":    u.ID,
		"username":   u.Username,
		"auth_type":  u.AuthType,
		"avatar_url": u.AvatarURL,
	})
}

// generateRandomState 生成随机状态码
func (h *OAuth2Handler) generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
