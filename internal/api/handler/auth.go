package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"auth-service/internal/domain/user"
	"auth-service/pkg/jwt"
	"auth-service/pkg/logger"
)

// LoginRequest 登录请求参数结构体
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"` // 用户名验证规则
	Password string `json:"password" binding:"required,min=6"`        // 密码验证规则
}

// RegisterRequest 注册请求参数结构体
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"` // 邮箱格式验证
}

// UserResponse 用户信息响应结构体
type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthHandler 认证处理器
type AuthHandler struct {
	userService *user.Service     // 依赖用户服务层
	jwtSecret   string            // JWT签名密钥
	logger      *logger.ZapLogger // 日志记录器
}

// NewAuthHandler 创建认证处理器实例
func NewAuthHandler(userService *user.Service, jwtSecret string, logger *logger.ZapLogger) *AuthHandler {
	return &AuthHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		logger:      logger,
	}
}

// Login 处理用户登录请求
// @Summary 用户登录
// @Description 通过用户名和密码获取JWT令牌
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} gin.H{token:string, user_id:uint, username:string}
// @Failure 400 {object} gin.H{error:string}
// @Failure 401 {object} gin.H{error:string}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	// 绑定并验证请求参数
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 调用服务层查询用户
	u, err := h.userService.GetByUsername(req.Username)
	if err != nil {
		// 记录详细错误信息到日志（便于调试）
		h.logger.Warn("用户登录失败：查询用户时发生错误",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		// 统一返回认证失败，避免泄露用户是否存在
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if !u.CheckPassword(req.Password) {
		// 记录密码验证失败（安全审计）
		h.logger.Warn("用户登录失败：密码错误",
			zap.String("username", req.Username),
			zap.Uint("user_id", u.ID),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成JWT令牌（有效期24小时）
	token, err := jwt.GenerateToken(u.ID, u.Username, h.jwtSecret, 24*time.Hour)
	if err != nil {
		// 记录令牌生成失败的详细错误
		h.logger.Error("JWT令牌生成失败",
			zap.String("username", req.Username),
			zap.Uint("user_id", u.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"}) // 对用户隐藏具体错误
		return
	}

	// 记录成功登录日志
	h.logger.Info("用户登录成功",
		zap.String("username", req.Username),
		zap.Uint("user_id", u.ID),
		zap.String("client_ip", c.ClientIP()),
	)

	// 返回登录结果
	c.JSON(http.StatusOK, gin.H{
		"token":    token,
		"user_id":  u.ID,
		"username": u.Username,
	})
}

// Register 处理用户注册请求
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册参数"
// @Success 201 {object} gin.H{message:string, user:UserResponse}
// @Failure 400 {object} gin.H{error:string}
// @Failure 500 {object} gin.H{error:string}
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录参数验证失败
		h.logger.Warn("用户注册失败：请求参数无效",
			zap.String("username", req.Username),
			zap.Error(err),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数: " + err.Error()})
		return
	}

	// 调用服务层注册用户
	newUser, err := h.userService.Register(&user.User{
		Username: req.Username,
		Password: req.Password, // 服务层会自动加密
		Email:    req.Email,
	})
	if err != nil {
		// 根据错误类型返回具体信息并记录日志
		switch err {
		case user.ErrUsernameExists:
			h.logger.Warn("用户注册失败：用户名已存在",
				zap.String("username", req.Username),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已被注册"})
		case user.ErrEmailExists:
			h.logger.Warn("用户注册失败：邮箱已存在",
				zap.String("username", req.Username),
				zap.String("email", req.Email),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "邮箱已被注册"})
		default:
			h.logger.Error("用户注册失败：未知错误",
				zap.String("username", req.Username),
				zap.String("email", req.Email),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"}) // 隐藏具体错误
		}
		return
	}

	// 记录成功注册日志
	h.logger.Info("用户注册成功",
		zap.String("username", req.Username),
		zap.String("email", req.Email),
		zap.Uint("user_id", newUser.ID),
		zap.String("client_ip", c.ClientIP()),
	)

	// 构造响应数据（过滤敏感字段）
	c.JSON(http.StatusCreated, gin.H{
		"message": "注册成功",
		"user": UserResponse{
			ID:        newUser.ID,
			Username:  newUser.Username,
			Email:     newUser.Email,
			CreatedAt: newUser.CreatedAt,
		},
	})
}

// GetCurrentUser 获取当前登录用户信息
// @Summary 获取当前用户信息
// @Description 需要JWT认证的接口，返回当前登录用户详情
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 401 {object} gin.H{error:string}
// @Failure 500 {object} gin.H{error:string}
// @Router /auth/user/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 从JWT中间件获取用户ID（假设中间件将用户ID存储在上下文）
	userID, exists := c.Get("userID")
	if !exists {
		h.logger.Warn("获取当前用户失败：JWT中间件未设置userID")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未授权访问"})
		return
	}

	// 调用服务层查询用户
	u, err := h.userService.GetByID(userID.(uint))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.logger.Warn("获取当前用户失败：用户不存在",
				zap.Uint("user_id", userID.(uint)),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
			return
		}
		h.logger.Error("获取当前用户失败：数据库查询错误",
			zap.Uint("user_id", userID.(uint)),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询用户失败"}) // 隐藏具体错误
		return
	}

	// 返回用户信息
	c.JSON(http.StatusOK, UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
	})
}
