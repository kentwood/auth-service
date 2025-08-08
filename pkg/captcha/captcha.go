package captcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"auth-service/internal/config"
	"auth-service/pkg/logger"
)

// HCaptchaService hCaptcha 验证服务
type HCaptchaService struct {
	config     *config.HCaptchaConfig
	httpClient *http.Client
	logger     *logger.ZapLogger
}

// HCaptchaVerifyRequest hCaptcha 验证请求
type HCaptchaVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
	RemoteIP string `json:"remoteip,omitempty"`
	SiteKey  string `json:"sitekey,omitempty"`
}

// HCaptchaVerifyResponse hCaptcha 验证响应
type HCaptchaVerifyResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Credit      bool     `json:"credit,omitempty"`
}

const hCaptchaVerifyURL = "https://hcaptcha.com/siteverify"

// NewHCaptchaService 创建 hCaptcha 验证服务
func NewHCaptchaService(cfg *config.HCaptchaConfig, logger *logger.ZapLogger) *HCaptchaService {
	return &HCaptchaService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// VerifyToken 验证 hCaptcha 令牌
func (s *HCaptchaService) VerifyToken(ctx context.Context, token, clientIP string) error {
	// 如果未启用验证，直接返回成功
	if !s.config.Enabled {
		s.logger.Debug("hCaptcha 验证已禁用，跳过验证")
		return nil
	}

	// 检查必要的配置
	if s.config.SecretKey == "" {
		return fmt.Errorf("hCaptcha SecretKey 未配置")
	}

	if token == "" {
		return fmt.Errorf("hCaptcha 令牌不能为空")
	}

	// 构建请求参数
	data := url.Values{
		"secret":   {s.config.SecretKey},
		"response": {token},
	}

	// 添加客户端IP（可选）
	if clientIP != "" {
		data.Set("remoteip", clientIP)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "POST", hCaptchaVerifyURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("创建 hCaptcha 验证请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 发送请求
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("hCaptcha 验证请求失败",
			logger.String("client_ip", clientIP),
			logger.Error(err),
		)
		return fmt.Errorf("hCaptcha 验证请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hCaptcha API 返回错误状态码: %d", resp.StatusCode)
	}

	// 解析响应
	var verifyResp HCaptchaVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return fmt.Errorf("解析 hCaptcha 响应失败: %w", err)
	}

	// 验证结果
	if !verifyResp.Success {
		s.logger.Warn("hCaptcha 验证失败",
			logger.String("client_ip", clientIP),
			logger.Any("error_codes", verifyResp.ErrorCodes),
		)

		// 根据错误代码返回具体错误信息
		errorMsg := s.getErrorMessage(verifyResp.ErrorCodes)
		return fmt.Errorf("hCaptcha 验证失败: %s", errorMsg)
	}

	s.logger.Debug("hCaptcha 验证成功",
		logger.String("client_ip", clientIP),
		logger.String("hostname", verifyResp.Hostname),
		logger.String("challenge_ts", verifyResp.ChallengeTS),
	)

	return nil
}

// getErrorMessage 根据错误代码获取错误消息
func (s *HCaptchaService) getErrorMessage(errorCodes []string) string {
	if len(errorCodes) == 0 {
		return "未知错误"
	}

	errorMessages := map[string]string{
		"missing-input-secret":             "缺少密钥参数",
		"invalid-input-secret":             "密钥参数无效或格式错误",
		"missing-input-response":           "缺少响应参数",
		"invalid-input-response":           "响应参数无效或格式错误",
		"bad-request":                      "请求格式错误",
		"invalid-or-already-seen-response": "响应参数无效或已被使用",
		"not-using-dummy-passcode":         "未使用测试通行码",
		"sitekey-secret-mismatch":          "站点密钥与密钥不匹配",
	}

	// 返回第一个错误的中文描述
	if msg, exists := errorMessages[errorCodes[0]]; exists {
		return msg
	}

	return fmt.Sprintf("验证码错误 (%s)", errorCodes[0])
}

// IsEnabled 检查 hCaptcha 是否启用
func (s *HCaptchaService) IsEnabled() bool {
	return s.config.Enabled
}

// GetSiteKey 获取站点密钥（用于前端）
func (s *HCaptchaService) GetSiteKey() string {
	return s.config.SiteKey
}
