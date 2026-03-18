package repository

// repository只做一件事: 和数据库交互
// 不处理 http 不处理json 不处理业务逻辑 不做密码校验 只负责查和修改数据

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
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
	INSERT INTO users( username, password, email,avatar_url, role, status)
	VALUES (?,?,?,?,?,?);
	`
	log.Printf("执行SQL插入: %s, %s, %s, %s, %d", user.Username, user.Password, user.Email, user.AvatarURL, user.Role, user.Status)
	// 用户注册的时候,可以暂不选择添加头像
	// 手动添加用户的时候,可以添加头像
	result, err := r.DB.Exec(query,
		user.Username,
		user.Password,
		user.Email,
		user.AvatarURL,
		user.Role,
		user.Status,
	)
	// created_at 字段已经设置了在插入的时候自动设置
	// updated_at 也设置了更新数据的时候自动更新
	// 那么下面代码就不用手动更新了
	if err != nil {
		log.Printf("SQL执行失败: %v", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("获取LastInsertId失败: %v", err)
		return 0, err
	}
	log.Printf("LastInsertId: %d", id)
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
		&user.UpdatedAt,
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
func (r *UserRepository) GetUserByID(id int64) (*model.User, error) {

	query := `
		SELECT id, username, password, email, avatar_url, role, status
		FROM users
		WHERE id = ?
`
	row := r.DB.QueryRow(query, id)

	var user model.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Email,
		&user.AvatarURL,
		&user.Role,
		&user.Status,
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
func (r *UserRepository) DeleteUser(userID int64) error {
	//测试
	log.Println("dao层:删除用户")

	// 删除的时候,是点击用户列表中的小垃圾桶,到时候直接传用户名就好了/或者是写邮箱,让用户名可以重复
	query := `
	DELETE FROM users
	WHERE id = ? ;
	`
	_, err := r.DB.Exec(query, userID)
	// 不管err是什么样,直接返回
	return err
}

// UpdateUser 更新用户状态
func (r *UserRepository) UpdateUser(id int64, username, password, email string, status int, avatarURL string, role string) error {

	var query string
	var args []interface{}

	if avatarURL != "" {
		query = `
		UPDATE users
		SET username = ?, password = ?, email = ?, status = ?, avatar_url = ?, role = ?, updated_at = current_timestamp
		WHERE id = ?;
		`
		args = []interface{}{username, password, email, status, avatarURL, role, id}
	} else {
		query = `
		UPDATE users
		SET username = ?, password = ?, email = ?, status = ?, role = ?, updated_at = current_timestamp
		WHERE id = ?;
		`
		args = []interface{}{username, password, email, status, role, id}
	}

	_, err := r.DB.Exec(query, args...)
	return err
}

// GetAllList 获取用户列表,带分页
func (r *UserRepository) GetAllList(offset, limit int) ([]model.User, error) {
	query := `
	SELECT id,username,email,avatar_url,role,status,created_at,updated_at
	FROM users
	ORDER BY id
	DESC LIMIT ?
	OFFSET ?;
`
	rows, err := r.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User

	for rows.Next() {
		var u model.User
		// 把数据库查询出来的每一列数据,一次赋值给结构体字段
		err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.AvatarURL,
			&u.Role,
			&u.Status,
			&u.CreatedAt,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// GetUserCount 获取用户数量
func (r *UserRepository) GetUserCount() (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM users
`
	err := r.DB.QueryRow(query).Scan(&count)
	return count, err

}

func (r *UserRepository) CreateLoginLog(userID int64, username string, success bool, ip, userAgent, message string) error {
	query := `
	INSERT INTO login_logs (user_id, username, success, ip, user_agent, message, created_at)
	VALUES (?, ?, ?, ?, ?, ?, ?)
`
	_, err := r.DB.Exec(query, userID, username, success, ip, userAgent, message, time.Now())
	return err
}

func (r *UserRepository) GetLoginCountInRange(start, end time.Time) (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM login_logs
	WHERE success = 1 AND created_at >= ? AND created_at < ?
`
	err := r.DB.QueryRow(query, start, end).Scan(&count)
	return count, err
}

func (r *UserRepository) GetDailyLogins(start, end time.Time) (map[string]int, error) {
	query := `
	SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS cnt
	FROM login_logs
	WHERE success = 1 AND created_at >= ? AND created_at < ?
	GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
	ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d')
`
	rows, err := r.DB.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result[day] = count
	}
	return result, nil
}

func (r *UserRepository) GetUserCountBefore(t time.Time) (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM users
	WHERE created_at < ?
`
	err := r.DB.QueryRow(query, t).Scan(&count)
	return count, err
}

func (r *UserRepository) GetUserCountInRange(start, end time.Time) (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM users
	WHERE created_at >= ? AND created_at < ?
`
	err := r.DB.QueryRow(query, start, end).Scan(&count)
	return count, err
}

func (r *UserRepository) GetDeactivatedCount() (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM users
	WHERE status = 0
`
	err := r.DB.QueryRow(query).Scan(&count)
	return count, err
}

func (r *UserRepository) GetDeactivatedCountInRange(start, end time.Time) (int, error) {
	var count int
	query := `
	SELECT COUNT(*) FROM users
	WHERE status = 0 AND updated_at >= ? AND updated_at < ?
`
	err := r.DB.QueryRow(query, start, end).Scan(&count)
	return count, err
}

func (r *UserRepository) GetDailyRegistrations(start, end time.Time) (map[string]int, error) {
	query := `
	SELECT DATE_FORMAT(created_at, '%Y-%m-%d') AS day, COUNT(*) AS cnt
	FROM users
	WHERE created_at >= ? AND created_at < ?
	GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d')
	ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d')
`
	rows, err := r.DB.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var day string
		var count int
		if err := rows.Scan(&day, &count); err != nil {
			return nil, err
		}
		result[day] = count
	}
	return result, nil
}
