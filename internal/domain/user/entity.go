package user

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User 领域实体：表示系统中的用户
type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Username  string         `gorm:"uniqueIndex;not null;size:50" json:"username"` // 用户名（唯一）
	Password  string         `gorm:"not null" json:"-"`                            // 密码（哈希存储，不返回给前端）
	Email     string         `gorm:"uniqueIndex;size:100" json:"email"`            // 邮箱（唯一）
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
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
