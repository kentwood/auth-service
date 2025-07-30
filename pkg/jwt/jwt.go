package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义JWT载荷，包含用户ID和用户名
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT令牌
// 参数：
//   - userID: 用户ID
//   - username: 用户名
//   - secret: 签名密钥
//   - expiration: 过期时间（如24*time.Hour）
//
// 返回：
//   - 生成的令牌字符串
//   - 错误信息
func GenerateToken(userID uint, username string, secret string, expiration time.Duration) (string, error) {
	// 设置过期时间
	expiresAt := time.Now().Add(expiration)

	// 创建自定义载荷
	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),  // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()), // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()), // 生效时间（立即生效）
			Issuer:    "auth-service",                 // 签发者
		},
	}

	// 创建令牌（使用HS256算法）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名令牌
	return token.SignedString([]byte(secret))
}

// ParseToken 解析并验证JWT令牌
// 参数：
//   - tokenString: 待解析的令牌字符串
//   - secret: 签名密钥（需与生成时一致）
//
// 返回：
//   - 解析后的自定义载荷
//   - 错误信息
func ParseToken(tokenString string, secret string) (*Claims, error) {
	// 解析令牌
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("不支持的签名算法")
			}
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// 验证令牌有效性并提取载荷
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的令牌")
}

// ValidateToken 仅验证令牌是否有效（不提取载荷）
func ValidateToken(tokenString string, secret string) bool {
	_, err := ParseToken(tokenString, secret)
	return err == nil
}
