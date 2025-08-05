package repository

import (
	"auth-service/internal/domain/user"

	"gorm.io/gorm"
)

// userRepository 仓库实现：基于GORM实现数据访问逻辑
type userRepository struct {
	db *gorm.DB // 数据库连接（通过依赖注入）
}

// NewUserRepository 创建仓库实例
func NewUserRepository(db *gorm.DB) user.Repository {
	return &userRepository{
		db: db,
	}
}

// Create 保存用户到数据库
func (r *userRepository) Create(u *user.User) error {
	return r.db.Create(u).Error
}

// FindByUsername 根据用户名查询用户
func (r *userRepository) FindByUsername(username string) (*user.User, error) {
	var u user.User
	result := r.db.Where("username = ?", username).First(&u)
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

// FindByID 根据ID查询用户
func (r *userRepository) FindByID(id uint) (*user.User, error) {
	var u user.User
	result := r.db.First(&u, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

// ExistsByUsername 检查用户名是否已存在
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	result := r.db.Model(&user.User{}).Where("username = ?", username).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// ExistsByEmail 检查邮箱是否已存在
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	result := r.db.Model(&user.User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		return false, result.Error
	}
	return count > 0, nil
}

// FindByGitHubID 根据 GitHub ID 查询用户
func (r *userRepository) FindByGitHubID(githubID int64) (*user.User, error) {
	var u user.User
	result := r.db.Where("github_id = ?", githubID).First(&u)
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

// FindByEmail 根据邮箱查询用户
func (r *userRepository) FindByEmail(email string) (*user.User, error) {
	var u user.User
	result := r.db.Where("email = ?", email).First(&u)
	if result.Error != nil {
		return nil, result.Error
	}
	return &u, nil
}

// Update 更新用户信息
func (r *userRepository) Update(u *user.User) error {
	return r.db.Save(u).Error
}
