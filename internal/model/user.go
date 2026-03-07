package model

import "time"

type User struct { // 直接模版渲染不需要加json,但是如果是返回json就需要加
	ID       int64
	Username string
	// password 不能加json,不然会暴露给前端
	Password  string
	Email     string
	AvatarURL string
	Role      string
	Status    int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RegisterRequest struct {
	Username string
	Password string
	Email    string
}
type LoginRequest struct {
	Username string
	Password string
}
type UserResponse struct {
	ID       int64
	Username string
	Email    string
	Role     string
}
