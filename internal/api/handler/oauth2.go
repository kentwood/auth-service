package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"auth-service/internal/config"
	"auth-service/internal/domain/user"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
	"auth-service/pkg/oauth2"
	"auth-service/pkg/redis"
	"auth-service/pkg/session"
)

// OAuth2Handler OAuth2 认证处理器
type OAuth2Handler struct {
	userService    *user.Service
	config         *config.Config
	logger         *logger.ZapLogger
	githubOAuth2   *oauth2.GitHubOAuth2Service
	sessionManager *session.Manager // 新增 Session 管理器
}

// NewOAuth2Handler 创建 OAuth2 处理器实例
func NewOAuth2Handler(userService *user.Service, cfg *config.Config, logger *logger.ZapLogger, githubOAuth2 *oauth2.GitHubOAuth2Service, redisClient *redis.Client) *OAuth2Handler {
	return &OAuth2Handler{
		userService:    userService,
		config:         cfg,
		logger:         logger,
		githubOAuth2:   githubOAuth2,
		sessionManager: session.NewManager(redisClient),
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

	// 创建 OAuth2 会话，将状态信息存储到 Redis
	sessionID, err := h.sessionManager.CreateOAuth2Session(
		c.Request.Context(),
		state,
		c.GetHeader("User-Agent"),
		c.ClientIP(),
	)
	if err != nil {
		h.logger.Error("创建 OAuth2 会话失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器内部错误"})
		return
	}

	// 将 session ID 存储在 httpOnly cookie 中
	c.SetCookie("oauth_session", sessionID, 600, "/", "", false, true) // 10分钟有效期

	// 获取授权 URL 并重定向
	authURL := h.githubOAuth2.GetAuthURL(state)
	h.logger.Info("发起 GitHub OAuth2 登录",
		zap.String("session_id", sessionID),
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
	// 1. 从 cookie 获取 session ID
	sessionID, err := c.Cookie("oauth_session")
	if err != nil {
		h.logger.Warn("OAuth2 回调缺少会话 ID",
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		h.redirectToError(c, "缺少会话信息")
		return
	}

	// 2. 验证状态码
	code := c.Query("code")
	receivedState := c.Query("state")

	if code == "" {
		h.logger.Warn("OAuth2 回调缺少授权码",
			zap.String("session_id", sessionID),
			zap.String("client_ip", c.ClientIP()),
		)
		h.redirectToError(c, "缺少授权码")
		return
	}

	// 3. 验证会话和状态码
	stateInfo, err := h.sessionManager.ValidateOAuth2Session(c.Request.Context(), sessionID, receivedState)
	if err != nil {
		h.logger.Warn("OAuth2 状态码验证失败",
			zap.String("session_id", sessionID),
			zap.String("received_state", receivedState),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		h.redirectToError(c, "状态码验证失败")
		return
	}

	// 4. 删除会话（一次性使用）
	if err := h.sessionManager.DeleteOAuth2Session(c.Request.Context(), sessionID); err != nil {
		h.logger.Warn("删除 OAuth2 会话失败", zap.Error(err))
		// 不影响主流程，只记录警告
	}

	// 5. 清除 cookie
	c.SetCookie("oauth_session", "", -1, "/", "", false, true)

	// 6. 交换授权码获取用户信息
	githubUser, err := h.githubOAuth2.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		h.logger.Error("GitHub OAuth2 授权失败",
			zap.String("code", code),
			zap.String("session_id", sessionID),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		h.redirectToError(c, "GitHub授权失败")
		return
	}

	// 7. 登录或注册用户
	u, err := h.userService.LoginWithGitHub(githubUser)
	if err != nil {
		h.logger.Error("GitHub 用户登录失败",
			zap.Int64("github_id", githubUser.ID),
			zap.String("github_login", githubUser.Login),
			zap.String("github_email", githubUser.Email),
			zap.String("session_id", sessionID),
			zap.String("client_ip", c.ClientIP()),
			zap.Error(err),
		)
		h.redirectToError(c, "用户登录失败")
		return
	}

	// 8. 生成 JWT 令牌
	token, err := jwt.GenerateToken(u.ID, u.Username, h.config.JWT.Secret, 24*time.Hour)
	if err != nil {
		h.logger.Error("生成 JWT 令牌失败",
			zap.Uint("user_id", u.ID),
			zap.String("username", u.Username),
			zap.String("session_id", sessionID),
			zap.Error(err),
		)
		h.redirectToError(c, "生成令牌失败")
		return
	}

	// 9. 记录登录成功日志
	h.logger.Info("GitHub OAuth2 登录成功",
		zap.Uint("user_id", u.ID),
		zap.String("username", u.Username),
		zap.String("auth_type", u.AuthType),
		zap.String("github_login", githubUser.Login),
		zap.String("session_id", sessionID),
		zap.String("stored_ip", stateInfo.ClientIP),
		zap.String("current_ip", c.ClientIP()),
	)

	// 10. 重定向到前端成功页面
	successURL := fmt.Sprintf("%s%s?token=%s&user_id=%d&username=%s&auth_type=%s",
		h.config.UI.BaseURL,
		h.config.UI.LoginSuccessPath,
		token,
		u.ID,
		u.Username,
		u.AuthType,
	)
	c.Redirect(http.StatusTemporaryRedirect, successURL)
}

// redirectToError 重定向到错误页面
func (h *OAuth2Handler) redirectToError(c *gin.Context, message string) {
	errorURL := fmt.Sprintf("%s%s?message=%s",
		h.config.UI.BaseURL,
		h.config.UI.LoginErrorPath,
		message,
	)
	c.Redirect(http.StatusTemporaryRedirect, errorURL)
}

// generateRandomState 生成随机状态码
func (h *OAuth2Handler) generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
