package database

import (
	"fmt"
)

// InitTable 初始化数据库表结构
func InitTable() error {
	// mysql中不支持-命名,如果需要,使用``包裹名称
	// 头像可以设置一个默认头像,是写在数据库里面还是写在注册的时候--可以写一个显示名称,就跟钉钉一样
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		username VARCHAR(50) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		email VARCHAR(100) UNIQUE,
		avatar_url VARCHAR(255) DEFAULT '',
		role VARCHAR(20) DEFAULT 'user',
		status TINYINT DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
	);`

	_, err := DB.Exec(createUserTable)
	if err != nil {
		return fmt.Errorf("创建 users 表失败:%w", err)
	}

	createLoginLogTable := `
	CREATE TABLE IF NOT EXISTS login_logs (
		id BIGINT PRIMARY KEY AUTO_INCREMENT,
		user_id BIGINT DEFAULT 0,
		username VARCHAR(50) DEFAULT '',
		success TINYINT DEFAULT 0,
		ip VARCHAR(64) DEFAULT '',
		user_agent VARCHAR(255) DEFAULT '',
		message VARCHAR(255) DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_login_user_id (user_id),
		INDEX idx_login_created_at (created_at)
	);`

	_, err = DB.Exec(createLoginLogTable)
	if err != nil {
		return fmt.Errorf("创建 login_logs 表失败:%w", err)
	}

	return nil
}
