package session

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"auth-service/pkg/redis"
)

// Manager Session 管理器
type Manager struct {
	redisClient *redis.Client
}

// OAuth2State OAuth2 状态信息
type OAuth2State struct {
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
	UserAgent string    `json:"user_agent,omitempty"`
	ClientIP  string    `json:"client_ip,omitempty"`
}

// NewManager 创建 Session 管理器
func NewManager(redisClient *redis.Client) *Manager {
	return &Manager{
		redisClient: redisClient,
	}
}

// CreateOAuth2Session 创建 OAuth2 会话
func (m *Manager) CreateOAuth2Session(ctx context.Context, state, userAgent, clientIP string) (string, error) {
	// 生成唯一的 session ID
	sessionID := uuid.New().String()

	// 创建状态信息
	stateInfo := OAuth2State{
		State:     state,
		CreatedAt: time.Now(),
		UserAgent: userAgent,
		ClientIP:  clientIP,
	}

	// 序列化为 JSON
	stateJSON, err := json.Marshal(stateInfo)
	if err != nil {
		return "", fmt.Errorf("序列化状态信息失败: %w", err)
	}

	// 存储到 Redis，设置 10 分钟过期时间
	key := fmt.Sprintf("oauth2:session:%s", sessionID)
	if err := m.redisClient.Set(ctx, key, string(stateJSON), 10*time.Minute); err != nil {
		return "", fmt.Errorf("存储会话到 Redis 失败: %w", err)
	}

	return sessionID, nil
}

// ValidateOAuth2Session 验证 OAuth2 会话
func (m *Manager) ValidateOAuth2Session(ctx context.Context, sessionID, receivedState string) (*OAuth2State, error) {
	// 从 Redis 获取状态信息
	key := fmt.Sprintf("oauth2:session:%s", sessionID)
	stateJSON, err := m.redisClient.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("获取会话信息失败: %w", err)
	}

	// 反序列化状态信息
	var stateInfo OAuth2State
	if err := json.Unmarshal([]byte(stateJSON), &stateInfo); err != nil {
		return nil, fmt.Errorf("解析状态信息失败: %w", err)
	}

	// 验证状态码
	if stateInfo.State != receivedState {
		return nil, fmt.Errorf("状态码不匹配: 期望 %s, 收到 %s", stateInfo.State, receivedState)
	}

	// 验证时间（可选的额外安全检查）
	if time.Since(stateInfo.CreatedAt) > 10*time.Minute {
		return nil, fmt.Errorf("会话已过期")
	}

	return &stateInfo, nil
}

// DeleteOAuth2Session 删除 OAuth2 会话（一次性使用）
func (m *Manager) DeleteOAuth2Session(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("oauth2:session:%s", sessionID)
	return m.redisClient.Del(ctx, key)
}

// CleanupExpiredSessions 清理过期会话（可以通过定时任务调用）
func (m *Manager) CleanupExpiredSessions(ctx context.Context) error {
	// Redis 会自动处理过期的键，这里可以添加额外的清理逻辑
	// 比如清理相关的业务数据等
	return nil
}
