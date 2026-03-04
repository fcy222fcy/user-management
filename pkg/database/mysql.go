package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func InitDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	var err error
	// sql.Open 只是初始化,不会真正连接
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("数据库连接错误:%w", err)
	}

	// PING 测试连接
	if err = DB.Ping(); err != nil {
		return fmt.Errorf("数据库连接失败:%w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)

	err = InitTable()
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
