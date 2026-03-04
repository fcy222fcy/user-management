package repository

// repository只做一件事: 和数据库交互
// 不处理 http 不处理json 不处理业务逻辑 不做密码校验 只负责查和修改数据

import (
	"database/sql"
	"errors"
	"fmt"
	"user-management/internal/model"
)

type UserRepository struct {
	DB *sql.DB
}

// NewUserRepository 创建 UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

// CreateUser 插入用户
func (r *UserRepository) CreateUser(user *model.User) (int64, error) {
	query := `
	INSERT INTO users( username, password, email, role, status) 
	VALUES (?,?,?,?,?);
	`
	// 用户登录的时候,可以暂不选择添加头像
	result, err := r.DB.Exec(query,
		user.Username,
		user.Password,
		user.Email,
		user.Role,
		user.Status,
	)
	// created_at 字段已经设置了在插入的时候自动设置
	// updated_at 也设置了更新数据的时候自动更新
	// 那么下面代码就不用手动更新了
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetUserByUsername 根据用户名获取用户结构体
func (r *UserRepository) GetUserByUsername(username string) (*model.User, error) {
	query := `
	SELECT id,username,password,email,role,status,created_at,updated_at
	FROM users 
	WHERE username = ?;
`
	row := r.DB.QueryRow(query, username)

	var user model.User
	// 把数据库查询出来的每一列数据,一次赋值给结构体字段
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.Role,
		&user.Status,
		&user.CreatedAt,
		&user.UpdateAt,
	)
	if err != nil {
		// 如果用户不存在,不要使用 err == sql.ErrNoRows
		if errors.Is(err, sql.ErrNoRows) { // 这里也可以返回空
			return nil, fmt.Errorf("用户不存在: %w ", err)
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

// DeleteUser 根据ID删除用户
func (r *UserRepository) DeleteUser(userID string) error {
	// 删除的时候,是点击用户列表中的小垃圾桶,到时候直接传用户名就好了/或者是写邮箱,让用户名可以重复
	query := `
	DELETE FROM users
	WHERE id = ? ;
	`
	_, err := r.DB.Exec(query, userID)
	// 不管err是什么样,直接返回
	return err
}

// UpdateUserAvatar 更新用户头像
func (r *UserRepository) UpdateUserAvatar(userID int64, avatarURL string) error {
	query := `
	UPDATE users 
	SET avatar_url = ?,updated_at = CURRENT_TIMESTAMP
	WHERE id = ?
	`
	_, err := r.DB.Exec(query, avatarURL, userID)
	return err
}

// UpdateUsername 更新用户名
func (r *UserRepository) UpdateUsername(userID int64, username string) error {
	query := `
	UPDATE users 
	SET username = ?,updated_at = CURRENT_TIMESTAMP
	WHERE id = ?;
	`
	_, err := r.DB.Exec(query, username, userID)
	return err
}

// UpdateUserEmail 更新用户邮箱
func (r *UserRepository) UpdateUserEmail(userID int64, email string) error {
	query := `
	UPDATE users
	SET email = ?,updated_at = current_timestamp
	WHERE id = ?;
	`
	_, err := r.DB.Exec(query, email, userID)
	return err
}

// UpdateUserRole 更新用户角色
func (r *UserRepository) UpdateUserRole(userID int64, role string) error {
	query := `
	UPDATE users
	SET role = ?,updated_at = current_timestamp
	WHERE id = ?;
	`
	_, err := r.DB.Exec(query, role, userID)
	return err
}

// UpdateUserStatus 更新用户状态
func (r *UserRepository) UpdateUserStatus(userID int64, status int) error {
	query := `
	UPDATE users
	SET status = ?,updated_at = current_timestamp
	WHERE  id = ?;
	`
	_, err := r.DB.Exec(query, status, userID)
	return err
}
