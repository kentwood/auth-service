package oauth2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"auth-service/internal/config"
)

// GitHubUser GitHub 用户信息
type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// GitHubOAuth2Service GitHub OAuth2 服务
type GitHubOAuth2Service struct {
	config *oauth2.Config
}

// NewGitHubOAuth2Service 创建 GitHub OAuth2 服务
func NewGitHubOAuth2Service(cfg *config.GitHubOAuth2Config) *GitHubOAuth2Service {
	conf := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	return &GitHubOAuth2Service{
		config: conf,
	}
}

// GetAuthURL 获取授权URL
func (s *GitHubOAuth2Service) GetAuthURL(state string) string {
	return s.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode 交换授权码获取用户信息
func (s *GitHubOAuth2Service) ExchangeCode(ctx context.Context, code string) (*GitHubUser, error) {
	token, err := s.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("交换令牌失败: %w", err)
	}

	// 使用令牌获取用户信息
	client := s.config.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API 返回错误状态码: %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	// 如果用户信息中没有邮箱，单独获取
	if user.Email == "" {
		if err := s.fetchUserEmail(ctx, client, &user); err != nil {
			return nil, err
		}
	}

	return &user, nil
}

// fetchUserEmail 获取用户邮箱
func (s *GitHubOAuth2Service) fetchUserEmail(ctx context.Context, client *http.Client, user *GitHubUser) error {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return fmt.Errorf("获取用户邮箱失败: %w", err)
	}
	defer resp.Body.Close()

	var emails []struct {
		Email   string `json:"email"`
		Primary bool   `json:"primary"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return fmt.Errorf("解析邮箱信息失败: %w", err)
	}

	// 找到主邮箱
	for _, email := range emails {
		if email.Primary {
			user.Email = email.Email
			break
		}
	}

	return nil
}
