package user

import (
	"errors"
	"fmt"

	"auth-service/pkg/oauth2"
)

// Service 领域服务：封装用户领域的业务逻辑
type Service struct {
	repo Repository // 依赖仓库接口（抽象），而非具体实现
}

// NewService 创建领域服务实例（通过依赖注入仓库接口）
func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// Register 注册新用户（业务流程：验证 → 检查唯一性 → 加密密码 → 保存）
func (s *Service) Register(u *User) (*User, error) {
	// 1. 基础属性验证（调用实体自身的验证方法）
	if err := u.Validate(); err != nil {
		return nil, err
	}

	// 2. 检查用户名唯一性（通过仓库接口查询）
	usernameExists, err := s.repo.ExistsByUsername(u.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, ErrUsernameExists
	}

	// 3. 检查邮箱唯一性（通过仓库接口查询）
	if u.Email != "" {
		emailExists, err := s.repo.ExistsByEmail(u.Email)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, ErrEmailExists
		}
	}

	// 4. 加密密码（调用实体自身的加密方法）
	if err := u.HashPassword(); err != nil {
		return nil, err
	}

	// 5. 保存用户（通过仓库接口持久化）
	if err := s.repo.Create(u); err != nil {
		return nil, err
	}

	return u, nil
}

// GetByUsername 根据用户名查询用户（供登录验证使用）
func (s *Service) GetByUsername(username string) (*User, error) {
	if username == "" {
		return nil, ErrUsernameEmpty
	}
	return s.repo.FindByUsername(username)
}

// GetByID 根据ID查询用户（供获取用户信息使用）
func (s *Service) GetByID(id uint) (*User, error) {
	if id == 0 {
		return nil, errors.New("用户ID无效")
	}
	return s.repo.FindByID(id)
}

// LoginWithGitHub 使用 GitHub OAuth2 登录或注册
func (s *Service) LoginWithGitHub(githubUser *oauth2.GitHubUser) (*User, error) {
	// 1. 先通过 GitHub ID 查找用户
	existingUser, err := s.repo.FindByGitHubID(githubUser.ID)
	if err == nil {
		// 用户已存在，更新信息并返回
		existingUser.AvatarURL = githubUser.AvatarURL
		if err := s.repo.Update(existingUser); err != nil {
			return nil, fmt.Errorf("更新用户信息失败: %w", err)
		}
		return existingUser, nil
	}

	// 2. 如果通过 GitHub ID 找不到，尝试通过邮箱查找
	if githubUser.Email != "" {
		existingUser, err := s.repo.FindByEmail(githubUser.Email)
		if err == nil {
			// 邮箱已存在，绑定 GitHub 账号
			existingUser.GitHubID = &githubUser.ID
			existingUser.AvatarURL = githubUser.AvatarURL
			existingUser.AuthType = "github"
			if err := s.repo.Update(existingUser); err != nil {
				return nil, fmt.Errorf("绑定 GitHub 账号失败: %w", err)
			}
			return existingUser, nil
		}
	}

	// 3. 用户不存在，创建新用户
	newUser := &User{
		Username:  githubUser.Login,
		Email:     githubUser.Email,
		GitHubID:  &githubUser.ID,
		AvatarURL: githubUser.AvatarURL,
		AuthType:  "github",
	}

	// 检查用户名是否已存在，如果存在则添加后缀
	if exists, _ := s.repo.ExistsByUsername(newUser.Username); exists {
		newUser.Username = fmt.Sprintf("%s_%d", githubUser.Login, githubUser.ID)
	}

	if err := s.repo.Create(newUser); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return newUser, nil
}
