package user

// UserRepository 仓库接口：定义用户数据访问的抽象方法
type UserRepository interface {
	Create(u *User) error                           // 保存用户
	FindByUsername(username string) (*User, error)  // 根据用户名查询
	FindByID(id uint) (*User, error)                // 根据ID查询
	ExistsByUsername(username string) (bool, error) // 检查用户名是否存在
	ExistsByEmail(email string) (bool, error)       // 检查邮箱是否存在
}
