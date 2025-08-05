package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User 用户实体
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password  string    `gorm:"size:255" json:"-"` // 对于OAuth2用户，可能为空
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// OAuth2 相关字段
	GitHubID  *int64 `gorm:"uniqueIndex" json:"github_id,omitempty"`   // GitHub 用户ID
	AvatarURL string `gorm:"size:255" json:"avatar_url,omitempty"`     // 头像URL
	AuthType  string `gorm:"size:20;default:'local'" json:"auth_type"` // 认证类型：local, github
}

// 领域错误定义：在领域层内部定义，供服务层使用
var (
	ErrUsernameEmpty  = errors.New("用户名不能为空")
	ErrPasswordEmpty  = errors.New("密码不能为空")
	ErrEmailInvalid   = errors.New("邮箱格式无效")
	ErrUsernameExists = errors.New("用户名已被注册")
	ErrEmailExists    = errors.New("邮箱已被注册")
)

// HashPassword 加密密码（实体自身行为）
func (u *User) HashPassword() error {
	if u.Password == "" {
		return ErrPasswordEmpty
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

// CheckPassword 验证密码（实体自身行为）
func (u *User) CheckPassword(rawPassword string) bool {
	if rawPassword == "" || u.Password == "" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(rawPassword)) == nil
}

// Validate 验证实体基础属性（领域规则）
func (u *User) Validate() error {
	if u.Username == "" {
		return ErrUsernameEmpty
	}
	if u.Password == "" {
		return ErrPasswordEmpty
	}
	// 简单邮箱格式校验（实际项目可使用正则）
	if u.Email != "" && (len(u.Email) < 5 || u.Email[len(u.Email)-4:] != ".com") {
		return ErrEmailInvalid
	}
	return nil
}

// IsOAuth2User 检查是否是 OAuth2 用户
func (u *User) IsOAuth2User() bool {
	return u.AuthType != "local"
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	if u.AuthType == "local" {
		return u.Password != ""
	}
	return true // OAuth2 用户总是可以登录
}
